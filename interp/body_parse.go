package interp

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"strings"

	message "github.com/emersion/go-message"
	"github.com/emersion/go-message/textproto"
)

// ContentTypeMatches reports whether the MIME content type ct matches any of
// the given patterns, following RFC 5173 §5.2 rules:
//   - pattern starting or ending with '/', or with multiple slashes → matches nothing
//   - "" (empty) → matches everything
//   - "type/subtype" → exact match
//   - "type" → matches any subtype
func ContentTypeMatches(contentType string, patterns []string) bool {
	// Normalize: strip parameters (e.g. "text/plain; charset=utf-8" → "text/plain")
	if i := strings.IndexByte(contentType, ';'); i >= 0 {
		contentType = strings.TrimSpace(contentType[:i])
	}
	contentType = strings.ToLower(contentType)

	for _, p := range patterns {
		p = strings.ToLower(p)

		// RFC 5173: begins or ends with '/', or multiple slashes → no match
		if strings.HasPrefix(p, "/") || strings.HasSuffix(p, "/") {
			continue
		}
		slashCount := strings.Count(p, "/")
		if slashCount > 1 {
			continue
		}

		if p == "" {
			return true
		}
		if slashCount == 1 {
			// full type/subtype match
			if contentType == p {
				return true
			}
		} else {
			// type-only match: any subtype; reject malformed ct with extra slashes
			if strings.HasPrefix(contentType, p+"/") && strings.Count(contentType, "/") == 1 {
				return true
			}
		}
	}
	return false
}

func skipHeader(r *bufio.Reader) error {
	for {
		line, err := r.ReadSlice('\n')
		if err != nil {
			return fmt.Errorf("skipHeader: %w", err)
		}
		// If line is empty (message uses LF delim) or contains only CR (messages uses CRLF delim)
		if len(line) == 0 || (len(line) == 1 || line[0] == '\r') {
			break
		}
	}
	return nil
}

// ErrNoBody should be returned by BodyMessage methods if processed message
// has no body and therefore cannot match any 'body' test.
var ErrNoBody = errors.New("BodyMessage: no body")

// ParseBodyRaw reads the raw (undecoded) body from a complete RFC 2822 message
// stream. It returns a reader starting after the first blank line (header/body
// separator). If no blank line is found (header-only message), it returns nil.
func ParseBodyRaw(ctx context.Context, rawMsg io.Reader) (*bufio.Reader, error) {
	buf := bufio.NewReader(rawMsg)

	if err := skipHeader(buf); err != nil {
		if errors.Is(err, io.EOF) {
			// Message without body (no trailing CRLF).
			return buf, ErrNoBody
		}
		return nil, err
	}

	return buf, nil
}

type htmlStripper struct {
	BR io.ByteReader

	inTag bool
}

func (hs *htmlStripper) Read(p []byte) (n int, err error) {
	for n < len(p) {
		b, err := hs.ReadByte()
		if err != nil {
			return n, err
		}

		p[n] = b
		n++
	}
	return n, nil
}

func (hs *htmlStripper) ReadByte() (byte, error) {
	for {
		b, err := hs.BR.ReadByte()
		if err != nil {
			return 0, err
		}

		if b == '<' && !hs.inTag {
			hs.inTag = true
			continue
		}
		if b == '>' && hs.inTag {
			hs.inTag = false
			continue
		}
		if hs.inTag {
			continue
		}

		return b, nil
	}
}

// ParseBodyParts parses a complete RFC 2822/MIME message and returns
// body parts whose content-type matches any of contentTypes, following
// RFC 5173 §5.2 rules.
//
// This is the helper for BodyParts intended for BodyMessage implementors
// that only have access to the raw message stream. This is intentionally
// naive implementation that keeps the matching message parts in memory for
// parsing  and returns byte readers. Care should be taken to avoid using
// this on large matching parts.
//
// The rawMsg reader should start at the beginning of the message (headers
// included), not after the separator line.
func ParseBodyParts(ctx context.Context, rawMsg io.Reader, contentTypes []string, maxBytes int) ([]BodyPart, error) {
	r := bufio.NewReader(rawMsg)

	hdr, err := textproto.ReadHeader(r)
	if err != nil {
		return nil, fmt.Errorf("ParseBodyParts: %w", err)
	}

	var results []BodyPart
	if err := walkMIMEParts(hdr, r, contentTypes, &results, maxBytes); err != nil {
		return nil, err
	}
	return results, nil
}

// walkMIMEParts recursively walks MIME structure and collects parts matching contentTypes.
// hdr is the header of the current entity, body is the decoded body reader.
func walkMIMEParts(
	hdr textproto.Header, body io.Reader,
	contentTypes []string, results *[]BodyPart,
	maxBytes int,
) error {
	msgHdr := message.Header{Header: hdr}
	ct, params, _ := msgHdr.ContentType()
	ct = strings.ToLower(ct)

	isMultipart := strings.HasPrefix(ct, "multipart/")
	isMessageRFC822 := ct == "message/rfc822" || ct == "message/global"

	if isMultipart {
		boundary := params["boundary"]

		if ContentTypeMatches(ct, contentTypes) {
			body = io.LimitReader(body, int64(maxBytes))

			// Read raw body to extract prologue/epilogue and recurse into parts
			rawBody, err := io.ReadAll(body)
			if err != nil {
				return nil
			}

			// Extract prologue and epilogue (RFC 5173 §5.2)
			prologue, epilogue := extractMultipartPrologueEpilogue(rawBody, boundary)
			var blob []byte
			blob = append(blob, prologue...)
			blob = append(blob, '\r', '\n')
			blob = append(blob, epilogue...)

			*results = append(*results, BodyPartBytes{
				ContentTypeValue: ct,
				Blob:             blob,
			})

			body = bytes.NewReader(rawBody)
		}

		// Recurse into subparts
		mr := textproto.NewMultipartReader(body, boundary)
		for {
			part, err := mr.NextPart()
			if errors.Is(err, io.EOF) {
				break
			}
			if err != nil {
				break // skip malformed parts
			}

			if err := walkMIMEParts(part.Header, part, contentTypes, results, maxBytes); err != nil {
				return err
			}
		}
		return nil
	}

	if isMessageRFC822 {
		buffered := bufio.NewReader(body)

		nestedHdr, err := textproto.ReadHeader(buffered)
		if err != nil {
			return err
		}

		// RFC 5173 §5.2: message/rfc822 - if it matches, return header as single string.
		// Nested body parts are recursed into if their type matches.
		if ContentTypeMatches(ct, contentTypes) {
			// Serialize the header as a single string (no trailing blank line)
			hdrStr, err := serializeHeader(nestedHdr)
			if err != nil {
				return err
			}

			*results = append(*results, BodyPartBytes{
				ContentTypeValue: ct,
				Blob:             []byte(hdrStr),
			})
		}

		if err := walkMIMEParts(nestedHdr, buffered, contentTypes, results, maxBytes); err != nil {
			return err
		}
		return nil
	}

	// Leaf part: decode and collect if matching
	if !ContentTypeMatches(ct, contentTypes) {
		return nil
	}

	// Apply transfer-encoding and charset decoding via go-message
	mh := message.Header{Header: hdr}
	ent, err := message.New(mh, body)
	if err != nil {
		// Skip parts with unknown or invalid encoding/charset
		return nil
	}

	limited := io.LimitReader(ent.Body, int64(maxBytes))
	decoded, err := io.ReadAll(limited)
	if err != nil {
		// Skip parts that fail to decode
		return nil
	}
	if len(decoded) > 0 {
		*results = append(*results, BodyPartBytes{
			ContentTypeValue: ct,
			Blob:             decoded,
		})
	}
	return nil
}

// extractMultipartPrologueEpilogue returns the prologue (before first boundary)
// and epilogue (after last boundary) of a raw multipart body.
func extractMultipartPrologueEpilogue(body []byte, boundary string) (prologue, epilogue []byte) {
	dashBoundary := []byte("--" + boundary)

	// Find first boundary occurrence
	firstIdx := bytes.Index(body, dashBoundary)
	if firstIdx < 0 {
		return nil, nil
	}
	// Prologue: everything before first boundary line
	prologue = bytes.TrimRight(body[:firstIdx], "\r\n")

	// Find the closing boundary "--boundary--"
	closingBoundary := append(dashBoundary, []byte("--")...)
	lastIdx := bytes.LastIndex(body, closingBoundary)
	if lastIdx < 0 {
		return prologue, nil
	}
	// Skip past the closing boundary line
	rest := body[lastIdx+len(closingBoundary):]
	// Skip CRLF after closing boundary
	rest = bytes.TrimLeft(rest, "\r\n")
	epilogue = bytes.TrimRight(rest, "\r\n")

	return prologue, epilogue
}

// serializeHeader converts a textproto.Header to a string representation
// suitable for matching (without trailing blank line).
// It preserves the original folding of multi-line header values.
func serializeHeader(hdr textproto.Header) (string, error) {
	var b strings.Builder
	fields := hdr.Fields()
	for fields.Next() {
		raw, err := fields.Raw()
		if err != nil {
			return "", err
		}
		// raw already uses CRLF; write as-is
		b.WriteString(string(raw))
	}
	return b.String(), nil
}

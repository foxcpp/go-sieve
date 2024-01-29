package lexer

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"strconv"
	"strings"
)

type Options struct {
	Filename   string
	NoPosition bool
	MaxTokens  int
}

func consumeCRLF(r *bufio.Reader, state *lexerState) error {
	b, err := r.ReadByte()
	if err != nil {
		return err
	}
	switch b {
	case '\r':
		b, err = r.ReadByte()
		if err != nil {
			return err
		}
		if b != '\n' {
			return fmt.Errorf("CR is not followed by LF")
		}
		fallthrough
	case '\n':
		state.Line++
		state.Col = 0
		return nil
	default:
		panic("consumeCRLF should not be called not on CR/LF")
	}
}

func Lex(r io.Reader, opts *Options) ([]Token, error) {
	if opts == nil {
		opts = &Options{}
	}
	toks, err := tokenStream(bufio.NewReader(r), opts)
	if err != nil {
		if err == io.EOF {
			return nil, io.ErrUnexpectedEOF
		}
		return nil, err
	}
	return toks, nil
}

type lexerState struct {
	Position
}

func tokenStream(r *bufio.Reader, opts *Options) ([]Token, error) {
	res := []Token{}
	state := &lexerState{}
	state.File = opts.Filename
	state.Line = 1
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if opts.NoPosition {
			state.Line = 0
			state.Col = 0
		} else {
			state.Col++
		}
		switch b {
		case 0:
			return nil, fmt.Errorf("go-sieve/lexer: NUL is not allowed in input stream")
		case '[':
			res = append(res, ListStart{state.Position})
		case ']':
			res = append(res, ListEnd{state.Position})
		case '{':
			res = append(res, BlockStart{state.Position})
		case '}':
			res = append(res, BlockEnd{state.Position})
		case '(':
			res = append(res, TestListStart{state.Position})
		case ')':
			res = append(res, TestListEnd{state.Position})
		case ',':
			res = append(res, Comma{state.Position})
		case ':':
			res = append(res, Colon{state.Position})
		case ';':
			res = append(res, Semicolon{state.Position})
		case ' ', '\t':
			continue
		case '\r', '\n':
			if err := r.UnreadByte(); err != nil {
				return nil, err
			}
			if err := consumeCRLF(r, state); err != nil {
				return nil, err
			}
		case '"':
			lineCol := state.Position
			str, err := quotedString(r, state)
			if err != nil {
				return nil, err
			}
			res = append(res, String{Position: lineCol, Text: str})
		case '#':
			if err := hashComment(r, state); err != nil {
				return nil, err
			}
		case '/':
			b2, err := r.ReadByte()
			if err != nil {
				return nil, err
			}
			state.Col++
			if b2 != '*' {
				return nil, fmt.Errorf("unexpected forward slash")
			}
			if err := multilineComment(r, state); err != nil {
				return nil, err
			}
		case 't':
			// "text:"
			lineCol := state.Position
			ext, err := r.Peek(4)
			if err != nil {
				return nil, err
			}
			if bytes.Equal(ext, []byte("ext:")) {
				if _, err := r.Discard(4); err != nil {
					return nil, err
				}
				state.Col += 4
				// we consume whitespace and then build the multiline string
			wsLoop:
				for {
					b, err := r.ReadByte()
					if err != nil {
						return nil, err
					}
					state.Col++
					switch b {
					case ' ', '\t':
						continue
					case '#':
						if err := hashComment(r, state); err != nil {
							return nil, err
						}
						break wsLoop
					case '\r', '\n':
						if err := r.UnreadByte(); err != nil {
							return nil, err
						}
						if err := consumeCRLF(r, state); err != nil {
							return nil, err
						}
						break wsLoop
					default:
						return nil, fmt.Errorf("unexpected character: %v", b)
					}
				}
				mlString, err := multilineString(r, state)
				if err != nil {
					return nil, err
				}
				res = append(res, String{Position: lineCol, Text: mlString})
				continue
			}
			// if that's not text: but something else
			fallthrough
		default:
			lineCol := state.Position

			if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') {
				str, err := identifier(r, string(b), state)
				if err != nil {
					return nil, err
				}
				res = append(res, Identifier{Position: lineCol, Text: str})
			} else if b >= '0' && b <= '9' {
				num, err := number(r, string(b), state)
				if err != nil {
					return nil, err
				}
				num.Position = lineCol
				res = append(res, num)
			} else {
				return nil, fmt.Errorf("unexpected character: %v", b)
			}
		}
		if opts.MaxTokens != 0 && len(res) > opts.MaxTokens {
			return nil, fmt.Errorf("too many tokens")
		}
	}
	return res, nil
}

func IsValidIdentifier(s string) bool {
	if len(s) == 0 {
		return false
	}

	first := s[0]
	if !(first >= 'a' && first <= 'z') && !(first >= 'A' && first <= 'Z') {
		return false
	}

	for _, chr := range s[1:] {
		switch {
		case chr >= 'a' && chr <= 'z':
		case chr >= 'A' && chr <= 'Z':
		case chr >= '0' && chr <= '9':
		case chr == '_':
		default:
			return false
		}
	}
	return true
}

func identifier(r *bufio.Reader, startWith string, state *lexerState) (string, error) {
	id := strings.Builder{}
	id.WriteString(startWith)
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return "", err
		}
		state.Col++
		//  identifier         = (ALPHA / "_") *(ALPHA / DIGIT / "_")
		if (b >= 'a' && b <= 'z') || (b >= 'A' && b <= 'Z') || (b >= '0' && b <= '9') || b == '_' {
			id.WriteByte(b)
		} else {
			if err := r.UnreadByte(); err != nil {
				return "", err
			}
			state.Col--
			break
		}
	}
	return id.String(), nil
}

func number(r *bufio.Reader, startWith string, state *lexerState) (Number, error) {
	num := strings.Builder{}
	num.WriteString(startWith)
	q := None
readLoop:
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return Number{}, err
		}
		state.Col++
		switch b {
		case 'K', 'G', 'M':
			q = Quantifier(b)
			break readLoop
		case 'k', 'g', 'm':
			q = Quantifier(b - 32 /* to upper */)
			break readLoop
		}
		if b >= '0' && b <= '9' {
			num.WriteByte(b)
		} else {
			if err := r.UnreadByte(); err != nil {
				return Number{}, err
			}
			state.Col--
			break readLoop
		}
	}

	numParsed, err := strconv.Atoi(num.String())
	if err != nil {
		return Number{}, err
	}
	return Number{Value: numParsed, Quantifier: q}, nil
}

func hashComment(r *bufio.Reader, state *lexerState) error {
	for {
		b, err := r.ReadByte()
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}
		state.Col++
		if b == '\r' || b == '\n' {
			if err := r.UnreadByte(); err != nil {
				return err
			}
			if err := consumeCRLF(r, state); err != nil {
				return err
			}
			break
		}
	}
	return nil
}

func multilineComment(r *bufio.Reader, state *lexerState) error {
	wasStar := false
	for {
		b, err := r.ReadByte()
		if err != nil {
			return err
		}
		state.Col++
		if b == '\n' {
			state.Line++
			state.Col = 0
		}
		if wasStar && b == '/' {
			return nil
		}
		wasStar = b == '*'
	}
}

func quotedString(r *bufio.Reader, state *lexerState) (string, error) {
	str := strings.Builder{}
	atBackslash := false
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		state.Col++
		switch b {
		case '\r', '\n':
			if err := r.UnreadByte(); err != nil {
				return "", err
			}
			if err := consumeCRLF(r, state); err != nil {
				return "", err
			}

			str.WriteByte('\r')
			str.WriteByte('\n')
		case '\\':
			if !atBackslash {
				atBackslash = true
				continue
			}
			str.WriteByte(b)
		case '"':
			if !atBackslash {
				return str.String(), nil
			}
			str.WriteByte(b)
		default:
			str.WriteByte(b)
		}
		atBackslash = false
	}
}

func multilineString(r *bufio.Reader, state *lexerState) (string, error) {
	atLF := false
	atLFHadDot := false
	var data strings.Builder
	for {
		b, err := r.ReadByte()
		if err != nil {
			return "", err
		}
		state.Col++
		// We also normalize LF into CRLF while reading multiline strings.
		switch b {
		case '.':
			if atLF {
				atLFHadDot = true
			} else {
				data.WriteByte('.')
				atLFHadDot = false
			}

			atLF = false
		case '\r', '\n':
			if err := r.UnreadByte(); err != nil {
				return "", err
			}
			if err := consumeCRLF(r, state); err != nil {
				return "", err
			}
			if atLFHadDot {
				return data.String(), nil
			}
			data.WriteByte('\r')
			data.WriteByte('\n')
			atLF = true
		default:
			if atLFHadDot {
				data.WriteByte('.')
			}
			atLF = false
			atLFHadDot = false
			data.WriteByte(b)
		}
	}
}

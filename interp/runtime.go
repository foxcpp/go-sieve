package interp

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/fs"
	"strings"

	"github.com/foxcpp/go-sieve/lexer"
)

type PolicyReader interface {
	RedirectAllowed(ctx context.Context, d *RuntimeData, addr string) (bool, error)
}

// Env provides access to named environment items for the
// environment test (RFC 5183). Implementations return ("", false) for unknown
// or unavailable items.
type Env interface {
	// GetEnvironment returns the value of the named item and whether it exists.
	GetEnvironment(name string) (value string, ok bool)
}

type Envelope interface {
	EnvelopeFrom() string
	EnvelopeTo() string
	AuthUsername() string
}

type Message interface {
	/*
		HeaderGet returns the header field value.

		RFC requires the following handling for encoded fields:

		      Comparisons are performed on octets.  Implementations convert text
		      from header fields in all charsets [MIME3] to Unicode, encoded as
		      UTF-8, as input to the comparator (see section 2.7.3).
		      Implementations MUST be capable of converting US-ASCII, ISO-8859-
		      1, the US-ASCII subset of ISO-8859-* character sets, and UTF-8.
		      Text that the implementation cannot convert to Unicode for any
		      reason MAY be treated as plain US-ASCII (including any [MIME3]
		      syntax) or processed according to local conventions.  An encoded
		      NUL octet (character zero) SHOULD NOT cause early termination of
		      the header content being compared against.
	*/
	HeaderGet(key string) ([]string, error)
	MessageSize() int
}

type BodyPart interface {
	ContentType() string
	Open(ctx context.Context) (io.ReadCloser, error)
}

// BodyPartBytes is the minimal implementation of BodyPart
// that uses []byte to store whole body part.
type BodyPartBytes struct {
	Blob             []byte
	ContentTypeValue string
}

func (b BodyPartBytes) ContentType() string {
	return b.ContentTypeValue
}

func (b BodyPartBytes) Open(ctx context.Context) (io.ReadCloser, error) {
	return io.NopCloser(bytes.NewReader(b.Blob)), nil
}

type BodyPartRaw struct {
	BodyMessage
}

func (BodyPartRaw) ContentType() string {
	return ""
}

func (b BodyPartRaw) Open(ctx context.Context) (io.ReadCloser, error) {
	r, err := b.BodyRaw(ctx)
	if err != nil {
		return nil, err
	}
	if r == nil {
		return nil, ErrNoBody
	}
	return io.NopCloser(r), nil
}

// BodyMessage is an optional extension of the Message interface that provides
// access to the message body for the body extension (RFC 5173).
//
// Implementations should implement this interface to support body tests.
// If the Message does not implement BodyMessage, body tests will always
// return false (no body available).
type BodyMessage interface {
	// BodyRaw returns the raw, undecoded body of the message as a stream.
	// This is used for the :raw body transform.
	// The body starts after the first empty line following the headers.
	// The returned reader must be closed by the caller.
	// If there is no body (header-only message without empty line separator),
	// it should return nil, nil.
	BodyRaw(ctx context.Context) (io.Reader, error)

	// BodyParts returns decoded MIME parts whose content-type matches any of
	// the given patterns. This is used for the :content and :text body transforms.
	//
	// Each pattern may be:
	//   - "" (empty string): matches all content types
	//   - "type" (no slash): matches any subtype of the given type
	//   - "type/subtype": matches exactly that content-type pair
	//   - patterns beginning/ending with '/' or with multiple slashes: match nothing
	// ContentTypeMatches function can be used to help determine whether each
	// content-type of each body part matches provided patterns.
	//
	// For multipart/* parts: returns the prologue and epilogue as separate strings.
	// For message/rfc822 parts: returns only the header as a single string.
	// For other parts: returns the decoded body content as UTF-8.
	//
	// Returns nil slice (not an error) if no matching parts are found.
	//
	// If Content-Transfer-Encoding is used for a part, the returned reader is expected to
	// be decoded accordingly (e.g. base64 or quoted-printable decoded).
	//
	// Since Sieve operates on UTF-8 or 7-bit ASCII strings, the returned readers
	// should provide decoded content as UTF-8, performing necessary decoding
	// if a non-Unicode charset is specified in Content-Type.
	BodyParts(ctx context.Context, contentTypes []string) ([]BodyPart, error)
}

type ExecuteTestMessage struct {
	Envelope Envelope
	Message  Message
	Flags    []string
}

// ExecuteTestEnvironment is a mock MTA interface used for integration tests. It is possible to test
// the actual execution of actions (fileinto, redirect, keep, etc.) using this interface in a particular
// MTA implementation.
type ExecuteTestEnvironment interface {
	CreateMailbox(name string) error
	// GetDefaultMailbox returns the mailbox name used by keep action.
	GetDefaultMailbox() string
	ExecuteActions(d *RuntimeData, actions []AppliedAction) error
	// GetSMTPMessage returns the message that was enqueued for delivery using redirect or other
	// similar command.
	GetSMTPMessage(index int) (*ExecuteTestMessage, error)
	HasSMTPMessage(index int) (bool, error)
	// GetMailboxMessage returns the message that was saved using either fileinfo or keep command.
	GetMailboxMessage(mailboxName string, index int) (*ExecuteTestMessage, error)
	HasMailboxMessage(mailboxName string, index int) (bool, error)
}

type TestRuntime struct {
	Name         string
	FailMessage  string
	FailAt       lexer.Position
	Script       *Script
	SavedScripts map[string][]byte
	MaxNesting   int
	Execute      ExecuteTestEnvironment

	// Msg, Envelope from RuntimeData saved here when test_message
	// is used.
	OriginalEnvelope Envelope
	OriginalMsg      Message
	OriginalFlags    Flags
}

type RuntimeData struct {
	Policy   PolicyReader
	Envelope Envelope
	Msg      Message
	Script   *Script
	Env      Env
	// For files accessible vis "include", "test_script_compile", etc.
	Namespace fs.FS

	ifResult bool

	// OnAction is called each time a Sieve script requests some action. If an error is
	// returned - script execution is aborted. Default handler (set by NewRuntimeData)
	// appends applied action to AppliedActions. If OnAction is replaced, it must
	// still append all actions to AppliedActions as it is used for some internal checks.
	OnAction       func(ctx context.Context, action AppliedAction, d *RuntimeData) error
	AppliedActions []AppliedAction

	RedirectAddr []string // Deprecated: Use AppliedActions instead.
	Mailboxes    []string // Deprecated: Use AppliedActions instead.
	Flags        []string // Default list of flags applied for actions.
	Keep         bool     // Deprecated: Use AppliedActions instead.
	ImplicitKeep bool     // Deprecated: Use AppliedActions instead.

	FlagAliases map[string]string

	MatchVariables []string
	Variables      map[string]string

	// vnd.dovecot.testsuite state, not intended for production use
	Test *TestRuntime
}

func (d *RuntimeData) Copy() *RuntimeData {
	newData := &RuntimeData{
		Policy:         d.Policy,
		Envelope:       d.Envelope,
		Msg:            d.Msg,
		Script:         d.Script,
		Env:            d.Env,
		Namespace:      d.Namespace,
		OnAction:       d.OnAction,
		AppliedActions: make([]AppliedAction, len(d.AppliedActions)),
		RedirectAddr:   make([]string, len(d.RedirectAddr)),
		Mailboxes:      make([]string, len(d.Mailboxes)),
		Keep:           d.Keep,
		ImplicitKeep:   d.ImplicitKeep,
		FlagAliases:    make(map[string]string, len(d.FlagAliases)),
		MatchVariables: make([]string, len(d.MatchVariables)),
		Variables:      make(map[string]string, len(d.Variables)),
		Test:           d.Test,
	}

	copy(newData.AppliedActions, d.AppliedActions)
	copy(newData.RedirectAddr, d.RedirectAddr)
	copy(newData.Mailboxes, d.Mailboxes)
	if d.Flags != nil {
		d.Flags = make([]string, len(d.Flags))
		copy(newData.Flags, d.Flags)
	}
	copy(newData.MatchVariables, d.MatchVariables)

	for k, v := range d.FlagAliases {
		newData.FlagAliases[k] = v
	}
	for k, v := range d.Variables {
		newData.Variables[k] = v
	}

	return newData
}

func (d *RuntimeData) MatchVariable(i int) string {
	if i >= len(d.MatchVariables) {
		return ""
	}
	return d.MatchVariables[i]
}

func (d *RuntimeData) Var(name string) (string, error) {
	namespace, name, ok := strings.Cut(strings.ToLower(name), ".")
	if !ok {
		name = namespace
		namespace = ""
	}

	switch namespace {
	case "envelope":
		// >  References to namespaces without a prior require statement for the
		// >  relevant extension MUST cause an error.
		if !d.Script.RequiresExtension("envelope") {
			return "", fmt.Errorf("require 'envelope' to use corresponding variables")
		}
		switch name {
		case "from":
			return d.Envelope.EnvelopeFrom(), nil
		case "to":
			return d.Envelope.EnvelopeTo(), nil
		case "auth":
			return d.Envelope.AuthUsername(), nil
		default:
			return "", nil
		}
	case "":
		// User variables.
		return d.Variables[name], nil
	default:
		return "", fmt.Errorf("unknown extension variable: %v", name)
	}
}

func (d *RuntimeData) SetVar(name, value string) error {
	if len(name) > d.Script.opts.MaxVariableNameLen {
		return fmt.Errorf("attempting to use a too long variable name: %v", name)
	}
	if len(value) > d.Script.opts.MaxVariableLen {
		until := d.Script.opts.MaxVariableLen
		// If this truncated an otherwise valid Unicode character,
		// remove the character altogether.
		for until > 0 && value[until] >= 128 && value[until] < 192 /* second or further octet of UTF-8 encoding */ {
			until--
		}

		value = value[:until]

	}

	namespace, name, ok := strings.Cut(strings.ToLower(name), ".")
	if !ok {
		name = namespace
		namespace = ""
	}

	switch namespace {
	case "envelope":
		return fmt.Errorf("cannot modify envelope. variables")
	case "":
		// User variables.
		d.Variables[name] = value
		return nil
	default:
		return fmt.Errorf("unknown extension variable: %v", name)
	}
}

func DefaultOnAction(ctx context.Context, action AppliedAction, d *RuntimeData) error {
	d.AppliedActions = append(d.AppliedActions, action)
	return nil
}

func NewRuntimeData(s *Script, p PolicyReader, e Envelope, m Message) *RuntimeData {
	return &RuntimeData{
		Script:       s,
		Policy:       p,
		Envelope:     e,
		Msg:          m,
		OnAction:     DefaultOnAction,
		ImplicitKeep: true,
		FlagAliases:  make(map[string]string),
		Variables:    map[string]string{},
	}
}

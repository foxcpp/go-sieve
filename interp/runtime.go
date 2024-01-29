package interp

import (
	"context"
	"fmt"
	"io/fs"
	"strings"

	"github.com/foxcpp/go-sieve/lexer"
)

type PolicyReader interface {
	RedirectAllowed(ctx context.Context, d *RuntimeData, addr string) (bool, error)
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

type RuntimeData struct {
	Policy   PolicyReader
	Envelope Envelope
	Msg      Message
	Script   *Script
	// For files accessible vis "include", "test_script_compile", etc.
	Namespace fs.FS

	ifResult bool

	RedirectAddr []string
	Mailboxes    []string
	Flags        []string
	Keep         bool
	ImplicitKeep bool

	FlagAliases map[string]string

	MatchVariables []string
	Variables      map[string]string

	// vnd.dovecot.testsuit state
	testName        string
	testFailMessage string // if set - test failed.
	testFailAt      lexer.Position
	testScript      *Script // script loaded using test_script_compile
	testMaxNesting  int     // max nesting for scripts loaded using test_script_compile
}

func (d *RuntimeData) Copy() *RuntimeData {
	newData := &RuntimeData{
		Policy:          d.Policy,
		Envelope:        d.Envelope,
		Msg:             d.Msg,
		Script:          d.Script,
		Namespace:       d.Namespace,
		RedirectAddr:    make([]string, len(d.RedirectAddr)),
		Mailboxes:       make([]string, len(d.Mailboxes)),
		Flags:           make([]string, len(d.Flags)),
		Keep:            d.Keep,
		ImplicitKeep:    d.ImplicitKeep,
		FlagAliases:     make(map[string]string, len(d.FlagAliases)),
		MatchVariables:  make([]string, len(d.MatchVariables)),
		Variables:       make(map[string]string, len(d.Variables)),
		testName:        d.testName,
		testFailMessage: d.testFailMessage,
		testFailAt:      d.testFailAt,
		testScript:      d.testScript,
		testMaxNesting:  d.testMaxNesting,
	}

	copy(newData.RedirectAddr, d.RedirectAddr)
	copy(newData.Mailboxes, d.Mailboxes)
	copy(newData.Flags, d.Flags)
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

func NewRuntimeData(s *Script, p PolicyReader, e Envelope, m Message) *RuntimeData {
	return &RuntimeData{
		Script:       s,
		Policy:       p,
		Envelope:     e,
		Msg:          m,
		ImplicitKeep: true,
		FlagAliases:  make(map[string]string),
		Variables:    map[string]string{},
	}
}

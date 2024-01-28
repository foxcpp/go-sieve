package interp

import (
	"context"
	"io/fs"
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

	Variables map[string]string

	// vnd.dovecot.testsuit state
	testName        string
	testFailMessage string  // if set - test failed.
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
		Variables:       make(map[string]string, len(d.Variables)),
		testName:        d.testName,
		testFailMessage: d.testFailMessage,
		testScript:      d.testScript,
		testMaxNesting:  d.testMaxNesting,
	}

	copy(newData.RedirectAddr, d.RedirectAddr)
	copy(newData.Mailboxes, d.Mailboxes)
	copy(newData.Flags, d.Flags)

	for k, v := range d.FlagAliases {
		newData.FlagAliases[k] = v
	}
	for k, v := range d.Variables {
		newData.Variables[k] = v
	}

	return newData
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

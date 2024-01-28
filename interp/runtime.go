package interp

import "context"

type PolicyReader interface {
	RedirectAllowed(ctx context.Context, d *RuntimeData, addr string) (bool, error)
}

type Message interface {
	EnvelopeFrom() string
	EnvelopeTo() string

	HeaderGet(key string) (string, bool, error)

	MessageSize() int
}

type Callback struct {
	RedirectAllowed func(ctx context.Context, d *RuntimeData, addr string) (bool, error)
	HeaderGet       func(value string) (string, bool, error)
}

type SMTPEnvelope struct {
	From string
	To   string
}

type RuntimeData struct {
	Policy   PolicyReader
	Msg      Message
	Script   *Script
	Callback Callback

	ifResult bool

	RedirectAddr []string
	Mailboxes    []string
	Flags        []string
	Keep         bool
	ImplicitKeep bool

	FlagAliases map[string]string
}

func NewRuntimeData(s *Script, p PolicyReader, m Message) *RuntimeData {
	return &RuntimeData{
		Script:       s,
		Policy:       p,
		Msg:          m,
		ImplicitKeep: true,
		FlagAliases:  make(map[string]string),
	}
}

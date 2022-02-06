package interp

import "context"

type Callback struct {
	RedirectAllowed func(ctx context.Context, d *RuntimeData, addr string) (bool, error)
	HeaderGet       func(value string) (string, bool, error)
}

type SMTPEnvelope struct {
	From string
	To   string
}

type RuntimeData struct {
	Script      *Script
	Callback    Callback
	SMTP        SMTPEnvelope
	MessageSize int

	ifResult bool

	RedirectAddr []string
	Mailboxes    []string
	Keep         bool
	ImplicitKeep bool
}

func NewRuntimeData(s *Script, p Callback) *RuntimeData {
	return &RuntimeData{
		Script:       s,
		Callback:     p,
		ImplicitKeep: true,
	}
}

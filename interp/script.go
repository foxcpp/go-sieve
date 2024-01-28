package interp

import (
	"context"
	"errors"
)

type Cmd interface {
	Execute(ctx context.Context, d *RuntimeData) error
}

type Options struct {
	MaxRedirects int

	// Enable vnd.dovecot.testsuite extension. Use for testing only.
	AllowDovecotTests bool
}

type Script struct {
	extensions map[string]struct{}
	cmd        []Cmd

	opts *Options
}

var ErrStop = errors.New("interpreter: stop called")

func (s Script) Extensions() []string {
	exts := make([]string, 0, len(s.extensions))
	for ext := range s.extensions {
		exts = append(exts, ext)
	}
	return exts
}

func (s Script) RequiresExtension(name string) bool {
	_, ok := s.extensions[name]
	return ok
}

func (s Script) Execute(ctx context.Context, d *RuntimeData) error {
	for _, c := range s.cmd {
		if err := c.Execute(ctx, d); err != nil {
			if errors.Is(err, ErrStop) {
				return nil
			}
			return err
		}
	}
	return nil
}

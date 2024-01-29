package interp

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/foxcpp/go-sieve/lexer"
)

type Cmd interface {
	Execute(ctx context.Context, d *RuntimeData) error
}

type Options struct {
	MaxRedirects int

	MaxVariableCount   int
	MaxVariableNameLen int
	MaxVariableLen     int

	// If specified - enables vnd.dovecot.testsuite extension
	// and will execute tests.
	T             *testing.T
	DisabledTests []string
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

func (s *Script) IsVarUsable(variableName string) (settable, gettable bool) {
	if len(variableName) > s.opts.MaxVariableNameLen {
		return false, false
	}

	namespace, name, ok := strings.Cut(strings.ToLower(variableName), ".")
	if !ok {
		name = namespace
		namespace = ""
	}

	if !lexer.IsValidIdentifier(name) {
		return false, false
	}

	switch namespace {
	case "envelope":
		if !s.RequiresExtension("envelope") {
			return false, false
		}
		return false, true
	case "":
		return true, true
	default:
		return false, false
	}
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

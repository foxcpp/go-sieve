package sieve

import (
	"io"

	"github.com/foxcpp/go-sieve/interp"
	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

type (
	Script      = interp.Script
	RuntimeData = interp.RuntimeData

	PolicyReader = interp.PolicyReader
	Message      = interp.Message
	Envelope     = interp.Envelope

	Options struct {
		Lexer  lexer.Options
		Parser parser.Options
		Interp interp.Options
	}
)

func DefaultOptions() Options {
	return Options{
		Lexer: lexer.Options{
			MaxTokens: 5000,
		},
		Parser: parser.Options{
			MaxBlockNesting: 15,
			MaxTestNesting:  15,
		},
		Interp: interp.Options{
			MaxRedirects:       5,
			MaxVariableCount:   128,
			MaxVariableNameLen: 32,
			MaxVariableLen:     4000,
		},
	}
}

func Load(r io.Reader, opts Options) (*Script, error) {
	toks, err := lexer.Lex(r, &opts.Lexer)
	if err != nil {
		return nil, err
	}

	cmds, err := parser.Parse(lexer.NewStream(toks), &opts.Parser)
	if err != nil {
		return nil, err
	}

	return interp.LoadScript(cmds, &opts.Interp)
}

func NewRuntimeData(s *Script, p interp.PolicyReader, e interp.Envelope, msg interp.Message) *interp.RuntimeData {
	return interp.NewRuntimeData(s, p, e, msg)
}

package sieve

import (
	"io"

	"github.com/foxcpp/go-sieve/interp"
	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

type (
	Script      = interp.Script
	RuntimeDate = interp.RuntimeData

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
			MaxRedirects: 5,
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

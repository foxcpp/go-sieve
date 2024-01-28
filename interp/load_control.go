package interp

import (
	"fmt"

	"github.com/foxcpp/go-sieve/parser"
)

func loadRequire(s *Script, pcmd parser.Cmd) (Cmd, error) {
	var exts []string
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				Optional: false,
				MatchStr: func(val []string) {
					exts = val
				},
				MinStrCount: 1,
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return nil, err
	}

	for _, ext := range exts {
		if ext == DovecotTestExtension {
			if s.opts.T == nil {
				return nil, fmt.Errorf("testing environment is not available, cannot use vnd.dovecot.testsuite")
			}
			s.extensions[DovecotTestExtension] = struct{}{}
			continue
		}

		if _, ok := supportedRequires[ext]; !ok {
			return nil, fmt.Errorf("loadRequire: unsupported extension: %v", ext)
		}
		s.extensions[ext] = struct{}{}
	}
	return nil, nil
}

func loadIf(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdIf{}
	err := LoadSpec(s, &Spec{
		AddTest: func(t Test) {
			cmd.Test = t
		},
		AddBlock: func(cmds []Cmd) {
			cmd.Block = cmds
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadElsif(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdElsif{}
	err := LoadSpec(s, &Spec{
		AddTest: func(t Test) {
			cmd.Test = t
		},
		AddBlock: func(cmds []Cmd) {
			cmd.Block = cmds
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadElse(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdElse{}
	err := LoadSpec(s, &Spec{
		AddBlock: func(cmds []Cmd) {
			cmd.Block = cmds
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadStop(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdStop{}
	err := LoadSpec(s, &Spec{}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

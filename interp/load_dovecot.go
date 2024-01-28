package interp

import (
	"fmt"

	"github.com/foxcpp/go-sieve/parser"
)

func loadDovecotTestSet(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension(DovecotTestExtension) || s.opts.T == nil {
		return nil, fmt.Errorf("testing environment is not enabled")
	}
	cmd := CmdDovecotTestSet{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.VariableName = val[0]
				},
			},
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.VariableValue = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadDovecotTestFail(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension(DovecotTestExtension) || s.opts.T == nil {
		return nil, fmt.Errorf("testing environment is not enabled")
	}
	cmd := CmdDovecotTestFail{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Message = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadDovecotTest(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension(DovecotTestExtension) || s.opts.T == nil {
		return nil, fmt.Errorf("testing environment is not enabled")
	}
	cmd := CmdDovecotTest{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.TestName = val[0]
				},
			},
		},
		AddBlock: func(cmds []Cmd) {
			cmd.Cmds = cmds
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadDovecotCompile(s *Script, test parser.Test) (Test, error) {
	loaded := TestDovecotCompile{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.ScriptPath = val[0]
				},
				MinStrCount: 1,
				MaxStrCount: 1,
			},
		},
	}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadDovecotRun(s *Script, test parser.Test) (Test, error) {
	loaded := TestDovecotRun{}
	err := LoadSpec(s, &Spec{}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadDovecotError(s *Script, test parser.Test) (Test, error) {
	loaded := TestDovecotTestError{}
	err := LoadSpec(s, &Spec{}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

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
				NoVariables: true,
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
	if err != nil {
		return nil, err
	}

	return cmd, nil
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
	cmd.At = pcmd.Position
	if err != nil {
		return nil, err
	}

	if !usedVarsAreValid(s, cmd.Message) {
		return nil, parser.ErrorAt(pcmd.Position, "invalid variable used: %v", cmd.Message)
	}

	return cmd, nil
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

func loadDovecotConfigSet(s *Script, pcmd parser.Cmd) (Cmd, error) {
	loaded := CmdDovecotConfigSet{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Key = val[0]
				},
				MinStrCount: 1,
				MaxStrCount: 1,
			},
			{
				MatchStr: func(val []string) {
					loaded.Value = val[0]
				},
				MinStrCount: 1,
				MaxStrCount: 1,
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return loaded, err
}

func loadDovecotConfigUnset(s *Script, pcmd parser.Cmd) (Cmd, error) {
	loaded := CmdDovecotConfigSet{
		Unset: true,
	}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Key = val[0]
				},
				MinStrCount: 1,
				MaxStrCount: 1,
			},
			{
				MatchStr: func(val []string) {
					loaded.Value = val[0]
				},
				MinStrCount: 1,
				MaxStrCount: 1,
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return loaded, err
}

func loadDovecotRun(s *Script, test parser.Test) (Test, error) {
	loaded := TestDovecotRun{}
	err := LoadSpec(s, &Spec{}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadDovecotError(s *Script, test parser.Test) (Test, error) {
	loaded := TestDovecotTestError{matcherTest: newMatcherTest()}
	err := LoadSpec(s, loaded.addSpecTags(&Spec{
		Tags: map[string]SpecTag{
			"index": {
				NeedsValue:  true,
				MinStrCount: 1,
				MaxStrCount: 1,
				NoVariables: true,
				MatchNum:    func(val int) {},
			},
		},
		Pos: []SpecPosArg{
			{
				MatchStr:    func(val []string) {},
				MinStrCount: 1,
			},
		},
	}), test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

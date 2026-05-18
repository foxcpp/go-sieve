package interp

import (
	"strings"

	"github.com/foxcpp/go-sieve/parser"
)

type Flags []string

func canonicalFlags(src []string, remove Flags, aliases map[string]string) Flags {
	if len(src) == 0 {
		return nil
	}

	// This does four things
	// * Translate space delimited lists of flags into separate flags
	// * Handle flag aliases
	// * Deduplicate
	// * Sort
	// * (optionally) remove flags

	toRemoveMap := make(map[string]struct{}, len(remove))
	for _, f := range remove {
		if fc, ok := aliases[f]; ok {
			toRemoveMap[fc] = struct{}{}
		} else {
			toRemoveMap[f] = struct{}{}
		}
	}

	c := make(Flags, 0, len(src))
	fm := make(map[string]struct{}, len(src))
	for _, fl := range src {
		if fl == "" {
			continue
		}
		for _, f := range strings.Fields(fl) {
			if _, ok := fm[f]; ok {
				continue
			}

			if fc, ok := aliases[f]; ok {
				f = fc
			}

			if _, toRemove := toRemoveMap[f]; toRemove {
				continue
			}

			c = append(c, f)
			fm[f] = struct{}{}
		}
	}

	return c
}

func loadFileInto(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("fileinto") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'fileinto")
	}
	cmd := CmdFileInto{}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"flags": {
				NeedsValue:  true,
				MinStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Flags = canonicalFlags(val, nil, nil)
				},
			},
			"copy": {
				NeedsValue: false,
				MatchBool: func() {
					cmd.Copy = true
				},
			},
		},
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Mailbox = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return nil, err
	}

	if !s.RequiresExtension("imap4flags") && cmd.Flags != nil {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}
	if cmd.Copy && !s.RequiresExtension("copy") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'copy'")
	}

	return cmd, nil
}

func loadRedirect(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdRedirect{}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"copy": {
				NeedsValue: false,
				MatchBool: func() {
					cmd.Copy = true
				},
			},
		},
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Addr = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return nil, err
	}

	if cmd.Copy && !s.RequiresExtension("copy") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'copy'")
	}

	return cmd, nil
}

func loadKeep(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdKeep{}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"flags": {
				NeedsValue:  true,
				MinStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Flags = canonicalFlags(val, nil, nil)
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return nil, err
	}

	if !s.RequiresExtension("imap4flags") && cmd.Flags != nil {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}

	return cmd, nil
}

func loadDiscard(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdDiscard{}
	err := LoadSpec(s, &Spec{}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadReject(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("reject") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'reject'")
	}
	cmd := CmdReject{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Reason = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func loadEReject(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("ereject") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'ereject'")
	}
	cmd := CmdEReject{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Reason = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return nil, err
	}
	return cmd, nil
}

func loadFlagCmd(s *Script, pcmd parser.Cmd) (variable string, flags []string, err error) {
	var arg1, arg2 []string
	err = LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					arg1 = val
				},
			},
			{
				MatchStr: func(val []string) {
					arg2 = val
				},
				Optional: true,
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	if err != nil {
		return "", nil, err
	}

	if len(arg2) == 0 {
		if len(arg1) == 0 {
			return "", nil, parser.ErrorAt(pcmd.Position, "missing required flags")
		}
		flags = canonicalFlags(arg1, nil, nil)
	} else {
		if len(arg1) != 1 {
			return "", nil, parser.ErrorAt(pcmd.Position, "expected only one string as a variable name")
		}
		variable = arg1[0]
		flags = canonicalFlags(arg2, nil, nil)
	}
	return variable, flags, nil
}

func loadSetFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}

	cmd := CmdSetFlag{}

	variable, flag, err := loadFlagCmd(s, pcmd)
	if err != nil {
		return nil, err
	}
	cmd.Variable = variable
	cmd.Flags = flag

	return cmd, nil
}

func loadAddFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}

	cmd := CmdAddFlag{}

	variable, flag, err := loadFlagCmd(s, pcmd)
	if err != nil {
		return nil, err
	}
	cmd.Variable = variable
	cmd.Flags = flag
	return cmd, nil
}

func loadRemoveFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}

	cmd := CmdRemoveFlag{}

	variable, flag, err := loadFlagCmd(s, pcmd)
	if err != nil {
		return nil, err
	}
	cmd.Variable = variable
	cmd.Flags = flag

	return cmd, nil
}

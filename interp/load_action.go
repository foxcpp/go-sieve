package interp

import (
	"sort"
	"strings"

	"github.com/foxcpp/go-sieve/parser"
)

type Flags []string

func canonicalFlags(src []string, remove Flags, aliases map[string]string) Flags {
	// This does four things
	// * Translate space delimited lists of flags into separate flags
	// * Handle flag aliases
	// * Deduplicate
	// * Sort
	// * (optionally) remove flags
	c := make(Flags, 0, len(src))
	fm := make(map[string]struct{})
	for _, fl := range src {
		for _, f := range strings.Split(fl, " ") {
			if fc, ok := aliases[f]; ok {
				fm[fc] = struct{}{}
			} else {
				fm[f] = struct{}{}
			}
		}
	}
	if remove != nil {
		for _, fl := range remove {
			for _, f := range strings.Split(fl, " ") {
				if fc, ok := aliases[f]; ok {
					delete(fm, fc)
				} else {
					delete(fm, f)
				}
			}
		}
	}
	for f := range fm {
		c = append(c, f)
	}
	sort.Strings(c)
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

	return cmd, nil
}

func loadRedirect(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdRedirect{}
	err := LoadSpec(s, &Spec{
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

func loadSetFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}
	cmd := CmdSetFlag{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
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

	return cmd, nil
}

func loadAddFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}
	cmd := CmdAddFlag{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
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

	return cmd, nil
}

func loadRemoveFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'imap4flags")
	}
	cmd := CmdRemoveFlag{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
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

	return cmd, nil
}

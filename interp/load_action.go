package interp

import (
	"fmt"
	"sort"
	"strings"

	"github.com/foxcpp/go-sieve/parser"
)

type Flags []string

func canonicalFlags(src []string, remove *Flags, aliases map[string]string) *Flags {
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
		for _, fl := range *remove {
			for _, f := range strings.Split(fl, " ") {
				if fc, ok := aliases[f]; ok {
					delete(fm, fc)
				} else {
					delete(fm, f)
				}
			}
		}
	}
	for f, _ := range fm {
		c = append(c, f)
	}
	sort.Strings(c)
	return &c
}

func loadFileInto(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("fileinto") {
		return nil, fmt.Errorf("require fileinto to use it")
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
	if !s.RequiresExtension("imap4flags") && cmd.Flags != nil {
		return nil, fmt.Errorf("require imap4flags to use it")
	}
	return cmd, err
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
	return cmd, err
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
	if !s.RequiresExtension("imap4flags") && cmd.Flags != nil {
		return nil, fmt.Errorf("require imap4flags to use it")
	}
	return cmd, err
}

func loadDiscard(s *Script, pcmd parser.Cmd) (Cmd, error) {
	cmd := CmdDiscard{}
	err := LoadSpec(s, &Spec{}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)
	return cmd, err
}

func loadSetFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, fmt.Errorf("require impa4flags to use it")
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
	return cmd, err
}

func loadAddFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, fmt.Errorf("require impa4flags to use it")
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
	return cmd, err
}

func loadRemoveFlag(s *Script, pcmd parser.Cmd) (Cmd, error) {
	if !s.RequiresExtension("imap4flags") {
		return nil, fmt.Errorf("require impa4flags to use it")
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
	return cmd, err
}

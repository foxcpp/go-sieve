package interp

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/foxcpp/go-sieve/parser"
)

func loadSet(script *Script, pcmd parser.Cmd) (Cmd, error) {
	if !script.RequiresExtension("variables") {
		return nil, parser.ErrorAt(pcmd.Position, "missing require 'variables'")
	}
	cmd := CmdSet{}

	// by precedence
	var modifiers = map[int]func(string) string{}
	var conflictingMods bool

	err := LoadSpec(script, &Spec{
		Tags: map[string]SpecTag{
			"length": {
				MatchBool: func() {
					if modifiers[10] != nil {
						conflictingMods = true
					}
					modifiers[10] = func(s string) string {
						// RFC mentions `characters' and not octets
						return strconv.Itoa(len([]rune(s)))
					}
				},
			},
			"quotewildcard": {
				MatchBool: func() {
					if modifiers[20] != nil {
						conflictingMods = true
					}
					modifiers[20] = func(s string) string {
						escaped := strings.Builder{}
						escaped.Grow(len(s))
						for _, chr := range s {
							switch chr {
							case '\\', '*', '?':
								escaped.WriteByte('\\')
								escaped.WriteRune(chr)
							default:
								escaped.WriteRune(chr)
							}
						}
						return escaped.String()
					}
				},
			},
			"upper": {
				MatchBool: func() {
					if modifiers[40] != nil {
						conflictingMods = true
					}
					modifiers[40] = func(s string) string {
						return strings.ToUpper(s)
					}
				},
			},
			"lower": {
				MatchBool: func() {
					if modifiers[40] != nil {
						conflictingMods = true
					}
					modifiers[40] = func(s string) string {
						return strings.ToLower(s)
					}
				},
			},
			"upperfirst": {
				MatchBool: func() {
					if modifiers[30] != nil {
						conflictingMods = true
					}
					modifiers[30] = func(s string) string {
						if len(s) == 0 {
							return s
						}
						first := s[0]
						if first >= 'a' && first <= 'z' {
							first -= 'a' - 'A'
						}
						return string(first) + s[1:]
					}
				},
			},
			"lowerfirst": {
				MatchBool: func() {
					if modifiers[30] != nil {
						conflictingMods = true
					}
					modifiers[30] = func(s string) string {
						if len(s) == 0 {
							return s
						}
						first := s[0]
						if first >= 'A' && first <= 'Z' {
							first += 'a' - 'A'
						}
						return string(first) + s[1:]
					}
				},
			},
		},
		Pos: []SpecPosArg{
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Name = strings.ToLower(val[0])
				},
			},
			{
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					cmd.Value = val[0]
				},
			},
		},
	}, pcmd.Position, pcmd.Args, pcmd.Tests, pcmd.Block)

	if conflictingMods {
		return nil, parser.ErrorAt(pcmd.Position, "conflicting value modifiers")
	}

	settable, _ := script.IsVarUsable(cmd.Name)
	if !settable {
		return nil, parser.ErrorAt(pcmd.Position, "cannot set this variable")
	}

	cmd.ModifyValue = func(s string) string {
		lastPrec := 9999
		for _, prec := range [4]int{40, 30, 20, 10} {
			fun := modifiers[prec]
			if fun != nil {
				s = fun(s)
				lastPrec = prec
			}
		}

		// If last run modifier was quotewildcard - check
		// whether created value would remain valid
		// if truncated to MaxVariableLen. If so, truncate
		// here and remove dangling backslashes (if any).
		if lastPrec == 20 {
			if len(s) > script.opts.MaxVariableLen {
				until := script.opts.MaxVariableLen

				// (Copy-pasted from RuntimeData.SetVar)
				// If this truncated an otherwise valid Unicode character,
				// remove the character altogether.
				for until > 0 && s[until] >= 128 && s[until] < 192 /* second or further octet of UTF-8 encoding */ {
					until--
				}

				if s[until-1] == '\\' {
					until--
				}

				s = s[:until]
			}
		}

		return s
	}

	return cmd, err
}

func loadStringTest(s *Script, test parser.Test) (Test, error) {
	if !s.RequiresExtension("variables") {
		return nil, fmt.Errorf("missing require 'variables'")
	}

	loaded := TestString{matcherTest: newMatcherTest()}
	var key []string
	err := LoadSpec(s, loaded.addSpecTags(&Spec{
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Source = val
				},
				MinStrCount: 1,
			},
			{
				MatchStr: func(val []string) {
					key = val
				},
				MinStrCount: 1,
			},
		},
	}), test.Position, test.Args, test.Tests, nil)
	if err != nil {
		return nil, err
	}

	if err := loaded.setKey(s, key); err != nil {
		return nil, err
	}

	return loaded, nil
}

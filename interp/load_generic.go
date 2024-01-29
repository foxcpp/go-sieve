package interp

import (
	"strings"

	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

type SpecTag struct {
	NeedsValue bool
	MatchStr   func(val []string)
	MatchNum   func(val int)
	MatchBool  func()

	// Checks for used string list.
	MinStrCount int
	MaxStrCount int

	// Toggle checks for valid variable names.
	NoVariables bool
}

type SpecPosArg struct {
	Optional bool
	MatchStr func(val []string)
	MatchNum func(i int)

	// Checks for used string list.
	MinStrCount int
	MaxStrCount int

	// Toggle checks for valid variable names.
	NoVariables bool
}

type Spec struct {
	Tags          map[string]SpecTag
	Pos           []SpecPosArg
	AddBlock      func([]Cmd)
	BlockOptional bool
	AddTest       func(Test)
	TestOptional  bool
	MultipleTests bool
}

func LoadSpec(s *Script, spec *Spec, position lexer.Position, args []parser.Arg, tests []parser.Test, block []parser.Cmd) error {
	var lastTag *SpecTag
	nextPosArg := 0
	for _, a := range args {
		switch a := a.(type) {
		case parser.StringArg:
			if lastTag != nil && lastTag.NeedsValue {
				if lastTag.MatchNum != nil {
					return lexer.ErrorAt(a, "LoadSpec: tagged argument requires a number, got string-list")
				} else if lastTag.MatchStr != nil {
					value := a.Value
					if s.RequiresExtension("encoded-character") {
						var err error
						value, err = decodeEncodedChars(value)
						if err != nil {
							return lexer.ErrorAt(position, "LoadSpec: malformed encoded character sequence: %v", err)
						}
					}
					if s.RequiresExtension("variables") && !lastTag.NoVariables {

					}

					lastTag.MatchStr([]string{value})
				} else {
					panic("missing matcher for SpecTag")
				}
				lastTag = nil
				continue
			}
			if nextPosArg >= len(spec.Pos) {
				return lexer.ErrorAt(a, "LoadSpec: too many arguments")
			}
			pos := spec.Pos[nextPosArg]
			if pos.MinStrCount > 1 {
				return lexer.ErrorAt(a, "LoadSpec: string-list required, got single string")
			}
			if pos.MatchNum != nil {
				return lexer.ErrorAt(a, "LoadSpec: argument requires a number, got string-list")
			} else if pos.MatchStr != nil {
				value := a.Value
				if s.RequiresExtension("encoded-character") {
					var err error
					value, err = decodeEncodedChars(value)
					if err != nil {
						return lexer.ErrorAt(position, "LoadSpec: malformed encoded character sequence: %v", err)
					}
				}

				pos.MatchStr([]string{value})
			} else {
				panic("no pos matcher")
			}
			nextPosArg++
		case parser.StringListArg:
			if lastTag != nil && lastTag.NeedsValue {
				if lastTag.MatchNum != nil {
					return lexer.ErrorAt(a, "LoadSpec: tagged argument requires a number, got string-list")
				} else if lastTag.MatchStr != nil {
					if (lastTag.MinStrCount != 0 && len(a.Value) < lastTag.MinStrCount) || (lastTag.MaxStrCount != 0 && len(a.Value) > lastTag.MaxStrCount) {
						return lexer.ErrorAt(a, "LoadSpec: wrong amount of string arguments")
					}

					value := a.Value
					if s.RequiresExtension("encoded-character") {
						for i := range value {
							var err error
							value[i], err = decodeEncodedChars(value[i])
							if err != nil {
								return lexer.ErrorAt(position, "LoadSpec: malformed encoded character sequence: %v", err)
							}
						}
					}

					lastTag.MatchStr(value)
				} else {
					panic("missing matcher for SpecTag")
				}
				lastTag = nil
				continue
			}

			if nextPosArg >= len(spec.Pos) {
				return lexer.ErrorAt(a, "LoadSpec: too many arguments")
			}
			pos := spec.Pos[nextPosArg]
			if (pos.MinStrCount != 0 && len(a.Value) < pos.MinStrCount) || (pos.MaxStrCount != 0 && len(a.Value) > pos.MaxStrCount) {
				return lexer.ErrorAt(a, "LoadSpec: wrong amount of string arguments")
			}
			if pos.MatchNum != nil {
				return lexer.ErrorAt(a, "LoadSpec: argument requires a number, got string-list")
			} else if pos.MatchStr != nil {
				value := a.Value
				if s.RequiresExtension("encoded-character") {
					for i := range value {
						var err error
						value[i], err = decodeEncodedChars(value[i])
						if err != nil {
							return lexer.ErrorAt(position, "LoadSpec: malformed encoded character sequence: %v", err)
						}
					}
				}

				pos.MatchStr(value)
			} else {
				panic("no pos matcher")
			}
			nextPosArg++
		case parser.NumberArg:
			if lastTag != nil && lastTag.NeedsValue {
				if lastTag.MatchStr != nil {
					return lexer.ErrorAt(a, "LoadSpec: tagged argument requires a string-list, got number")
				} else if lastTag.MatchNum != nil {
					lastTag.MatchNum(a.Value)
				} else {
					panic("missing matcher for SpecTag")
				}
				lastTag = nil
				continue
			}

			if nextPosArg >= len(spec.Pos) {
				return lexer.ErrorAt(a, "LoadSpec: too many arguments")
			}
			pos := spec.Pos[nextPosArg]
			if pos.MatchStr != nil {
				return lexer.ErrorAt(a, "LoadSpec: argument requires a string-list, got number")
			} else if pos.MatchNum != nil {
				pos.MatchNum(a.Value)
			} else {
				panic("no pos matcher")
			}
			nextPosArg++
		case parser.TagArg:
			if lastTag != nil && lastTag.NeedsValue {
				return lexer.ErrorAt(a, "LoadSpec: tagged argument requires a value")
			}
			tag, ok := spec.Tags[strings.ToLower(a.Value)]
			if !ok {
				return lexer.ErrorAt(a, "LoadSpec: unknown tagged argument: %v", a.Value)
			}
			if tag.NeedsValue {
				lastTag = &tag
			} else {
				tag.MatchBool()
			}
		}
	}
	for i := nextPosArg; i < len(spec.Pos); i++ {
		if !spec.Pos[i].Optional {
			return lexer.ErrorAt(position, "LoadSpec: %d argument is required", i+1)
		}
	}

	if spec.AddTest == nil {
		if len(tests) != 0 {
			return lexer.ErrorAt(position, "LoadSpec: no tests allowed")
		}
	} else {
		if len(tests) == 0 && !spec.TestOptional {
			return lexer.ErrorAt(position, "LoadSpec: at least one test required")
		}
		if len(tests) > 1 && !spec.MultipleTests {
			return lexer.ErrorAt(position, "LoadSpec: only one test allowed")
		}
		for _, t := range tests {
			loadedTest, err := LoadTest(s, t)
			if err != nil {
				return err
			}
			spec.AddTest(loadedTest)
		}
	}
	if spec.AddBlock != nil {
		if block != nil {
			loadedCmds, err := LoadBlock(s, block)
			if err != nil {
				return err
			}
			spec.AddBlock(loadedCmds)
		} else if !spec.BlockOptional {
			return lexer.ErrorAt(position, "LoadSpec: block is required")
		}
	} else if block != nil {
		return lexer.ErrorAt(position, "LoadSpec: no block allowed")
	}
	return nil
}

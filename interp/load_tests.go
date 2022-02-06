package interp

import (
	"fmt"

	"github.com/foxcpp/go-sieve/parser"
)

func loadAddressTest(s *Script, test parser.Test) (Test, error) {
	loaded := AddressTest{
		Comparator:  ComparatorOctet,
		AddressPart: All,
		Match:       MatchIs,
	}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"comparator": {
				NeedsValue:  true,
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					loaded.Comparator = Comparator(val[0])
				},
			},
			"is": {
				MatchBool: func() {
					loaded.Match = MatchIs
				},
			},
			"contains": {
				MatchBool: func() {
					loaded.Match = MatchContains
				},
			},
			"matches": {
				MatchBool: func() {
					loaded.Match = MatchMatches
				},
			},
			"all": {
				MatchBool: func() {
					loaded.AddressPart = All
				},
			},
			"localpart": {
				MatchBool: func() {
					loaded.AddressPart = LocalPart
				},
			},
			"domain": {
				MatchBool: func() {
					loaded.AddressPart = Domain
				},
			},
		},
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Header = val
				},
				MinStrCount: 1,
			},
			{
				MatchStr: func(val []string) {
					loaded.Key = val
				},
				MinStrCount: 1,
			},
		},
	}, test.Position, test.Args, test.Tests, nil)
	switch loaded.Comparator {
	case ComparatorOctet, ComparatorUnicodeCaseMap,
		ComparatorASCIICaseMap, ComparatorASCIINumeric:
	default:
		return nil, fmt.Errorf("unsupported comparator: %v", loaded.Comparator)
	}
	return loaded, err
}

func loadAllOfTest(s *Script, test parser.Test) (Test, error) {
	loaded := AllOfTest{}
	err := LoadSpec(s, &Spec{
		AddTest: func(t Test) {
			loaded.Tests = append(loaded.Tests, t)
		},
		MultipleTests: true,
	}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadAnyOfTest(s *Script, test parser.Test) (Test, error) {
	loaded := AnyOfTest{}
	err := LoadSpec(s, &Spec{
		AddTest: func(t Test) {
			loaded.Tests = append(loaded.Tests, t)
		},
		MultipleTests: true,
	}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadEnvelopeTest(s *Script, test parser.Test) (Test, error) {
	if !s.RequiresExtension("envelope") {
		return nil, fmt.Errorf("require envelope to use it")
	}
	loaded := EnvelopeTest{
		Comparator:  ComparatorOctet,
		AddressPart: All,
		Match:       MatchIs,
	}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"comparator": {
				NeedsValue:  true,
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					loaded.Comparator = Comparator(val[0])
				},
			},
			"is": {
				MatchBool: func() {
					loaded.Match = MatchIs
				},
			},
			"contains": {
				MatchBool: func() {
					loaded.Match = MatchContains
				},
			},
			"matches": {
				MatchBool: func() {
					loaded.Match = MatchMatches
				},
			},
			"all": {
				MatchBool: func() {
					loaded.AddressPart = All
				},
			},
			"localpart": {
				MatchBool: func() {
					loaded.AddressPart = LocalPart
				},
			},
			"domain": {
				MatchBool: func() {
					loaded.AddressPart = Domain
				},
			},
		},
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Field = val
				},
				MinStrCount: 1,
			},
			{
				MatchStr: func(val []string) {
					loaded.Key = val
				},
				MinStrCount: 1,
			},
		},
	}, test.Position, test.Args, test.Tests, nil)
	if err != nil {
		return nil, err
	}
	switch loaded.Comparator {
	case ComparatorOctet, ComparatorUnicodeCaseMap,
		ComparatorASCIICaseMap, ComparatorASCIINumeric:
	default:
		return nil, fmt.Errorf("unsupported comparator: %v", loaded.Comparator)
	}
	return loaded, nil
}

func loadExistsTest(s *Script, test parser.Test) (Test, error) {
	loaded := ExistsTest{}
	err := LoadSpec(s, &Spec{
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Fields = val
				},
				MinStrCount: 1,
			},
		},
	}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadFalseTest(s *Script, test parser.Test) (Test, error) {
	loaded := FalseTest{}
	err := LoadSpec(s, &Spec{}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadTrueTest(s *Script, test parser.Test) (Test, error) {
	loaded := TrueTest{}
	err := LoadSpec(s, &Spec{}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadHeaderTest(s *Script, test parser.Test) (Test, error) {
	loaded := HeaderTest{
		Comparator: ComparatorOctet,
		Match:      MatchIs,
	}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"comparator": {
				NeedsValue:  true,
				MinStrCount: 1,
				MaxStrCount: 1,
				MatchStr: func(val []string) {
					loaded.Comparator = Comparator(val[0])
				},
			},
			"is": {
				MatchBool: func() {
					loaded.Match = MatchIs
				},
			},
			"contains": {
				MatchBool: func() {
					loaded.Match = MatchContains
				},
			},
			"matches": {
				MatchBool: func() {
					loaded.Match = MatchMatches
				},
			},
		},
		Pos: []SpecPosArg{
			{
				MatchStr: func(val []string) {
					loaded.Header = val
				},
				MinStrCount: 1,
			},
			{
				MatchStr: func(val []string) {
					loaded.Key = val
				},
				MinStrCount: 1,
			},
		},
	}, test.Position, test.Args, test.Tests, nil)
	if err != nil {
		return nil, err
	}
	switch loaded.Comparator {
	case ComparatorOctet, ComparatorUnicodeCaseMap,
		ComparatorASCIICaseMap, ComparatorASCIINumeric:
	default:
		return nil, fmt.Errorf("unsupported comparator: %v", loaded.Comparator)
	}
	return loaded, nil
}

func loadNotTest(s *Script, test parser.Test) (Test, error) {
	loaded := NotTest{}
	err := LoadSpec(s, &Spec{
		AddTest: func(t Test) {
			loaded.Test = t
		},
	}, test.Position, test.Args, test.Tests, nil)
	return loaded, err
}

func loadSizeTest(s *Script, test parser.Test) (Test, error) {
	loaded := SizeTest{}
	err := LoadSpec(s, &Spec{
		Tags: map[string]SpecTag{
			"under": {
				MatchBool: func() { loaded.Under = true },
			},
			"over": {
				MatchBool: func() { loaded.Over = true },
			},
		},
		Pos: []SpecPosArg{
			{
				MatchNum: func(i int) {
					loaded.Size = i
				},
			},
		},
	}, test.Position, test.Args, test.Tests, nil)
	if loaded.Under == loaded.Over {
		return nil, fmt.Errorf("loadSizeTest: either under or over is required")
	}
	return loaded, err
}

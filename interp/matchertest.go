package interp

import (
	"fmt"
	"strconv"
)

// matcherTest contains code shared between tests
// such as 'header', 'address', 'envelope', 'string' -
// all tests that compare some values from message
// with pre-defined "key"
type matcherTest struct {
	Comparator Comparator
	Match      Match
	Relational Relational
	Key        []string

	// Used for keys without variables.
	keyCompiled []CompiledMatcher

	matchCnt int
}

func newMatcherTest() matcherTest {
	return matcherTest{
		Comparator: DefaultComparator,
		Match:      MatchIs,
	}
}

func (t *matcherTest) addSpecTags(s *Spec) *Spec {
	if s.Tags == nil {
		s.Tags = make(map[string]SpecTag, 4)
	}
	s.Tags["comparator"] = SpecTag{
		NeedsValue:  true,
		MinStrCount: 1,
		MaxStrCount: 1,
		MatchStr: func(val []string) {
			t.Comparator = Comparator(val[0])
		},
		NoVariables: true,
	}
	s.Tags["is"] = SpecTag{
		MatchBool: func() {
			t.Match = MatchIs
			t.matchCnt++
		},
	}
	s.Tags["contains"] = SpecTag{
		MatchBool: func() {
			t.Match = MatchContains
			t.matchCnt++
		},
	}
	s.Tags["matches"] = SpecTag{
		MatchBool: func() {
			t.Match = MatchMatches
			t.matchCnt++
		},
	}
	s.Tags["value"] = SpecTag{
		NeedsValue:  true,
		MinStrCount: 1,
		MaxStrCount: 1,
		NoVariables: true,
		MatchStr: func(val []string) {
			t.Match = MatchValue
			t.matchCnt++
			t.Relational = Relational(val[0])
		},
	}
	s.Tags["count"] = SpecTag{
		NeedsValue:  true,
		MinStrCount: 1,
		MaxStrCount: 1,
		NoVariables: true,
		MatchStr: func(val []string) {
			t.Match = MatchCount
			t.matchCnt++
			t.Relational = Relational(val[0])
		},
	}
	return s
}

func (t *matcherTest) setKey(s *Script, k []string) error {
	t.Key = k

	if t.matchCnt > 1 {
		return fmt.Errorf("multiple match-types are not allowed")
	}

	if t.Match == MatchCount || t.Match == MatchValue {
		if !s.RequiresExtension("relational") {
			return fmt.Errorf("missing require 'relational'")
		}
		switch t.Relational {
		case RelGreaterThan, RelGreaterOrEqual,
			RelLessThan, RelLessOrEqual, RelEqual,
			RelNotEqual:
		default:
			return fmt.Errorf("unknown relational operator: %v", t.Relational)
		}
	}

	caseFold := false
	octet := false
	switch t.Comparator {
	case ComparatorOctet:
		octet = true
	case ComparatorUnicodeCaseMap:
		caseFold = true
	case ComparatorASCIICaseMap:
		octet = true
		caseFold = true
	case ComparatorASCIINumeric:
	default:
		return fmt.Errorf("unsupported comparator: %v", t.Comparator)
	}

	if t.Match == MatchMatches {
		t.keyCompiled = make([]CompiledMatcher, len(t.Key))
		for i := range t.Key {
			if len(usedVars(s, t.Key[i])) > 0 {
				continue
			}

			var err error
			t.keyCompiled[i], err = compileMatcher(t.Key[i], octet, caseFold)
			if err != nil {
				return fmt.Errorf("malformed pattern (%v): %v", t.Key[i], err)
			}
		}
	}

	if t.Match == MatchCount && t.Comparator != ComparatorASCIINumeric {
		return fmt.Errorf("non-numeric comparators cannot be used with :count")
	}

	return nil
}

func (t *matcherTest) isCount() bool {
	return t.Match == MatchCount
}

func (t *matcherTest) countMatches(d *RuntimeData, value uint64) bool {
	if !t.isCount() {
		panic("countMatches can be called only with MatchCount matcher")
	}

	for _, k := range t.Key {
		kNum, err := strconv.ParseUint(expandVars(d, k), 10, 64)
		if err != nil {
			continue
		}

		if t.Relational.CompareUint64(value, kNum) {
			return true
		}
	}

	return false
}

func (t *matcherTest) tryMatch(d *RuntimeData, source string) (bool, error) {
	for i, key := range t.Key {
		var (
			ok      bool
			matches []string
			err     error
		)
		if t.keyCompiled != nil && t.keyCompiled[i].IsLoaded() {
			ok, matches, err = t.keyCompiled[i].Match(source)
		} else {
			key = expandVars(d, key)
			ok, matches, err = testString(t.Comparator, t.Match, t.Relational, source, expandVars(d, key))
		}
		if err != nil {
			return false, err
		}
		if ok {
			if t.Match == MatchMatches {
				d.MatchVariables = matches
			}
			return true, nil
		}
	}
	return false, nil
}

package interp

import (
	"fmt"
)

// matcherTest contains code shared between tests
// such as 'header', 'address', 'envelope', 'string' -
// all tests that compare some values from message
// with pre-defined "key"
type matcherTest struct {
	comparator Comparator
	match      Match
	key        []string

	// Used for keys without
	keyCompiled []CompiledMatcher

	matchCnt int
}

func newMatcherTest() matcherTest {
	return matcherTest{
		comparator: DefaultComparator,
		match:      MatchIs,
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
			t.comparator = Comparator(val[0])
		},
	}
	s.Tags["is"] = SpecTag{
		MatchBool: func() {
			t.match = MatchIs
			t.matchCnt++
		},
	}
	s.Tags["contains"] = SpecTag{
		MatchBool: func() {
			t.match = MatchContains
			t.matchCnt++
		},
	}
	s.Tags["matches"] = SpecTag{
		MatchBool: func() {
			t.match = MatchMatches
			t.matchCnt++
		},
	}
	return s
}

func (t *matcherTest) setKey(s *Script, k []string) error {
	t.key = k

	if t.matchCnt > 1 {
		return fmt.Errorf("multiple match-types are not allowed")
	}

	caseFold := false
	octet := false
	switch t.comparator {
	case ComparatorOctet:
		octet = true
	case ComparatorUnicodeCaseMap:
		caseFold = true
	case ComparatorASCIICaseMap:
		octet = true
		caseFold = true
	case ComparatorASCIINumeric:
	default:
		return fmt.Errorf("unsupported comparator: %v", t.comparator)
	}

	if t.match == MatchMatches {
		t.keyCompiled = make([]CompiledMatcher, len(t.key))
		for i := range t.key {
			if len(usedVars(s, t.key[i])) > 0 {
				continue
			}

			var err error
			t.keyCompiled[i], err = compileMatcher(t.key[i], octet, caseFold)
			if err != nil {
				return fmt.Errorf("malformed pattern (%v): %v", t.key[i], err)
			}
		}
	}

	return nil
}

func (t *matcherTest) tryMatch(d *RuntimeData, source string) (bool, error) {
	for i, key := range t.key {
		var (
			ok      bool
			matches []string
			err     error
		)
		if t.keyCompiled != nil && t.keyCompiled[i] != nil {
			ok, matches, err = t.keyCompiled[i](source)
		} else {
			key = expandVars(d, key)
			ok, matches, err = testString(t.comparator, t.match, source, expandVars(d, key))
		}
		if err != nil {
			return false, err
		}
		if ok {
			if t.match == MatchMatches {
				d.MatchVariables = matches
			}
			return true, nil
		}
	}
	return false, nil
}

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
	comparator Comparator
	match      Match
	relational Relational
	key        []string

	// Used for keys without variables.
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
		NoVariables: true,
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
	s.Tags["value"] = SpecTag{
		NeedsValue:  true,
		MinStrCount: 1,
		MaxStrCount: 1,
		NoVariables: true,
		MatchStr: func(val []string) {
			t.match = MatchValue
			t.matchCnt++
			t.relational = Relational(val[0])
		},
	}
	s.Tags["count"] = SpecTag{
		NeedsValue:  true,
		MinStrCount: 1,
		MaxStrCount: 1,
		NoVariables: true,
		MatchStr: func(val []string) {
			t.match = MatchCount
			t.matchCnt++
			t.relational = Relational(val[0])
		},
	}
	return s
}

func (t *matcherTest) setKey(s *Script, k []string) error {
	t.key = k

	if t.matchCnt > 1 {
		return fmt.Errorf("multiple match-types are not allowed")
	}

	if t.match == MatchCount || t.match == MatchValue {
		if !s.RequiresExtension("relational") {
			return fmt.Errorf("missing require 'relational'")
		}
		switch t.relational {
		case RelGreaterThan, RelGreaterOrEqual,
			RelLessThan, RelLessOrEqual, RelEqual,
			RelNotEqual:
		default:
			return fmt.Errorf("unknown relational operator: %v", t.relational)
		}
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

	if t.match == MatchCount && t.comparator != ComparatorASCIINumeric {
		return fmt.Errorf("non-numeric comparators cannot be used with :count")
	}

	return nil
}

func (t *matcherTest) isCount() bool {
	return t.match == MatchCount
}

func (t *matcherTest) countMatches(d *RuntimeData, value uint64) bool {
	if !t.isCount() {
		panic("countMatches can be called only with MatchCount matcher")
	}
	
	for _, k := range t.key {
		kNum, err := strconv.ParseUint(expandVars(d, k), 10, 64)
		if err != nil {
			continue
		}

		if t.relational.CompareUint64(value, kNum) {
			return true
		}
	}

	return false
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
			ok, matches, err = testString(t.comparator, t.match, t.relational, source, expandVars(d, key))
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

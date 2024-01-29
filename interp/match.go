package interp

import (
	"regexp"
	"strings"

	"rsc.io/binaryregexp"
)

func foldASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

func patternToRegex(pattern string, caseFold bool) string {
	result := strings.Builder{}
	if caseFold {
		result.WriteString(`(?i)`)
	}
	result.WriteRune('^')
	escaped := false
	for _, chr := range pattern {
		if !escaped {
			switch chr {
			case '\\':
				escaped = true
			case '?':
				result.WriteString(`(.)`)
			case '*':
				result.WriteString(`(.*?)`)
			case '.', '+', '(', ')', '|', '[', ']', '{', '}', '^', '$':
				result.WriteRune('\\')
				fallthrough
			default:
				result.WriteRune(chr)
			}
		} else {
			switch chr {
			case '\\', '?', '*', '.', '+', '(', ')', '|', '[', ']', '{', '}', '^', '$':
				result.WriteRune('\\')
				fallthrough
			default:
				result.WriteRune(chr)
			}

			escaped = false
		}
	}

	// Such regex won't compile.
	if escaped {
		return result.String()
	}

	result.WriteRune('$')

	return result.String()
}

type CompiledMatcher func(value string) (bool, []string, error)

// compileMatcher returns a function that will check whether pre-defined pattern matches the passed
// value. It is preferable to use compileMatcher over matchOctet, matchUnicode if
// pattern does not change often (e.g. does not depend on any variables).
func compileMatcher(pattern string, octet bool, caseFold bool) (CompiledMatcher, error) {
	if octet {
		regex, err := binaryregexp.Compile(patternToRegex(pattern, caseFold))
		if err != nil {
			return nil, err
		}

		return func(value string) (bool, []string, error) {
			matches := regex.FindStringSubmatch(value)
			return len(matches) != 0, matches, nil
		}, nil
	}

	regex, err := regexp.Compile(patternToRegex(pattern, caseFold))
	if err != nil {
		return nil, err
	}

	return func(value string) (bool, []string, error) {
		matches := regex.FindStringSubmatch(value)
		return len(matches) != 0, matches, nil
	}, nil
}

func matchOctet(pattern, value string, caseFold bool) (bool, []string, error) {
	regex, err := binaryregexp.Compile(patternToRegex(pattern, caseFold))
	if err != nil {
		return false, nil, err
	}

	matches := regex.FindStringSubmatch(value)
	return len(matches) != 0, matches, nil
}

func matchUnicode(pattern, value string, caseFold bool) (bool, []string, error) {
	regex, err := regexp.Compile(patternToRegex(pattern, caseFold))
	if err != nil {
		return false, nil, err
	}

	matches := regex.FindStringSubmatch(value)
	return len(matches) != 0, matches, nil
}

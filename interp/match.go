package interp

import (
	"unicode"

	"github.com/foxcpp/go-sieve/interp/match"
)

func foldASCII(b byte) byte {
	if b >= 'A' && b <= 'Z' {
		return b + 32
	}
	return b
}

func matchOctet(pattern, value string, caseFold bool) (bool, []string, error) {
	var fold func(b byte) byte
	if caseFold {
		fold = foldASCII
	}

	ok, matches, err := match.Match([]byte(pattern), []byte(value), fold)
	if err != nil {
		return false, nil, err
	}
	matchesDec := make([]string, len(matches))
	for i := range matches {
		matchesDec[i] = string(matches[i])
	}
	return ok, matchesDec, nil
}

func matchUnicode(pattern, value string, caseFold bool) (bool, []string, error) {
	var fold func(r rune) rune
	if caseFold {
		fold = unicode.SimpleFold
	}

	ok, matches, err := match.Match([]rune(pattern), []rune(value), fold)
	if err != nil {
		return false, nil, err
	}
	matchesDec := make([]string, len(matches))
	for i := range matches {
		matchesDec[i] = string(matches[i])
	}
	return ok, matchesDec, nil
}

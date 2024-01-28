package interp

import "github.com/foxcpp/go-sieve/interp/match"

func matchOctet(pattern, value string) (bool, error) {
	return match.Match([]byte(pattern), []byte(value))
}

func matchUnicode(pattern, value string) (bool, error) {
	return match.Match([]rune(pattern), []rune(value))
}

// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package match

import (
	"errors"
)

// ErrBadPattern indicates a pattern was malformed.
var ErrBadPattern = errors.New("syntax error in pattern")

// Match is a simplified version of path.Match that removes support
// for character classes and operates on already decoded input ([]byte or []rune).
func Match[T rune | byte](pattern, name []T, caseFold func(T) T) (matched bool, matches [][]T, err error) {
	matches = append(matches, name)
Pattern:
	for len(pattern) > 0 {
		var star bool
		var chunk []T
		star, chunk, pattern = scanChunk(pattern)
		if star && len(chunk) == 0 {
			matches = append(matches, name)
			return true, matches, nil
		}
		// Look for match at current position.
		t, subMatches, ok, err := matchChunk(chunk, name, caseFold)
		// if we're the last chunk, make sure we've exhausted the name
		// otherwise we'll give a false result even if we could still match
		// using the star
		matches = append(matches, subMatches...)
		if ok && (len(t) == 0 || len(pattern) > 0) {
			name = t
			continue
		}
		if err != nil {
			return false, matches, err
		}
		if star {
			// Look for match skipping i+1 bytes.
			for i := 0; i < len(name); i++ {
				t, subMatches, ok, err := matchChunk(chunk, name[i+1:], caseFold)
				if ok {
					// if we're the last chunk, make sure we exhausted the name
					if len(pattern) == 0 && len(t) > 0 {
						continue
					}
					matches = append(matches, subMatches...)
					matches = append(matches, name[:i+1])
					name = t
					continue Pattern
				}
				if err != nil {
					return false, matches, err
				}
			}
		}
		// Before returning false with no error,
		// check that the remainder of the pattern is syntactically valid.
		for len(pattern) > 0 {
			_, chunk, pattern = scanChunk(pattern)
			if _, _, _, err := matchChunk(chunk, []T{}, caseFold); err != nil {
				return false, matches, err
			}
		}
		return false, matches, nil
	}
	return len(name) == 0, matches, nil
}

// scanChunk gets the next segment of pattern, which is a non-star string
// possibly preceded by a star.
func scanChunk[T rune | byte](pattern []T) (star bool, chunk, rest []T) {
	for len(pattern) > 0 && pattern[0] == '*' {
		pattern = pattern[1:]
		star = true
	}
	var i int
Scan:
	for i = 0; i < len(pattern); i++ {
		switch pattern[i] {
		case '\\':
			// error check handled in matchChunk: bad pattern.
			if i+1 < len(pattern) {
				i++
			}
		case '*':
			break Scan
		}
	}
	return star, pattern[0:i], pattern[i:]
}

// matchChunk checks whether chunk matches the beginning of s.
// If so, it returns the remainder of s (after the match).
// Chunk is all single-character operators: literals, char classes, and ?.
func matchChunk[T rune | byte](chunk, s []T, fold func(T) T) (rest []T, matches [][]T, ok bool, err error) {
	// failed records whether the match has failed.
	// After the match fails, the loop continues on processing chunk,
	// checking that the pattern is well-formed but no longer reading s.
	failed := false
	for len(chunk) > 0 {
		if !failed && len(s) == 0 {
			failed = true
		}
		switch chunk[0] {
		case '?':
			if !failed {
				matches = append(matches, []T{s[0]})
				s = s[1:]
			}
			chunk = chunk[1:]

		case '\\':
			chunk = chunk[1:]
			if len(chunk) == 0 {
				return []T{}, matches, false, ErrBadPattern
			}
			fallthrough

		default:
			if !failed {
				if fold != nil {
					if fold(chunk[0]) != fold(s[0]) {
						failed = true
					}
				} else {
					if chunk[0] != s[0] {
						failed = true
					}
				}
				s = s[1:]
			}
			chunk = chunk[1:]
		}
	}
	if failed {
		return []T{}, matches, false, nil
	}
	return s, matches, true, nil
}

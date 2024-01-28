package tests

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve"
)

func TestMatchContains(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "match-types", "contains.svtest"))
}

func TestMatchIs(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "match-types", "is.svtest"))
}

func TestMatchMatches(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "match-types", "matches.svtest"))
}

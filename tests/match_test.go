package tests

import (
	"path/filepath"
	"testing"
)

func TestMatchContains(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "match-types", "contains.svtest"))
}

func TestMatchIs(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "match-types", "is.svtest"))
}

func TestMatchMatches(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "match-types", "matches.svtest"))
}

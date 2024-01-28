package tests

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve"
)

func TestComparatorsOctet(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "comparators", "i-octet.svtest"))
}

func TestComparatorsASCIICasemap(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "comparators", "i-ascii-casemap.svtest"))
}

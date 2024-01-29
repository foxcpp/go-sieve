package tests

import (
	"path/filepath"
	"testing"
)

func TestComparatorsOctet(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "comparators", "i-octet.svtest"))
}

func TestComparatorsASCIICasemap(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "comparators", "i-ascii-casemap.svtest"))
}

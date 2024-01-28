package tests

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve"
)

func TestTestsuite(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "testsuite.svtest"))
}

func TestLexer(t *testing.T) {
	t.Skip("requires variables extension")
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "lexer.svtest"))
}

func TestControlIf(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "control-if.svtest"))
}

func TestControlStop(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "control-stop.svtest"))
}

func TestTestAddress(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-address.svtest"))
}

func TestTestAllof(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-allof.svtest"))
}

func TestTestAnyof(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-anyof.svtest"))
}

func TestTestExists(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-exists.svtest"))
}

func TestTestHeader(t *testing.T) {
	t.Skip("requires variables extension")
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-header.svtest"))
}

func TestTestSize(t *testing.T) {
	sieve.RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-size.svtest"))
}

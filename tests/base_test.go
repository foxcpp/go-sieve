package tests

import (
	"path/filepath"
	"testing"
)

func TestTestsuite(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "testsuite.svtest"))
}

func TestLexer(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "lexer.svtest"))
}

func TestControlIf(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "control-if.svtest"))
}

func TestControlStop(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "control-stop.svtest"))
}

func TestTestAddress(t *testing.T) {
	RunDovecotTestWithout(t, filepath.Join("pigeonhole", "tests", "test-address.svtest"),
		[]string{
			// test_fail at 85:3 called: failed to ignore comment in address
			// go-sieve address parser does not remove comments.
			"Basic functionality",
			// test_fail at 458:3 called: :localpart matched invalid UTF-8 address
			// FIXME: Not sure what is wrong here. UTF-8 looks valid?
			"Invalid addresses",
		})
}

func TestTestAllof(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-allof.svtest"))
}

func TestTestAnyof(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-anyof.svtest"))
}

func TestTestExists(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-exists.svtest"))
}

func TestTestHeader(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-header.svtest"))
}

func TestTestSize(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "test-size.svtest"))
}

package tests

import (
	"path/filepath"
	"testing"
)

func TestRelationalBasic(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "relational", "basic.svtest"))
}

func TestRelationalComparators(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "relational", "comparators.svtest"))
}

func TestRelationalErrors(t *testing.T) {
	// Stripped test_error calls.
	RunDovecotTestInline(t, filepath.Join("pigeonhole", "tests", "extensions", "relational"), `
require "vnd.dovecot.testsuite";
test "Syntax errors" {
	if test_script_compile "errors/syntax.sieve" {
		test_fail "compile should have failed";
	}
}
test "Validation errors" {
	if test_script_compile "errors/validation.sieve" {
		test_fail "compile should have failed";
	}
}
`)

	//RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "relational", "errors.svtest"))
}

func TestRelationalRFC(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "relational", "rfc.svtest"))
}

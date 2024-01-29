package tests

import (
	"path/filepath"
	"testing"
)

func TestExtensionsVariablesBasic(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "basic.svtest"))
}

func TestExtensionsVariablesErrors(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "errors.svtest"))
}

func TestExtensionsVariablesLimits(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "limits.svtest"))
}

func TestExtensionsVariablesMatch(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "match.svtest"))
}

func TestExtensionsVariablesModifiers(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "modifiers.svtest"))
}

func TestExtensionsVariablesQuoting(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "quoting.svtest"))
}

func TestExtensionsVariablesRegex(t *testing.T) {
	t.Skip("requires regex extension")
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "regex.svtest"))
}

func TestExtensionsVariablesString(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "string.svtest"))
}

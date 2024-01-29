package tests

import (
	"path/filepath"
	"testing"
)

func TestExtensionsVariablesBasic(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "basic.svtest"))
}

func TestExtensionsVariablesErrors(t *testing.T) {
	t.Skip("requires relational extension")
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "errors.svtest"))
}

func TestExtensionsVariablesLimits(t *testing.T) {
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "limits.svtest"))
}

func TestExtensionsVariablesMatch(t *testing.T) {
	RunDovecotTestWithout(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "match.svtest"),
		[]string{
			// :matches is non-conforming - * is greedy but shouldn't be.
			"RFC - non greedy",
			"RFC - example",
			"Words sep ?",
			"Letters words *? - 1",
			"Letters words *? - 2",
			"Letters words *? backtrack",
			"Letters words *? first",
		})
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
	t.Skip("requires relational extension")
	RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "variables", "string.svtest"))
}

package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestExecuteExamples(t *testing.T) {
	t.Skip("Not compatible yet")
	
	svTest, err := os.ReadFile(filepath.Join("pigeonhole", "tests", "execute", "examples.svtest"))
	if err != nil {
		t.Fatal(err)
	}
	svTestPatched := strings.ReplaceAll(string(svTest), "../../examples/", "examples/")

	RunDovecotTestInline(t, "pigeonhole", svTestPatched)
}

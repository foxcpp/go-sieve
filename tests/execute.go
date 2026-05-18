package tests

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/foxcpp/go-sieve/interp"
)

func RunExecuteTests(t *testing.T, env interp.ExecuteTestEnvironment) {
	t.Run("examples", func(t *testing.T) {
		t.Skip("needs reject extension")

		svTest, err := os.ReadFile(filepath.Join("pigeonhole", "tests", "execute", "examples.svtest"))
		if err != nil {
			t.Fatal(err)
		}
		svTestPatched := strings.ReplaceAll(string(svTest), "../../examples/", "examples/")

		RunDovecotTestInline(t, "pigeonhole", svTestPatched,
			ExecuteTestRuntime(env),
		)
	})
	t.Run("extensions/imap4flags", func(t *testing.T) {
		t.Skip("fileinto deduplication is not implemented in go-sieve")

		RunDovecotTest(t, filepath.Join("pigeonhole", "tests", "extensions", "imap4flags", "execute.svtest"),
			ExecuteTestRuntime(env),
		)
	})
	t.Run("extensions/reject", func(t *testing.T) {
		RunDovecotTest(t,
			filepath.Join("pigeonhole", "tests", "extensions", "reject", "execute.svtest"),
			ExecuteTestRuntime(env),
		)
	})
}

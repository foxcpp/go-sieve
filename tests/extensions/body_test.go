package extensions

import (
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve/tests"
)

func TestBodyExtension(t *testing.T) {
	base := filepath.Join("..", "pigeonhole", "tests", "extensions", "body")

	t.Run("basic", func(t *testing.T) {
		tests.RunDovecotTest(t, filepath.Join(base, "basic.svtest"))
	})
	t.Run("raw", func(t *testing.T) {
		tests.RunDovecotTest(t, filepath.Join(base, "raw.svtest"))
	})
	t.Run("content", func(t *testing.T) {
		tests.RunDovecotTest(t, filepath.Join(base, "content.svtest"))
	})
	t.Run("text", func(t *testing.T) {
		tests.RunDovecotTest(t, filepath.Join(base, "text.svtest"))
	})
	t.Run("match-values", func(t *testing.T) {
		tests.RunDovecotTest(t, filepath.Join(base, "match-values.svtest"))
	})
	t.Run("errors", func(t *testing.T) {
		tests.RunDovecotTest(t, filepath.Join(base, "errors.svtest"))
	})
}

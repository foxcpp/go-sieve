package sieve

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve/interp"
)

func RunDovecotTest(t *testing.T, path string) {
	svScript, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	opts := DefaultOptions()
	opts.Interp.T = t

	script, err := Load(bytes.NewReader(svScript), opts)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Empty data.
	data := NewRuntimeData(script, interp.DummyPolicy{},
		interp.EnvelopeStatic{}, interp.MessageStatic{})
	data.Namespace = os.DirFS(filepath.Dir(path))

	err = script.Execute(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
}

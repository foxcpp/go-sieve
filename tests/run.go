package tests

import (
	"bytes"
	"context"
	"os"
	"path/filepath"
	"testing"

	"github.com/foxcpp/go-sieve"
	"github.com/foxcpp/go-sieve/interp"
)

func RunDovecotTestWithout(t *testing.T, path string, disabledTests []string) {
	svScript, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}

	opts := sieve.DefaultOptions()
	opts.Interp.T = t
	opts.Interp.DisabledTests = disabledTests

	script, err := sieve.Load(bytes.NewReader(svScript), opts)
	if err != nil {
		t.Fatal(err)
	}

	ctx := context.Background()

	// Empty data.
	data := sieve.NewRuntimeData(script, interp.DummyPolicy{},
		interp.EnvelopeStatic{}, interp.MessageStatic{})
	data.Namespace = os.DirFS(filepath.Dir(path))

	err = script.Execute(ctx, data)
	if err != nil {
		t.Fatal(err)
	}
}

func RunDovecotTest(t *testing.T, path string) {
	RunDovecotTestWithout(t, path, nil)
}

//go:build go1.18
// +build go1.18

package lexer

import (
	"reflect"
	"strings"
	"testing"
)

func FuzzLex(f *testing.F) {
	f.Add(``)
	f.Add(`"hello"`)
	f.Add(`[ "hello", "there"]`)
	f.Add(`{ id1 id2 }`)
	f.Add(`{ id1 # comment id2 }`)
	f.Add(`{}`)
	f.Add(`"multi
line
string"`)
	f.Add(`
there are 
also
/* multi
line
comments */`)
	f.Add(`[ "hello" ] # comment parsing should also work
/* also a comment
whatever # aaaa
"still a comment"
{}
*/
{ identifier :size 123K } `)
	f.Fuzz(func(t *testing.T, script string) {
		toks, err := Lex(strings.NewReader(script), &Options{NoPosition: true})
		if err != nil {
			t.Skip(err)
		}
		out := strings.Builder{}
		if err := Write(&out, toks); err != nil {
			t.Fatal("Write should succeed for any Lex output:", err)
		}
		toks2, err := Lex(strings.NewReader(out.String()), &Options{NoPosition: true})
		if err != nil {
			t.Fatal("Lex should succeed for any Write output:", err)
		}
		if !reflect.DeepEqual(toks, toks2) {
			t.Log("Two Lex calls produced inconsistent output")
			t.Log("First: ", toks)
			t.Log("Second:", toks2)
			t.FailNow()
		}
	})
}

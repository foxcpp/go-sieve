package lexer

import (
	"reflect"
	"strings"
	"testing"
)

func testLexer(t *testing.T, script string, tokens []Token) {
	t.Run("case", func(t *testing.T) {
		actualTokens, err := Lex(strings.NewReader(script), &Options{})
		if err != nil {
			if tokens == nil {
				return
			}
			t.Error("Unexpected error:", err)
			return
		}
		if tokens == nil {
			t.Error("Unexpected success:", actualTokens)
			return
		}
		if !reflect.DeepEqual(tokens, actualTokens) {
			t.Log("Wrong lexer output:")
			t.Logf("Actual:   %#v", actualTokens)
			t.Logf("Expected: %#v", tokens)
			t.Fail()
		}
	})
}

func TestLex(t *testing.T) {
	testLexer(t, ``, []Token{})
	testLexer(t, `[]`, []Token{ListStart{Position: LineCol(1, 1)}, ListEnd{Position: LineCol(1, 2)}})
	testLexer(t, `[ "hello1" , "hello2" ]`, []Token{
		ListStart{Position: LineCol(1, 1)},
		String{Text: "hello1", Position: LineCol(1, 3)},
		Comma{Position: LineCol(1, 12)},
		String{Text: "hello2", Position: LineCol(1, 14)},
		ListEnd{LineCol(1, 23)},
	})
	testLexer(t, `"multi
line
string"`, []Token{String{Text: "multi\r\nline\r\nstring", Position: LineCol(1, 1)}})
	testLexer(t, `" and so it goes... `, nil) // lexer error
	testLexer(t, `[ "hello" ] id`, []Token{
		ListStart{Position: LineCol(1, 1)},
		String{Text: "hello", Position: LineCol(1, 3)},
		ListEnd{Position: LineCol(1, 11)},
		Identifier{Text: "id", Position: LineCol(1, 13)},
	})
	testLexer(t, `[ "hello" ]
/* also a comment
whatever # aaaa
"still a comment"
{}
*/
{ identifier :size 123K }`, []Token{
		ListStart{Position: LineCol(1, 1)},
		String{Text: "hello", Position: LineCol(1, 3)},
		ListEnd{Position: LineCol(1, 11)},
		BlockStart{Position: LineCol(7, 1)},
		Identifier{Text: "identifier", Position: LineCol(7, 3)},
		Colon{Position: LineCol(7, 14)},
		Identifier{Text: "size", Position: LineCol(7, 15)},
		Number{Value: 123, Quantifier: Kilo, Position: LineCol(7, 20)},
		BlockEnd{Position: LineCol(7, 25)},
	})
	testLexer(t, `set "message" text:
From: sirius@example.org
To: nico@frop.example.com
Subject: Frop!

Frop!
.
`, []Token{
		Identifier{Text: "set", Position: LineCol(1, 1)},
		String{Text: "message", Position: LineCol(1, 5)},
		String{Text: "From: sirius@example.org\r\n" +
			"To: nico@frop.example.com\r\n" +
			"Subject: Frop!\r\n" +
			"\r\n" +
			"Frop!\r\n", Position: LineCol(1, 15)},
	})
	testLexer(t, `set "message" text:
From: sirius@example.org
To: nico@frop.example.com
Subject: Frop!

..
Frop!
.
`, []Token{
		Identifier{Text: "set", Position: LineCol(1, 1)},
		String{Text: "message", Position: LineCol(1, 5)},
		String{Text: "From: sirius@example.org\r\n" +
			"To: nico@frop.example.com\r\n" +
			"Subject: Frop!\r\n" +
			"\r\n" +
			".\r\n" +
			"Frop!\r\n", Position: LineCol(1, 15)},
	})
}

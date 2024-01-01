package interp

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/foxcpp/go-sieve/lexer"
	"github.com/foxcpp/go-sieve/parser"
)

func testCmdLoader(t *testing.T, s *Script, in string, out []Cmd) {
	t.Run("case", func(t *testing.T) {
		toks, err := lexer.Lex(strings.NewReader(in), &lexer.Options{})
		if err != nil {
			t.Fatal("Lexer failed:", err)
		}
		inCmds, err := parser.Parse(lexer.NewStream(toks), &parser.Options{})
		if err != nil {
			t.Fatal("Parser failed:", err)
		}

		if testing.Verbose() {
			t.Log("Parse tree:")
			t.Log(spew.Sdump(inCmds))
		}

		actualCmd, err := LoadBlock(s, inCmds)
		if err != nil {
			if out != nil {
				t.Error("Unexpected error:", err)
			}
			return
		}
		if out == nil {
			t.Error("Unexpected success:", actualCmd)
			return
		}
		if !reflect.DeepEqual(actualCmd, out) {
			t.Log("Wrong LoadBlock output")
			t.Log("Actual:  ", actualCmd)
			t.Log("Expected:", out)
			t.Fail()
		}
	})
}

func TestLoadBlock(t *testing.T) {
	s := &Script{
		extensions: supportedExtensions,
	}
	testCmdLoader(t, s, `require ["envelope"];`, []Cmd{})
	testCmdLoader(t, s, `if true { }`, []Cmd{CmdIf{
		Test:  TrueTest{},
		Block: []Cmd{},
	}})
	testCmdLoader(t, s, `require "envelope";
require "fileinto";
if envelope :is "from" "test@example.org" {
	fileinto "hell";
}
`, []Cmd{
		CmdIf{
			Test: EnvelopeTest{
				Comparator:  ComparatorOctet,
				Match:       MatchIs,
				AddressPart: All,
				Field:       []string{"from"},
				Key:         []string{"test@example.org"},
			},
			Block: []Cmd{
				CmdFileInto{Mailbox: "hell"},
			},
		},
	})
	testCmdLoader(t, s, `require "imap4flags";
require "fileinto";
fileinto :flags "flag1 flag2" "hell";
keep :flags ["flag1", "flag2"];
setflag ["flag2", "flag1", "flag2"];
addflag ["flag2", "flag1"];
removeflag "flag2";
`, []Cmd{
		CmdFileInto{
			Mailbox: "hell",
			Flags:   &Flags{"flag1", "flag2"},
		},
		CmdKeep{
			Flags: &Flags{"flag1", "flag2"},
		},
		CmdSetFlag{
			Flags: &Flags{"flag1", "flag2"},
		},
		CmdAddFlag{
			Flags: &Flags{"flag1", "flag2"},
		},
		CmdRemoveFlag{
			Flags: &Flags{"flag2"},
		},
	})
}

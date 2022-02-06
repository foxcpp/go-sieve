package parser

import (
	"reflect"
	"strings"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/foxcpp/go-sieve/lexer"
)

const exampleScript = ` #
    # Example Sieve Filter
    # Declare any optional features or extension used by the script
    #
    require ["fileinto"];

    #
    # Handle messages from known mailing lists
    # Move messages from IETF filter discussion list to filter mailbox
    #
    if header :is "Sender" "owner-ietf-mta-filters@imc.org"
            {
            fileinto "filter";  # move to "filter" mailbox
            }
    #
    # Keep all messages to or from people in my company
    #
    elsif address :DOMAIN :is ["From", "To"] "example.com"
            {
            keep;               # keep in "In" mailbox
            }

    #
    # Try and catch unsolicited email.  If a message is not to me,
    # or it contains a subject known to be spam, file it away.
    #
    elsif anyof (NOT address :all :contains
                   ["To", "Cc", "Bcc"] "me@example.com",
                 header :matches "subject"
                   ["*make*money*fast*", "*university*dipl*mas*"])
            {
            fileinto "spam";   # move to "spam" mailbox
            }
    else
            {
            # Move all other (non-company) mail to "personal"
            # mailbox.
            fileinto "personal";
            }
`

func testParse(t *testing.T, script string, cmds []Cmd) {
	toks, err := lexer.Lex(strings.NewReader(script), &lexer.Options{
		NoPosition: true,
	})
	if err != nil {
		t.Fatal("Lexer failed:", err)
	}
	s := lexer.NewStream(toks)
	actualCmds, err := parse(s, 0, &Options{})
	if err != nil {
		t.Error("parse failed:", err)
		return
	}
	if err != nil {
		if cmds == nil {
			return
		}
		t.Error("Unexpected failure:", err)
		return
	}
	if cmds == nil {
		t.Error("Unexpected success:", actualCmds)
		return
	}
	if !reflect.DeepEqual(cmds, actualCmds) {
		t.Log("Wrong parse result")
		t.Log("Expected:")
		t.Log(spew.Sdump(cmds))
		t.Log("Actual:")
		t.Log(spew.Sdump(actualCmds))
		t.Fail()
	}
}

func TestParser(t *testing.T) {
	testParse(t, exampleScript, []Cmd{
		{
			Id: "require",
			Args: []Arg{
				StringListArg{Value: []string{"fileinto"}},
			},
		},
		{
			Id: "if",
			Tests: []Test{
				{
					Id: "header",
					Args: []Arg{
						TagArg{Value: "is"},
						StringArg{Value: "Sender"},
						StringArg{Value: "owner-ietf-mta-filters@imc.org"},
					},
				},
			},
			Block: []Cmd{
				{
					Id: "fileinto",
					Args: []Arg{
						StringArg{Value: "filter"},
					},
				},
			},
		},
		{
			Id: "elsif",
			Tests: []Test{
				{
					Id: "address",
					Args: []Arg{
						TagArg{Value: "DOMAIN"},
						TagArg{Value: "is"},
						StringListArg{Value: []string{"From", "To"}},
						StringArg{Value: "example.com"},
					},
				},
			},
			Block: []Cmd{
				{
					Id: "keep",
				},
			},
		},
		{
			Id: "elsif",
			Tests: []Test{
				{
					Id: "anyof",
					Tests: []Test{
						{
							Id: "NOT",
							Tests: []Test{
								{
									Id: "address",
									Args: []Arg{
										TagArg{Value: "all"},
										TagArg{Value: "contains"},
										StringListArg{Value: []string{"To", "Cc", "Bcc"}},
										StringArg{Value: "me@example.com"},
									},
								},
							},
						},
						{
							Id: "header",
							Args: []Arg{
								TagArg{Value: "matches"},
								StringArg{Value: "subject"},
								StringListArg{Value: []string{"*make*money*fast*", "*university*dipl*mas*"}},
							},
						},
					},
				},
			},
			Block: []Cmd{
				{
					Id: "fileinto",
					Args: []Arg{
						StringArg{Value: "spam"},
					},
				},
			},
		},
		{
			Id: "else",
			Block: []Cmd{
				{
					Id: "fileinto",
					Args: []Arg{
						StringArg{Value: "personal"},
					},
				},
			},
		},
	})
}

package parser

import (
	"reflect"
	"strings"
	"testing"

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
	toks, err := lexer.Lex(strings.NewReader(script), &lexer.Options{})
	if err != nil {
		t.Fatal("Lexer failed:", err)
	}
	s := lexer.NewStream(toks)
	actualCmds, err := parse(s, 0)
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
		t.Logf("Expected: %+v", cmds)
		t.Logf("Actual:   %+v", actualCmds)
		t.Fail()
	}
}

func TestParser(t *testing.T) {
	testParse(t, exampleScript, []Cmd{})
}

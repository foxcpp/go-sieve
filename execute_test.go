package sieve

import (
	"bufio"
	"bytes"
	"context"
	"net/textproto"
	"reflect"
	"strings"
	"testing"

	"github.com/foxcpp/go-sieve/interp"
)

var eml string = `Date: Tue, 1 Apr 1997 09:06:31 -0800 (PST)
From: coyote@desert.example.org
To: roadrunner@acme.example.com
Subject: I have a present for you

Look, I'm sorry about the whole anvil thing, and I really
didn't mean to try and drop it on you from the top of the
cliff.  I want to try to make it up to you.  I've got some
great birdseed over here at my place--top of the line
stuff--and if you come by, I'll have it all wrapped up
for you.  I'm really sorry for all the problems I've caused
for you over the years, but I know we can work this out.
--
Wile E. Coyote   "Super Genius"   coyote@desert.example.org
`

func testExecute(t *testing.T, in string, eml string, intendedResult []interp.AppliedAction) {
	t.Helper()

	msgHdr, err := textproto.NewReader(bufio.NewReader(strings.NewReader(eml))).ReadMIMEHeader()
	if err != nil {
		t.Fatal(err)
	}

	script := bufio.NewReader(strings.NewReader(in))

	loadedScript, err := Load(script, DefaultOptions())
	if err != nil {
		t.Fatal(err)
	}
	env := interp.EnvelopeStatic{
		From: "from@test.com",
		To:   "to@test.com",
	}
	msg := interp.MessageStatic{
		Size:   len(eml),
		Header: msgHdr,
	}
	data := interp.NewRuntimeData(loadedScript, interp.DummyPolicy{},
		env, msg)

	ctx := context.Background()
	if err := loadedScript.Execute(ctx, data); err != nil {
		t.Fatal(err)
	}

	if !reflect.DeepEqual(data.AppliedActions, intendedResult) {
		t.Log("Wrong Execute output")
		t.Logf("Actual:   %#v", data.AppliedActions)
		t.Logf("Expected: %#v", intendedResult)
		t.Fail()
	}

	parentFailed := t.Failed()

	t.Run("binary reloaded", func(t *testing.T) {
		if parentFailed {
			t.Skip("skipping binary reloaded if regular execution fails too")
		}

		savedScript, err := loadedScript.Save()
		restoredScript, err := RestoreSaved(bytes.NewReader(savedScript))
		if err != nil {
			t.Fatal(err)
		}

		env := interp.EnvelopeStatic{
			From: "from@test.com",
			To:   "to@test.com",
		}
		msg := interp.MessageStatic{
			Size:   len(eml),
			Header: msgHdr,
		}
		data := interp.NewRuntimeData(restoredScript, interp.DummyPolicy{},
			env, msg)

		ctx := context.Background()
		if err := restoredScript.Execute(ctx, data); err != nil {
			t.Fatal(err)
		}

		if !reflect.DeepEqual(data.AppliedActions, intendedResult) {
			t.Log("Wrong Execute output")
			t.Logf("Actual:   %#v", data.AppliedActions)
			t.Logf("Expected: %#v", intendedResult)
			t.Fail()
		}
	})
}

func TestFileinto(t *testing.T) {
	testExecute(t, `require ["fileinto"];
	fileinto "test";
`, eml,
		[]interp.AppliedAction{
			interp.ActionFileInto{Mailbox: "test"},
		})
	testExecute(t, `require ["fileinto"];
		fileinto "test";
		fileinto "test2";
	`, eml,
		[]interp.AppliedAction{
			interp.ActionFileInto{Mailbox: "test"},
			interp.ActionFileInto{Mailbox: "test2"},
		},
	)
}

func TestFlags(t *testing.T) {
	t.Run("flag2 flag3", func(t *testing.T) {
		testExecute(t, `require ["fileinto", "imap4flags"];
	setflag ["flag1", "flag2"];
	addflag ["flag2", "flag3"];
	removeflag ["flag1"];
	fileinto "test";
`, eml,
			[]interp.AppliedAction{
				interp.ActionFileInto{
					Mailbox: "test",
					Flags:   []string{"flag2", "flag3"},
				},
			},
		)
	})

	t.Run("flag2", func(t *testing.T) {
		testExecute(t, `require ["fileinto", "imap4flags"];
		addflag ["flag2", "flag3"];
		removeflag ["flag3", "flag4"];
		fileinto "test";
	`, eml,
			[]interp.AppliedAction{
				interp.ActionFileInto{
					Mailbox: "test",
					Flags:   []string{"flag2"},
				},
			},
		)
	})
}

package sieve

import (
	"bufio"
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

type result struct {
	redirect     []string
	fileinto     []string
	implicitKeep bool
	keep         bool
	flags        []string
}

func testExecute(t *testing.T, in string, eml string, intendedResult result) {
	t.Run("case", func(t *testing.T) {

		msgHdr, err := textproto.NewReader(bufio.NewReader(strings.NewReader(eml))).ReadMIMEHeader()
		if err != nil {
			t.Fatal(err)
		}

		script := bufio.NewReader(strings.NewReader(in))

		loadedScript, err := Load(script, DefaultOptions())
		if err != nil {
			t.Fatal(err)
		}
		data := interp.NewRuntimeData(loadedScript, interp.Callback{
			RedirectAllowed: func(ctx context.Context, d *interp.RuntimeData, addr string) (bool, error) {
				return true, nil
			},
			HeaderGet: func(key string) (string, bool, error) {
				vals, ok := msgHdr[key]
				if !ok {
					return "", false, nil
				}
				return vals[0], true, nil
			},
		})
		data.MessageSize = len(eml)
		data.SMTP.From = "from@test.com"
		data.SMTP.To = "to@test.com"

		ctx := context.Background()
		if err := loadedScript.Execute(ctx, data); err != nil {
			t.Fatal(err)
		}

		r := result{
			redirect:     data.RedirectAddr,
			fileinto:     data.Mailboxes,
			keep:         data.Keep,
			implicitKeep: data.ImplicitKeep,
			flags:        data.Flags,
		}

		if !reflect.DeepEqual(r, intendedResult) {
			t.Log("Wrong Execute output")
			t.Log("Actual:  ", r)
			t.Log("Expected:", intendedResult)
			t.Fail()
		}
	})
}

func TestFileinto(t *testing.T) {
	testExecute(t, `require ["fileinto"];
	fileinto "test";
`, eml,
		result{
			fileinto: []string{"test"},
		})
	testExecute(t, `require ["fileinto"];
		fileinto "test";
		fileinto "test2";
	`, eml,
		result{
			fileinto: []string{"test", "test2"},
		})
}

func TestFlags(t *testing.T) {
	testExecute(t, `require ["fileinto", "imap4flags"];
	setflag ["flag1", "flag2"];
	addflag ["flag2", "flag3"];
	removeflag ["flag1"];
	fileinto "test";
`, eml,
		result{
			fileinto: []string{"test"},
			flags:    []string{"flag2", "flag3"},
		})
	testExecute(t, `require ["fileinto", "imap4flags"];
		addflag ["flag2", "flag3"];
		removeflag ["flag3", "flag4"];
		fileinto "test";
	`, eml,
		result{
			fileinto: []string{"test"},
			flags:    []string{"flag2"},
		})
}

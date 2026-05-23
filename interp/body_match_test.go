package interp

import (
	"context"
	"io"
	"strings"
	"testing"
)

type testBodyMessage struct {
	parts []BodyPart
	raw   io.Reader
}

func (m testBodyMessage) HeaderGet(string) ([]string, error) {
	return nil, nil
}

func (m testBodyMessage) MessageSize() int {
	return 0
}

func (m testBodyMessage) BodyRaw(context.Context) (io.Reader, error) {
	return m.raw, nil
}

func (m testBodyMessage) BodyParts(context.Context, []string) ([]BodyPart, error) {
	return m.parts, nil
}

type countedReadCloser struct {
	io.Reader
	onClose func()
}

func (c countedReadCloser) Close() error {
	c.onClose()
	return nil
}

type countedBodyPart struct {
	contentType string
	content     string
	closeCount  *int
}

func (p countedBodyPart) ContentType() string {
	return p.contentType
}

func (p countedBodyPart) Open(context.Context) (io.ReadCloser, error) {
	return countedReadCloser{
		Reader: strings.NewReader(p.content),
		onClose: func() {
			(*p.closeCount)++
		},
	}, nil
}

func TestBodyPartRawOpenNilReader(t *testing.T) {
	part := BodyPartRaw{BodyMessage: testBodyMessage{raw: nil}}

	_, err := part.Open(context.Background())
	if err == nil {
		t.Fatal("expected error for nil raw body reader")
	}
	if err != ErrNoBody {
		t.Fatalf("expected ErrNoBody, got: %v", err)
	}
}

func TestBodyTextStripHTMLDoesNotLeakAcrossParts(t *testing.T) {
	d := &RuntimeData{
		Msg: testBodyMessage{parts: []BodyPart{
			BodyPartBytes{ContentTypeValue: "text/html", Blob: []byte("<p>ignored</p>")},
			BodyPartBytes{ContentTypeValue: "text/plain", Blob: []byte("literal <b> marker")},
		}},
		Script: &Script{},
	}

	test := BodyTest{
		matcherTest: matcherTest{
			Comparator: ComparatorOctet,
			Match:      MatchContains,
			Key:        []string{"<b>"},
		},
		Transform: BodyTransformText,
	}

	ok, err := test.Check(context.Background(), d)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected match in plain text part after HTML part")
	}
}

func TestTryMatchBodyPartClosesReadersOnNonMatch(t *testing.T) {
	closeCount := 0
	part := countedBodyPart{
		contentType: "text/plain",
		content:     "hello",
		closeCount:  &closeCount,
	}

	matcher := matcherTest{
		Comparator: ComparatorOctet,
		Match:      MatchIs,
		Key:        []string{"nomatch", "hello"},
	}

	ok, err := matcher.tryMatchBodyPart(context.Background(), &RuntimeData{Script: &Script{}}, part, false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected second key to match")
	}
	if closeCount != 2 {
		t.Fatalf("expected 2 closes, got %d", closeCount)
	}
}

func TestTryMatchBodyPartASCIINumericIs(t *testing.T) {
	matcher := matcherTest{
		Comparator: ComparatorASCIINumeric,
		Match:      MatchIs,
		Key:        []string{"2"},
	}

	part := BodyPartBytes{
		ContentTypeValue: "text/plain",
		Blob:             []byte("002 trailing text"),
	}

	ok, err := matcher.tryMatchBodyPart(context.Background(), &RuntimeData{Script: &Script{}}, part, false)
	if err != nil {
		t.Fatal(err)
	}
	if !ok {
		t.Fatal("expected numeric :is to match equivalent numeric prefix")
	}
}

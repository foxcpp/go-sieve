package interp

import (
	"bytes"
	"io"
	"reflect"
	"testing"
)

func TestHtmlStripper_ReadByte(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{
			name:  "simple html",
			input: "<p>Hello, <b>world</b>!</p>",
			want:  "Hello, world!",
		},
		{
			name:  "html with attributes",
			input: `<a href="https://example.com">Link</a>`,
			want:  "Link",
		},
		{
			name:  "html with nested tags",
			input: "<div><p>Nested <span>tags</span></p></div>",
			want:  "Nested tags",
		},
		{
			name:  "malformed html",
			input: "<p>Hello, <b>worldb>!</p>",
			want:  "Hello, worldb>!",
		},
		{
			name:  "malformed html 2",
			input: "<p>Hello, <b>world!<<<</p>",
			want:  "Hello, world!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			stripper := &htmlStripper{BR: bytes.NewReader([]byte(tt.input))}
			actual, err := io.ReadAll(stripper)
			if (err != nil) != tt.wantErr {
				t.Errorf("Read() error = %v, wantErr %v", err, tt.wantErr)
			}
			if !reflect.DeepEqual(actual, []byte(tt.want)) {
				t.Errorf("Read() actual = %v, want %v", string(actual), tt.want)
			}
		})
	}
}

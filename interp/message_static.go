package interp

import (
	"context"
	"net/textproto"
)

type DummyPolicy struct {
}

func (d DummyPolicy) RedirectAllowed(_ context.Context, _ *RuntimeData, _ string) (bool, error) {
	return true, nil
}

type MessageHeader interface {
	Values(key string) []string
	Set(key, value string)
	Del(key string)
}

var (
	_ MessageHeader = textproto.MIMEHeader{}
)

// MessageStatic is a simple Message interface implementation
// that just keeps all data in memory in a Go struct.
type MessageStatic struct {
	SMTPFrom string
	SMTPTo   string
	Size     int
	Header   MessageHeader
}

func (m MessageStatic) EnvelopeFrom() string {
	return m.SMTPFrom
}

func (m MessageStatic) EnvelopeTo() string {
	return m.SMTPTo
}

func (m MessageStatic) HeaderGet(key string) (string, bool, error) {
	values := m.Header.Values(key)
	if len(values) == 0 {
		return "", false, nil
	}
	return values[0], true, nil
}

func (m MessageStatic) MessageSize() int {
	return m.Size
}

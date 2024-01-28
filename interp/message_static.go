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

type EnvelopeStatic struct {
	From string
	To   string
	Auth string
}

func (m EnvelopeStatic) EnvelopeFrom() string {
	return m.From
}

func (m EnvelopeStatic) EnvelopeTo() string {
	return m.To
}

func (m EnvelopeStatic) AuthUsername() string {
	return m.Auth
}

// MessageStatic is a simple Message interface implementation
// that just keeps all data in memory in a Go struct.
type MessageStatic struct {
	Size   int
	Header MessageHeader
}

func (m MessageStatic) HeaderGet(key string) ([]string, error) {
	return m.Header.Values(key), nil
}

func (m MessageStatic) MessageSize() int {
	return m.Size
}

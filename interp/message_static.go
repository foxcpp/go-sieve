package interp

import (
	"bytes"
	"context"
	"io"
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
//
// It also implements BodyMessage: set RawMessage to the full RFC 2822 message
// bytes (including headers). If RawMessage is nil, body tests will return false.
type MessageStatic struct {
	Size   int
	Header MessageHeader
	// RawMessage is the complete raw message (headers + body) used by body tests.
	// If nil, body tests always return false.
	RawMessage []byte
}

func (m MessageStatic) HeaderGet(key string) ([]string, error) {
	return m.Header.Values(key), nil
}

func (m MessageStatic) MessageSize() int {
	return m.Size
}

func (m MessageStatic) BodyRaw(ctx context.Context) (io.Reader, error) {
	if m.RawMessage == nil {
		return nil, nil
	}
	return ParseBodyRaw(ctx, bytes.NewReader(m.RawMessage))
}

func (m MessageStatic) BodyParts(ctx context.Context, contentTypes []string) ([]BodyPart, error) {
	if m.RawMessage == nil {
		return nil, nil
	}
	return ParseBodyParts(ctx, bytes.NewReader(m.RawMessage), contentTypes, len(m.RawMessage))
}

// MapEnv is a simple Env implementation backed by a map.
type MapEnv map[string]string

func (e MapEnv) GetEnvironment(name string) (string, bool) {
	v, ok := e[name]
	return v, ok
}

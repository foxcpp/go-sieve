package interp

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
)

type Test interface {
	Check(ctx context.Context, d *RuntimeData) (bool, error)
}

type AddressTest struct {
	Comparator  Comparator
	AddressPart AddressPart
	Match       Match

	Header []string
	Key    []string
}

var allowedAddrHeaders = map[string]struct{}{
	"from":        {},
	"to":          {},
	"cc":          {},
	"bcc":         {},
	"sender":      {},
	"resent-from": {},
	"resent-to":   {},
}

func (a AddressTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	for _, hdr := range a.Header {
		if _, ok := allowedAddrHeaders[hdr]; !ok {
			continue
		}

		value, ok, err := d.Callback.HeaderGet(hdr)
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		addrList, err := mail.ParseAddressList(value)
		if err != nil {
			return false, nil
		}

		for _, k := range a.Key {
			ok, err := testAddress(a.AddressPart, a.Comparator, a.Match, addrList, k)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}

type AllOfTest struct {
	Tests []Test
}

func (a AllOfTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	for _, t := range a.Tests {
		ok, err := t.Check(ctx, d)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

type AnyOfTest struct {
	Tests []Test
}

func (a AnyOfTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	for _, t := range a.Tests {
		ok, err := t.Check(ctx, d)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type EnvelopeTest struct {
	Comparator  Comparator
	AddressPart AddressPart
	Match       Match

	Field []string
	Key   []string
}

func (e EnvelopeTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	for _, field := range e.Field {
		var value string
		switch strings.ToLower(field) {
		case "from":
			value = d.SMTP.From
		case "to":
			value = d.SMTP.To
		default:
			return false, fmt.Errorf("envelope: unsupported envelope-part: %v", field)
		}

		for _, k := range e.Key {
			ok, err := testAddress(e.AddressPart, e.Comparator, e.Match, []*mail.Address{
				{Address: value},
			}, k)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}

type ExistsTest struct {
	Fields []string
}

func (e ExistsTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	for _, field := range e.Fields {
		_, ok, err := d.Callback.HeaderGet(field)
		if err != nil {
			return false, err
		}
		if !ok {
			return false, nil
		}
	}
	return true, nil
}

type FalseTest struct{}

func (f FalseTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	return false, nil
}

type TrueTest struct{}

func (t TrueTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	return true, nil
}

type HeaderTest struct {
	Comparator Comparator
	Match      Match

	Header []string
	Key    []string
}

func (h HeaderTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	for _, hdr := range h.Header {
		value, ok, err := d.Callback.HeaderGet(hdr)
		if err != nil {
			return false, err
		}
		if !ok {
			continue
		}

		for _, k := range h.Key {
			ok, err := testString(h.Comparator, h.Match, value, k)
			if err != nil {
				return false, err
			}
			if ok {
				return true, nil
			}
		}
	}
	return false, nil
}

type NotTest struct {
	Test Test
}

func (n NotTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	ok, err := n.Test.Check(ctx, d)
	if err != nil {
		return false, err
	}
	return !ok, nil
}

type SizeTest struct {
	Size  int
	Over  bool
	Under bool
}

func (s SizeTest) Check(ctx context.Context, d *RuntimeData) (bool, error) {
	if s.Over && d.MessageSize > s.Size {
		return true, nil
	}
	if s.Under && d.MessageSize < s.Size {
		return true, nil
	}
	return false, nil
}

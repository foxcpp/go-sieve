package interp

import (
	"context"
	"fmt"
	"strings"

	"github.com/emersion/go-message/mail"
)

type Test interface {
	Check(ctx context.Context, d *RuntimeData) (bool, error)
}

type AddressTest struct {
	matcherTest

	AddressPart AddressPart
	Header      []string
}

var allowedAddrHeaders = map[string]struct{}{
	// Required by Sieve.
	"from":        {},
	"to":          {},
	"cc":          {},
	"bcc":         {},
	"sender":      {},
	"resent-from": {},
	"resent-to":   {},
	// Misc (RFC 2822)
	"reply-to":        {},
	"resent-reply-to": {},
	"resent-sender":   {},
	"resent-cc":       {},
	"resent-bcc":      {},
	// Non-standard (RFC 2076, draft-palme-mailext-headers-08.txt)
	"for-approval":                       {},
	"for-handling":                       {},
	"for-comment":                        {},
	"apparently-to":                      {},
	"errors-to":                          {},
	"delivered-to":                       {},
	"return-receipt-to":                  {},
	"x-admin":                            {},
	"read-receipt-to":                    {},
	"x-confirm-reading-to":               {},
	"return-receipt-requested":           {},
	"registered-mail-reply-requested-by": {},
	"mail-followup-to":                   {},
	"mail-reply-to":                      {},
	"abuse-reports-to":                   {},
	"x-complaints-to":                    {},
	"x-report-abuse-to":                  {},
	"x-beenthere":                        {},
	"x-original-from":                    {},
	"x-original-to":                      {},
}

func (a AddressTest) Check(_ context.Context, d *RuntimeData) (bool, error) {
	for _, hdr := range a.Header {
		hdr = strings.ToLower(hdr)
		hdr = expandVars(d, hdr)

		if _, ok := allowedAddrHeaders[hdr]; !ok {
			continue
		}

		values, err := d.Msg.HeaderGet(hdr)
		if err != nil {
			return false, err
		}

		for _, value := range values {
			addrList, err := mail.ParseAddressList(value)
			if err != nil {
				return false, nil
			}

			ok, err := testAddress(d, a.matcherTest, a.AddressPart, addrList)
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
	matcherTest

	AddressPart AddressPart
	Field       []string
}

func (e EnvelopeTest) Check(_ context.Context, d *RuntimeData) (bool, error) {
	for _, field := range e.Field {
		var value string
		switch strings.ToLower(expandVars(d, field)) {
		case "from":
			value = d.Envelope.EnvelopeFrom()
		case "to":
			value = d.Envelope.EnvelopeTo()
		case "auth":
			value = d.Envelope.AuthUsername()
		default:
			return false, fmt.Errorf("envelope: unsupported envelope-part: %v", field)
		}

		ok, err := testAddress(d, e.matcherTest, e.AddressPart, []*mail.Address{
			{Address: value},
		})
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

type ExistsTest struct {
	Fields []string
}

func (e ExistsTest) Check(_ context.Context, d *RuntimeData) (bool, error) {
	for _, field := range e.Fields {
		values, err := d.Msg.HeaderGet(expandVars(d, field))
		if err != nil {
			return false, err
		}
		if len(values) == 0 {
			return false, nil
		}
	}
	return true, nil
}

type FalseTest struct{}

func (f FalseTest) Check(context.Context, *RuntimeData) (bool, error) {
	return false, nil
}

type TrueTest struct{}

func (t TrueTest) Check(context.Context, *RuntimeData) (bool, error) {
	return true, nil
}

type HeaderTest struct {
	matcherTest

	Header []string
}

func (h HeaderTest) Check(_ context.Context, d *RuntimeData) (bool, error) {
	for _, hdr := range h.Header {
		values, err := d.Msg.HeaderGet(expandVars(d, hdr))
		if err != nil {
			return false, err
		}

		for _, value := range values {
			ok, err := h.matcherTest.tryMatch(d, value)
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

func (s SizeTest) Check(_ context.Context, d *RuntimeData) (bool, error) {
	if s.Over && d.Msg.MessageSize() > s.Size {
		return true, nil
	}
	if s.Under && d.Msg.MessageSize() < s.Size {
		return true, nil
	}
	return false, nil
}

package interp

import (
	"errors"
	"fmt"
	"net/mail"
	"strconv"
	"strings"
	"unicode"
)

type Match string

const (
	MatchContains Match = "contains"
	MatchIs       Match = "is"
	MatchMatches  Match = "matches"
)

type Comparator string

const (
	ComparatorOctet          Comparator = "i;octet"
	ComparatorASCIICaseMap   Comparator = "i;ascii-casemap"
	ComparatorASCIINumeric   Comparator = "i;ascii-numeric"
	ComparatorUnicodeCaseMap Comparator = "i;unicode-casemap"
)

type AddressPart string

const (
	LocalPart AddressPart = "localpart"
	Domain    AddressPart = "domain"
	All       AddressPart = "all"
)

func split(addr string) (mailbox, domain string, err error) {
	if strings.EqualFold(addr, "postmaster") {
		return addr, "", nil
	}

	indx := strings.LastIndexByte(addr, '@')
	if indx == -1 {
		return "", "", errors.New("address: missing at-sign")
	}
	mailbox = addr[:indx]
	domain = addr[indx+1:]
	if mailbox == "" {
		return "", "", errors.New("address: empty local-part")
	}
	if domain == "" {
		return "", "", errors.New("address: empty domain")
	}
	return
}

var ErrComparatorMatchUnsupported = fmt.Errorf("match-comparator combination not supported")

func numericValue(s string) *uint64 {
	if len(s) == 0 {
		return nil
	}
	runes := []rune(s)
	if !unicode.IsDigit(runes[0]) {
		return nil
	}
	var sl string
	for i, r := range runes {
		if !unicode.IsDigit(r) {
			sl = string(runes[:i])
		}
	}
	digit, _ := strconv.ParseUint(sl, 10, 64)
	return &digit
}

func testString(comparator Comparator, match Match, value, key string) (bool, error) {
	switch comparator {
	case ComparatorOctet:
		switch match {
		case MatchContains:
			return strings.Contains(value, key), nil
		case MatchIs:
			return value == key, nil
		case MatchMatches:
			return matchOctet(key, value)
		}
	case ComparatorASCIINumeric:
		switch match {
		case MatchContains:
			return false, ErrComparatorMatchUnsupported
		case MatchIs:
			lhsNum := numericValue(value)
			rhsNum := numericValue(key)
			if lhsNum == nil || rhsNum == nil {
				return false, nil
			}
			return *lhsNum == *rhsNum, nil
		case MatchMatches:
			return false, ErrComparatorMatchUnsupported
		}
	case ComparatorASCIICaseMap:
		value = strings.ToLower(value)
		key = strings.ToLower(key)
		switch match {
		case MatchContains:
			return strings.Contains(value, key), nil
		case MatchIs:
			return value == key, nil
		case MatchMatches:
			return matchUnicode(value, key)
		}
	case ComparatorUnicodeCaseMap:
		value = strings.ToLower(value)
		key = strings.ToLower(key)
		switch match {
		case MatchContains:
			return strings.Contains(value, key), nil
		case MatchIs:
			return value == key, nil
		case MatchMatches:
			return matchUnicode(value, key)
		}
	}
	return false, nil
}

func testAddress(part AddressPart, comparator Comparator, match Match, headerVal []*mail.Address, addrValue string) (bool, error) {
	for _, addr := range headerVal {
		var valueToCompare string
		if addr.Address != "" {
			switch part {
			case LocalPart:
				localPart, _, err := split(addr.Address)
				if err != nil {
					continue
				}
				valueToCompare = localPart
			case Domain:
				_, domain, err := split(addr.Address)
				if err != nil {
					continue
				}
				valueToCompare = domain
			case All:
				valueToCompare = addr.Address
			}
		}

		ok, err := testString(comparator, match, valueToCompare, addrValue)
		if err != nil {
			return false, err
		}
		if ok {
			return true, nil
		}
	}
	return false, nil
}

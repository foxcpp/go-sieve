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

	DefaultComparator = ComparatorASCIICaseMap
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

func testString(comparator Comparator, match Match, value, key string) (bool, []string, error) {
	switch comparator {
	case ComparatorOctet:
		switch match {
		case MatchContains:
			return strings.Contains(value, key), nil, nil
		case MatchIs:
			return value == key, nil, nil
		case MatchMatches:
			return matchOctet(key, value, false)
		}
	case ComparatorASCIINumeric:
		switch match {
		case MatchContains:
			return false, nil, ErrComparatorMatchUnsupported
		case MatchIs:
			lhsNum := numericValue(value)
			rhsNum := numericValue(key)
			if lhsNum == nil || rhsNum == nil {
				return false, nil, nil
			}
			return *lhsNum == *rhsNum, nil, nil
		case MatchMatches:
			return false, nil, ErrComparatorMatchUnsupported
		}
	case ComparatorASCIICaseMap:
		switch match {
		case MatchContains:
			value = strings.ToLower(value)
			key = strings.ToLower(key)
			return strings.Contains(value, key), nil, nil
		case MatchIs:
			value = strings.ToLower(value)
			key = strings.ToLower(key)
			return value == key, nil, nil
		case MatchMatches:
			return matchOctet(key, value, true)
		}
	case ComparatorUnicodeCaseMap:
		switch match {
		case MatchContains:
			value = strings.ToLower(value)
			key = strings.ToLower(key)
			return strings.Contains(value, key), nil, nil
		case MatchIs:
			return strings.EqualFold(value, key), nil, nil
		case MatchMatches:
			return matchUnicode(key, value, true)
		}
	}
	return false, nil, nil
}

func testAddress(part AddressPart, comparator Comparator, match Match, headerVal []*mail.Address, addrValue string) (bool, []string, error) {
	for _, addr := range headerVal {
		if addr.Address == "<>" {
			addr.Address = ""
		}

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

		ok, matches, err := testString(comparator, match, valueToCompare, addrValue)
		if err != nil {
			return false, nil, err
		}
		if ok {
			return true, matches, nil
		}
	}
	return false, nil, nil
}

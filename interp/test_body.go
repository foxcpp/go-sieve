package interp

import (
	"bufio"
	"io"
	"math"
	"regexp"
	"unicode"
)

type asciiLowerByteReader struct {
	b io.ByteReader
	r io.Reader
}

func (r asciiLowerByteReader) ReadByte() (byte, error) {
	c, err := r.b.ReadByte()
	if err != nil {
		return 0, err
	}

	if 'A' <= c && c <= 'Z' {
		c += 'a' - 'A'
	}

	return c, nil
}

func (r asciiLowerByteReader) Read(p []byte) (int, error) {
	n, err := r.r.Read(p)
	if n != 0 {
		for i, c := range p[:n] {
			if 'A' <= c && c <= 'Z' {
				p[i] += 'a' - 'A'
			}
		}
	}
	return n, err
}

func testReader(comparator Comparator, match Match, valueReader io.Reader, key string) (bool, error) {
	if comparator == ComparatorASCIINumeric {
		switch match {
		case MatchContains, MatchMatches:
			return false, ErrComparatorMatchUnsupported
		case MatchIs:
			lhsNum, err := numericValueReader(valueReader)
			if err != nil {
				return false, err
			}
			rhsNum := numericValue(key)
			return RelEqual.CompareNumericValue(lhsNum, rhsNum), nil
		case MatchValue:
			panic("testReader does not support relational matching")
		case MatchCount:
			panic("testReader should not be used with MatchCount")
		}
	}

	var (
		regex    string
		octet    = comparator.IsOctet()
		caseFold = comparator.IsCaseMap()
	)
	switch match {
	case MatchContains:
		regex = regexp.QuoteMeta(key)
	case MatchIs:
		regex = "^" + regexp.QuoteMeta(key) + "$"
	case MatchMatches:
		regex = patternToRegex(key, caseFold)
	case MatchValue:
		panic("testReader does not support relational matching")
	case MatchCount:
		panic("testReader should not be used with MatchCount")
	}

	if caseFold {
		if octet {
			br, ok := valueReader.(io.ByteReader)
			if !ok {
				br = bufio.NewReader(valueReader)
			}
			valueReader = asciiLowerByteReader{b: br}

			regex = toLowerASCII(regex)
		} else {
			regex = "(?i)" + regex
		}
	}

	matcher, err := compileMatcherRegex(regex, octet)
	if err != nil {
		return false, err
	}

	return matcher.MatchReader(valueReader)
}

func numericValueReader(r io.Reader) (*uint64, error) {
	br := bufio.NewReader(r)

	first, _, err := br.ReadRune()
	if err != nil {
		if err == io.EOF {
			return nil, nil
		}
		return nil, err
	}

	if !unicode.IsDigit(first) {
		return nil, nil
	}
	if first < '0' || first > '9' {
		return nil, nil
	}

	value := uint64(first - '0')
	overflow := false

	for {
		r, _, err := br.ReadRune()
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}

		if !unicode.IsDigit(r) {
			break
		}

		if r < '0' || r > '9' {
			overflow = true
			continue
		}

		digit := uint64(r - '0')
		if !overflow {
			if value > (math.MaxUint64-digit)/10 {
				overflow = true
				continue
			}
			value = value*10 + digit
		}
	}

	if overflow {
		return nil, nil
	}

	return &value, nil
}

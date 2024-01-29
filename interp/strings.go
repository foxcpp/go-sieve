package interp

import (
	"encoding/hex"
	"fmt"
	"regexp"
	"strconv"
	"strings"
)

/*
	blank                = WSP / CRLF
	encoded-arb-octets   = "${hex:" hex-pair-seq "}"
	hex-pair-seq         = *blank hex-pair *(1*blank hex-pair) *blank
	hex-pair             = 1*2HEXDIG

This whitespace handling is giving me headaches. [ \t\r\n]* is "blank" above.
*/
var encodedHexRegex = regexp.MustCompile(
	`(?i)\${(` +
		`hex:[ \t\r\n]*([0-9a-f]{1,2}(?:[ \t\r\n]+[0-9a-f]{1,2})*)[ \t\r\n]*|` +
		`unicode:[ \t\r\n]*([0-9a-f]+(?:[ \t\r\n]+[0-9a-f]+)*)[ \t\r\n]*)}`)

var removeWSP = strings.NewReplacer(
	"\t", "", "\r", "", "\n", "", " ", "",
)

var normalizeWSP = strings.NewReplacer(
	"\t", "", "\r", "", "\n", "",
)

func decodeEncodedChars(s string) (string, error) {
	var lastErr error
	decoded := encodedHexRegex.ReplaceAllStringFunc(s, func(match string) string {
		if strings.HasPrefix(strings.ToLower(match), "${hex:") {
			hexString := removeWSP.Replace(match[6 : len(match)-1])
			decoded, err := hex.DecodeString(hexString)
			if err != nil {
				lastErr = err
				return ""
			}
			return string(decoded)
		}

		cpString := strings.Split(normalizeWSP.Replace(match[10:len(match)-1]), " ")
		replacement := strings.Builder{}
		replacement.Grow(len(cpString))
		for _, part := range cpString {
			if part != "" {
				value, err := strconv.ParseInt(part, 16, 32)
				if err != nil {
					lastErr = err
					return ""
				}
				/*
						RFC 5228 Section 2.4.2.4:
					   It is an error for a script to use a hexadecimal value that isn't in
					   either the range 0 to D7FF or the range E000 to 10FFFF.  (The range
					   D800 to DFFF is excluded as those character numbers are only used as
					   part of the UTF-16 encoding form and are not applicable to the UTF-8
					   encoding that the syntax here represents.)
				*/
				if !(value >= 0 && value <= 0xD7FF) && !(value >= 0xE000 && value <= 0x10FFFF) {
					lastErr = fmt.Errorf("encoded unicode keypoint is out of range")
					return ""
				}
				replacement.WriteRune(rune(value))
			}
		}
		return replacement.String()
	})
	if lastErr != nil {
		return "", lastErr
	}
	return decoded, nil
}

package hfid

import (
	"fmt"
	"strings"
)

// Encoding The characters to use to encode HFIDs
type Encoding string

// NumericEncoding an Encoding that contains only numbers
const NumericEncoding = "0123456789"

// DefaultEncoding Contains numbers and capital English letters only.
const DefaultEncoding = "0123456789ABCDEFGHIJKLMNOPQRSTUVWXYZ"

// Valid checks whether the Encoding is a valid one or not. An Encoding is considered valid if it has at least 3 characters
// and no character is repeated more than once.
func (e Encoding) Valid() error {
	if len(e) < 3 {
		return fmt.Errorf("encoding should contain at least 3 characters")
	}
	seenRunes := map[rune]bool{}
	for _, r := range e {
		if seenRunes[r] {
			return fmt.Errorf("encoding must have unique characters but '%s' was duplicated in '%s'", string(r), e)
		}
		seenRunes[r] = true
	}
	return nil
}

// Encode a number into a string using this Encoding
func (e Encoding) Encode(number int64) (string, error) {
	if err := e.Valid(); err != nil {
		return "", fmt.Errorf("cannot encode '%d' using an invalid Encoding '%s': %s", number, e, err)
	}

	if number < 0 {
		return "", fmt.Errorf("cannot encode negative number %d", number)
	}

	// Figure out how many encoded bits are needed.
	nBits := 1
	for maxN, _ := Pow(len(e), nBits); maxN <= number; nBits++ {
		maxN, _ = Pow(len(e), nBits+1)
	}

	result := ""
	for i := nBits - 1; i >= 0; i-- {
		m, _ := Pow(len(e), i)
		digitIndex := number / m
		result += string(e[digitIndex])
		number -= m * digitIndex
	}
	return result, nil
}

// Decode a number to a string using this Encoding
func (e Encoding) Decode(s string) (int64, error) {
	if err := e.Valid(); err != nil {
		return int64(0), fmt.Errorf("cannot decode '%s' using an invalid Encoding '%s': %s", s, e, err)
	}
	result := int64(0)
	for i := 0; i < len(s); i++ {
		m := int64(strings.IndexRune(string(e), rune(s[i])))
		if m < 0 {
			return int64(0), fmt.Errorf("invalid character '%s' encountered while decoding '%s' using '%s'", string(s[i]), s, e)
		}
		m2, err := Pow(len(e), len(s)-i-1)
		increment := m2 * int64(m)
		result += increment
		// Overflow detection
		if err != nil || increment/m2 != m || result < 0 {
			return 0, fmt.Errorf("overflow error occured while decoding %s using encoding '%s'", s, e)
		}
	}
	return result, nil
}

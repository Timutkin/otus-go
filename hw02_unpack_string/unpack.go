package hw02unpackstring

import (
	"errors"
	"strconv"
	"strings"
	"unicode"
)

var ErrInvalidString = errors.New("invalid string")

func Unpack(str string) (string, error) {
	if len(str) == 0 {
		return "", nil
	}

	runes := []rune(str)
	if len(runes) == 1 {
		if unicode.IsDigit(runes[0]) || string(runes[0]) == `\` {
			return "", ErrInvalidString
		}
		return string(runes[0]), nil
	}

	var res strings.Builder
	i := 0
	for i < len(runes) {
		if unicode.IsDigit(runes[i]) {
			return "", ErrInvalidString
		}

		if i == len(runes)-1 {
			if string(runes[i]) == `\` {
				return "", ErrInvalidString
			}
			res.WriteRune(runes[i])
			break
		}

		current := string(runes[i])
		if current == `\` {
			offset, s, err := processNextSymbolsAfterEscape(runes, i)
			if err != nil {
				return s, err
			}
			res.WriteString(s)
			i += offset
			continue
		}

		if unicode.IsDigit(runes[i+1]) {
			res.WriteString(getRepeatedString(current, runes[i+1]))
			i += 2
		} else {
			res.WriteRune(runes[i])
			i++
		}
	}
	return res.String(), nil
}

func processNextSymbolsAfterEscape(runes []rune, i int) (offset int, str string, err error) {
	next := string(runes[i+1])
	if next == `\` || unicode.IsDigit(runes[i+1]) {
		if i+2 < len(runes) && unicode.IsDigit(runes[i+2]) {
			return 3, getRepeatedString(next, runes[i+2]), nil
		}
		return 2, next, nil
	}
	return 0, "", ErrInvalidString
}

func getRepeatedString(str string, count rune) string {
	digit, _ := strconv.Atoi(string(count))
	return strings.Repeat(str, digit)
}

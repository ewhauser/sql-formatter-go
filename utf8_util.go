package sqlformatter

import "unicode/utf8"

func utf8DecodeRuneInString(s string) (rune, int) {
	return utf8.DecodeRuneInString(s)
}

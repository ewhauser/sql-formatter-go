package sqlformatter

import (
	"strings"
	"unicode"
)

type IdentifierMatcher struct {
	chars IdentChars
}

func NewIdentifierMatcher(chars *IdentChars) *IdentifierMatcher {
	return &IdentifierMatcher{chars: derefIdentChars(chars)}
}

func (m *IdentifierMatcher) Match(input string, index int) (string, bool) {
	if index >= len(input) {
		return "", false
	}
	r, size := utf8DecodeRuneInString(input[index:])
	if !m.isFirstChar(r) {
		return "", false
	}
	pos := index + size
	lastWasDash := false
	for pos < len(input) {
		r2, size2 := utf8DecodeRuneInString(input[pos:])
		if r2 == '-' && m.chars.Dashes {
			lastWasDash = true
			pos += size2
			continue
		}
		if !m.isRestChar(r2) {
			break
		}
		lastWasDash = false
		pos += size2
	}
	if lastWasDash {
		// identifiers cannot end with dash
		return "", false
	}
	return input[index:pos], true
}

func (m *IdentifierMatcher) isFirstChar(r rune) bool {
	if m.chars.AllowFirstCharNumber && unicode.IsDigit(r) {
		return true
	}
	if isLetterOrUnderscore(r) || unicode.IsMark(r) {
		return true
	}
	if strings.ContainsRune(m.chars.First, r) {
		return true
	}
	return false
}

func (m *IdentifierMatcher) isRestChar(r rune) bool {
	if isLetterOrUnderscore(r) || unicode.IsDigit(r) || unicode.IsMark(r) {
		return true
	}
	if strings.ContainsRune(m.chars.Rest, r) {
		return true
	}
	return false
}

func isLetterOrUnderscore(r rune) bool {
	return unicode.IsLetter(r) || r == '_'
}

package sqlformatter

import (
	"regexp"
	"strings"
	"unicode"
)

type ReservedWordMatcher struct {
	re         *regexp.Regexp
	identChars IdentChars
}

func NewReservedWordMatcher(words []string, identChars *IdentChars) *ReservedWordMatcher {
	if len(words) == 0 {
		return &ReservedWordMatcher{re: regexp.MustCompile(`^\b$`)}
	}
	items := make([]string, len(words))
	copy(items, words)
	SortByLengthDesc(items)
	patterns := make([]string, 0, len(items))
	for _, word := range items {
		escaped := EscapeRegExp(word)
		escaped = strings.ReplaceAll(escaped, " ", "\\s+")
		patterns = append(patterns, escaped)
	}
	pattern := "(?i)^(?:" + strings.Join(patterns, "|") + ")"
	return &ReservedWordMatcher{re: regexp.MustCompile(pattern), identChars: derefIdentChars(identChars)}
}

func (m *ReservedWordMatcher) Match(input string, index int) (string, bool) {
	if index > len(input) {
		return "", false
	}
	loc := m.re.FindStringIndex(input[index:])
	if loc == nil || loc[0] != 0 {
		return "", false
	}
	match := input[index : index+loc[1]]
	nextIndex := index + loc[1]
	if nextIndex < len(input) {
		nextRune, _ := utf8DecodeRuneInString(input[nextIndex:])
		if isIdentifierContinuation(nextRune, m.identChars) {
			return "", false
		}
	}
	// also ensure a word boundary in terms of unicode letters/digits/underscore
	if nextIndex < len(input) {
		nextRune, _ := utf8DecodeRuneInString(input[nextIndex:])
		if unicode.IsLetter(nextRune) || unicode.IsDigit(nextRune) || nextRune == '_' {
			return "", false
		}
	}
	return match, true
}

func derefIdentChars(chars *IdentChars) IdentChars {
	if chars == nil {
		return IdentChars{}
	}
	return *chars
}

func isIdentifierContinuation(r rune, chars IdentChars) bool {
	if r == '-' && chars.Dashes {
		return true
	}
	if strings.ContainsRune(chars.Rest, r) {
		return true
	}
	return false
}

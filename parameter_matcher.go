package sqlformatter

import (
	"strings"
	"unicode"
)

type ParameterMatcher struct {
	prefixes []string
	ident    *IdentifierMatcher
}

func NewParameterMatcher(prefixes []string, chars *IdentChars, identTypes []QuoteType) *ParameterMatcher {
	return &ParameterMatcher{prefixes: prefixes, ident: NewIdentifierMatcher(chars)}
}

func (m *ParameterMatcher) Match(input string, index int) (string, bool) {
	for _, prefix := range m.prefixes {
		if strings.HasPrefix(input[index:], prefix) {
			start := index + len(prefix)
			if ident, ok := m.ident.Match(input, start); ok {
				return input[index : start+len(ident)], true
			}
		}
	}
	return "", false
}

type QuotedParameterMatcher struct {
	prefixes []string
	quotes   *QuoteMatcher
}

func NewQuotedParameterMatcher(prefixes []string, identTypes []QuoteType) *QuotedParameterMatcher {
	return &QuotedParameterMatcher{prefixes: prefixes, quotes: NewQuoteMatcher(identTypes)}
}

func (m *QuotedParameterMatcher) Match(input string, index int) (string, bool) {
	for _, prefix := range m.prefixes {
		if strings.HasPrefix(input[index:], prefix) {
			start := index + len(prefix)
			if quoted, ok := m.quotes.Match(input, start); ok {
				return input[index : start+len(quoted)], true
			}
		}
	}
	return "", false
}

type NumberedParameterMatcher struct {
	prefixes []string
}

func NewNumberedParameterMatcher(prefixes []string) *NumberedParameterMatcher {
	return &NumberedParameterMatcher{prefixes: prefixes}
}

func (m *NumberedParameterMatcher) Match(input string, index int) (string, bool) {
	for _, prefix := range m.prefixes {
		if strings.HasPrefix(input[index:], prefix) {
			start := index + len(prefix)
			j := start
			for j < len(input) {
				r, size := utf8DecodeRuneInString(input[j:])
				if !unicode.IsDigit(r) {
					break
				}
				j += size
			}
			if j > start {
				return input[index:j], true
			}
		}
	}
	return "", false
}

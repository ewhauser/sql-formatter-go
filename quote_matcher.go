package sqlformatter

import (
	"strings"
	"unicode"
	"unicode/utf8"
)

type QuoteMatcher struct {
	quoteTypes []QuoteType
	forIdentifiers bool
}

func NewQuoteMatcher(types []QuoteType) *QuoteMatcher {
	return &QuoteMatcher{quoteTypes: types}
}

func (m *QuoteMatcher) Match(input string, index int) (string, bool) {
	for _, qt := range m.quoteTypes {
		if matched, ok := matchQuoteType(input, index, qt); ok {
			return matched, true
		}
	}
	return "", false
}

func matchQuoteType(input string, index int, qt QuoteType) (string, bool) {
	switch v := qt.(type) {
	case PlainQuoteType:
		return matchPlainQuote(input, index, string(v))
	case PrefixedQuoteType:
		return matchPrefixedQuote(input, index, v)
	case *PrefixedQuoteType:
		return matchPrefixedQuote(input, index, *v)
	case RegexPattern:
		return matchRegexPattern(input, index, v.Regex)
	case *RegexPattern:
		return matchRegexPattern(input, index, v.Regex)
	case string:
		return matchPlainQuote(input, index, v)
	default:
		return "", false
	}
}

func matchPrefixedQuote(input string, index int, qt PrefixedQuoteType) (string, bool) {
	prefixes := qt.Prefixes
	if len(prefixes) == 0 {
		return matchPlainQuote(input, index, string(qt.Quote))
	}
	// Try with each prefix (case-insensitive) and optional prefix when not required
	for _, prefix := range prefixes {
		if hasPrefixFold(input[index:], prefix) {
			start := index + len(prefix)
			if matched, ok := matchPlainQuote(input, start, string(qt.Quote)); ok {
				return input[index : start+len(matched)], true
			}
		}
	}
	if !qt.RequirePrefix {
		return matchPlainQuote(input, index, string(qt.Quote))
	}
	return "", false
}

func matchRegexPattern(input string, index int, pattern string) (string, bool) {
	re := PatternToRegex(pattern, false)
	loc := re.FindStringIndex(input[index:])
	if loc == nil || loc[0] != 0 {
		return "", false
	}
	return input[index : index+loc[1]], true
}

func matchPlainQuote(input string, index int, quote string) (string, bool) {
	switch quote {
	case "''-qq":
		return matchQuotedString(input, index, '\'', true, false)
	case "''-bs":
		return matchQuotedString(input, index, '\'', false, true)
	case "''-qq-bs":
		return matchQuotedString(input, index, '\'', true, true)
	case "''-raw":
		return matchQuotedString(input, index, '\'', false, false)
	case "\"\"-qq":
		return matchQuotedString(input, index, '"', true, false)
	case "\"\"-bs":
		return matchQuotedString(input, index, '"', false, true)
	case "\"\"-qq-bs":
		return matchQuotedString(input, index, '"', true, true)
	case "\"\"-raw":
		return matchQuotedString(input, index, '"', false, false)
	case "$$":
		return matchDollarQuoted(input, index)
	case "``":
		return matchQuotedString(input, index, '`', true, false)
	case "[]":
		return matchBracketQuoted(input, index)
	default:
		return "", false
	}
}

func matchQuotedString(input string, index int, quoteChar byte, allowRepeatQuote bool, allowBackslash bool) (string, bool) {
	if index >= len(input) || input[index] != quoteChar {
		return "", false
	}
	i := index + 1
	for i < len(input) {
		ch := input[i]
		if allowBackslash && ch == '\\' {
			i += 2
			continue
		}
		if ch == quoteChar {
			if allowRepeatQuote && i+1 < len(input) && input[i+1] == quoteChar {
				i += 2
				continue
			}
			return input[index : i+1], true
		}
		i++
	}
	return "", false
}

func matchBracketQuoted(input string, index int) (string, bool) {
	if index >= len(input) || input[index] != '[' {
		return "", false
	}
	for i := index + 1; i < len(input); i++ {
		if input[i] == ']' {
			if i+1 < len(input) && input[i+1] == ']' {
				i++
				continue
			}
			return input[index : i+1], true
		}
	}
	return "", false
}

func matchDollarQuoted(input string, index int) (string, bool) {
	if index >= len(input) || input[index] != '$' {
		return "", false
	}
	// find tag end
	i := index + 1
	for i < len(input) {
		r, size := utf8.DecodeRuneInString(input[i:])
		if r == '$' {
			tag := input[index : i+1]
			end := strings.Index(input[i+1:], tag)
			if end == -1 {
				return "", false
			}
			endPos := i + 1 + end + len(tag)
			return input[index:endPos], true
		}
		if !(unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_') {
			return "", false
		}
		i += size
	}
	return "", false
}

func hasPrefixFold(s, prefix string) bool {
	if len(s) < len(prefix) {
		return false
	}
	return strings.EqualFold(s[:len(prefix)], prefix)
}

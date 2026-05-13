package sqlformatter

import (
	"errors"
	"fmt"
	"regexp"
	"unicode"
)

type Matcher interface {
	Match(input string, index int) (string, bool)
}

type RegexMatcher struct {
	re *regexpWrapper
}

func NewRegexMatcher(re *regexpWrapper) *RegexMatcher {
	return &RegexMatcher{re: re}
}

func (m *RegexMatcher) Match(input string, index int) (string, bool) {
	return m.re.MatchAt(input, index)
}

type ByteMatcher byte

func (m ByteMatcher) Match(input string, index int) (string, bool) {
	if index < len(input) && input[index] == byte(m) {
		return input[index : index+1], true
	}
	return "", false
}

type TokenRule struct {
	Type  TokenType
	Regex Matcher
	Text  func(raw string) string
	Key   func(raw string) string
}

type TokenizerEngine struct {
	rules       []TokenRule
	dialectName string
	input       string
	index       int
}

func NewTokenizerEngine(rules []TokenRule, dialectName string) *TokenizerEngine {
	return &TokenizerEngine{rules: rules, dialectName: dialectName}
}

func (t *TokenizerEngine) Tokenize(input string) ([]Token, error) {
	t.input = input
	t.index = 0
	tokens := make([]Token, 0, estimatedTokenCapacity(input))

	for t.index < len(t.input) {
		precedingWhitespace := t.getWhitespace()
		if t.index >= len(t.input) {
			break
		}
		token, ok := t.getNextToken()
		if !ok {
			return nil, t.createParseError()
		}
		token.PrecedingWhitespace = precedingWhitespace
		tokens = append(tokens, token)
	}
	return tokens, nil
}

func estimatedTokenCapacity(input string) int {
	if len(input) == 0 {
		return 0
	}
	estimate := len(input) / 4
	if estimate < 8 {
		return 8
	}
	return estimate
}

func (t *TokenizerEngine) createParseError() error {
	text := t.input[t.index:]
	if len(text) > 10 {
		text = text[:10]
	}
	line, col := lineColFromIndex(t.input, t.index)
	return errors.New(fmt.Sprintf("Parse error: Unexpected \"%s\" at line %d column %d.\n%s", text, line, col, t.dialectInfo()))
}

func (t *TokenizerEngine) dialectInfo() string {
	if t.dialectName == "sql" {
		return "This likely happens because you're using the default \"sql\" dialect.\nIf possible, please select a more specific dialect (like sqlite, postgresql, etc)."
	}
	return fmt.Sprintf("SQL dialect used: \"%s\".", t.dialectName)
}

func (t *TokenizerEngine) getWhitespace() string {
	start := t.index
	for t.index < len(t.input) {
		r := rune(t.input[t.index])
		if r == '\r' || r == '\n' || r == '\t' || r == ' ' {
			t.index++
			continue
		}
		if unicode.IsSpace(r) {
			t.index++
			continue
		}
		break
	}
	if t.index > start {
		return t.input[start:t.index]
	}
	return ""
}

func (t *TokenizerEngine) getNextToken() (Token, bool) {
	for _, rule := range t.rules {
		if rule.Regex == nil {
			continue
		}
		matched, ok := rule.Regex.Match(t.input, t.index)
		if ok {
			raw := matched
			text := raw
			if rule.Text != nil {
				text = rule.Text(raw)
			}
			token := Token{Type: rule.Type, Raw: raw, Text: text, Start: t.index}
			if rule.Key != nil {
				token.Key = rule.Key(raw)
			}
			t.index += len(raw)
			return token, true
		}
	}
	return Token{}, false
}

// regexpWrapper keeps Go regex and avoids repeated substring allocations.
type regexpWrapper struct {
	re *regexp.Regexp
}

func newRegexpWrapper(re *regexp.Regexp) *regexpWrapper {
	return &regexpWrapper{re: re}
}

func (r *regexpWrapper) MatchAt(input string, index int) (string, bool) {
	if index > len(input) {
		return "", false
	}
	loc := r.re.FindStringIndex(input[index:])
	if loc == nil || loc[0] != 0 {
		return "", false
	}
	return input[index : index+loc[1]], true
}

package sqlformatter

import "strings"

type ParenMatcher struct {
	open     bool
	patterns []string
}

func NewParenMatcher(open bool, extra []string) *ParenMatcher {
	pairs := []string{"()"}
	pairs = append(pairs, extra...)
	patterns := make([]string, 0, len(pairs))
	for _, pair := range pairs {
		if len(pair) != 2 {
			continue
		}
		if open {
			patterns = append(patterns, string(pair[0]))
		} else {
			patterns = append(patterns, string(pair[1]))
		}
	}
	return &ParenMatcher{open: open, patterns: patterns}
}

func (m *ParenMatcher) Match(input string, index int) (string, bool) {
	if index >= len(input) {
		return "", false
	}
	for _, pat := range m.patterns {
		if strings.HasPrefix(input[index:], pat) {
			return pat, true
		}
	}
	return "", false
}

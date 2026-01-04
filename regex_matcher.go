package sqlformatter

import "regexp"

type RegexPatternMatcher struct {
	re *regexpWrapper
}

func NewRegexPatternMatcher(re *regexp.Regexp) *RegexPatternMatcher {
	return &RegexPatternMatcher{re: newRegexpWrapper(re)}
}

func (m *RegexPatternMatcher) Match(input string, index int) (string, bool) {
	return m.re.MatchAt(input, index)
}

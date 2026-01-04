package sqlformatter

import "strings"

type LineCommentMatcher struct {
	prefixes []string
}

func NewLineCommentMatcher(prefixes []string) *LineCommentMatcher {
	return &LineCommentMatcher{prefixes: prefixes}
}

func (m *LineCommentMatcher) Match(input string, index int) (string, bool) {
	if index >= len(input) {
		return "", false
	}
	for _, prefix := range m.prefixes {
		if strings.HasPrefix(input[index:], prefix) {
			// consume until end of line or input
			end := index + len(prefix)
			for end < len(input) {
				ch := input[end]
				if ch == '\n' || ch == '\r' {
					break
				}
				end++
			}
			return input[index:end], true
		}
	}
	return "", false
}

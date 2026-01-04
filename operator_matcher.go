package sqlformatter

import "strings"

type OperatorMatcher struct {
	operators []string
}

func NewOperatorMatcher(ops []string) *OperatorMatcher {
	opsCopy := make([]string, len(ops))
	copy(opsCopy, ops)
	SortByLengthDesc(opsCopy)
	return &OperatorMatcher{operators: opsCopy}
}

func (m *OperatorMatcher) Match(input string, index int) (string, bool) {
	if index >= len(input) {
		return "", false
	}
	for _, op := range m.operators {
		if strings.HasPrefix(input[index:], op) {
			return op, true
		}
	}
	return "", false
}

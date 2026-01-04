package sqlformatter

// NestedCommentMatcher matches nested /* */ comments.
type NestedCommentMatcher struct{}

func (m NestedCommentMatcher) Match(input string, index int) (string, bool) {
	if index+1 >= len(input) {
		return "", false
	}
	if input[index:index+2] != "/*" {
		return "", false
	}
	depth := 0
	for i := index; i < len(input)-1; i++ {
		if input[i:i+2] == "/*" {
			depth++
			i++
			continue
		}
		if input[i:i+2] == "*/" {
			depth--
			i++
			if depth == 0 {
				return input[index : i+1], true
			}
			continue
		}
	}
	// Unterminated comment should not be treated as a comment
	return "", false
}

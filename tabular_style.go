package sqlformatter

import "strings"

func toTabularFormat(tokenText string, indentStyle IndentStyle) string {
	if indentStyle == IndentStyleStandard {
		return tokenText
	}
	tail := []string{}
	if len(tokenText) >= 10 && strings.Contains(tokenText, " ") {
		parts := strings.Split(tokenText, " ")
		tokenText = parts[0]
		tail = parts[1:]
	}
	if indentStyle == IndentStyleTabularLeft {
		tokenText = tokenText + strings.Repeat(" ", max(9-len(tokenText), 0))
	} else {
		if len(tokenText) < 9 {
			tokenText = strings.Repeat(" ", 9-len(tokenText)) + tokenText
		}
	}
	if len(tail) == 0 {
		return tokenText
	}
	return tokenText + " " + strings.Join(tail, " ")
}

func isTabularToken(tokenType TokenType) bool {
	return IsLogicalOperator(tokenType) || tokenType == TokenReservedClause || tokenType == TokenReservedSelect || tokenType == TokenReservedSetOperation || tokenType == TokenReservedJoin || tokenType == TokenLimit
}

func max(a, b int) int {
	if a > b {
		return a
	}
	return b
}

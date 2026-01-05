package sqlformatter

func DisambiguateTokens(tokens []Token) []Token {
	mapTokensInPlace(tokens, propertyNameKeywordToIdent)
	mapTokensInPlace(tokens, funcNameToIdent)
	mapTokensInPlace(tokens, clauseAfterJoinToIdent)
	mapTokensInPlace(tokens, dataTypeToParameterizedDataType)
	mapTokensInPlace(tokens, identToArrayIdent)
	mapTokensInPlace(tokens, dataTypeToArrayKeyword)
	return tokens
}

func mapTokens(tokens []Token, fn func(Token, int, []Token) Token) []Token {
	out := make([]Token, len(tokens))
	for i, token := range tokens {
		out[i] = fn(token, i, tokens)
	}
	return out
}

func mapTokensInPlace(tokens []Token, fn func(Token, int, []Token) Token) {
	for i := range tokens {
		tokens[i] = fn(tokens[i], i, tokens)
	}
}

func propertyNameKeywordToIdent(token Token, i int, tokens []Token) Token {
	if IsReserved(token.Type) {
		if prev := prevNonCommentToken(tokens, i); prev.Type != "" && prev.Type == TokenPropertyAccessOperator {
			return Token{Type: TokenIdentifier, Raw: token.Raw, Text: token.Raw, Start: token.Start, PrecedingWhitespace: token.PrecedingWhitespace}
		}
		if next := nextNonCommentToken(tokens, i); next.Type != "" && next.Type == TokenPropertyAccessOperator {
			return Token{Type: TokenIdentifier, Raw: token.Raw, Text: token.Raw, Start: token.Start, PrecedingWhitespace: token.PrecedingWhitespace}
		}
	}
	return token
}

func funcNameToIdent(token Token, i int, tokens []Token) Token {
	if token.Type == TokenReservedFunctionName {
		next := nextNonCommentToken(tokens, i)
		if next.Type == "" || !isOpenParen(next) {
			return Token{Type: TokenIdentifier, Raw: token.Raw, Text: token.Raw, Start: token.Start, PrecedingWhitespace: token.PrecedingWhitespace}
		}
	}
	return token
}

// clauseAfterJoinToIdent converts clause tokens to identifiers when they appear
// right after a JOIN keyword. This handles cases like "LEFT JOIN comment ON ..."
// where "comment" is a table name, not the start of "COMMENT ON" clause.
func clauseAfterJoinToIdent(token Token, i int, tokens []Token) Token {
	if token.Type == TokenReservedClause {
		prev := prevNonCommentToken(tokens, i)
		if prev.Type == TokenReservedJoin {
			return Token{Type: TokenIdentifier, Raw: token.Raw, Text: token.Raw, Start: token.Start, PrecedingWhitespace: token.PrecedingWhitespace}
		}
	}
	return token
}

func dataTypeToParameterizedDataType(token Token, i int, tokens []Token) Token {
	if token.Type == TokenReservedDataType || token.Type == TokenReservedDataTypePhrase {
		next := nextNonCommentToken(tokens, i)
		if next.Type != "" && isOpenParen(next) {
			token.Type = TokenReservedParameterizedDataType
		}
	}
	return token
}

func identToArrayIdent(token Token, i int, tokens []Token) Token {
	if token.Type == TokenIdentifier {
		next := nextNonCommentToken(tokens, i)
		if next.Type != "" && isOpenBracket(next) {
			token.Type = TokenArrayIdentifier
		}
	}
	return token
}

func dataTypeToArrayKeyword(token Token, i int, tokens []Token) Token {
	if token.Type == TokenReservedDataType {
		next := nextNonCommentToken(tokens, i)
		if next.Type != "" && isOpenBracket(next) {
			token.Type = TokenArrayKeyword
		}
	}
	return token
}

func prevNonCommentToken(tokens []Token, index int) Token {
	return nextNonCommentToken(tokens, index, -1)
}

func nextNonCommentToken(tokens []Token, index int, dir ...int) Token {
	direction := 1
	if len(dir) > 0 {
		direction = dir[0]
	}
	for i := 1; ; i++ {
		idx := index + i*direction
		if idx < 0 || idx >= len(tokens) {
			return Token{}
		}
		if !isComment(tokens[idx]) {
			return tokens[idx]
		}
	}
}

func isOpenParen(t Token) bool { return t.Type == TokenOpenParen && t.Text == "(" }

func isOpenBracket(t Token) bool { return t.Type == TokenOpenParen && t.Text == "[" }

func isComment(t Token) bool { return t.Type == TokenBlockComment || t.Type == TokenLineComment }

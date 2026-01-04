package sqlformatter

type TokenType string

const (
	TokenQuotedIdentifier        TokenType = "QUOTED_IDENTIFIER"
	TokenIdentifier              TokenType = "IDENTIFIER"
	TokenString                  TokenType = "STRING"
	TokenVariable                TokenType = "VARIABLE"
	TokenReservedDataType         TokenType = "RESERVED_DATA_TYPE"
	TokenReservedParameterizedDataType TokenType = "RESERVED_PARAMETERIZED_DATA_TYPE"
	TokenReservedKeyword          TokenType = "RESERVED_KEYWORD"
	TokenReservedFunctionName     TokenType = "RESERVED_FUNCTION_NAME"
	TokenReservedKeywordPhrase    TokenType = "RESERVED_KEYWORD_PHRASE"
	TokenReservedDataTypePhrase   TokenType = "RESERVED_DATA_TYPE_PHRASE"
	TokenReservedSetOperation     TokenType = "RESERVED_SET_OPERATION"
	TokenReservedClause           TokenType = "RESERVED_CLAUSE"
	TokenReservedSelect           TokenType = "RESERVED_SELECT"
	TokenReservedJoin             TokenType = "RESERVED_JOIN"
	TokenArrayIdentifier          TokenType = "ARRAY_IDENTIFIER"
	TokenArrayKeyword             TokenType = "ARRAY_KEYWORD"
	TokenCase                     TokenType = "CASE"
	TokenEnd                      TokenType = "END"
	TokenWhen                     TokenType = "WHEN"
	TokenElse                     TokenType = "ELSE"
	TokenThen                     TokenType = "THEN"
	TokenLimit                    TokenType = "LIMIT"
	TokenBetween                  TokenType = "BETWEEN"
	TokenAnd                      TokenType = "AND"
	TokenOr                       TokenType = "OR"
	TokenXor                      TokenType = "XOR"
	TokenOperator                 TokenType = "OPERATOR"
	TokenComma                    TokenType = "COMMA"
	TokenAsterisk                 TokenType = "ASTERISK"
	TokenPropertyAccessOperator   TokenType = "PROPERTY_ACCESS_OPERATOR"
	TokenOpenParen                TokenType = "OPEN_PAREN"
	TokenCloseParen               TokenType = "CLOSE_PAREN"
	TokenLineComment              TokenType = "LINE_COMMENT"
	TokenBlockComment             TokenType = "BLOCK_COMMENT"
	TokenDisableComment           TokenType = "DISABLE_COMMENT"
	TokenNumber                   TokenType = "NUMBER"
	TokenNamedParameter           TokenType = "NAMED_PARAMETER"
	TokenQuotedParameter          TokenType = "QUOTED_PARAMETER"
	TokenNumberedParameter        TokenType = "NUMBERED_PARAMETER"
	TokenPositionalParameter      TokenType = "POSITIONAL_PARAMETER"
	TokenCustomParameter          TokenType = "CUSTOM_PARAMETER"
	TokenDelimiter                TokenType = "DELIMITER"
	TokenEOF                      TokenType = "EOF"
)

type Token struct {
	Type               TokenType
	Raw                string
	Text               string
	Key                string
	Start              int
	PrecedingWhitespace string
}

func CreateEofToken(index int) Token {
	return Token{
		Type: TokenEOF,
		Raw:  "«EOF»",
		Text: "«EOF»",
		Start: index,
	}
}

var EOFToken = CreateEofToken(int(^uint(0) >> 1))

func IsReserved(t TokenType) bool {
	switch t {
	case TokenReservedDataType,
		TokenReservedKeyword,
		TokenReservedFunctionName,
		TokenReservedKeywordPhrase,
		TokenReservedDataTypePhrase,
		TokenReservedClause,
		TokenReservedSelect,
		TokenReservedSetOperation,
		TokenReservedJoin,
		TokenArrayKeyword,
		TokenCase,
		TokenEnd,
		TokenWhen,
		TokenElse,
		TokenThen,
		TokenLimit,
		TokenBetween,
		TokenAnd,
		TokenOr,
		TokenXor:
		return true
	default:
		return false
	}
}

func IsLogicalOperator(t TokenType) bool {
	return t == TokenAnd || t == TokenOr || t == TokenXor
}

func IsTokenArray(token Token) bool {
	return token.Type == TokenReservedDataType && token.Text == "ARRAY"
}

func IsTokenBy(token Token) bool {
	return token.Type == TokenReservedKeyword && token.Text == "BY"
}

func IsTokenSet(token Token) bool {
	return token.Type == TokenReservedClause && token.Text == "SET"
}

func IsTokenStruct(token Token) bool {
	return token.Type == TokenReservedDataType && token.Text == "STRUCT"
}

func IsTokenWindow(token Token) bool {
	return token.Type == TokenReservedClause && token.Text == "WINDOW"
}

func IsTokenValues(token Token) bool {
	return token.Type == TokenReservedClause && token.Text == "VALUES"
}

package sqlformatter

type IdentChars struct {
	First                string
	Rest                 string
	Dashes               bool
	AllowFirstCharNumber bool
}

type PlainQuoteType string

type PrefixedQuoteType struct {
	Quote        PlainQuoteType
	Prefixes     []string
	RequirePrefix bool
}

type RegexPattern struct {
	Regex string
}

type QuoteType interface{}

type VariableType interface{}

type ParamTypes struct {
	Positional bool
	Numbered   []string
	Named      []string
	Quoted     []string
	Custom     []CustomParameter
}

type CustomParameter struct {
	Regex string
	Key   func(text string) string
}

type TokenizerOptions struct {
	ReservedSelect            []string
	ReservedClauses           []string
	SupportsXor               bool
	ReservedSetOperations     []string
	ReservedJoins             []string
	ReservedKeywordPhrases    []string
	ReservedDataTypePhrases   []string
	ReservedFunctionNames     []string
	ReservedDataTypes         []string
	ReservedKeywords          []string
	StringTypes               []QuoteType
	IdentTypes                []QuoteType
	VariableTypes             []VariableType
	ExtraParens               []string
	ParamTypes                *ParamTypes
	LineCommentTypes          []string
	NestedBlockComments       bool
	IdentChars                *IdentChars
	ParamChars                *IdentChars
	Operators                 []string
	PropertyAccessOperators   []string
	OperatorKeyword           bool
	UnderscoresInNumbers      bool
	PostProcess               func(tokens []Token) []Token
}

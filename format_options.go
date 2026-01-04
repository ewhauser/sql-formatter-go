package sqlformatter

type IndentStyle string

const (
	IndentStyleStandard     IndentStyle = "standard"
	IndentStyleTabularLeft  IndentStyle = "tabularLeft"
	IndentStyleTabularRight IndentStyle = "tabularRight"
)

type KeywordCase string

type IdentifierCase = KeywordCase

type DataTypeCase = KeywordCase

type FunctionCase = KeywordCase

type LogicalOperatorNewline string

const (
	KeywordCasePreserve KeywordCase = "preserve"
	KeywordCaseUpper    KeywordCase = "upper"
	KeywordCaseLower    KeywordCase = "lower"
)

const (
	LogicalOperatorNewlineBefore LogicalOperatorNewline = "before"
	LogicalOperatorNewlineAfter  LogicalOperatorNewline = "after"
)

type FormatOptions struct {
	TabWidth               int
	UseTabs                bool
	KeywordCase            KeywordCase
	IdentifierCase         IdentifierCase
	DataTypeCase           DataTypeCase
	FunctionCase           FunctionCase
	IndentStyle            IndentStyle
	LogicalOperatorNewline LogicalOperatorNewline
	ExpressionWidth        int
	ExpressionWidthSet     bool
	LinesBetweenQueries    int
	LinesBetweenQueriesSet bool
	DenseOperators         bool
	NewlineBeforeSemicolon bool
	Params                 ParamItemsOrList
	ParamTypes             *ParamTypes
}

type FormatOptionsWithLanguage struct {
	FormatOptions
	Language SqlLanguage
}

type FormatOptionsWithDialect struct {
	FormatOptions
	Dialect DialectOptions
}

package sqlformatter

type NodeType string

const (
	NodeStatement          NodeType = "statement"
	NodeClause             NodeType = "clause"
	NodeSetOperation       NodeType = "set_operation"
	NodeFunctionCall       NodeType = "function_call"
	NodeParameterizedDataType NodeType = "parameterized_data_type"
	NodeArraySubscript     NodeType = "array_subscript"
	NodePropertyAccess     NodeType = "property_access"
	NodeParenthesis        NodeType = "parenthesis"
	NodeBetweenPredicate   NodeType = "between_predicate"
	NodeCaseExpression     NodeType = "case_expression"
	NodeCaseWhen           NodeType = "case_when"
	NodeCaseElse           NodeType = "case_else"
	NodeLimitClause        NodeType = "limit_clause"
	NodeAllColumnsAsterisk NodeType = "all_columns_asterisk"
	NodeLiteral            NodeType = "literal"
	NodeIdentifier         NodeType = "identifier"
	NodeKeyword            NodeType = "keyword"
	NodeDataType           NodeType = "data_type"
	NodeParameter          NodeType = "parameter"
	NodeOperator           NodeType = "operator"
	NodeComma              NodeType = "comma"
	NodeLineComment        NodeType = "line_comment"
	NodeBlockComment       NodeType = "block_comment"
	NodeDisableComment     NodeType = "disable_comment"
)

type AstNode interface{}

type BaseNode struct {
	LeadingComments  []CommentNode
	TrailingComments []CommentNode
}

type StatementNode struct {
	BaseNode
	Type         NodeType
	Children     []AstNode
	HasSemicolon bool
}

type ClauseNode struct {
	BaseNode
	Type    NodeType
	NameKw  KeywordNode
	Children []AstNode
}

type SetOperationNode struct {
	BaseNode
	Type    NodeType
	NameKw  KeywordNode
	Children []AstNode
}

type FunctionCallNode struct {
	BaseNode
	Type       NodeType
	NameKw     KeywordNode
	Parenthesis ParenthesisNode
}

type ParameterizedDataTypeNode struct {
	BaseNode
	Type      NodeType
	DataType  DataTypeNode
	Parenthesis ParenthesisNode
}

type ArraySubscriptNode struct {
	BaseNode
	Type       NodeType
	Array      AstNode
	Parenthesis ParenthesisNode
}

type PropertyAccessNode struct {
	BaseNode
	Type     NodeType
	Object   AstNode
	Operator string
	Property AstNode
}

type ParenthesisNode struct {
	BaseNode
	Type      NodeType
	Children  []AstNode
	OpenParen string
	CloseParen string
}

type BetweenPredicateNode struct {
	BaseNode
	Type      NodeType
	BetweenKw KeywordNode
	Expr1     []AstNode
	AndKw     KeywordNode
	Expr2     []AstNode
}

type CaseExpressionNode struct {
	BaseNode
	Type    NodeType
	CaseKw  KeywordNode
	EndKw   KeywordNode
	Expr    []AstNode
	Clauses []AstNode
}

type CaseWhenNode struct {
	BaseNode
	Type      NodeType
	WhenKw    KeywordNode
	ThenKw    KeywordNode
	Condition []AstNode
	Result    []AstNode
}

type CaseElseNode struct {
	BaseNode
	Type   NodeType
	ElseKw KeywordNode
	Result []AstNode
}

type LimitClauseNode struct {
	BaseNode
	Type    NodeType
	LimitKw KeywordNode
	Count   []AstNode
	Offset  []AstNode
}

type AllColumnsAsteriskNode struct {
	BaseNode
	Type NodeType
}

type LiteralNode struct {
	BaseNode
	Type NodeType
	Text string
}

type IdentifierNode struct {
	BaseNode
	Type   NodeType
	Quoted bool
	Text   string
}

type DataTypeNode struct {
	BaseNode
	Type NodeType
	Text string
	Raw  string
}

type KeywordNode struct {
	BaseNode
	Type     NodeType
	TokenType TokenType
	Text     string
	Raw      string
}

type ParameterNode struct {
	BaseNode
	Type NodeType
	Key  string
	Text string
}

type OperatorNode struct {
	BaseNode
	Type NodeType
	Text string
}

type CommaNode struct {
	BaseNode
	Type NodeType
}

type LineCommentNode struct {
	BaseNode
	Type              NodeType
	Text              string
	PrecedingWhitespace string
}

type BlockCommentNode struct {
	BaseNode
	Type              NodeType
	Text              string
	PrecedingWhitespace string
}

type DisableCommentNode struct {
	BaseNode
	Type              NodeType
	Text              string
	PrecedingWhitespace string
}

type CommentNode interface{}

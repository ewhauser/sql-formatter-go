package sqlformatter

import (
	"regexp"
	"strings"
)

type ExpressionFormatter struct {
	cfg        FormatOptions
	dialectCfg ProcessedDialectFormatOptions
	params     *Params
	layout     LayoutWriter
	inline     bool
	nodes      []AstNode
	index      int
}

type ExpressionFormatterParams struct {
	Cfg        FormatOptions
	DialectCfg ProcessedDialectFormatOptions
	Params     *Params
	Layout     LayoutWriter
	Inline     bool
}

func NewExpressionFormatter(p ExpressionFormatterParams) *ExpressionFormatter {
	return &ExpressionFormatter{
		cfg:        p.Cfg,
		dialectCfg: p.DialectCfg,
		params:     p.Params,
		layout:     p.Layout,
		inline:     p.Inline,
		index:      -1,
	}
}

func (f *ExpressionFormatter) Format(nodes []AstNode) LayoutWriter {
	f.nodes = nodes
	for f.index = 0; f.index < len(f.nodes); f.index++ {
		f.formatNode(f.nodes[f.index])
	}
	return f.layout
}

func (f *ExpressionFormatter) formatNode(node AstNode) {
	f.formatComments(getLeadingComments(node))
	f.formatNodeWithoutComments(node)
	f.formatComments(getTrailingComments(node))
}

func (f *ExpressionFormatter) formatNodeWithoutComments(node AstNode) {
	switch n := node.(type) {
	case *FunctionCallNode:
		f.formatFunctionCall(n)
	case *ParameterizedDataTypeNode:
		f.formatParameterizedDataType(n)
	case *ArraySubscriptNode:
		f.formatArraySubscript(n)
	case *PropertyAccessNode:
		f.formatPropertyAccess(n)
	case *ParenthesisNode:
		f.formatParenthesis(n)
	case *BetweenPredicateNode:
		f.formatBetweenPredicate(n)
	case *CaseExpressionNode:
		f.formatCaseExpression(n)
	case *CaseWhenNode:
		f.formatCaseWhen(n)
	case *CaseElseNode:
		f.formatCaseElse(n)
	case *ClauseNode:
		f.formatClause(n)
	case *SetOperationNode:
		f.formatSetOperation(n)
	case *LimitClauseNode:
		f.formatLimitClause(n)
	case *AllColumnsAsteriskNode:
		f.formatAllColumnsAsterisk(n)
	case *LiteralNode:
		f.formatLiteral(n)
	case *IdentifierNode:
		f.formatIdentifier(n)
	case *ParameterNode:
		f.formatParameter(n)
	case *OperatorNode:
		f.formatOperator(n)
	case *CommaNode:
		f.formatComma(n)
	case *LineCommentNode:
		f.formatLineComment(n)
	case *BlockCommentNode:
		f.formatBlockComment(n)
	case *DisableCommentNode:
		f.formatDisableComment(n)
	case *DataTypeNode:
		f.formatDataType(n)
	case *KeywordNode:
		f.formatKeywordNode(n)
	}
}

func (f *ExpressionFormatter) formatFunctionCall(node *FunctionCallNode) {
	f.withCommentsKeyword(&node.NameKw, func() {
		f.layout.Add(f.showFunctionKw(&node.NameKw))
	})
	f.formatNode(&node.Parenthesis)
}

func (f *ExpressionFormatter) formatParameterizedDataType(node *ParameterizedDataTypeNode) {
	f.withCommentsDataType(&node.DataType, func() {
		f.layout.Add(f.showDataType(&node.DataType))
	})
	f.formatNode(&node.Parenthesis)
}

func (f *ExpressionFormatter) formatArraySubscript(node *ArraySubscriptNode) {
	var (
		formattedArray     string
		spaceBeforeBracket bool
	)
	switch arr := node.Array.(type) {
	case *DataTypeNode:
		formattedArray = f.showDataType(arr)
	case *ParameterizedDataTypeNode:
		formattedArray = f.showParameterizedDataTypeInline(arr)
		spaceBeforeBracket = true
	case *KeywordNode:
		formattedArray = f.showKw(arr)
	case *IdentifierNode:
		formattedArray = f.showIdentifier(arr)
	default:
		formattedArray = ""
	}
	f.withComments(node.Array, func() {
		if spaceBeforeBracket {
			f.layout.Add(formattedArray, Space)
			return
		}
		f.layout.Add(formattedArray)
	})
	f.formatNode(&node.Parenthesis)
}

func (f *ExpressionFormatter) showParameterizedDataTypeInline(node *ParameterizedDataTypeNode) string {
	inlineLayout := f.formatInlineExpression([]AstNode{node})
	if inlineLayout != nil {
		return strings.TrimRight(inlineLayout.ToString(), " ")
	}
	layout := NewLayout(NewIndentation(f.layout.GetIndentation().GetSingleIndent()))
	formatter := NewExpressionFormatter(ExpressionFormatterParams{Cfg: f.cfg, DialectCfg: f.dialectCfg, Params: f.params, Layout: layout, Inline: true})
	formatter.Format([]AstNode{node})
	return strings.TrimRight(layout.ToString(), " ")
}

func (f *ExpressionFormatter) formatPropertyAccess(node *PropertyAccessNode) {
	f.formatNode(node.Object)
	f.layout.Add(NoSpace, node.Operator)
	f.formatNode(node.Property)
}

func (f *ExpressionFormatter) formatParenthesis(node *ParenthesisNode) {
	inlineLayout := f.formatInlineExpression(node.Children)
	if inlineLayout != nil {
		f.layout.Add(node.OpenParen)
		for _, item := range inlineLayout.GetLayoutItems() {
			f.layout.Add(item)
		}
		f.layout.Add(NoSpace, node.CloseParen, Space)
		return
	}
	f.layout.Add(node.OpenParen, Newline)
	if isTabularStyle(f.cfg) {
		f.layout.Add(Indent)
		f.layout = f.formatSubExpression(node.Children)
	} else {
		f.layout.GetIndentation().IncreaseBlockLevel()
		f.layout.Add(Indent)
		f.layout = f.formatSubExpression(node.Children)
		f.layout.GetIndentation().DecreaseBlockLevel()
	}
	f.layout.Add(Newline, Indent, node.CloseParen, Space)
}

func (f *ExpressionFormatter) formatBetweenPredicate(node *BetweenPredicateNode) {
	f.layout.Add(f.showKw(&node.BetweenKw), Space)
	f.layout = f.formatSubExpression(node.Expr1)
	f.layout.Add(NoSpace, Space, f.showNonTabularKw(&node.AndKw), Space)
	f.layout = f.formatSubExpression(node.Expr2)
	f.layout.Add(Space)
}

func (f *ExpressionFormatter) formatCaseExpression(node *CaseExpressionNode) {
	f.formatNode(&node.CaseKw)
	f.layout.GetIndentation().IncreaseBlockLevel()
	f.layout = f.formatSubExpression(node.Expr)
	f.layout = f.formatSubExpression(node.Clauses)
	f.layout.GetIndentation().DecreaseBlockLevel()
	f.layout.Add(Newline, Indent)
	f.formatNode(&node.EndKw)
}

func (f *ExpressionFormatter) formatCaseWhen(node *CaseWhenNode) {
	f.layout.Add(Newline, Indent)
	f.formatNode(&node.WhenKw)
	f.layout = f.formatSubExpression(node.Condition)
	f.formatNode(&node.ThenKw)
	f.layout = f.formatSubExpression(node.Result)
}

func (f *ExpressionFormatter) formatCaseElse(node *CaseElseNode) {
	f.layout.Add(Newline, Indent)
	f.formatNode(&node.ElseKw)
	f.layout = f.formatSubExpression(node.Result)
}

func (f *ExpressionFormatter) formatClause(node *ClauseNode) {
	if f.isOnelineClause(node) {
		f.formatClauseInOnelineStyle(node)
	} else if isTabularStyle(f.cfg) {
		f.formatClauseInTabularStyle(node)
	} else {
		f.formatClauseInIndentedStyle(node)
	}
}

func (f *ExpressionFormatter) isOnelineClause(node *ClauseNode) bool {
	if isTabularStyle(f.cfg) {
		return f.dialectCfg.TabularOnelineClauses[node.NameKw.Text]
	}
	return f.dialectCfg.OnelineClauses[node.NameKw.Text]
}

func (f *ExpressionFormatter) formatClauseInIndentedStyle(node *ClauseNode) {
	f.layout.Add(Newline, Indent, f.showKw(&node.NameKw), Newline)
	f.layout.GetIndentation().IncreaseTopLevel()
	f.layout.Add(Indent)
	f.layout = f.formatSubExpression(node.Children)
	f.layout.GetIndentation().DecreaseTopLevel()
}

func (f *ExpressionFormatter) formatClauseInOnelineStyle(node *ClauseNode) {
	f.layout.Add(Newline, Indent, f.showKw(&node.NameKw), Space)
	f.layout = f.formatSubExpression(node.Children)
}

func (f *ExpressionFormatter) formatClauseInTabularStyle(node *ClauseNode) {
	f.layout.Add(Newline, Indent, f.showKw(&node.NameKw), Space)
	f.layout.GetIndentation().IncreaseTopLevel()
	f.layout = f.formatSubExpression(node.Children)
	f.layout.GetIndentation().DecreaseTopLevel()
}

func (f *ExpressionFormatter) formatSetOperation(node *SetOperationNode) {
	f.layout.Add(Newline, Indent, f.showKw(&node.NameKw), Newline)
	f.layout.Add(Indent)
	f.layout = f.formatSubExpression(node.Children)
}

func (f *ExpressionFormatter) formatLimitClause(node *LimitClauseNode) {
	f.withCommentsKeyword(&node.LimitKw, func() {
		f.layout.Add(Newline, Indent, f.showKw(&node.LimitKw))
	})
	f.layout.GetIndentation().IncreaseTopLevel()
	if isTabularStyle(f.cfg) {
		f.layout.Add(Space)
	} else {
		f.layout.Add(Newline, Indent)
	}
	if len(node.Offset) > 0 {
		f.layout = f.formatSubExpression(node.Offset)
		f.layout.Add(NoSpace, ",", Space)
		f.layout = f.formatSubExpression(node.Count)
	} else {
		f.layout = f.formatSubExpression(node.Count)
	}
	f.layout.GetIndentation().DecreaseTopLevel()
}

func (f *ExpressionFormatter) formatAllColumnsAsterisk(_ *AllColumnsAsteriskNode) {
	f.layout.Add("*", Space)
}

func (f *ExpressionFormatter) formatLiteral(node *LiteralNode) {
	f.layout.Add(node.Text, Space)
}

func (f *ExpressionFormatter) formatIdentifier(node *IdentifierNode) {
	f.layout.Add(f.showIdentifier(node), Space)
}

func (f *ExpressionFormatter) formatParameter(node *ParameterNode) {
	key := node.Key
	text := node.Text
	value := text
	if f.params != nil {
		value = f.params.Get(key, text)
	}
	f.layout.Add(value, Space)
}

func (f *ExpressionFormatter) formatOperator(node *OperatorNode) {
	text := node.Text
	if f.cfg.DenseOperators || containsString(f.dialectCfg.AlwaysDenseOperators, text) {
		f.layout.Add(NoSpace, text)
	} else if text == ":" {
		f.layout.Add(NoSpace, text, Space)
	} else {
		f.layout.Add(text, Space)
	}
}

func (f *ExpressionFormatter) formatComma(_ *CommaNode) {
	if !f.inline {
		f.layout.Add(NoSpace, ",", Newline, Indent)
	} else {
		f.layout.Add(NoSpace, ",", Space)
	}
}

func (f *ExpressionFormatter) withComments(node AstNode, fn func()) {
	f.formatComments(getLeadingComments(node))
	fn()
	f.formatComments(getTrailingComments(node))
}

func (f *ExpressionFormatter) withCommentsKeyword(node *KeywordNode, fn func()) {
	f.formatComments(node.LeadingComments)
	fn()
	f.formatComments(node.TrailingComments)
}

func (f *ExpressionFormatter) withCommentsDataType(node *DataTypeNode, fn func()) {
	f.formatComments(node.LeadingComments)
	fn()
	f.formatComments(node.TrailingComments)
}

func (f *ExpressionFormatter) formatComments(comments []CommentNode) {
	for _, com := range comments {
		switch c := com.(type) {
		case *LineCommentNode:
			f.formatLineComment(c)
		case *BlockCommentNode:
			f.formatBlockComment(c)
		case *DisableCommentNode:
			f.formatDisableComment(c)
		}
	}
}

func (f *ExpressionFormatter) formatDisableComment(node *DisableCommentNode) {
	if IsMultiline(node.Text) || IsMultiline(node.PrecedingWhitespace) {
		f.layout.Add(Newline, Indent, node.Text, Newline, Indent)
	} else {
		f.layout.Add(node.Text, Space)
	}
}

func (f *ExpressionFormatter) formatLineComment(node *LineCommentNode) {
	if IsMultiline(node.PrecedingWhitespace) {
		f.layout.Add(Newline, Indent, node.Text, MandatoryNewline, Indent)
	} else if len(f.layout.GetLayoutItems()) > 0 {
		f.layout.Add(NoNewline, Space, node.Text, MandatoryNewline, Indent)
	} else {
		f.layout.Add(node.Text, MandatoryNewline, Indent)
	}
}

func (f *ExpressionFormatter) formatBlockComment(node *BlockCommentNode) {
	if f.isMultilineBlockComment(node) {
		for _, line := range f.splitBlockComment(node.Text) {
			f.layout.Add(Newline, Indent, line)
		}
		f.layout.Add(Newline, Indent)
	} else {
		f.layout.Add(node.Text, Space)
	}
}

func (f *ExpressionFormatter) isMultilineBlockComment(node *BlockCommentNode) bool {
	return IsMultiline(node.Text) || IsMultiline(node.PrecedingWhitespace)
}

func (f *ExpressionFormatter) isDocComment(comment string) bool {
	lines := strings.Split(comment, "\n")
	if len(lines) == 0 {
		return false
	}
	first := lines[0]
	last := lines[len(lines)-1]
	if !regexp.MustCompile(`^/\*\*?$`).MatchString(first) {
		return false
	}
	for _, line := range lines[1 : len(lines)-1] {
		if !regexp.MustCompile(`^\s*\*`).MatchString(line) {
			return false
		}
	}
	return regexp.MustCompile(`^\s*\*/$`).MatchString(last)
}

func (f *ExpressionFormatter) splitBlockComment(comment string) []string {
	lines := strings.Split(comment, "\n")
	if f.isDocComment(comment) {
		out := make([]string, 0, len(lines))
		for _, line := range lines {
			if regexp.MustCompile(`^\s*\*`).MatchString(line) {
				out = append(out, " "+strings.TrimLeft(line, " \t"))
			} else {
				out = append(out, line)
			}
		}
		return out
	}
	out := make([]string, 0, len(lines))
	for _, line := range lines {
		out = append(out, strings.TrimLeft(line, " \t"))
	}
	return out
}

func (f *ExpressionFormatter) formatSubExpression(nodes []AstNode) LayoutWriter {
	return NewExpressionFormatter(ExpressionFormatterParams{Cfg: f.cfg, DialectCfg: f.dialectCfg, Params: f.params, Layout: f.layout, Inline: f.inline}).Format(nodes)
}

func (f *ExpressionFormatter) formatInlineExpression(nodes []AstNode) LayoutWriter {
	oldIndex := 0
	if f.params != nil {
		oldIndex = f.params.GetPositionalParameterIndex()
	}
	inlineLayout := NewInlineLayout(f.cfg.ExpressionWidth)
	formatter := NewExpressionFormatter(ExpressionFormatterParams{Cfg: f.cfg, DialectCfg: f.dialectCfg, Params: f.params, Layout: inlineLayout, Inline: true})
	var inlineErr bool
	defer func() {
		if r := recover(); r != nil {
			if _, ok := r.(InlineLayoutError); ok {
				if f.params != nil {
					f.params.SetPositionalParameterIndex(oldIndex)
				}
				inlineErr = true
			} else {
				panic(r)
			}
		}
	}()
	formatter.Format(nodes)
	if inlineErr {
		return nil
	}
	return inlineLayout
}

func (f *ExpressionFormatter) formatKeywordNode(node *KeywordNode) {
	switch node.TokenType {
	case TokenReservedJoin:
		f.formatJoin(node)
	case TokenAnd, TokenOr, TokenXor:
		f.formatLogicalOperator(node)
	default:
		f.formatKeyword(node)
	}
}

func (f *ExpressionFormatter) formatJoin(node *KeywordNode) {
	if isTabularStyle(f.cfg) {
		f.layout.GetIndentation().DecreaseTopLevel()
		f.layout.Add(Newline, Indent, f.showKw(node), Space)
		f.layout.GetIndentation().IncreaseTopLevel()
	} else {
		f.layout.Add(Newline, Indent, f.showKw(node), Space)
	}
}

func (f *ExpressionFormatter) formatKeyword(node *KeywordNode) {
	f.layout.Add(f.showKw(node), Space)
}

func (f *ExpressionFormatter) formatLogicalOperator(node *KeywordNode) {
	if f.cfg.LogicalOperatorNewline == LogicalOperatorNewlineBefore {
		if isTabularStyle(f.cfg) {
			f.layout.GetIndentation().DecreaseTopLevel()
			f.layout.Add(Newline, Indent, f.showKw(node), Space)
			f.layout.GetIndentation().IncreaseTopLevel()
		} else {
			f.layout.Add(Newline, Indent, f.showKw(node), Space)
		}
	} else {
		f.layout.Add(f.showKw(node), Newline, Indent)
	}
}

func (f *ExpressionFormatter) formatDataType(node *DataTypeNode) {
	f.layout.Add(f.showDataType(node), Space)
}

func (f *ExpressionFormatter) showKw(node *KeywordNode) string {
	if isTabularToken(node.TokenType) {
		return toTabularFormat(f.showNonTabularKw(node), f.cfg.IndentStyle)
	}
	return f.showNonTabularKw(node)
}

func (f *ExpressionFormatter) showNonTabularKw(node *KeywordNode) string {
	switch f.cfg.KeywordCase {
	case KeywordCasePreserve:
		return EqualizeWhitespace(node.Raw)
	case KeywordCaseUpper:
		return node.Text
	case KeywordCaseLower:
		return strings.ToLower(node.Text)
	default:
		return node.Text
	}
}

func (f *ExpressionFormatter) showFunctionKw(node *KeywordNode) string {
	if isTabularToken(node.TokenType) {
		return toTabularFormat(f.showNonTabularFunctionKw(node), f.cfg.IndentStyle)
	}
	return f.showNonTabularFunctionKw(node)
}

func (f *ExpressionFormatter) showNonTabularFunctionKw(node *KeywordNode) string {
	switch f.cfg.FunctionCase {
	case KeywordCasePreserve:
		return EqualizeWhitespace(node.Raw)
	case KeywordCaseUpper:
		return node.Text
	case KeywordCaseLower:
		return strings.ToLower(node.Text)
	default:
		return node.Text
	}
}

func (f *ExpressionFormatter) showIdentifier(node *IdentifierNode) string {
	if node.Quoted {
		return node.Text
	}
	switch f.cfg.IdentifierCase {
	case KeywordCasePreserve:
		return node.Text
	case KeywordCaseUpper:
		return strings.ToUpper(node.Text)
	case KeywordCaseLower:
		return strings.ToLower(node.Text)
	default:
		return node.Text
	}
}

func (f *ExpressionFormatter) showDataType(node *DataTypeNode) string {
	switch f.cfg.DataTypeCase {
	case KeywordCasePreserve:
		return EqualizeWhitespace(node.Raw)
	case KeywordCaseUpper:
		return node.Text
	case KeywordCaseLower:
		return strings.ToLower(node.Text)
	default:
		return node.Text
	}
}

// comment accessors
func getLeadingComments(node AstNode) []CommentNode {
	switch n := node.(type) {
	case *ClauseNode:
		return n.LeadingComments
	case *SetOperationNode:
		return n.LeadingComments
	case *FunctionCallNode:
		return n.LeadingComments
	case *ParameterizedDataTypeNode:
		return n.LeadingComments
	case *ArraySubscriptNode:
		return n.LeadingComments
	case *PropertyAccessNode:
		return n.LeadingComments
	case *ParenthesisNode:
		return n.LeadingComments
	case *BetweenPredicateNode:
		return n.LeadingComments
	case *CaseExpressionNode:
		return n.LeadingComments
	case *CaseWhenNode:
		return n.LeadingComments
	case *CaseElseNode:
		return n.LeadingComments
	case *LimitClauseNode:
		return n.LeadingComments
	case *AllColumnsAsteriskNode:
		return n.LeadingComments
	case *LiteralNode:
		return n.LeadingComments
	case *IdentifierNode:
		return n.LeadingComments
	case *KeywordNode:
		return n.LeadingComments
	case *DataTypeNode:
		return n.LeadingComments
	case *ParameterNode:
		return n.LeadingComments
	case *OperatorNode:
		return n.LeadingComments
	case *CommaNode:
		return n.LeadingComments
	case *LineCommentNode:
		return n.LeadingComments
	case *BlockCommentNode:
		return n.LeadingComments
	case *DisableCommentNode:
		return n.LeadingComments
	default:
		return nil
	}
}

func getTrailingComments(node AstNode) []CommentNode {
	switch n := node.(type) {
	case *ClauseNode:
		return n.TrailingComments
	case *SetOperationNode:
		return n.TrailingComments
	case *FunctionCallNode:
		return n.TrailingComments
	case *ParameterizedDataTypeNode:
		return n.TrailingComments
	case *ArraySubscriptNode:
		return n.TrailingComments
	case *PropertyAccessNode:
		return n.TrailingComments
	case *ParenthesisNode:
		return n.TrailingComments
	case *BetweenPredicateNode:
		return n.TrailingComments
	case *CaseExpressionNode:
		return n.TrailingComments
	case *CaseWhenNode:
		return n.TrailingComments
	case *CaseElseNode:
		return n.TrailingComments
	case *LimitClauseNode:
		return n.TrailingComments
	case *AllColumnsAsteriskNode:
		return n.TrailingComments
	case *LiteralNode:
		return n.TrailingComments
	case *IdentifierNode:
		return n.TrailingComments
	case *KeywordNode:
		return n.TrailingComments
	case *DataTypeNode:
		return n.TrailingComments
	case *ParameterNode:
		return n.TrailingComments
	case *OperatorNode:
		return n.TrailingComments
	case *CommaNode:
		return n.TrailingComments
	case *LineCommentNode:
		return n.TrailingComments
	case *BlockCommentNode:
		return n.TrailingComments
	case *DisableCommentNode:
		return n.TrailingComments
	default:
		return nil
	}
}

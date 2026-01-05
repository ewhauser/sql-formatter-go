package sqlformatter

import (
	"fmt"
)

type Parser struct {
	tokens []Token
	index  int
}

func NewParser(tokenizer *Tokenizer) *Parser {
	return &Parser{}
}

func (p *Parser) Parse(sql string, tokenizer *Tokenizer, paramTypesOverrides *ParamTypes) ([]*StatementNode, error) {
	if tokenizer == nil {
		return nil, fmt.Errorf("tokenizer is nil")
	}
	tokens, err := tokenizer.Tokenize(sql, paramTypesOverrides)
	if err != nil {
		return nil, err
	}
	tokens = DisambiguateTokens(tokens)
	// append EOF token
	tokens = append(tokens, CreateEofToken(len(sql)))
	p.tokens = tokens
	p.index = 0
	return p.parseMain()
}

func (p *Parser) parseMain() ([]*StatementNode, error) {
	statements := []*StatementNode{}
	for {
		if p.peek().Type == TokenEOF {
			break
		}
		stmt, err := p.parseStatement()
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
		if p.peek().Type == TokenEOF {
			break
		}
	}
	if len(statements) == 0 {
		return statements, nil
	}
	last := statements[len(statements)-1]
	if !last.HasSemicolon && len(last.Children) == 0 {
		return statements[:len(statements)-1], nil
	}
	return statements, nil
}

func (p *Parser) parseStatement() (*StatementNode, error) {
	children, err := p.parseExpressionsOrClauses(TokenDelimiter, TokenEOF)
	if err != nil {
		return nil, err
	}
	hasSemicolon := false
	if p.peek().Type == TokenDelimiter {
		hasSemicolon = true
		p.consume()
	} else if p.peek().Type == TokenEOF {
		// ok
	} else {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	return &StatementNode{Type: NodeStatement, Children: children, HasSemicolon: hasSemicolon}, nil
}

func (p *Parser) parseExpressionsOrClauses(stopTypes ...TokenType) ([]AstNode, error) {
	expressions := []AstNode{}
	clauses := []AstNode{}
	for {
		if p.isStop(stopTypes...) {
			break
		}
		if p.isClauseStart(p.peek()) {
			break
		}
		node, ok, err := p.parseFreeFormSQL()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		expressions = append(expressions, node)
	}
	for {
		if p.isStop(stopTypes...) {
			break
		}
		node, ok, err := p.parseClause()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		clauses = append(clauses, node)
	}
	return append(expressions, clauses...), nil
}

func (p *Parser) parseClause() (AstNode, bool, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenLimit:
		node, err := p.parseLimitClause()
		return node, err == nil, err
	case TokenReservedSelect:
		node, err := p.parseSelectClause()
		return node, err == nil, err
	case TokenReservedClause:
		node, err := p.parseOtherClause()
		return node, err == nil, err
	case TokenReservedSetOperation:
		node, err := p.parseSetOperation()
		return node, err == nil, err
	default:
		return nil, false, nil
	}
}

func (p *Parser) parseLimitClause() (*LimitClauseNode, error) {
	limitTok := p.consume()
	trailing := p.parseComments()
	limitKw := KeywordNode{Type: NodeKeyword, TokenType: limitTok.Type, Text: limitTok.Text, Raw: limitTok.Raw}
	limitKw = addTrailingCommentsKeyword(limitKw, trailing)

	expr1, err := p.parseExpressionChainTrailing()
	if err != nil {
		return nil, err
	}

	var offset []AstNode
	count := expr1
	if p.peek().Type == TokenComma {
		p.consume()
		exp2 := []AstNode{}
		for {
			node, ok, err := p.parseFreeFormSQL()
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
			exp2 = append(exp2, node)
		}
		offset = expr1
		count = exp2
	}

	return &LimitClauseNode{Type: NodeLimitClause, LimitKw: limitKw, Count: count, Offset: offset}, nil
}

func (p *Parser) parseSelectClause() (*ClauseNode, error) {
	selectTok := p.consume()
	nameKw := KeywordNode{Type: NodeKeyword, TokenType: selectTok.Type, Text: selectTok.Text, Raw: selectTok.Raw}
	children := []AstNode{}
	if p.peek().Type == TokenAsterisk {
		p.consume()
		children = append(children, &AllColumnsAsteriskNode{Type: NodeAllColumnsAsterisk})
		for {
			node, ok, err := p.parseFreeFormSQL()
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
			children = append(children, node)
		}
	} else {
		// attempt to parse any free form SQL elements
		for {
			if p.isClauseStart(p.peek()) || p.isStop(TokenDelimiter, TokenEOF, TokenCloseParen) {
				break
			}
			node, ok, err := p.parseAsterisklessFreeFormSQL()
			if err != nil {
				return nil, err
			}
			if !ok {
				break
			}
			children = append(children, node)
			// allow further free form SQL items
			for {
				next, ok, err := p.parseFreeFormSQL()
				if err != nil {
					return nil, err
				}
				if !ok {
					break
				}
				children = append(children, next)
			}
			break
		}
	}
	return &ClauseNode{Type: NodeClause, NameKw: nameKw, Children: children}, nil
}

func (p *Parser) parseOtherClause() (*ClauseNode, error) {
	clauseTok := p.consume()
	nameKw := KeywordNode{Type: NodeKeyword, TokenType: clauseTok.Type, Text: clauseTok.Text, Raw: clauseTok.Raw}
	children := []AstNode{}
	for {
		if p.isClauseStart(p.peek()) || p.isStop(TokenDelimiter, TokenEOF, TokenCloseParen) {
			break
		}
		node, ok, err := p.parseFreeFormSQL()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		children = append(children, node)
	}
	return &ClauseNode{Type: NodeClause, NameKw: nameKw, Children: children}, nil
}

func (p *Parser) parseSetOperation() (*SetOperationNode, error) {
	opTok := p.consume()
	nameKw := KeywordNode{Type: NodeKeyword, TokenType: opTok.Type, Text: opTok.Text, Raw: opTok.Raw}
	children := []AstNode{}
	for {
		if p.isClauseStart(p.peek()) || p.isStop(TokenDelimiter, TokenEOF, TokenCloseParen) {
			break
		}
		node, ok, err := p.parseFreeFormSQL()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		children = append(children, node)
	}
	return &SetOperationNode{Type: NodeSetOperation, NameKw: nameKw, Children: children}, nil
}

func (p *Parser) parseExpressionChainTrailing() ([]AstNode, error) {
	items := []AstNode{}
	for {
		expr, ok, err := p.parseExpression()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		trailing := p.parseComments()
		expr = addTrailingComments(expr, trailing)
		items = append(items, expr)
		// continue if next token can start expression
		if !p.canStartExpression(p.peek()) {
			break
		}
	}
	return items, nil
}

func (p *Parser) parseExpression() (AstNode, bool, error) {
	if p.peek().Type == TokenAnd || p.peek().Type == TokenOr || p.peek().Type == TokenXor {
		kw := p.consume()
		return &KeywordNode{Type: NodeKeyword, TokenType: kw.Type, Text: kw.Text, Raw: kw.Raw}, true, nil
	}
	return p.parseAndlessExpression()
}

func (p *Parser) parseAndlessExpression() (AstNode, bool, error) {
	if p.peek().Type == TokenAsterisk {
		tok := p.consume()
		return &OperatorNode{Type: NodeOperator, Text: tok.Text}, true, nil
	}
	return p.parseAsterisklessAndlessExpression()
}

func (p *Parser) parseAsterisklessAndlessExpression() (AstNode, bool, error) {
	if p.peek().Type == TokenBetween {
		node, err := p.parseBetweenPredicate()
		return node, err == nil, err
	}
	if p.peek().Type == TokenCase {
		node, err := p.parseCaseExpression()
		return node, err == nil, err
	}
	return p.parseAtomicExpression()
}

func (p *Parser) parseFreeFormSQL() (AstNode, bool, error) {
	if p.peek().Type == TokenAsterisk {
		p.consume()
		return &OperatorNode{Type: NodeOperator, Text: "*"}, true, nil
	}
	return p.parseAsterisklessFreeFormSQL()
}

func (p *Parser) parseAsterisklessFreeFormSQL() (AstNode, bool, error) {
	// logic operator
	if p.peek().Type == TokenAnd || p.peek().Type == TokenOr || p.peek().Type == TokenXor {
		kw := p.consume()
		return &KeywordNode{Type: NodeKeyword, TokenType: kw.Type, Text: kw.Text, Raw: kw.Raw}, true, nil
	}
	// comma
	if p.peek().Type == TokenComma {
		p.consume()
		return &CommaNode{Type: NodeComma}, true, nil
	}
	// comment
	if p.isCommentToken(p.peek()) {
		node := p.parseCommentNode()
		return node, true, nil
	}
	// other keyword
	if p.peek().Type == TokenWhen || p.peek().Type == TokenThen || p.peek().Type == TokenElse || p.peek().Type == TokenEnd {
		kw := p.consume()
		return &KeywordNode{Type: NodeKeyword, TokenType: kw.Type, Text: kw.Text, Raw: kw.Raw}, true, nil
	}
	return p.parseAsterisklessAndlessExpression()
}

func (p *Parser) parseAtomicExpression() (AstNode, bool, error) {
	var base AstNode
	// array subscript
	if p.peek().Type == TokenArrayIdentifier || p.peek().Type == TokenArrayKeyword {
		node, ok, err := p.parseArraySubscript()
		if err != nil || !ok {
			return node, ok, err
		}
		base = node
		goto propertyAccess
	}
	// function call
	if p.peek().Type == TokenReservedFunctionName {
		node, ok, err := p.parseFunctionCall()
		if err != nil || !ok {
			return node, ok, err
		}
		base = node
		goto propertyAccess
	}
	// parameterized data type
	if p.peek().Type == TokenReservedParameterizedDataType {
		node, ok, err := p.parseParameterizedDataType()
		if err != nil || !ok {
			return node, ok, err
		}
		base = node
		goto arraySuffix
	}
	// parenthesis or brackets
	if p.peek().Type == TokenOpenParen {
		node, err := p.parseParenthesis()
		if err != nil {
			return nil, false, err
		}
		base = node
		goto propertyAccess
	}
	// operator
	if p.peek().Type == TokenOperator {
		tok := p.consume()
		base = &OperatorNode{Type: NodeOperator, Text: tok.Text}
		goto propertyAccess
	}
	// identifier
	if p.peek().Type == TokenIdentifier || p.peek().Type == TokenQuotedIdentifier || p.peek().Type == TokenVariable {
		tok := p.consume()
		quoted := tok.Type != TokenIdentifier
		base = &IdentifierNode{Type: NodeIdentifier, Quoted: quoted, Text: tok.Text}
		goto propertyAccess
	}
	// parameter
	if p.isParameterToken(p.peek()) {
		tok := p.consume()
		base = &ParameterNode{Type: NodeParameter, Key: tok.Key, Text: tok.Text}
		goto propertyAccess
	}
	// literal
	if p.peek().Type == TokenNumber || p.peek().Type == TokenString {
		tok := p.consume()
		base = &LiteralNode{Type: NodeLiteral, Text: tok.Text}
		goto propertyAccess
	}
	// data type
	if p.peek().Type == TokenReservedDataType || p.peek().Type == TokenReservedDataTypePhrase {
		tok := p.consume()
		base = &DataTypeNode{Type: NodeDataType, Text: tok.Text, Raw: tok.Raw}
		goto arraySuffix
	}
	// keyword
	if p.peek().Type == TokenReservedKeyword || p.peek().Type == TokenReservedKeywordPhrase || p.peek().Type == TokenReservedJoin {
		tok := p.consume()
		base = &KeywordNode{Type: NodeKeyword, TokenType: tok.Type, Text: tok.Text, Raw: tok.Raw}
		goto arraySuffix
	}

	return nil, false, nil

arraySuffix:
	if base == nil {
		return nil, false, nil
	}
	if _, ok := base.(*ParameterizedDataTypeNode); ok {
		if p.peek().Type == TokenOpenParen && p.peek().Text == "[" {
			parens, err := p.parseSquareBrackets()
			if err != nil {
				return nil, false, err
			}
			base = &ArraySubscriptNode{Type: NodeArraySubscript, Array: base, Parenthesis: *parens}
		}
	}

propertyAccess:
	if base == nil {
		return nil, false, nil
	}
	chain, err := p.parsePropertyAccessChain(base)
	if err != nil {
		return nil, false, err
	}
	return chain, true, nil
}

func (p *Parser) parsePropertyAccessChain(node AstNode) (AstNode, error) {
	for {
		next, commentCount := p.peekAfterComments()
		if next.Type != TokenPropertyAccessOperator {
			break
		}
		trailing := p.consumeComments(commentCount)
		if len(trailing) > 0 {
			node = addTrailingComments(node, trailing)
		}
		opTok := p.consume()
		leading := p.parseComments()
		prop, err := p.parsePropertyAccessProperty()
		if err != nil {
			return nil, err
		}
		prop = addLeadingComments(prop, leading)
		node = &PropertyAccessNode{Type: NodePropertyAccess, Object: node, Operator: opTok.Text, Property: prop}
	}
	return node, nil
}

func (p *Parser) parsePropertyAccessProperty() (AstNode, error) {
	tok := p.peek()
	switch tok.Type {
	case TokenAsterisk:
		p.consume()
		return &AllColumnsAsteriskNode{Type: NodeAllColumnsAsterisk}, nil
	case TokenArrayIdentifier, TokenArrayKeyword:
		node, ok, err := p.parseArraySubscript()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("Parse error: Invalid SQL")
		}
		return node, nil
	case TokenReservedFunctionName:
		// Handle function calls in property access (e.g., sqlc.arg())
		node, ok, err := p.parseFunctionCall()
		if err != nil || !ok {
			return nil, fmt.Errorf("Parse error: Invalid SQL")
		}
		return node, nil
	case TokenIdentifier, TokenQuotedIdentifier, TokenVariable:
		p.consume()
		quoted := tok.Type != TokenIdentifier
		return &IdentifierNode{Type: NodeIdentifier, Quoted: quoted, Text: tok.Text}, nil
	case TokenNamedParameter, TokenQuotedParameter, TokenNumberedParameter, TokenPositionalParameter, TokenCustomParameter:
		p.consume()
		return &ParameterNode{Type: NodeParameter, Key: tok.Key, Text: tok.Text}, nil
	default:
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
}

func (p *Parser) parseArraySubscript() (AstNode, bool, error) {
	tok := p.consume()
	trailing := p.parseComments()
	var array AstNode
	switch tok.Type {
	case TokenArrayIdentifier:
		array = &IdentifierNode{Type: NodeIdentifier, Quoted: false, Text: tok.Text}
	case TokenArrayKeyword:
		array = &KeywordNode{Type: NodeKeyword, TokenType: tok.Type, Text: tok.Text, Raw: tok.Raw}
	default:
		return nil, false, nil
	}
	array = addTrailingComments(array, trailing)
	parens, err := p.parseSquareBrackets()
	if err != nil {
		return nil, false, err
	}
	return &ArraySubscriptNode{Type: NodeArraySubscript, Array: array, Parenthesis: *parens}, true, nil
}

func (p *Parser) parseFunctionCall() (AstNode, bool, error) {
	nameTok := p.consume()
	trailing := p.parseComments()
	nameKw := KeywordNode{Type: NodeKeyword, TokenType: nameTok.Type, Text: nameTok.Text, Raw: nameTok.Raw}
	nameKw = addTrailingCommentsKeyword(nameKw, trailing)
	parens, err := p.parseParenthesis()
	if err != nil {
		return nil, false, err
	}
	return &FunctionCallNode{Type: NodeFunctionCall, NameKw: nameKw, Parenthesis: *parens}, true, nil
}

func (p *Parser) parseParameterizedDataType() (AstNode, bool, error) {
	nameTok := p.consume()
	trailing := p.parseComments()
	dataType := DataTypeNode{Type: NodeDataType, Text: nameTok.Text, Raw: nameTok.Raw}
	dataType = addTrailingCommentsDataType(dataType, trailing)
	parens, err := p.parseParenthesis()
	if err != nil {
		return nil, false, err
	}
	return &ParameterizedDataTypeNode{Type: NodeParameterizedDataType, DataType: dataType, Parenthesis: *parens}, true, nil
}

func (p *Parser) parseParenthesis() (*ParenthesisNode, error) {
	openTok := p.consume()
	open := openTok.Text
	var close string
	if open == "(" {
		close = ")"
		children, err := p.parseExpressionsOrClauses(TokenCloseParen)
		if err != nil {
			return nil, err
		}
		if p.peek().Type != TokenCloseParen || p.peek().Text != close {
			return nil, fmt.Errorf("Parse error: Invalid SQL")
		}
		p.consume()
		return &ParenthesisNode{Type: NodeParenthesis, Children: children, OpenParen: open, CloseParen: close}, nil
	}
	if open == "{" {
		close = "}"
		children, err := p.parseFreeFormListUntilClose(close)
		if err != nil {
			return nil, err
		}
		return &ParenthesisNode{Type: NodeParenthesis, Children: children, OpenParen: open, CloseParen: close}, nil
	}
	if open == "[" {
		close = "]"
		children, err := p.parseFreeFormListUntilClose(close)
		if err != nil {
			return nil, err
		}
		return &ParenthesisNode{Type: NodeParenthesis, Children: children, OpenParen: open, CloseParen: close}, nil
	}
	return nil, fmt.Errorf("Parse error: Invalid SQL")
}

func (p *Parser) parseSquareBrackets() (*ParenthesisNode, error) {
	if p.peek().Type != TokenOpenParen || p.peek().Text != "[" {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	return p.parseParenthesis()
}

func (p *Parser) parseFreeFormListUntilClose(close string) ([]AstNode, error) {
	children := []AstNode{}
	for {
		if p.peek().Type == TokenCloseParen && p.peek().Text == close {
			p.consume()
			break
		}
		node, ok, err := p.parseFreeFormSQL()
		if err != nil {
			return nil, err
		}
		if !ok {
			return nil, fmt.Errorf("Parse error: Invalid SQL")
		}
		children = append(children, node)
	}
	return children, nil
}

func (p *Parser) parseBetweenPredicate() (*BetweenPredicateNode, error) {
	betweenTok := p.consume()
	leading := p.parseComments()
	expr1, err := p.parseAndlessExpressionChain()
	if err != nil {
		return nil, err
	}
	trail := p.parseComments()
	andTok := p.expect(TokenAnd)
	if andTok.Type == "" {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	leading2 := p.parseComments()
	expr2, ok, err := p.parseAndlessExpression()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	betweenKw := KeywordNode{Type: NodeKeyword, TokenType: betweenTok.Type, Text: betweenTok.Text, Raw: betweenTok.Raw}
	betweenKw = addTrailingCommentsKeyword(betweenKw, leading)
	expr1 = addCommentsToArray(expr1, leading, trail)
	andKw := KeywordNode{Type: NodeKeyword, TokenType: andTok.Type, Text: andTok.Text, Raw: andTok.Raw}
	expr2Node := addLeadingComments(expr2, leading2)
	return &BetweenPredicateNode{Type: NodeBetweenPredicate, BetweenKw: betweenKw, Expr1: expr1, AndKw: andKw, Expr2: []AstNode{expr2Node}}, nil
}

func (p *Parser) parseAndlessExpressionChain() ([]AstNode, error) {
	items := []AstNode{}
	first, ok, err := p.parseAndlessExpression()
	if err != nil {
		return nil, err
	}
	if !ok {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	items = append(items, first)
	for {
		next, commentCount := p.peekAfterComments()
		if !p.canStartAndlessExpression(next) {
			break
		}
		leading := p.consumeComments(commentCount)
		expr, ok, err := p.parseAndlessExpression()
		if err != nil {
			return nil, err
		}
		if !ok {
			break
		}
		expr = addLeadingComments(expr, leading)
		items = append(items, expr)
	}
	return items, nil
}

func (p *Parser) parseCaseExpression() (*CaseExpressionNode, error) {
	caseTok := p.consume()
	trailing := p.parseComments()
	caseKw := KeywordNode{Type: NodeKeyword, TokenType: caseTok.Type, Text: caseTok.Text, Raw: caseTok.Raw}
	caseKw = addTrailingCommentsKeyword(caseKw, trailing)

	expr := []AstNode{}
	if p.canStartExpression(p.peek()) {
		expList, err := p.parseExpressionChainTrailing()
		if err != nil {
			return nil, err
		}
		expr = expList
	}

	clauses := []AstNode{}
	for {
		if p.peek().Type == TokenWhen {
			clause, err := p.parseCaseWhen()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
			continue
		}
		if p.peek().Type == TokenElse {
			clause, err := p.parseCaseElse()
			if err != nil {
				return nil, err
			}
			clauses = append(clauses, clause)
			continue
		}
		break
	}
	endTok := p.expect(TokenEnd)
	if endTok.Type == "" {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	endKw := KeywordNode{Type: NodeKeyword, TokenType: endTok.Type, Text: endTok.Text, Raw: endTok.Raw}
	return &CaseExpressionNode{Type: NodeCaseExpression, CaseKw: caseKw, EndKw: endKw, Expr: expr, Clauses: clauses}, nil
}

func (p *Parser) parseCaseWhen() (*CaseWhenNode, error) {
	whenTok := p.consume()
	trailing := p.parseComments()
	cond, err := p.parseExpressionChainTrailing()
	if err != nil {
		return nil, err
	}
	thenTok := p.expect(TokenThen)
	if thenTok.Type == "" {
		return nil, fmt.Errorf("Parse error: Invalid SQL")
	}
	thenTrailing := p.parseComments()
	result, err := p.parseExpressionChainTrailing()
	if err != nil {
		return nil, err
	}
	whenKw := KeywordNode{Type: NodeKeyword, TokenType: whenTok.Type, Text: whenTok.Text, Raw: whenTok.Raw}
	whenKw = addTrailingCommentsKeyword(whenKw, trailing)
	thenKw := KeywordNode{Type: NodeKeyword, TokenType: thenTok.Type, Text: thenTok.Text, Raw: thenTok.Raw}
	thenKw = addTrailingCommentsKeyword(thenKw, thenTrailing)
	return &CaseWhenNode{Type: NodeCaseWhen, WhenKw: whenKw, ThenKw: thenKw, Condition: cond, Result: result}, nil
}

func (p *Parser) parseCaseElse() (*CaseElseNode, error) {
	elseTok := p.consume()
	trailing := p.parseComments()
	result, err := p.parseExpressionChainTrailing()
	if err != nil {
		return nil, err
	}
	elseKw := KeywordNode{Type: NodeKeyword, TokenType: elseTok.Type, Text: elseTok.Text, Raw: elseTok.Raw}
	elseKw = addTrailingCommentsKeyword(elseKw, trailing)
	return &CaseElseNode{Type: NodeCaseElse, ElseKw: elseKw, Result: result}, nil
}

func (p *Parser) parseCommentNode() AstNode {
	tok := p.consume()
	switch tok.Type {
	case TokenLineComment:
		return &LineCommentNode{Type: NodeLineComment, Text: tok.Text, PrecedingWhitespace: tok.PrecedingWhitespace}
	case TokenBlockComment:
		return &BlockCommentNode{Type: NodeBlockComment, Text: tok.Text, PrecedingWhitespace: tok.PrecedingWhitespace}
	case TokenDisableComment:
		return &DisableCommentNode{Type: NodeDisableComment, Text: tok.Text, PrecedingWhitespace: tok.PrecedingWhitespace}
	default:
		return &BlockCommentNode{Type: NodeBlockComment, Text: tok.Text, PrecedingWhitespace: tok.PrecedingWhitespace}
	}
}

func (p *Parser) parseComments() []CommentNode {
	comments := []CommentNode{}
	for p.isCommentToken(p.peek()) {
		comments = append(comments, p.parseCommentNode())
	}
	return comments
}

func (p *Parser) peekAfterComments() (Token, int) {
	idx := p.index
	count := 0
	for idx < len(p.tokens) && p.isCommentToken(p.tokens[idx]) {
		idx++
		count++
	}
	if idx >= len(p.tokens) {
		return Token{}, count
	}
	return p.tokens[idx], count
}

func (p *Parser) consumeComments(count int) []CommentNode {
	if count <= 0 {
		return nil
	}
	comments := make([]CommentNode, 0, count)
	for i := 0; i < count; i++ {
		if !p.isCommentToken(p.peek()) {
			break
		}
		comments = append(comments, p.parseCommentNode())
	}
	return comments
}

func (p *Parser) isCommentToken(tok Token) bool {
	return tok.Type == TokenLineComment || tok.Type == TokenBlockComment || tok.Type == TokenDisableComment
}

func (p *Parser) canStartExpression(tok Token) bool {
	switch tok.Type {
	case TokenAnd, TokenOr, TokenXor,
		TokenAsterisk,
		TokenArrayIdentifier, TokenArrayKeyword,
		TokenReservedFunctionName, TokenReservedParameterizedDataType,
		TokenOpenParen,
		TokenOperator,
		TokenIdentifier, TokenQuotedIdentifier, TokenVariable,
		TokenNamedParameter, TokenQuotedParameter, TokenNumberedParameter, TokenPositionalParameter, TokenCustomParameter,
		TokenNumber, TokenString,
		TokenReservedDataType, TokenReservedDataTypePhrase,
		TokenReservedKeyword, TokenReservedKeywordPhrase, TokenReservedJoin,
		TokenBetween, TokenCase,
		TokenWhen, TokenThen, TokenElse, TokenEnd,
		TokenComma,
		TokenLineComment, TokenBlockComment, TokenDisableComment:
		return true
	default:
		return false
	}
}

func (p *Parser) canStartAndlessExpression(tok Token) bool {
	switch tok.Type {
	case TokenAsterisk,
		TokenArrayIdentifier, TokenArrayKeyword,
		TokenReservedFunctionName, TokenReservedParameterizedDataType,
		TokenOpenParen,
		TokenOperator,
		TokenIdentifier, TokenQuotedIdentifier, TokenVariable,
		TokenNamedParameter, TokenQuotedParameter, TokenNumberedParameter, TokenPositionalParameter, TokenCustomParameter,
		TokenNumber, TokenString,
		TokenReservedDataType, TokenReservedDataTypePhrase,
		TokenReservedKeyword, TokenReservedKeywordPhrase, TokenReservedJoin,
		TokenBetween, TokenCase,
		TokenWhen, TokenThen, TokenElse, TokenEnd,
		TokenComma,
		TokenLineComment, TokenBlockComment, TokenDisableComment:
		return true
	default:
		return false
	}
}

func (p *Parser) isClauseStart(tok Token) bool {
	return tok.Type == TokenLimit || tok.Type == TokenReservedSelect || tok.Type == TokenReservedClause || tok.Type == TokenReservedSetOperation
}

func (p *Parser) isStop(stopTypes ...TokenType) bool {
	if len(stopTypes) == 0 {
		return false
	}
	for _, t := range stopTypes {
		if p.peek().Type == t {
			return true
		}
	}
	return false
}

func (p *Parser) peek() Token {
	if p.index >= len(p.tokens) {
		return Token{}
	}
	return p.tokens[p.index]
}

func (p *Parser) consume() Token {
	if p.index >= len(p.tokens) {
		return Token{}
	}
	ok := p.tokens[p.index]
	p.index++
	return ok
}

func (p *Parser) expect(t TokenType) Token {
	if p.peek().Type != t {
		return Token{}
	}
	return p.consume()
}

func (p *Parser) isParameterToken(tok Token) bool {
	switch tok.Type {
	case TokenNamedParameter, TokenQuotedParameter, TokenNumberedParameter, TokenPositionalParameter, TokenCustomParameter:
		return true
	default:
		return false
	}
}

// comment attachment helpers

func addLeadingComments(node AstNode, comments []CommentNode) AstNode {
	if len(comments) == 0 {
		return node
	}
	switch n := node.(type) {
	case *ClauseNode:
		n.LeadingComments = comments
	case *SetOperationNode:
		n.LeadingComments = comments
	case *FunctionCallNode:
		n.LeadingComments = comments
	case *ParameterizedDataTypeNode:
		n.LeadingComments = comments
	case *ArraySubscriptNode:
		n.LeadingComments = comments
	case *PropertyAccessNode:
		n.LeadingComments = comments
	case *ParenthesisNode:
		n.LeadingComments = comments
	case *BetweenPredicateNode:
		n.LeadingComments = comments
	case *CaseExpressionNode:
		n.LeadingComments = comments
	case *CaseWhenNode:
		n.LeadingComments = comments
	case *CaseElseNode:
		n.LeadingComments = comments
	case *LimitClauseNode:
		n.LeadingComments = comments
	case *AllColumnsAsteriskNode:
		n.LeadingComments = comments
	case *LiteralNode:
		n.LeadingComments = comments
	case *IdentifierNode:
		n.LeadingComments = comments
	case *KeywordNode:
		n.LeadingComments = comments
	case *DataTypeNode:
		n.LeadingComments = comments
	case *ParameterNode:
		n.LeadingComments = comments
	case *OperatorNode:
		n.LeadingComments = comments
	case *CommaNode:
		n.LeadingComments = comments
	case *LineCommentNode:
		n.LeadingComments = comments
	case *BlockCommentNode:
		n.LeadingComments = comments
	case *DisableCommentNode:
		n.LeadingComments = comments
	}
	return node
}

func addTrailingComments(node AstNode, comments []CommentNode) AstNode {
	if len(comments) == 0 {
		return node
	}
	switch n := node.(type) {
	case *ClauseNode:
		n.TrailingComments = comments
	case *SetOperationNode:
		n.TrailingComments = comments
	case *FunctionCallNode:
		n.TrailingComments = comments
	case *ParameterizedDataTypeNode:
		n.TrailingComments = comments
	case *ArraySubscriptNode:
		n.TrailingComments = comments
	case *PropertyAccessNode:
		n.TrailingComments = comments
	case *ParenthesisNode:
		n.TrailingComments = comments
	case *BetweenPredicateNode:
		n.TrailingComments = comments
	case *CaseExpressionNode:
		n.TrailingComments = comments
	case *CaseWhenNode:
		n.TrailingComments = comments
	case *CaseElseNode:
		n.TrailingComments = comments
	case *LimitClauseNode:
		n.TrailingComments = comments
	case *AllColumnsAsteriskNode:
		n.TrailingComments = comments
	case *LiteralNode:
		n.TrailingComments = comments
	case *IdentifierNode:
		n.TrailingComments = comments
	case *KeywordNode:
		n.TrailingComments = comments
	case *DataTypeNode:
		n.TrailingComments = comments
	case *ParameterNode:
		n.TrailingComments = comments
	case *OperatorNode:
		n.TrailingComments = comments
	case *CommaNode:
		n.TrailingComments = comments
	case *LineCommentNode:
		n.TrailingComments = comments
	case *BlockCommentNode:
		n.TrailingComments = comments
	case *DisableCommentNode:
		n.TrailingComments = comments
	}
	return node
}

func addTrailingCommentsKeyword(node KeywordNode, comments []CommentNode) KeywordNode {
	if len(comments) == 0 {
		return node
	}
	node.TrailingComments = comments
	return node
}

func addTrailingCommentsDataType(node DataTypeNode, comments []CommentNode) DataTypeNode {
	if len(comments) == 0 {
		return node
	}
	node.TrailingComments = comments
	return node
}

func addCommentsToArray(nodes []AstNode, leading []CommentNode, trailing []CommentNode) []AstNode {
	if len(nodes) == 0 {
		return nodes
	}
	if len(leading) > 0 {
		nodes[0] = addLeadingComments(nodes[0], leading)
	}
	if len(trailing) > 0 {
		nodes[len(nodes)-1] = addTrailingComments(nodes[len(nodes)-1], trailing)
	}
	return nodes
}

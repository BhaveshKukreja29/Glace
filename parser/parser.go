package parser

import (
	"fmt"

	"github.com/glace-lang/glace/ast"
	"github.com/glace-lang/glace/lexer"
)

// Parser converts a slice of tokens into an AST.
type Parser struct {
	tokens  []lexer.Token
	current int
	errors  []string
}

// New creates a new Parser for the given token slice.
func New(tokens []lexer.Token) *Parser {
	return &Parser{
		tokens: tokens,
		errors: make([]string, 0),
	}
}

// Parse parses the token stream and returns the root Program node.
//
// TODO: Implement recursive descent parsing.
// The general approach:
//   1. Loop calling parseStatement() until TOKEN_EOF
//   2. Skip TOKEN_NEWLINE between statements
//   3. Collect parse errors, attempt to synchronize on errors
func Parse(tokens []lexer.Token) (*ast.Program, []string) {
	p := New(tokens)
	program := p.parseProgram()
	return program, p.errors
}

func (p *Parser) parseProgram() *ast.Program {
	program := &ast.Program{
		Statements: make([]ast.Statement, 0),
	}

	for !p.isAtEnd() {
		// Skip newlines between statements
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}

		stmt := p.parseStatement()
		if stmt != nil {
			program.Statements = append(program.Statements, stmt)
		}
	}

	return program
}

// ---------------------------------------------------------------------------
// Statement Parsing
// ---------------------------------------------------------------------------

func (p *Parser) parseStatement() ast.Statement {
	switch p.peek().Type {
	case lexer.TOKEN_LET:
		return p.parseLetStatement()
	case lexer.TOKEN_MUT:
		return p.parseMutStatement()
	case lexer.TOKEN_FN:
		// Check if it's a declaration (fn name(...)) vs expression (fn(...))
		if p.peekNext().Type == lexer.TOKEN_IDENT {
			return p.parseFnDeclaration()
		}
		return p.parseExpressionStatement()
	case lexer.TOKEN_RETURN:
		return p.parseReturnStatement()
	case lexer.TOKEN_IF:
		return p.parseIfStatement()
	case lexer.TOKEN_LOOP:
		return p.parseLoopStatement()
	// --- GLACE-003: Parse break statement ---
	case lexer.TOKEN_BREAK:
		return p.parseBreakStatement()
	// --- GLACE-003: Parse continue statement ---
	case lexer.TOKEN_CONTINUE:
		return p.parseContinueStatement()
	case lexer.TOKEN_MATCH:
		return p.parseMatchStatement()
	case lexer.TOKEN_TEST:
		return p.parseTestBlock()
	default:
		return p.parseExpressionOrAssignment()
	}
}

// --- Statement stubs — IMPLEMENT EACH OF THESE ---

func (p *Parser) parseLetStatement() ast.Statement {
	pos := p.advance().Pos
	if !p.expect(lexer.TOKEN_IDENT) {
		return nil
	}
	name := p.tokens[p.current-1].Literal
	if !p.expect(lexer.TOKEN_ASSIGN) {
		return nil
	}
	value := p.parseExpression(PREC_LOWEST)
	if value == nil {
		return nil
	}
	return &ast.LetStatement{Pos: pos, Name: name, Value: value}
}

func (p *Parser) parseMutStatement() ast.Statement {
	pos := p.advance().Pos
	if !p.expect(lexer.TOKEN_IDENT) {
		return nil
	}
	name := p.tokens[p.current-1].Literal
	if !p.expect(lexer.TOKEN_ASSIGN) {
		return nil
	}
	value := p.parseExpression(PREC_LOWEST)
	if value == nil {
		return nil
	}
	return &ast.MutStatement{Pos: pos, Name: name, Value: value}
}

func (p *Parser) parseFnDeclaration() ast.Statement {
	pos := p.advance().Pos
	if !p.expect(lexer.TOKEN_IDENT) {
		return nil
	}
	name := p.tokens[p.current-1].Literal
	if !p.expect(lexer.TOKEN_LPAREN) {
		return nil
	}
	params := []string{}
	for p.peek().Type != lexer.TOKEN_RPAREN && !p.isAtEnd() {
		if !p.expect(lexer.TOKEN_IDENT) {
			return nil
		}
		params = append(params, p.tokens[p.current-1].Literal)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	if !p.expect(lexer.TOKEN_RPAREN) {
		return nil
	}
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &ast.FnDeclaration{Pos: pos, Name: name, Params: params, Body: body}
}

func (p *Parser) parseReturnStatement() ast.Statement {
	pos := p.advance().Pos
	values := []ast.Expression{}
	if p.peek().Type != lexer.TOKEN_NEWLINE && p.peek().Type != lexer.TOKEN_EOF && p.peek().Type != lexer.TOKEN_RBRACE {
		val := p.parseExpression(PREC_LOWEST)
		if val != nil {
			values = append(values, val)
			for p.peek().Type == lexer.TOKEN_COMMA {
				p.advance()
				val := p.parseExpression(PREC_LOWEST)
				if val != nil {
					values = append(values, val)
				}
			}
		}
	}
	return &ast.ReturnStatement{Pos: pos, Values: values}
}

func (p *Parser) parseIfStatement() ast.Statement {
	pos := p.advance().Pos
	condition := p.parseExpression(PREC_LOWEST)
	if condition == nil {
		return nil
	}
	consequence := p.parseBlock()
	if consequence == nil {
		return nil
	}
	var elifClauses []ast.ElifClause
	for p.peek().Type == lexer.TOKEN_ELIF {
		p.advance()
		elifCond := p.parseExpression(PREC_LOWEST)
		if elifCond == nil {
			return nil
		}
		elifBody := p.parseBlock()
		if elifBody == nil {
			return nil
		}
		elifClauses = append(elifClauses, ast.ElifClause{
			Condition:   elifCond,
			Consequence: elifBody,
		})
	}
	var alternative *ast.BlockStatement
	if p.peek().Type == lexer.TOKEN_ELSE {
		p.advance()
		alternative = p.parseBlock()
		if alternative == nil {
			return nil
		}
	}
	return &ast.IfStatement{
		Pos:         pos,
		Condition:   condition,
		Consequence: consequence,
		ElifClauses: elifClauses,
		Alternative: alternative,
	}
}

func (p *Parser) parseLoopStatement() ast.Statement {
	pos := p.advance().Pos
	var condition ast.Expression
	var iterator string
	var iterable ast.Expression
	if p.peek().Type == lexer.TOKEN_IDENT && p.peekNext().Type == lexer.TOKEN_IN {
		iterator = p.advance().Literal
		p.advance()
		iterable = p.parseExpression(PREC_LOWEST)
		if iterable == nil {
			return nil
		}
	} else if p.peek().Type != lexer.TOKEN_LBRACE {
		condition = p.parseExpression(PREC_LOWEST)
	}
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &ast.LoopStatement{
		Pos:       pos,
		Condition: condition,
		Iterator:  iterator,
		Iterable:  iterable,
		Body:      body,
	}
}

// --- GLACE-003: Parse break statement ---
func (p *Parser) parseBreakStatement() ast.Statement {
	pos := p.advance().Pos
	return &ast.BreakStatement{Pos: pos}
}

// --- GLACE-003: Parse continue statement ---
func (p *Parser) parseContinueStatement() ast.Statement {
	pos := p.advance().Pos
	return &ast.ContinueStatement{Pos: pos}
}

func (p *Parser) parseMatchStatement() ast.Statement {
	pos := p.advance().Pos
	subject := p.parseExpression(PREC_LOWEST)
	if subject == nil {
		return nil
	}
	if !p.expect(lexer.TOKEN_LBRACE) {
		return nil
	}
	var arms []ast.MatchArm
	p.skipNewlines()
	for p.peek().Type != lexer.TOKEN_RBRACE && !p.isAtEnd() {
		pattern := p.parseExpression(PREC_LOWEST)
		if pattern == nil {
			return nil
		}
		var guard ast.Expression
		if p.peek().Type == lexer.TOKEN_IF {
			p.advance()
			guard = p.parseExpression(PREC_LOWEST)
			if guard == nil {
				return nil
			}
		}
		if !p.expect(lexer.TOKEN_ARROW) {
			return nil
		}
		bodyExpr := p.parseExpression(PREC_LOWEST)
		if bodyExpr == nil {
			return nil
		}
		body := &ast.BlockStatement{
			Pos: pos,
			Statements: []ast.Statement{
				&ast.ExpressionStatement{
					Pos:        pos,
					Expression: bodyExpr,
				},
			},
		}
		arms = append(arms, ast.MatchArm{
			Pattern: pattern,
			Guard:   guard,
			Body:    body,
		})
		p.skipNewlines()
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.MatchStatement{
		Pos:     pos,
		Subject: subject,
		Arms:    arms,
	}
}

func (p *Parser) parseTestBlock() ast.Statement {
	pos := p.advance().Pos
	if !p.expect(lexer.TOKEN_STRING) {
		return nil
	}
	description := p.tokens[p.current-1].Literal
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &ast.TestBlock{
		Pos:         pos,
		Description: description,
		Body:        body,
	}
}

func (p *Parser) parseExpressionStatement() ast.Statement {
	pos := p.peek().Pos
	expr := p.parseExpression(PREC_LOWEST)
	if expr == nil {
		return nil
	}
	return &ast.ExpressionStatement{Pos: pos, Expression: expr}
}

func (p *Parser) parseExpressionOrAssignment() ast.Statement {
	// Parse expression first; if followed by '=', convert to assignment
	pos := p.peek().Pos
	expr := p.parseExpression(PREC_LOWEST)
	if expr == nil {
		return nil
	}

	// Check for assignment: ident = expr  or  expr[index] = expr
	if p.peek().Type == lexer.TOKEN_ASSIGN {
		p.advance() // consume '='
		value := p.parseExpression(PREC_LOWEST)

		switch target := expr.(type) {
		case *ast.Identifier:
			return &ast.AssignStatement{Pos: pos, Name: target.Name, Value: value}
		case *ast.IndexExpression:
			return &ast.IndexAssignStatement{Pos: pos, Left: target.Left, Index: target.Index, Value: value}
		default:
			p.addError("invalid assignment target")
			return nil
		}
	}

	return &ast.ExpressionStatement{Pos: pos, Expression: expr}
}

func (p *Parser) parseBlock() *ast.BlockStatement {
	// TODO: expect '{', parse statements until '}', return BlockStatement
	pos := p.peek().Pos
	if !p.expect(lexer.TOKEN_LBRACE) {
		return nil
	}
	block := &ast.BlockStatement{Pos: pos, Statements: make([]ast.Statement, 0)}

	p.skipNewlines()
	for !p.isAtEnd() && p.peek().Type != lexer.TOKEN_RBRACE {
		stmt := p.parseStatement()
		if stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.skipNewlines()
	}

	p.expect(lexer.TOKEN_RBRACE)
	return block
}

// ---------------------------------------------------------------------------
// Expression Parsing (Pratt Parser)
// ---------------------------------------------------------------------------

// parseExpression is the Pratt parser entry point.
func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	token := p.peek()
	prefix := p.getPrefixParser(token.Type)
	if prefix == nil {
		p.addError(fmt.Sprintf("no prefix parser for %v", token.Type))
		return nil
	}
	left := prefix()
	if left == nil {
		return nil
	}
	for precedence < p.getPrecedence(p.peek().Type) {
		infix := p.getInfixParser(p.peek().Type)
		if infix == nil {
			break
		}
		left = infix(left)
		if left == nil {
			return nil
		}
	}
	return left
}

type prefixParseFn func() ast.Expression
type infixParseFn func(ast.Expression) ast.Expression

func (p *Parser) getPrefixParser(tokenType lexer.TokenType) prefixParseFn {
	switch tokenType {
	case lexer.TOKEN_INT:
		return p.parseIntegerLiteral
	case lexer.TOKEN_FLOAT:
		return p.parseFloatLiteral
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral
	case lexer.TOKEN_TRUE, lexer.TOKEN_FALSE:
		return p.parseBooleanLiteral
	case lexer.TOKEN_NONE:
		return p.parseNoneLiteral
	case lexer.TOKEN_IDENT:
		return p.parseIdentifier
	case lexer.TOKEN_LPAREN:
		return p.parseGrouping
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral
	case lexer.TOKEN_LBRACE:
		return p.parseMapLiteral
	case lexer.TOKEN_FN:
		return p.parseFnLiteral
	case lexer.TOKEN_NOT, lexer.TOKEN_MINUS:
		return p.parseUnaryExpression
	default:
		return nil
	}
}

func (p *Parser) getInfixParser(tokenType lexer.TokenType) infixParseFn {
	switch tokenType {
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS, lexer.TOKEN_STAR, lexer.TOKEN_SLASH,
		lexer.TOKEN_PERCENT, lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE,
		lexer.TOKEN_GTE, lexer.TOKEN_EQ, lexer.TOKEN_NEQ, lexer.TOKEN_AND,
		lexer.TOKEN_OR:
		return p.parseBinaryExpression
	case lexer.TOKEN_LPAREN:
		return p.parseCallExpression
	case lexer.TOKEN_LBRACKET:
		return p.parseIndexExpression
	case lexer.TOKEN_DOT:
		return p.parseDotExpression
	case lexer.TOKEN_QMARK:
		return p.parseSafeAccessExpression
	case lexer.TOKEN_PIPE:
		return p.parsePipelineExpression
	case lexer.TOKEN_DOTDOT:
		return p.parseRangeExpression
	case lexer.TOKEN_COALESCE:
		return p.parseCoalesceExpression
	default:
		return nil
	}
}

func (p *Parser) getPrecedence(tokenType lexer.TokenType) Precedence {
	switch tokenType {
	case lexer.TOKEN_PIPE:
		return PREC_PIPELINE
	case lexer.TOKEN_COALESCE:
		return PREC_COALESCE
	case lexer.TOKEN_OR:
		return PREC_OR
	case lexer.TOKEN_AND:
		return PREC_AND
	case lexer.TOKEN_EQ, lexer.TOKEN_NEQ:
		return PREC_EQUALITY
	case lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE:
		return PREC_COMPARISON
	case lexer.TOKEN_DOTDOT:
		return PREC_RANGE
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS:
		return PREC_ADDITION
	case lexer.TOKEN_STAR, lexer.TOKEN_SLASH, lexer.TOKEN_PERCENT:
		return PREC_MULTIPLY
	case lexer.TOKEN_LPAREN, lexer.TOKEN_LBRACKET, lexer.TOKEN_DOT, lexer.TOKEN_QMARK:
		return PREC_CALL
	default:
		return PREC_LOWEST
	}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	tok := p.advance()
	var value int64
	fmt.Sscanf(tok.Literal, "%d", &value)
	return &ast.IntegerLiteral{Pos: tok.Pos, Value: value}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	tok := p.advance()
	var value float64
	fmt.Sscanf(tok.Literal, "%f", &value)
	return &ast.FloatLiteral{Pos: tok.Pos, Value: value}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	tok := p.advance()
	return &ast.StringLiteral{Pos: tok.Pos, Value: tok.Literal}
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	tok := p.advance()
	return &ast.BooleanLiteral{Pos: tok.Pos, Value: tok.Type == lexer.TOKEN_TRUE}
}

func (p *Parser) parseNoneLiteral() ast.Expression {
	tok := p.advance()
	return &ast.NoneLiteral{Pos: tok.Pos}
}

func (p *Parser) parseIdentifier() ast.Expression {
	tok := p.advance()
	return &ast.Identifier{Pos: tok.Pos, Name: tok.Literal}
}

func (p *Parser) parseGrouping() ast.Expression {
	p.advance()
	expr := p.parseExpression(PREC_LOWEST)
	p.expect(lexer.TOKEN_RPAREN)
	return expr
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	pos := p.advance().Pos
	elements := []ast.Expression{}
	for p.peek().Type != lexer.TOKEN_RBRACKET && !p.isAtEnd() {
		elem := p.parseExpression(PREC_LOWEST)
		if elem != nil {
			elements = append(elements, elem)
		}
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.ArrayLiteral{Pos: pos, Elements: elements}
}

func (p *Parser) parseMapLiteral() ast.Expression {
	pos := p.advance().Pos
	keys := []ast.Expression{}
	values := []ast.Expression{}
	for p.peek().Type != lexer.TOKEN_RBRACE && !p.isAtEnd() {
		key := p.parseExpression(PREC_LOWEST)
		if key == nil {
			return nil
		}
		keys = append(keys, key)
		if !p.expect(lexer.TOKEN_COLON) {
			return nil
		}
		value := p.parseExpression(PREC_LOWEST)
		if value == nil {
			return nil
		}
		values = append(values, value)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RBRACE)
	return &ast.MapLiteral{Pos: pos, Keys: keys, Values: values}
}

func (p *Parser) parseFnLiteral() ast.Expression {
	pos := p.advance().Pos
	p.expect(lexer.TOKEN_LPAREN)
	params := []string{}
	for p.peek().Type != lexer.TOKEN_RPAREN && !p.isAtEnd() {
		if !p.expect(lexer.TOKEN_IDENT) {
			return nil
		}
		params = append(params, p.tokens[p.current-1].Literal)
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	body := p.parseBlock()
	if body == nil {
		return nil
	}
	return &ast.FnLiteral{Pos: pos, Params: params, Body: body}
}

func (p *Parser) parseUnaryExpression() ast.Expression {
	tok := p.advance()
	operand := p.parseExpression(PREC_UNARY)
	if operand == nil {
		return nil
	}
	return &ast.UnaryExpression{Pos: tok.Pos, Operator: tok.Literal, Operand: operand}
}

func (p *Parser) parseBinaryExpression(left ast.Expression) ast.Expression {
	tok := p.advance()
	precedence := p.getPrecedence(tok.Type)
	right := p.parseExpression(precedence)
	if right == nil {
		return nil
	}
	return &ast.BinaryExpression{Pos: tok.Pos, Left: left, Operator: tok.Literal, Right: right}
}

func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	args := []ast.Expression{}
	for p.peek().Type != lexer.TOKEN_RPAREN && !p.isAtEnd() {
		arg := p.parseExpression(PREC_LOWEST)
		if arg != nil {
			args = append(args, arg)
		}
		if p.peek().Type == lexer.TOKEN_COMMA {
			p.advance()
		}
	}
	p.expect(lexer.TOKEN_RPAREN)
	return &ast.CallExpression{Pos: pos, Function: left, Arguments: args}
}

func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	index := p.parseExpression(PREC_LOWEST)
	if index == nil {
		return nil
	}
	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.IndexExpression{Pos: pos, Left: left, Index: index}
}

func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	if !p.expect(lexer.TOKEN_IDENT) {
		return nil
	}
	field := p.tokens[p.current-1].Literal
	return &ast.DotExpression{Pos: pos, Left: left, Field: field}
}

func (p *Parser) parseSafeAccessExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	if !p.expect(lexer.TOKEN_DOT) {
		return nil
	}
	if !p.expect(lexer.TOKEN_IDENT) {
		return nil
	}
	field := p.tokens[p.current-1].Literal
	return &ast.SafeAccessExpression{Pos: pos, Left: left, Field: field}
}

func (p *Parser) parsePipelineExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	right := p.parseExpression(PREC_PIPELINE)
	if right == nil {
		return nil
	}
	if call, ok := right.(*ast.CallExpression); ok {
		return &ast.PipelineExpression{Pos: pos, Left: left, Right: call}
	}
	p.addError("pipeline right-hand side must be a function call")
	return nil
}

func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	end := p.parseExpression(PREC_RANGE)
	if end == nil {
		return nil
	}
	var step ast.Expression
	if p.peek().Type == lexer.TOKEN_COLON {
		p.advance()
		step = p.parseExpression(PREC_RANGE)
		if step == nil {
			return nil
		}
	}
	return &ast.RangeExpression{Pos: pos, Start: left, End: end, Step: step}
}

func (p *Parser) parseCoalesceExpression(left ast.Expression) ast.Expression {
	pos := p.advance().Pos
	right := p.parseExpression(PREC_COALESCE)
	if right == nil {
		return nil
	}
	return &ast.CoalesceExpression{Pos: pos, Left: left, Right: right}
}

// ---------------------------------------------------------------------------
// Helper Methods
// ---------------------------------------------------------------------------

// peek returns the current token without advancing.
func (p *Parser) peek() lexer.Token {
	if p.current >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.current]
}

// peekNext returns the token after the current one.
func (p *Parser) peekNext() lexer.Token {
	if p.current+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.current+1]
}

// advance consumes and returns the current token.
func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	p.current++
	return tok
}

// expect consumes the current token if it matches the expected type.
// Returns true if matched, false otherwise (and adds an error).
func (p *Parser) expect(expected lexer.TokenType) bool {
	if p.peek().Type == expected {
		p.advance()
		return true
	}
	p.addError(fmt.Sprintf("expected %v, got %v at %s",
		expected, p.peek().Type, p.peek().Pos))
	return false
}

// isAtEnd returns true if we've reached EOF.
func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.TOKEN_EOF
}

// skipNewlines advances past any TOKEN_NEWLINE tokens.
func (p *Parser) skipNewlines() {
	for p.peek().Type == lexer.TOKEN_NEWLINE {
		p.advance()
	}
}

// addError records a parse error.
func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

// Errors returns the accumulated parse errors.
func (p *Parser) Errors() []string {
	return p.errors
}

// synchronize advances to the next statement boundary for error recovery.
func (p *Parser) synchronize() {
	for !p.isAtEnd() {
		if p.peek().Type == lexer.TOKEN_NEWLINE {
			p.advance()
			return
		}
		switch p.peek().Type {
		case lexer.TOKEN_LET, lexer.TOKEN_MUT, lexer.TOKEN_FN,
			lexer.TOKEN_RETURN, lexer.TOKEN_IF, lexer.TOKEN_LOOP,
			lexer.TOKEN_MATCH, lexer.TOKEN_TEST:
			return
		}
		p.advance()
	}
}

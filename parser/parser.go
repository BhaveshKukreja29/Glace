package parser

import (
	"fmt"
	"strconv"

	"github.com/glace-lang/glace/ast"
	"github.com/glace-lang/glace/lexer"
)

// Parser converts a token stream into an AST via recursive descent
// with Pratt parsing for expressions.
type Parser struct {
	tokens  []lexer.Token
	current int
	errors  []string
}

func New(tokens []lexer.Token) *Parser {
	return &Parser{tokens: tokens, errors: make([]string, 0)}
}

// Parse is the top-level entry point.
func Parse(tokens []lexer.Token) (*ast.Program, []string) {
	p := New(tokens)
	prog := p.parseProgram()
	return prog, p.errors
}

func (p *Parser) parseProgram() *ast.Program {
	prog := &ast.Program{Statements: make([]ast.Statement, 0)}
	for !p.isAtEnd() {
		p.skipNewlines()
		if p.isAtEnd() {
			break
		}
		if stmt := p.parseStatement(); stmt != nil {
			prog.Statements = append(prog.Statements, stmt)
		}
	}
	return prog
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
	case lexer.TOKEN_BREAK:
		return p.parseBreakStatement()
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

// let <ident> = <expr>
func (p *Parser) parseLetStatement() ast.Statement {
	pos := p.advance().Pos // consume 'let'
	name := p.advance()    // consume identifier
	if name.Type != lexer.TOKEN_IDENT {
		p.addError(fmt.Sprintf("expected identifier after 'let', got %q at %s", name.Literal, name.Pos))
		p.synchronize()
		return nil
	}
	if !p.expect(lexer.TOKEN_ASSIGN) {
		p.synchronize()
		return nil
	}
	value := p.parseExpression(PREC_LOWEST)
	return &ast.LetStatement{Pos: pos, Name: name.Literal, Value: value}
}

// mut <ident> = <expr>
func (p *Parser) parseMutStatement() ast.Statement {
	pos := p.advance().Pos
	name := p.advance()
	if name.Type != lexer.TOKEN_IDENT {
		p.addError(fmt.Sprintf("expected identifier after 'mut', got %q at %s", name.Literal, name.Pos))
		p.synchronize()
		return nil
	}
	if !p.expect(lexer.TOKEN_ASSIGN) {
		p.synchronize()
		return nil
	}
	value := p.parseExpression(PREC_LOWEST)
	return &ast.MutStatement{Pos: pos, Name: name.Literal, Value: value}
}

// fn <name>(<params>) <block>  |  fn <name>(<params>) => <expr>
func (p *Parser) parseFnDeclaration() ast.Statement {
	pos := p.advance().Pos // consume 'fn'
	name := p.advance()    // consume name
	params := p.parseParams()

	// Arrow form: fn name(params) => expr
	if p.peek().Type == lexer.TOKEN_ARROW {
		p.advance() // consume '=>'
		expr := p.parseExpression(PREC_LOWEST)
		body := &ast.BlockStatement{
			Pos: expr.TokenPos(),
			Statements: []ast.Statement{
				&ast.ReturnStatement{Pos: expr.TokenPos(), Values: []ast.Expression{expr}},
			},
		}
		return &ast.FnDeclaration{Pos: pos, Name: name.Literal, Params: params, Body: body}
	}

	body := p.parseBlock()
	return &ast.FnDeclaration{Pos: pos, Name: name.Literal, Params: params, Body: body}
}

// return [<expr> {, <expr>}]
func (p *Parser) parseReturnStatement() ast.Statement {
	pos := p.advance().Pos // consume 'return'
	values := make([]ast.Expression, 0)

	// Return with no value if next token is newline, '}', or EOF.
	if p.isStatementEnd() {
		return &ast.ReturnStatement{Pos: pos, Values: values}
	}

	values = append(values, p.parseExpression(PREC_LOWEST))
	for p.peek().Type == lexer.TOKEN_COMMA {
		p.advance() // consume ','
		values = append(values, p.parseExpression(PREC_LOWEST))
	}
	return &ast.ReturnStatement{Pos: pos, Values: values}
}

// if <cond> <block> {elif <cond> <block>} [else <block>]
func (p *Parser) parseIfStatement() ast.Statement {
	pos := p.advance().Pos // consume 'if'
	cond := p.parseExpression(PREC_LOWEST)
	consequence := p.parseBlock()

	stmt := &ast.IfStatement{
		Pos:         pos,
		Condition:   cond,
		Consequence: consequence,
	}

	p.skipNewlines()
	for p.peek().Type == lexer.TOKEN_ELIF {
		p.advance() // consume 'elif'
		elifCond := p.parseExpression(PREC_LOWEST)
		elifBody := p.parseBlock()
		stmt.ElifClauses = append(stmt.ElifClauses, ast.ElifClause{
			Condition:   elifCond,
			Consequence: elifBody,
		})
		p.skipNewlines()
	}

	if p.peek().Type == lexer.TOKEN_ELSE {
		p.advance() // consume 'else'
		stmt.Alternative = p.parseBlock()
	}

	return stmt
}

// loop <block>  |  loop <cond> <block>  |  loop <ident> in <expr> <block>
func (p *Parser) parseLoopStatement() ast.Statement {
	pos := p.advance().Pos // consume 'loop'
	stmt := &ast.LoopStatement{Pos: pos}

	// Infinite loop: loop { ... }
	if p.peek().Type == lexer.TOKEN_LBRACE {
		stmt.Body = p.parseBlock()
		return stmt
	}

	// Peek ahead: if ident followed by 'in', it's a for-in loop.
	if p.peek().Type == lexer.TOKEN_IDENT && p.peekNext().Type == lexer.TOKEN_IN {
		stmt.Iterator = p.advance().Literal // consume ident
		p.advance()                         // consume 'in'
		stmt.Iterable = p.parseExpression(PREC_LOWEST)
		stmt.Body = p.parseBlock()
		return stmt
	}

	// Conditional loop: loop <cond> { ... }
	stmt.Condition = p.parseExpression(PREC_LOWEST)
	stmt.Body = p.parseBlock()
	return stmt
}

func (p *Parser) parseBreakStatement() ast.Statement {
	return &ast.BreakStatement{Pos: p.advance().Pos}
}

func (p *Parser) parseContinueStatement() ast.Statement {
	return &ast.ContinueStatement{Pos: p.advance().Pos}
}

// match <expr> { <arms> }
func (p *Parser) parseMatchStatement() ast.Statement {
	pos := p.advance().Pos // consume 'match'
	subject := p.parseExpression(PREC_LOWEST)

	if !p.expect(lexer.TOKEN_LBRACE) {
		return nil
	}

	arms := make([]ast.MatchArm, 0)
	p.skipNewlines()
	for !p.isAtEnd() && p.peek().Type != lexer.TOKEN_RBRACE {
		arm := p.parseMatchArm()
		arms = append(arms, arm)
		p.skipNewlines()
	}

	p.expect(lexer.TOKEN_RBRACE)
	return &ast.MatchStatement{Pos: pos, Subject: subject, Arms: arms}
}

func (p *Parser) parseMatchArm() ast.MatchArm {
	arm := ast.MatchArm{}

	// Parse pattern. '_' is a wildcard.
	if p.peek().Type == lexer.TOKEN_IDENT && p.peek().Literal == "_" {
		arm.Pattern = &ast.WildcardExpression{Pos: p.advance().Pos}
	} else {
		arm.Pattern = p.parseExpression(PREC_LOWEST)
	}

	// Optional guard: if <cond>
	if p.peek().Type == lexer.TOKEN_IF {
		p.advance() // consume 'if'
		arm.Guard = p.parseExpression(PREC_LOWEST)
	}

	p.expect(lexer.TOKEN_ARROW) // consume '=>'

	// Arm body: block or single-expression wrapped in a block.
	if p.peek().Type == lexer.TOKEN_LBRACE {
		arm.Body = p.parseBlock()
	} else {
		expr := p.parseExpression(PREC_LOWEST)
		arm.Body = &ast.BlockStatement{
			Pos: expr.TokenPos(),
			Statements: []ast.Statement{
				&ast.ReturnStatement{Pos: expr.TokenPos(), Values: []ast.Expression{expr}},
			},
		}
	}
	return arm
}

// test "<description>" <block>
func (p *Parser) parseTestBlock() ast.Statement {
	pos := p.advance().Pos // consume 'test'
	if p.peek().Type != lexer.TOKEN_STRING {
		p.addError(fmt.Sprintf("expected string after 'test', got %q at %s", p.peek().Literal, p.peek().Pos))
		p.synchronize()
		return nil
	}
	desc := p.advance().Literal
	body := p.parseBlock()
	return &ast.TestBlock{Pos: pos, Description: desc, Body: body}
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
	pos := p.peek().Pos
	expr := p.parseExpression(PREC_LOWEST)
	if expr == nil {
		return nil
	}

	if p.peek().Type == lexer.TOKEN_ASSIGN {
		p.advance() // consume '='
		value := p.parseExpression(PREC_LOWEST)
		switch target := expr.(type) {
		case *ast.Identifier:
			return &ast.AssignStatement{Pos: pos, Name: target.Name, Value: value}
		case *ast.IndexExpression:
			return &ast.IndexAssignStatement{Pos: pos, Left: target.Left, Index: target.Index, Value: value}
		default:
			p.addError(fmt.Sprintf("invalid assignment target at %s", pos))
			return nil
		}
	}

	return &ast.ExpressionStatement{Pos: pos, Expression: expr}
}

func (p *Parser) parseBlock() *ast.BlockStatement {
	pos := p.peek().Pos
	if !p.expect(lexer.TOKEN_LBRACE) {
		return nil
	}
	block := &ast.BlockStatement{Pos: pos, Statements: make([]ast.Statement, 0)}

	p.skipNewlines()
	for !p.isAtEnd() && p.peek().Type != lexer.TOKEN_RBRACE {
		if stmt := p.parseStatement(); stmt != nil {
			block.Statements = append(block.Statements, stmt)
		}
		p.skipNewlines()
	}

	p.expect(lexer.TOKEN_RBRACE)
	return block
}

// ---------------------------------------------------------------------------
// Expression Parsing (Pratt)
// ---------------------------------------------------------------------------

func (p *Parser) parseExpression(prec Precedence) ast.Expression {
	left := p.parsePrefixExpression()
	if left == nil {
		return nil
	}

	for !p.isAtEnd() && prec < p.peekPrecedence() {
		left = p.parseInfixExpression(left)
		if left == nil {
			return nil
		}
	}
	return left
}

// parsePrefixExpression dispatches to the correct prefix handler.
func (p *Parser) parsePrefixExpression() ast.Expression {
	switch p.peek().Type {
	case lexer.TOKEN_INT:
		return p.parseIntLiteral()
	case lexer.TOKEN_FLOAT:
		return p.parseFloatLiteral()
	case lexer.TOKEN_STRING:
		return p.parseStringLiteral()
	case lexer.TOKEN_TRUE, lexer.TOKEN_FALSE:
		return p.parseBoolLiteral()
	case lexer.TOKEN_NONE:
		return p.parseNoneLiteral()
	case lexer.TOKEN_IDENT:
		return p.parseIdentifier()
	case lexer.TOKEN_LPAREN:
		return p.parseGroupedExpression()
	case lexer.TOKEN_LBRACKET:
		return p.parseArrayLiteral()
	case lexer.TOKEN_LBRACE:
		return p.parseMapLiteral()
	case lexer.TOKEN_FN:
		return p.parseFnLiteral()
	case lexer.TOKEN_MINUS, lexer.TOKEN_NOT:
		return p.parseUnaryExpression()
	default:
		p.addError(fmt.Sprintf("unexpected token %q at %s", p.peek().Literal, p.peek().Pos))
		p.advance()
		return nil
	}
}

// parseInfixExpression dispatches to the correct infix handler.
func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	switch p.peek().Type {
	case lexer.TOKEN_PLUS, lexer.TOKEN_MINUS, lexer.TOKEN_STAR,
		lexer.TOKEN_SLASH, lexer.TOKEN_PERCENT,
		lexer.TOKEN_EQ, lexer.TOKEN_NEQ,
		lexer.TOKEN_LT, lexer.TOKEN_GT, lexer.TOKEN_LTE, lexer.TOKEN_GTE,
		lexer.TOKEN_AND, lexer.TOKEN_OR:
		return p.parseBinaryExpression(left)
	case lexer.TOKEN_LPAREN:
		return p.parseCallExpression(left)
	case lexer.TOKEN_LBRACKET:
		return p.parseIndexExpression(left)
	case lexer.TOKEN_DOT:
		return p.parseDotExpression(left)
	case lexer.TOKEN_QMARK:
		return p.parseSafeAccessExpression(left)
	case lexer.TOKEN_PIPE:
		return p.parsePipelineExpression(left)
	case lexer.TOKEN_DOTDOT:
		return p.parseRangeExpression(left)
	case lexer.TOKEN_COALESCE:
		return p.parseCoalesceExpression(left)
	default:
		return left
	}
}

// ---------------------------------------------------------------------------
// Prefix Parsers
// ---------------------------------------------------------------------------

func (p *Parser) parseIntLiteral() ast.Expression {
	tok := p.advance()
	val, err := strconv.ParseInt(tok.Literal, 10, 64)
	if err != nil {
		p.addError(fmt.Sprintf("invalid integer %q at %s", tok.Literal, tok.Pos))
		return nil
	}
	return &ast.IntegerLiteral{Pos: tok.Pos, Value: val}
}

func (p *Parser) parseFloatLiteral() ast.Expression {
	tok := p.advance()
	val, err := strconv.ParseFloat(tok.Literal, 64)
	if err != nil {
		p.addError(fmt.Sprintf("invalid float %q at %s", tok.Literal, tok.Pos))
		return nil
	}
	return &ast.FloatLiteral{Pos: tok.Pos, Value: val}
}

func (p *Parser) parseStringLiteral() ast.Expression {
	tok := p.advance()
	return &ast.StringLiteral{Pos: tok.Pos, Value: tok.Literal}
}

func (p *Parser) parseBoolLiteral() ast.Expression {
	tok := p.advance()
	return &ast.BooleanLiteral{Pos: tok.Pos, Value: tok.Type == lexer.TOKEN_TRUE}
}

func (p *Parser) parseNoneLiteral() ast.Expression {
	return &ast.NoneLiteral{Pos: p.advance().Pos}
}

func (p *Parser) parseIdentifier() ast.Expression {
	tok := p.advance()
	return &ast.Identifier{Pos: tok.Pos, Name: tok.Literal}
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	p.advance() // consume '('
	expr := p.parseExpression(PREC_LOWEST)
	p.expect(lexer.TOKEN_RPAREN)
	return expr
}

func (p *Parser) parseArrayLiteral() ast.Expression {
	tok := p.advance() // consume '['
	elements := p.parseExpressionList(lexer.TOKEN_RBRACKET)
	return &ast.ArrayLiteral{Pos: tok.Pos, Elements: elements}
}

func (p *Parser) parseMapLiteral() ast.Expression {
	tok := p.advance() // consume '{'
	keys := make([]ast.Expression, 0)
	values := make([]ast.Expression, 0)

	p.skipNewlines()
	for !p.isAtEnd() && p.peek().Type != lexer.TOKEN_RBRACE {
		key := p.parseExpression(PREC_LOWEST)
		p.expect(lexer.TOKEN_COLON)
		val := p.parseExpression(PREC_LOWEST)
		keys = append(keys, key)
		values = append(values, val)

		if p.peek().Type != lexer.TOKEN_RBRACE {
			p.expect(lexer.TOKEN_COMMA)
		}
		p.skipNewlines()
	}

	p.expect(lexer.TOKEN_RBRACE)
	return &ast.MapLiteral{Pos: tok.Pos, Keys: keys, Values: values}
}

// fn(<params>) => <expr>  |  fn(<params>) <block>
func (p *Parser) parseFnLiteral() ast.Expression {
	tok := p.advance() // consume 'fn'
	params := p.parseParams()

	if p.peek().Type == lexer.TOKEN_ARROW {
		p.advance() // consume '=>'
		expr := p.parseExpression(PREC_LOWEST)
		body := &ast.BlockStatement{
			Pos: expr.TokenPos(),
			Statements: []ast.Statement{
				&ast.ReturnStatement{Pos: expr.TokenPos(), Values: []ast.Expression{expr}},
			},
		}
		return &ast.FnLiteral{Pos: tok.Pos, Params: params, Body: body}
	}

	body := p.parseBlock()
	return &ast.FnLiteral{Pos: tok.Pos, Params: params, Body: body}
}

func (p *Parser) parseUnaryExpression() ast.Expression {
	tok := p.advance() // consume operator
	operand := p.parseExpression(PREC_UNARY)
	return &ast.UnaryExpression{Pos: tok.Pos, Operator: tok.Literal, Operand: operand}
}

// ---------------------------------------------------------------------------
// Infix Parsers
// ---------------------------------------------------------------------------

func (p *Parser) parseBinaryExpression(left ast.Expression) ast.Expression {
	tok := p.advance() // consume operator
	prec := tokenPrecedence(tok.Type)
	right := p.parseExpression(prec)
	return &ast.BinaryExpression{Pos: tok.Pos, Left: left, Operator: tok.Literal, Right: right}
}

// <left>(<args>)
func (p *Parser) parseCallExpression(left ast.Expression) ast.Expression {
	tok := p.advance() // consume '('
	args := p.parseExpressionList(lexer.TOKEN_RPAREN)
	return &ast.CallExpression{Pos: tok.Pos, Function: left, Arguments: args}
}

// <left>[<index>]
func (p *Parser) parseIndexExpression(left ast.Expression) ast.Expression {
	tok := p.advance() // consume '['
	index := p.parseExpression(PREC_LOWEST)
	p.expect(lexer.TOKEN_RBRACKET)
	return &ast.IndexExpression{Pos: tok.Pos, Left: left, Index: index}
}

// <left>.<field>
func (p *Parser) parseDotExpression(left ast.Expression) ast.Expression {
	p.advance() // consume '.'
	field := p.advance()
	if field.Type != lexer.TOKEN_IDENT {
		p.addError(fmt.Sprintf("expected field name after '.', got %q at %s", field.Literal, field.Pos))
		return left
	}
	return &ast.DotExpression{Pos: field.Pos, Left: left, Field: field.Literal}
}

// <left>?.<field>
func (p *Parser) parseSafeAccessExpression(left ast.Expression) ast.Expression {
	p.advance() // consume '?.'
	field := p.advance()
	if field.Type != lexer.TOKEN_IDENT {
		p.addError(fmt.Sprintf("expected field name after '?.', got %q at %s", field.Literal, field.Pos))
		return left
	}
	return &ast.SafeAccessExpression{Pos: field.Pos, Left: left, Field: field.Literal}
}

// <left> |> <call>
func (p *Parser) parsePipelineExpression(left ast.Expression) ast.Expression {
	tok := p.advance() // consume '|>'
	right := p.parseExpression(PREC_PIPELINE)

	call, ok := right.(*ast.CallExpression)
	if !ok {
		p.addError(fmt.Sprintf("right side of |> must be a function call at %s", tok.Pos))
		return left
	}
	return &ast.PipelineExpression{Pos: tok.Pos, Left: left, Right: call}
}

// <left>..<right> [step <expr>]
func (p *Parser) parseRangeExpression(left ast.Expression) ast.Expression {
	tok := p.advance() // consume '..'
	end := p.parseExpression(PREC_RANGE)
	rangeExpr := &ast.RangeExpression{Pos: tok.Pos, Start: left, End: end}

	if p.peek().Type == lexer.TOKEN_STEP {
		p.advance() // consume 'step'
		rangeExpr.Step = p.parseExpression(PREC_RANGE)
	}
	return rangeExpr
}

// <left> ?? <right>
func (p *Parser) parseCoalesceExpression(left ast.Expression) ast.Expression {
	tok := p.advance() // consume '??'
	right := p.parseExpression(PREC_COALESCE)
	return &ast.CoalesceExpression{Pos: tok.Pos, Left: left, Right: right}
}

// ---------------------------------------------------------------------------
// Helpers
// ---------------------------------------------------------------------------

// parseParams parses a parenthesized, comma-separated list of identifiers.
func (p *Parser) parseParams() []string {
	p.expect(lexer.TOKEN_LPAREN)
	params := make([]string, 0)

	if p.peek().Type != lexer.TOKEN_RPAREN {
		params = append(params, p.advance().Literal)
		for p.peek().Type == lexer.TOKEN_COMMA {
			p.advance() // consume ','
			params = append(params, p.advance().Literal)
		}
	}

	p.expect(lexer.TOKEN_RPAREN)
	return params
}

// parseExpressionList parses comma-separated expressions until the end token.
func (p *Parser) parseExpressionList(end lexer.TokenType) []ast.Expression {
	list := make([]ast.Expression, 0)
	p.skipNewlines()

	if p.peek().Type == end {
		p.advance()
		return list
	}

	list = append(list, p.parseExpression(PREC_LOWEST))
	for p.peek().Type == lexer.TOKEN_COMMA {
		p.advance() // consume ','
		p.skipNewlines()
		list = append(list, p.parseExpression(PREC_LOWEST))
	}

	p.skipNewlines()
	p.expect(end)
	return list
}

// peekPrecedence returns the precedence of the current token.
func (p *Parser) peekPrecedence() Precedence {
	return tokenPrecedence(p.peek().Type)
}

// tokenPrecedence maps a token type to its precedence level.
func tokenPrecedence(t lexer.TokenType) Precedence {
	switch t {
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

// isStatementEnd returns true if the current token terminates a statement.
func (p *Parser) isStatementEnd() bool {
	t := p.peek().Type
	return t == lexer.TOKEN_NEWLINE || t == lexer.TOKEN_EOF || t == lexer.TOKEN_RBRACE
}

func (p *Parser) peek() lexer.Token {
	if p.current >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.current]
}

func (p *Parser) peekNext() lexer.Token {
	if p.current+1 >= len(p.tokens) {
		return lexer.Token{Type: lexer.TOKEN_EOF}
	}
	return p.tokens[p.current+1]
}

func (p *Parser) advance() lexer.Token {
	tok := p.peek()
	p.current++
	return tok
}

func (p *Parser) expect(expected lexer.TokenType) bool {
	if p.peek().Type == expected {
		p.advance()
		return true
	}
	p.addError(fmt.Sprintf("expected %v, got %v at %s", expected, p.peek().Type, p.peek().Pos))
	return false
}

func (p *Parser) isAtEnd() bool {
	return p.peek().Type == lexer.TOKEN_EOF
}

func (p *Parser) skipNewlines() {
	for p.peek().Type == lexer.TOKEN_NEWLINE {
		p.advance()
	}
}

func (p *Parser) addError(msg string) {
	p.errors = append(p.errors, msg)
}

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

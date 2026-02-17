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

// --- Statement stubs — IMPLEMENT EACH OF THESE ---

func (p *Parser) parseLetStatement() ast.Statement {
	// TODO: consume TOKEN_LET, expect TOKEN_IDENT, expect TOKEN_ASSIGN, parse expression
	pos := p.advance().Pos
	_ = pos
	p.addError("parseLetStatement not implemented")
	return nil
}

func (p *Parser) parseMutStatement() ast.Statement {
	// TODO: consume TOKEN_MUT, expect TOKEN_IDENT, expect TOKEN_ASSIGN, parse expression
	pos := p.advance().Pos
	_ = pos
	p.addError("parseMutStatement not implemented")
	return nil
}

func (p *Parser) parseFnDeclaration() ast.Statement {
	// TODO: consume TOKEN_FN, TOKEN_IDENT, parse params, parse block or => expr
	pos := p.advance().Pos
	_ = pos
	p.addError("parseFnDeclaration not implemented")
	return nil
}

func (p *Parser) parseReturnStatement() ast.Statement {
	// TODO: consume TOKEN_RETURN, parse optional comma-separated expressions
	pos := p.advance().Pos
	_ = pos
	p.addError("parseReturnStatement not implemented")
	return nil
}

func (p *Parser) parseIfStatement() ast.Statement {
	// TODO: consume TOKEN_IF, parse condition, parse block, handle elif/else
	pos := p.advance().Pos
	_ = pos
	p.addError("parseIfStatement not implemented")
	return nil
}

func (p *Parser) parseLoopStatement() ast.Statement {
	// TODO: consume TOKEN_LOOP, detect loop form (infinite, conditional, for-in)
	pos := p.advance().Pos
	_ = pos
	p.addError("parseLoopStatement not implemented")
	return nil
}

func (p *Parser) parseBreakStatement() ast.Statement {
	pos := p.advance().Pos
	return &ast.BreakStatement{Pos: pos}
}

func (p *Parser) parseContinueStatement() ast.Statement {
	pos := p.advance().Pos
	return &ast.ContinueStatement{Pos: pos}
}

func (p *Parser) parseMatchStatement() ast.Statement {
	// TODO: consume TOKEN_MATCH, parse subject expr, parse arms inside braces
	pos := p.advance().Pos
	_ = pos
	p.addError("parseMatchStatement not implemented")
	return nil
}

func (p *Parser) parseTestBlock() ast.Statement {
	// TODO: consume TOKEN_TEST, expect string literal, parse block
	pos := p.advance().Pos
	_ = pos
	p.addError("parseTestBlock not implemented")
	return nil
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
//
// TODO: Implement prefix and infix parsing.
// Prefix handlers: literals, identifiers, unary ops, grouping, arrays, maps, fn literals
// Infix handlers: binary ops, call, index, dot, safe access, pipeline, range, coalesce
func (p *Parser) parseExpression(precedence Precedence) ast.Expression {
	// --- IMPLEMENT THIS ---
	// 1. Get the current token
	// 2. Look up the prefix parse function for its type
	// 3. Call prefix function → left expression
	// 4. While next token has higher precedence:
	//    a. Look up infix parse function
	//    b. Call infix(left) → new left
	// 5. Return left

	p.addError("parseExpression not implemented")
	p.advance()
	return nil
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

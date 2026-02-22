package ast

import "github.com/glace-lang/glace/lexer"

// ---------------------------------------------------------------------------
// Core Interfaces
// ---------------------------------------------------------------------------

// Node is the base interface for all AST nodes.
type Node interface {
	TokenPos() lexer.Position // position of the defining token
	String() string           // pretty-print for debugging
}

// Statement nodes do not produce a value.
type Statement interface {
	Node
	stmtNode() // marker method
}

// Expression nodes produce a value when evaluated.
type Expression interface {
	Node
	exprNode() // marker method
}

// ---------------------------------------------------------------------------
// Program (root node)
// ---------------------------------------------------------------------------

// Program is the root of every Glace AST — a list of statements.
type Program struct {
	Statements []Statement
}

func (p *Program) TokenPos() lexer.Position {
	if len(p.Statements) > 0 {
		return p.Statements[0].TokenPos()
	}
	return lexer.Position{}
}
func (p *Program) String() string { return "Program" }

// ---------------------------------------------------------------------------
// Statements
// ---------------------------------------------------------------------------

// LetStatement: let x = <expr>
type LetStatement struct {
	Pos   lexer.Position
	Name  string
	Value Expression
}

func (s *LetStatement) stmtNode()                  {}
func (s *LetStatement) TokenPos() lexer.Position    { return s.Pos }
func (s *LetStatement) String() string              { return "LetStatement(" + s.Name + ")" }

// MutStatement: mut x = <expr>
type MutStatement struct {
	Pos   lexer.Position
	Name  string
	Value Expression
}

func (s *MutStatement) stmtNode()                   {}
func (s *MutStatement) TokenPos() lexer.Position     { return s.Pos }
func (s *MutStatement) String() string               { return "MutStatement(" + s.Name + ")" }

// AssignStatement: x = <expr>
type AssignStatement struct {
	Pos   lexer.Position
	Name  string
	Value Expression
}

func (s *AssignStatement) stmtNode()                {}
func (s *AssignStatement) TokenPos() lexer.Position { return s.Pos }
func (s *AssignStatement) String() string           { return "AssignStatement(" + s.Name + ")" }

// IndexAssignStatement: arr[i] = <expr>
type IndexAssignStatement struct {
	Pos   lexer.Position
	Left  Expression // the array/map expression
	Index Expression // the index expression
	Value Expression
}

func (s *IndexAssignStatement) stmtNode()                {}
func (s *IndexAssignStatement) TokenPos() lexer.Position { return s.Pos }
func (s *IndexAssignStatement) String() string           { return "IndexAssignStatement" }

// ExpressionStatement wraps an expression used as a statement.
type ExpressionStatement struct {
	Pos        lexer.Position
	Expression Expression
}

func (s *ExpressionStatement) stmtNode()                {}
func (s *ExpressionStatement) TokenPos() lexer.Position { return s.Pos }
func (s *ExpressionStatement) String() string           { return "ExpressionStatement" }

// ReturnStatement: return <expr>, <expr>, ...
type ReturnStatement struct {
	Pos    lexer.Position
	Values []Expression // zero or more return values
}

func (s *ReturnStatement) stmtNode()                {}
func (s *ReturnStatement) TokenPos() lexer.Position { return s.Pos }
func (s *ReturnStatement) String() string           { return "ReturnStatement" }

// BlockStatement: { ... }
type BlockStatement struct {
	Pos        lexer.Position
	Statements []Statement
}

func (s *BlockStatement) stmtNode()                {}
func (s *BlockStatement) TokenPos() lexer.Position { return s.Pos }
func (s *BlockStatement) String() string           { return "BlockStatement" }

// IfStatement: if <cond> { ... } elif <cond> { ... } else { ... }
type IfStatement struct {
	Pos         lexer.Position
	Condition   Expression
	Consequence *BlockStatement
	ElifClauses []ElifClause
	Alternative *BlockStatement // else block, may be nil
}

// ElifClause represents a single elif branch.
type ElifClause struct {
	Condition   Expression
	Consequence *BlockStatement
}

func (s *IfStatement) stmtNode()                {}
func (s *IfStatement) TokenPos() lexer.Position { return s.Pos }
func (s *IfStatement) String() string           { return "IfStatement" }

// LoopStatement: loop { ... } | loop <cond> { ... } | loop x in <expr> { ... }
type LoopStatement struct {
	Pos       lexer.Position
	Condition Expression      // nil for infinite loop
	Iterator  string          // "" if not a for-in loop
	Iterable  Expression      // nil if not a for-in loop
	Body      *BlockStatement
}

func (s *LoopStatement) stmtNode()                {}
func (s *LoopStatement) TokenPos() lexer.Position { return s.Pos }
func (s *LoopStatement) String() string           { return "LoopStatement" }

// --- GLACE-003: BreakStatement represents a break statement ---
// BreakStatement: break
type BreakStatement struct {
	Pos lexer.Position
}

func (s *BreakStatement) stmtNode()                {}
func (s *BreakStatement) TokenPos() lexer.Position { return s.Pos }
func (s *BreakStatement) String() string           { return "BreakStatement" }

// --- GLACE-003: ContinueStatement represents a continue statement ---
// ContinueStatement: continue
type ContinueStatement struct {
	Pos lexer.Position
}

func (s *ContinueStatement) stmtNode()                {}
func (s *ContinueStatement) TokenPos() lexer.Position { return s.Pos }
func (s *ContinueStatement) String() string           { return "ContinueStatement" }

// FnDeclaration: fn name(params) { ... }  or  fn name(params) => <expr>
type FnDeclaration struct {
	Pos    lexer.Position
	Name   string
	Params []string
	Body   *BlockStatement // block body
}

func (s *FnDeclaration) stmtNode()                {}
func (s *FnDeclaration) TokenPos() lexer.Position { return s.Pos }
func (s *FnDeclaration) String() string           { return "FnDeclaration(" + s.Name + ")" }

// MatchStatement: match <expr> { <arms> }
type MatchStatement struct {
	Pos     lexer.Position
	Subject Expression
	Arms    []MatchArm
}

// MatchArm: <pattern> [if <guard>] => <expr> | <block>
type MatchArm struct {
	Pattern Expression     // literal, ident, range, or wildcard
	Guard   Expression     // optional if-guard, may be nil
	Body    *BlockStatement // the arm body
}

func (s *MatchStatement) stmtNode()                {}
func (s *MatchStatement) TokenPos() lexer.Position { return s.Pos }
func (s *MatchStatement) String() string           { return "MatchStatement" }

// TestBlock: test "description" { ... }
type TestBlock struct {
	Pos         lexer.Position
	Description string
	Body        *BlockStatement
}

func (s *TestBlock) stmtNode()                {}
func (s *TestBlock) TokenPos() lexer.Position { return s.Pos }
func (s *TestBlock) String() string           { return "TestBlock(" + s.Description + ")" }

// ---------------------------------------------------------------------------
// Expressions
// ---------------------------------------------------------------------------

// IntegerLiteral: 42
type IntegerLiteral struct {
	Pos   lexer.Position
	Value int64
}

func (e *IntegerLiteral) exprNode()                {}
func (e *IntegerLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *IntegerLiteral) String() string           { return "IntegerLiteral" }

// FloatLiteral: 3.14
type FloatLiteral struct {
	Pos   lexer.Position
	Value float64
}

func (e *FloatLiteral) exprNode()                {}
func (e *FloatLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *FloatLiteral) String() string           { return "FloatLiteral" }

// StringLiteral: "hello"
type StringLiteral struct {
	Pos   lexer.Position
	Value string
}

func (e *StringLiteral) exprNode()                {}
func (e *StringLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *StringLiteral) String() string           { return "StringLiteral" }

// StringInterpolation: "hello ${name}, you are ${age} years old"
type StringInterpolation struct {
	Pos   lexer.Position
	Parts []Expression // alternating StringLiteral and expressions
}

func (e *StringInterpolation) exprNode()                {}
func (e *StringInterpolation) TokenPos() lexer.Position { return e.Pos }
func (e *StringInterpolation) String() string           { return "StringInterpolation" }

// BooleanLiteral: true, false
type BooleanLiteral struct {
	Pos   lexer.Position
	Value bool
}

func (e *BooleanLiteral) exprNode()                {}
func (e *BooleanLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *BooleanLiteral) String() string           { return "BooleanLiteral" }

// NoneLiteral: none
type NoneLiteral struct {
	Pos lexer.Position
}

func (e *NoneLiteral) exprNode()                {}
func (e *NoneLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *NoneLiteral) String() string           { return "NoneLiteral" }

// Identifier: variable_name
type Identifier struct {
	Pos  lexer.Position
	Name string
}

func (e *Identifier) exprNode()                {}
func (e *Identifier) TokenPos() lexer.Position { return e.Pos }
func (e *Identifier) String() string           { return "Identifier(" + e.Name + ")" }

// BinaryExpression: left <op> right
type BinaryExpression struct {
	Pos      lexer.Position
	Left     Expression
	Operator string
	Right    Expression
}

func (e *BinaryExpression) exprNode()                {}
func (e *BinaryExpression) TokenPos() lexer.Position { return e.Pos }
func (e *BinaryExpression) String() string           { return "BinaryExpr(" + e.Operator + ")" }

// UnaryExpression: <op> operand (prefix)
type UnaryExpression struct {
	Pos      lexer.Position
	Operator string
	Operand  Expression
}

func (e *UnaryExpression) exprNode()                {}
func (e *UnaryExpression) TokenPos() lexer.Position { return e.Pos }
func (e *UnaryExpression) String() string           { return "UnaryExpr(" + e.Operator + ")" }

// CallExpression: callee(args...)
type CallExpression struct {
	Pos       lexer.Position
	Function  Expression   // the function being called
	Arguments []Expression
}

func (e *CallExpression) exprNode()                {}
func (e *CallExpression) TokenPos() lexer.Position { return e.Pos }
func (e *CallExpression) String() string           { return "CallExpression" }

// IndexExpression: left[index]
type IndexExpression struct {
	Pos   lexer.Position
	Left  Expression
	Index Expression
}

func (e *IndexExpression) exprNode()                {}
func (e *IndexExpression) TokenPos() lexer.Position { return e.Pos }
func (e *IndexExpression) String() string           { return "IndexExpression" }

// DotExpression: left.field
type DotExpression struct {
	Pos   lexer.Position
	Left  Expression
	Field string
}

func (e *DotExpression) exprNode()                {}
func (e *DotExpression) TokenPos() lexer.Position { return e.Pos }
func (e *DotExpression) String() string           { return "DotExpr(." + e.Field + ")" }

// SafeAccessExpression: left?.field
type SafeAccessExpression struct {
	Pos   lexer.Position
	Left  Expression
	Field string
}

func (e *SafeAccessExpression) exprNode()                {}
func (e *SafeAccessExpression) TokenPos() lexer.Position { return e.Pos }
func (e *SafeAccessExpression) String() string           { return "SafeAccess(?." + e.Field + ")" }

// ArrayLiteral: [elem1, elem2, ...]
type ArrayLiteral struct {
	Pos      lexer.Position
	Elements []Expression
}

func (e *ArrayLiteral) exprNode()                {}
func (e *ArrayLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *ArrayLiteral) String() string           { return "ArrayLiteral" }

// MapLiteral: {"key": value, ...}
type MapLiteral struct {
	Pos    lexer.Position
	Keys   []Expression
	Values []Expression
}

func (e *MapLiteral) exprNode()                {}
func (e *MapLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *MapLiteral) String() string           { return "MapLiteral" }

// FnLiteral: fn(params) => <expr>  or  fn(params) { ... }
type FnLiteral struct {
	Pos    lexer.Position
	Params []string
	Body   *BlockStatement
}

func (e *FnLiteral) exprNode()                {}
func (e *FnLiteral) TokenPos() lexer.Position { return e.Pos }
func (e *FnLiteral) String() string           { return "FnLiteral" }

// RangeExpression: start..end [step s]
type RangeExpression struct {
	Pos   lexer.Position
	Start Expression
	End   Expression
	Step  Expression // may be nil
}

func (e *RangeExpression) exprNode()                {}
func (e *RangeExpression) TokenPos() lexer.Position { return e.Pos }
func (e *RangeExpression) String() string           { return "RangeExpression" }

// PipelineExpression: left |> right
type PipelineExpression struct {
	Pos   lexer.Position
	Left  Expression
	Right *CallExpression // must be a function call
}

func (e *PipelineExpression) exprNode()                {}
func (e *PipelineExpression) TokenPos() lexer.Position { return e.Pos }
func (e *PipelineExpression) String() string           { return "PipelineExpression" }

// CoalesceExpression: left ?? right
type CoalesceExpression struct {
	Pos   lexer.Position
	Left  Expression
	Right Expression
}

func (e *CoalesceExpression) exprNode()                {}
func (e *CoalesceExpression) TokenPos() lexer.Position { return e.Pos }
func (e *CoalesceExpression) String() string           { return "CoalesceExpression" }

// WildcardExpression: _ (used in match arms)
type WildcardExpression struct {
	Pos lexer.Position
}

func (e *WildcardExpression) exprNode()                {}
func (e *WildcardExpression) TokenPos() lexer.Position { return e.Pos }
func (e *WildcardExpression) String() string           { return "Wildcard" }

package lexer

import "fmt"

// TokenType represents the type of a lexical token.
type TokenType int

const (
	// Special
	TOKEN_ILLEGAL TokenType = iota
	TOKEN_EOF
	TOKEN_NEWLINE

	// Literals
	TOKEN_INT    // 42
	TOKEN_FLOAT  // 3.14
	TOKEN_STRING // "hello"
	TOKEN_IDENT  // variable_name

	// Operators
	TOKEN_PLUS     // +
	TOKEN_MINUS    // -
	TOKEN_STAR     // *
	TOKEN_SLASH    // /
	TOKEN_PERCENT  // %
	TOKEN_ASSIGN   // =
	TOKEN_EQ       // ==
	TOKEN_NEQ      // !=
	TOKEN_LT       // <
	TOKEN_GT       // >
	TOKEN_LTE      // <=
	TOKEN_GTE      // >=
	TOKEN_AND      // &&
	TOKEN_OR       // ||
	TOKEN_NOT      // !
	TOKEN_PIPE     // |>
	TOKEN_DOTDOT   // ..
	TOKEN_ARROW    // =>
	TOKEN_QMARK    // ?.
	TOKEN_COALESCE // ??

	// Delimiters
	TOKEN_LPAREN   // (
	TOKEN_RPAREN   // )
	TOKEN_LBRACE   // {
	TOKEN_RBRACE   // }
	TOKEN_LBRACKET // [
	TOKEN_RBRACKET // ]
	TOKEN_COMMA    // ,
	TOKEN_DOT      // .
	TOKEN_COLON    // :

	// String Interpolation
	TOKEN_STRING_START // "hello ${
	TOKEN_STRING_MID   // } middle ${
	TOKEN_STRING_END   // } tail"

	// Keywords
	TOKEN_LET
	TOKEN_MUT
	TOKEN_FN
	TOKEN_RETURN
	TOKEN_IF
	TOKEN_ELIF
	TOKEN_ELSE
	TOKEN_LOOP
	TOKEN_IN
	TOKEN_BREAK
	TOKEN_CONTINUE
	TOKEN_MATCH
	TOKEN_TRUE
	TOKEN_FALSE
	TOKEN_NONE
	TOKEN_TEST
	TOKEN_STEP
	TOKEN_IMPORT
)

// tokenNames maps TokenType to a human-readable name.
var tokenNames = map[TokenType]string{
	TOKEN_ILLEGAL:      "ILLEGAL",
	TOKEN_EOF:          "EOF",
	TOKEN_NEWLINE:      "NEWLINE",
	TOKEN_INT:          "INT",
	TOKEN_FLOAT:        "FLOAT",
	TOKEN_STRING:       "STRING",
	TOKEN_IDENT:        "IDENT",
	TOKEN_PLUS:         "+",
	TOKEN_MINUS:        "-",
	TOKEN_STAR:         "*",
	TOKEN_SLASH:        "/",
	TOKEN_PERCENT:      "%",
	TOKEN_ASSIGN:       "=",
	TOKEN_EQ:           "==",
	TOKEN_NEQ:          "!=",
	TOKEN_LT:           "<",
	TOKEN_GT:           ">",
	TOKEN_LTE:          "<=",
	TOKEN_GTE:          ">=",
	TOKEN_AND:          "&&",
	TOKEN_OR:           "||",
	TOKEN_NOT:          "!",
	TOKEN_PIPE:         "|>",
	TOKEN_DOTDOT:       "..",
	TOKEN_ARROW:        "=>",
	TOKEN_QMARK:        "?.",
	TOKEN_COALESCE:     "??",
	TOKEN_LPAREN:       "(",
	TOKEN_RPAREN:       ")",
	TOKEN_LBRACE:       "{",
	TOKEN_RBRACE:       "}",
	TOKEN_LBRACKET:     "[",
	TOKEN_RBRACKET:     "]",
	TOKEN_COMMA:        ",",
	TOKEN_DOT:          ".",
	TOKEN_COLON:        ":",
	TOKEN_STRING_START: "STRING_START",
	TOKEN_STRING_MID:   "STRING_MID",
	TOKEN_STRING_END:   "STRING_END",
	TOKEN_LET:          "let",
	TOKEN_MUT:          "mut",
	TOKEN_FN:           "fn",
	TOKEN_RETURN:       "return",
	TOKEN_IF:           "if",
	TOKEN_ELIF:         "elif",
	TOKEN_ELSE:         "else",
	TOKEN_LOOP:         "loop",
	TOKEN_IN:           "in",
	TOKEN_BREAK:        "break",
	TOKEN_CONTINUE:     "continue",
	TOKEN_MATCH:        "match",
	TOKEN_TRUE:         "true",
	TOKEN_FALSE:        "false",
	TOKEN_NONE:         "none",
	TOKEN_TEST:         "test",
	TOKEN_STEP:         "step",
	TOKEN_IMPORT:       "import",
}

// Keywords maps keyword strings to their TokenType.
var Keywords = map[string]TokenType{
	"let":      TOKEN_LET,
	"mut":      TOKEN_MUT,
	"fn":       TOKEN_FN,
	"return":   TOKEN_RETURN,
	"if":       TOKEN_IF,
	"elif":     TOKEN_ELIF,
	"else":     TOKEN_ELSE,
	"loop":     TOKEN_LOOP,
	"in":       TOKEN_IN,
	"break":    TOKEN_BREAK,
	"continue": TOKEN_CONTINUE,
	"match":    TOKEN_MATCH,
	"true":     TOKEN_TRUE,
	"false":    TOKEN_FALSE,
	"none":     TOKEN_NONE,
	"test":     TOKEN_TEST,
	"step":     TOKEN_STEP,
	"import":   TOKEN_IMPORT,
}

// LookupIdent returns the TokenType for an identifier string.
// If the identifier is a keyword, returns the keyword token type.
// Otherwise returns TOKEN_IDENT.
func LookupIdent(ident string) TokenType {
	if tok, ok := Keywords[ident]; ok {
		return tok
	}
	return TOKEN_IDENT
}

// Position represents a location in the source code.
type Position struct {
	File   string
	Line   int
	Column int
}

func (p Position) String() string {
	if p.File != "" {
		return fmt.Sprintf("%s:%d:%d", p.File, p.Line, p.Column)
	}
	return fmt.Sprintf("%d:%d", p.Line, p.Column)
}

// Token represents a single lexical token.
type Token struct {
	Type    TokenType
	Literal string
	Pos     Position
}

func (t Token) String() string {
	name, ok := tokenNames[t.Type]
	if !ok {
		name = "UNKNOWN"
	}
	return fmt.Sprintf("%s(%q at %s)", name, t.Literal, t.Pos)
}

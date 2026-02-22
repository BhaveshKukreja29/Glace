package lexer

// Keywords maps keyword strings to their token types.
var keywords = map[string]TokenType{
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
	"import":   TOKEN_IMPORT,
}

// Lexer performs lexical analysis on Glace source code,
// converting a string of source text into a slice of Tokens.
type Lexer struct {
	source  string
	file    string // filename for error reporting
	tokens  []Token
	start   int // start of current token
	current int // current position in source
	line    int
	column  int
}

// New creates a new Lexer for the given source code.
// The file parameter is used for error position reporting.
func New(source string, file string) *Lexer {
	return &Lexer{
		source: source,
		file:   file,
		tokens: make([]Token, 0),
		line:   1,
		column: 1,
	}
}

// Tokenize scans the entire source and returns the list of tokens.
// The last token is always TOKEN_EOF.
func (l *Lexer) Tokenize() []Token {
	for !l.isAtEnd() {
		l.start = l.current
		ch := l.peek()

		// Whitespace
		if ch == ' ' || ch == '\t' || ch == '\r' {
			l.advance()
			continue
		}

		// Newline
		if ch == '\n' {
			l.advance()
			l.addToken(TOKEN_NEWLINE, "\n")
			l.line++
			l.column = 1
			continue
		}

		// Comments
		if ch == '/' && l.peekNext() == '/' {
			l.advance() // consume '/'
			l.advance() // consume '/'
			// Skip to end of line
			for !l.isAtEnd() && l.peek() != '\n' {
				l.advance()
			}
			continue
		}

		// Numbers
		if isDigit(ch) {
			l.scanNumber()
			continue
		}

		// Strings
		if ch == '"' {
			l.scanString()
			continue
		}

		// Identifiers and keywords
		if isAlpha(ch) {
			l.scanIdentifier()
			continue
		}

		// Two-character operators
		next := l.peekNext()
		twoChar := string([]byte{ch, next})

		switch twoChar {
		case "==":
			l.advance()
			l.advance()
			l.addToken(TOKEN_EQ, "==")
			continue
		case "!=":
			l.advance()
			l.advance()
			l.addToken(TOKEN_NEQ, "!=")
			continue
		case "<=":
			l.advance()
			l.advance()
			l.addToken(TOKEN_LTE, "<=")
			continue
		case ">=":
			l.advance()
			l.advance()
			l.addToken(TOKEN_GTE, ">=")
			continue
		case "&&":
			l.advance()
			l.advance()
			l.addToken(TOKEN_AND, "&&")
			continue
		case "||":
			l.advance()
			l.advance()
			l.addToken(TOKEN_OR, "||")
			continue
		case "|>":
			l.advance()
			l.advance()
			l.addToken(TOKEN_PIPE, "|>")
			continue
		case "..":
			l.advance()
			l.advance()
			l.addToken(TOKEN_DOTDOT, "..")
			continue
		case "=>":
			l.advance()
			l.advance()
			l.addToken(TOKEN_ARROW, "=>")
			continue
		case "??":
			l.advance()
			l.advance()
			l.addToken(TOKEN_COALESCE, "??")
			continue
		case "?.":
			l.advance()
			l.advance()
			l.addToken(TOKEN_QMARK, "?.")
			continue
		}

		// Single-character tokens
		l.advance()
		switch ch {
		case '+':
			l.addToken(TOKEN_PLUS, "+")
		case '-':
			l.addToken(TOKEN_MINUS, "-")
		case '*':
			l.addToken(TOKEN_STAR, "*")
		case '/':
			l.addToken(TOKEN_SLASH, "/")
		case '%':
			l.addToken(TOKEN_PERCENT, "%")
		case '=':
			l.addToken(TOKEN_ASSIGN, "=")
		case '<':
			l.addToken(TOKEN_LT, "<")
		case '>':
			l.addToken(TOKEN_GT, ">")
		case '!':
			l.addToken(TOKEN_NOT, "!")
		case '(':
			l.addToken(TOKEN_LPAREN, "(")
		case ')':
			l.addToken(TOKEN_RPAREN, ")")
		case '{':
			l.addToken(TOKEN_LBRACE, "{")
		case '}':
			l.addToken(TOKEN_RBRACE, "}")
		case '[':
			l.addToken(TOKEN_LBRACKET, "[")
		case ']':
			l.addToken(TOKEN_RBRACKET, "]")
		case ',':
			l.addToken(TOKEN_COMMA, ",")
		case ':':
			l.addToken(TOKEN_COLON, ":")
		case '.':
			l.addToken(TOKEN_DOT, ".")
		case '?':
			l.addToken(TOKEN_QMARK, "?")
		default:
			// Unknown character, skip it
		}
	}

	l.addToken(TOKEN_EOF, "")
	return l.tokens
}

// scanNumber scans an integer or floating-point literal.
func (l *Lexer) scanNumber() {
	for isDigit(l.peek()) {
		l.advance()
	}

	// Check for decimal point
	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance() // consume '.'
		for isDigit(l.peek()) {
			l.advance()
		}
		literal := l.source[l.start:l.current]
		l.addToken(TOKEN_FLOAT, literal)
	} else {
		literal := l.source[l.start:l.current]
		l.addToken(TOKEN_INT, literal)
	}
}

// scanString scans a string literal with optional interpolation.
func (l *Lexer) scanString() {
	l.advance() // consume opening '"'

	for !l.isAtEnd() && l.peek() != '"' {
		if l.peek() == '\n' {
			l.line++
			l.column = 0
		}
		l.advance()
	}

	if l.isAtEnd() {
		// Unterminated string
		return
	}

	l.advance() // consume closing '"'
	literal := l.source[l.start+1 : l.current-1] // exclude quotes
	l.addToken(TOKEN_STRING, literal)
}

// scanIdentifier scans an identifier or keyword.
func (l *Lexer) scanIdentifier() {
	for isAlphaNumeric(l.peek()) {
		l.advance()
	}

	literal := l.source[l.start:l.current]
	tokenType := TOKEN_IDENT

	// Check if it's a keyword
	if kw, isKeyword := keywords[literal]; isKeyword {
		tokenType = kw
	}

	l.addToken(tokenType, literal)
}

// --- Helper methods (implement these) ---

// advance consumes the current character and returns it.
func (l *Lexer) advance() byte {
	ch := l.source[l.current]
	l.current++
	l.column++
	return ch
}

// peek returns the current character without consuming it.
func (l *Lexer) peek() byte {
	if l.current >= len(l.source) {
		return 0
	}
	return l.source[l.current]
}

// peekNext returns the character after the current one.
func (l *Lexer) peekNext() byte {
	if l.current+1 >= len(l.source) {
		return 0
	}
	return l.source[l.current+1]
}

// isAtEnd returns true if we've consumed all source characters.
func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

// addToken appends a token to the token list.
func (l *Lexer) addToken(tokenType TokenType, literal string) {
	l.tokens = append(l.tokens, Token{
		Type:    tokenType,
		Literal: literal,
		Pos: Position{
			File:   l.file,
			Line:   l.line,
			Column: l.column - len(literal),
		},
	})
}

// isDigit returns true if the byte is an ASCII digit.
func isDigit(ch byte) bool {
	return ch >= '0' && ch <= '9'
}

// isAlpha returns true if the byte is a letter or underscore.
func isAlpha(ch byte) bool {
	return (ch >= 'a' && ch <= 'z') ||
		(ch >= 'A' && ch <= 'Z') ||
		ch == '_'
}

// isAlphaNumeric returns true if the byte is a letter, digit, or underscore.
func isAlphaNumeric(ch byte) bool {
	return isAlpha(ch) || isDigit(ch)
}

package lexer

// Lexer performs lexical analysis on Glace source code.
type Lexer struct {
	source  string
	file    string
	tokens  []Token
	start   int
	current int
	line    int
	column  int
}

// New creates a new Lexer for the given source code.
func New(source string, file string) *Lexer {
	return &Lexer{
		source: source,
		file:   file,
		tokens: make([]Token, 0),
		line:   1,
		column: 1,
	}
}

// Tokenize scans the source and returns a slice of tokens[cite: 20].
func (l *Lexer) Tokenize() []Token {
	for !l.isAtEnd() {
		l.start = l.current
		l.scanToken()
	}

	l.addToken(TOKEN_EOF, "")
	return l.tokens
}

func (l *Lexer) scanToken() {
	ch := l.advance()

	switch ch {
	case ' ', '\t', '\r':
		return
	case '\n':
		l.line++
		l.column = 1
		l.addToken(TOKEN_NEWLINE, "\n")
	case '(': l.addToken(TOKEN_LPAREN, "(")
	case ')': l.addToken(TOKEN_RPAREN, ")")
	case '{': l.addToken(TOKEN_LBRACE, "{")
	case '}': 
		l.addToken(TOKEN_RBRACE, "}")
		// After a closing brace, check if we are inside a string interpolation
		// to resume scanning the string tail.
		if l.peek() == '"' {
			l.start = l.current
			l.advance() // consume "
			l.addToken(TOKEN_STRING_END, "\"")
		}
	case '[': l.addToken(TOKEN_LBRACKET, "[")
	case ']': l.addToken(TOKEN_RBRACKET, "]")
	case ',': l.addToken(TOKEN_COMMA, ",")
	case ':': l.addToken(TOKEN_COLON, ":")
	case '+': l.addToken(TOKEN_PLUS, "+")
	case '-': l.addToken(TOKEN_MINUS, "-")
	case '*': l.addToken(TOKEN_STAR, "*")
	case '%': l.addToken(TOKEN_PERCENT, "%")
	case '.':
		if l.match('.') {
			l.addToken(TOKEN_DOTDOT, "..")
		} else {
			l.addToken(TOKEN_DOT, ".")
		}
	case '/':
		if l.match('/') {
			for l.peek() != '\n' && !l.isAtEnd() {
				l.advance()
			}
		} else {
			l.addToken(TOKEN_SLASH, "/")
		}
	case '=':
		if l.match('=') {
			l.addToken(TOKEN_EQ, "==")
		} else if l.match('>') {
			l.addToken(TOKEN_ARROW, "=>")
		} else {
			l.addToken(TOKEN_ASSIGN, "=")
		}
	case '!':
		if l.match('=') {
			l.addToken(TOKEN_NEQ, "!=")
		} else {
			l.addToken(TOKEN_NOT, "!")
		}
	case '<':
		if l.match('=') {
			l.addToken(TOKEN_LTE, "<=")
		} else {
			l.addToken(TOKEN_LT, "<")
		}
	case '>':
		if l.match('=') {
			l.addToken(TOKEN_GTE, ">=")
		} else {
			l.addToken(TOKEN_GT, ">")
		}
	case '|':
		if l.match('>') {
			l.addToken(TOKEN_PIPE, "|>") // Pipeline operator 
		} else {
			l.addToken(TOKEN_ILLEGAL, "|")
		}
	case '?':
		if l.match('.') {
			l.addToken(TOKEN_QMARK, "?.") // Safe access [cite: 402]
		} else if l.match('?') {
			l.addToken(TOKEN_COALESCE, "??") // Coalesce [cite: 402]
		} else {
			l.addToken(TOKEN_ILLEGAL, "?")
		}
	case '"':
		l.scanString()
	
	default:
		if isDigit(ch) {
			l.scanNumber()
		} else if isAlpha(ch) {
			l.scanIdentifier()
		} else {
			l.addToken(TOKEN_ILLEGAL, string(ch))
		}
	}
}

func (l *Lexer) scanIdentifier() {
	for isAlphaNumeric(l.peek()) {
		l.advance()
	}
	text := l.source[l.start:l.current]
	l.addToken(LookupIdent(text), text) // Check keywords [cite: 19]
}

func (l *Lexer) scanNumber() {
	for isDigit(l.peek()) {
		l.advance()
	}
	if l.peek() == '.' && isDigit(l.peekNext()) {
		l.advance()
		for isDigit(l.peek()) {
			l.advance()
		}
		l.addToken(TOKEN_FLOAT, l.source[l.start:l.current])
		return
	}
	l.addToken(TOKEN_INT, l.source[l.start:l.current])
}

func (l *Lexer) scanString() {
	for !l.isAtEnd() && l.peek() != '"' {
		if l.peek() == '$' && l.peekNext() == '{' {
			// Part 1: Emit STRING_START [cite: 106]
			l.addToken(TOKEN_STRING_START, l.source[l.start+1:l.current])
			
			// Part 2: Consume and emit interpolation start
			l.advance() // $
			l.advance() // {
			l.addToken(TOKEN_LBRACE, "${")
			
			// Return to Tokenize loop to handle expression inside 
			return 
		}
		if l.peek() == '\n' {
			l.line++
			l.column = 1
		}
		l.advance()
	}

	if l.isAtEnd() {
		l.addToken(TOKEN_ILLEGAL, "unterminated string")
		return
	}

	l.advance() // Closing "
	value := l.source[l.start+1 : l.current-1]
	l.addToken(TOKEN_STRING, value)
}

// Low-level Helpers
func (l *Lexer) advance() byte {
	ch := l.source[l.current]
	l.current++
	l.column++
	return ch
}

func (l *Lexer) match(expected byte) bool {
	if l.isAtEnd() || l.source[l.current] != expected {
		return false
	}
	l.current++
	l.column++
	return true
}

func (l *Lexer) peek() byte {
	if l.isAtEnd() { return 0 }
	return l.source[l.current]
}

func (l *Lexer) peekNext() byte {
	if l.current+1 >= len(l.source) { return 0 }
	return l.source[l.current+1]
}

func (l *Lexer) isAtEnd() bool {
	return l.current >= len(l.source)
}

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

func isDigit(ch byte) bool { return ch >= '0' && ch <= '9' }
func isAlpha(ch byte) bool { return (ch >= 'a' && ch <= 'z') || (ch >= 'A' && ch <= 'Z') || ch == '_' }
func isAlphaNumeric(ch byte) bool { return isAlpha(ch) || isDigit(ch) }
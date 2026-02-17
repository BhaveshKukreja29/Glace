package lexer

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
//
// TODO: Implement the full scanner. This stub returns only EOF.
func (l *Lexer) Tokenize() []Token {
	// --- IMPLEMENT THIS ---
	// Walk through l.source character by character.
	// Use l.advance(), l.peek(), l.peekNext() helpers.
	// For each character, determine the token type:
	//   - Single char: +, -, *, /, %, (, ), {, }, [, ], ,, :
	//   - Two char:    ==, !=, <=, >=, |>, .., =>, ?., ??
	//   - Numbers:     scanNumber()
	//   - Strings:     scanString() (handle ${} interpolation)
	//   - Identifiers: scanIdentifier() (check Keywords map)
	//   - Newlines:    emit TOKEN_NEWLINE (skip consecutive blanks)
	//   - Whitespace:  skip (spaces, tabs)
	//   - Comments:    // single-line comments, skip to end of line
	//
	// Call l.addToken(tokenType, literal) for each token found.

	l.addToken(TOKEN_EOF, "")
	return l.tokens
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

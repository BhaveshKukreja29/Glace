package lexer
import "fmt"

import (
	"testing"
)

func TestTokenize(t *testing.T) {
	input := `let x = 10
loop i in 0..5 {
    print("val: ${i}")
} |> next()`

	expected := []struct {
		expectedType    TokenType
		expectedLiteral string
	}{
		{TOKEN_LET, "let"},
		{TOKEN_IDENT, "x"},
		{TOKEN_ASSIGN, "="},
		{TOKEN_INT, "10"},
		{TOKEN_NEWLINE, "\n"},
		{TOKEN_LOOP, "loop"},
		{TOKEN_IDENT, "i"},
		{TOKEN_IN, "in"},
		{TOKEN_INT, "0"},
		{TOKEN_DOTDOT, ".."},
		{TOKEN_INT, "5"},
		{TOKEN_LBRACE, "{"},
		{TOKEN_NEWLINE, "\n"},
		{TOKEN_IDENT, "print"},
		{TOKEN_LPAREN, "("},
		{TOKEN_STRING_START, "val: "},
		{TOKEN_LBRACE, "${"},
		{TOKEN_IDENT, "i"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_STRING_END, "\""},
		{TOKEN_RPAREN, ")"},
		{TOKEN_NEWLINE, "\n"},
		{TOKEN_RBRACE, "}"},
		{TOKEN_PIPE, "|>"},
		{TOKEN_IDENT, "next"},
		{TOKEN_LPAREN, "("},
		{TOKEN_RPAREN, ")"},
		{TOKEN_EOF, ""},
	}

	l := New(input, "test.glace")
	tokens := l.Tokenize()

	if len(tokens) != len(expected) {
		t.Fatalf("expected %d tokens, got %d", len(expected), len(tokens))
	}

	for i, tok := range tokens {
		fmt.Printf("Token struct %d: %+v\n", i, tok)
		if tok.Type != expected[i].expectedType {
			t.Errorf("tests[%d] - tokentype wrong. expected=%v, got=%v", i, expected[i].expectedType, tok.Type)
		}
		if tok.Literal != expected[i].expectedLiteral {
			t.Errorf("tests[%d] - literal wrong. expected=%q, got=%q", i, expected[i].expectedLiteral, tok.Literal)
		}
	}
}
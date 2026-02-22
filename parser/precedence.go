package parser

// Precedence levels for Pratt parsing, from lowest to highest.
type Precedence int

const (
	PREC_LOWEST     Precedence = iota
	PREC_PIPELINE   // |>
	PREC_COALESCE   // ??
	PREC_OR         // ||
	PREC_AND        // &&
	PREC_EQUALITY   // == !=
	PREC_COMPARISON // < > <= >=
	PREC_RANGE      // ..
	PREC_ADDITION   // + -
	PREC_MULTIPLY   // * / %
	PREC_UNARY      // ! -x
	PREC_CALL       // fn() [] . ?.
)

// NOTE: The parser implementer should use these precedence levels
// in the Pratt parser's infix dispatch to determine binding strength.
//
// Example usage in parseExpression:
//
//   func (p *Parser) infixPrecedence(tokenType TokenType) Precedence {
//       switch tokenType {
//       case TOKEN_PIPE:     return PREC_PIPELINE
//       case TOKEN_COALESCE: return PREC_COALESCE
//       case TOKEN_OR:       return PREC_OR
//       case TOKEN_AND:      return PREC_AND
//       case TOKEN_EQ, TOKEN_NEQ: return PREC_EQUALITY
//       case TOKEN_LT, TOKEN_GT, TOKEN_LTE, TOKEN_GTE: return PREC_COMPARISON
//       case TOKEN_DOTDOT:   return PREC_RANGE
//       case TOKEN_PLUS, TOKEN_MINUS: return PREC_ADDITION
//       case TOKEN_STAR, TOKEN_SLASH, TOKEN_PERCENT: return PREC_MULTIPLY
//       case TOKEN_LPAREN, TOKEN_LBRACKET, TOKEN_DOT, TOKEN_QMARK: return PREC_CALL
//       default: return PREC_LOWEST
//       }
//   }

package parser

// Precedence levels for Pratt parsing, lowest to highest.
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

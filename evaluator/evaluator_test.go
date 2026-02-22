package evaluator

import (
	"fmt"
	"io/ioutil"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/glace-lang/glace/lexer"
	"github.com/glace-lang/glace/parser"
)

// ---------------------------------------------------------------------------
// Test Helper Functions
// ---------------------------------------------------------------------------

// testEval lexes, parses, and evaluates the input string.
// Returns the result value or error from evaluation.
func testEval(input string) (Value, error) {
	l := lexer.New(input, "test")
	tokens := l.Tokenize()
	program, parseErrors := parser.Parse(tokens)

	// Check for parse errors
	if len(parseErrors) > 0 {
		return nil, fmt.Errorf("parse error: %s", strings.Join(parseErrors, ", "))
	}

	// Create environment with builtins
	env := NewEnclosedEnvironment(nil)
	RegisterBuiltins(env)

	// Evaluate
	return Eval(program, env)
}

// assertIntValue checks that val is an IntValue with the expected value.
func assertIntValue(t *testing.T, val Value, expected int64) {
	t.Helper()
	intVal, ok := val.(*IntValue)
	if !ok {
		t.Fatalf("expected IntValue, got %T", val)
	}
	if intVal.Value != expected {
		t.Fatalf("expected %d, got %d", expected, intVal.Value)
	}
}

// assertFloatValue checks that val is a FloatValue with the expected value.
func assertFloatValue(t *testing.T, val Value, expected float64) {
	t.Helper()
	floatVal, ok := val.(*FloatValue)
	if !ok {
		t.Fatalf("expected FloatValue, got %T", val)
	}
	if floatVal.Value != expected {
		t.Fatalf("expected %f, got %f", expected, floatVal.Value)
	}
}

// assertBoolValue checks that val is a BoolValue with the expected value.
func assertBoolValue(t *testing.T, val Value, expected bool) {
	t.Helper()
	boolVal, ok := val.(*BoolValue)
	if !ok {
		t.Fatalf("expected BoolValue, got %T", val)
	}
	if boolVal.Value != expected {
		t.Fatalf("expected %t, got %t", expected, boolVal.Value)
	}
}

// assertStringValue checks that val is a StringValue with the expected value.
func assertStringValue(t *testing.T, val Value, expected string) {
	t.Helper()
	stringVal, ok := val.(*StringValue)
	if !ok {
		t.Fatalf("expected StringValue, got %T", val)
	}
	if stringVal.Value != expected {
		t.Fatalf("expected %q, got %q", expected, stringVal.Value)
	}
}

// assertNone checks that val is a NoneValue.
func assertNone(t *testing.T, val Value) {
	t.Helper()
	_, ok := val.(*NoneValue)
	if !ok {
		t.Fatalf("expected NoneValue, got %T", val)
	}
}

// assertError checks that err is non-nil and contains the substring.
func assertError(t *testing.T, err error, substring string) {
	t.Helper()
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), substring) {
		t.Fatalf("expected error containing %q, got %q", substring, err.Error())
	}
}

// ---------------------------------------------------------------------------
// Unit Tests
// ---------------------------------------------------------------------------

// TestIntegerArithmetic tests basic integer arithmetic operations.
func TestIntegerArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{"1 + 2", 3},
		{"10 - 3", 7},
		{"4 * 5", 20},
		{"10 / 3", 3},
		{"7 % 3", 1},
		{"-5", -5},
		{"2 + 3 * 4", 14}, // precedence test
		{"(2 + 3) * 4", 20},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertIntValue(t, val, test.expected)
		})
	}
}

// TestFloatArithmetic tests floating-point arithmetic.
func TestFloatArithmetic(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1.5 + 2.5", 4.0},
		{"10.0 - 3.5", 6.5},
		{"2.5 * 2.0", 5.0},
		{"10.0 / 2.5", 4.0},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertFloatValue(t, val, test.expected)
		})
	}
}

// TestMixedNumericOps tests mixed integer and float operations.
func TestMixedNumericOps(t *testing.T) {
	tests := []struct {
		input    string
		expected float64
	}{
		{"1 + 2.5", 3.5},
		{"10 / 3.0", 3.3333333},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			floatVal, ok := val.(*FloatValue)
			if !ok {
				t.Fatalf("expected FloatValue, got %T", val)
			}
			// Use approximate equality for floats
			diff := floatVal.Value - test.expected
			if diff < 0 {
				diff = -diff
			}
			if diff > 0.01 {
				t.Fatalf("expected %f, got %f", test.expected, floatVal.Value)
			}
		})
	}
}

// TestComparisonOps tests comparison operators.
func TestComparisonOps(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"5 < 10", true},
		{"10 < 5", false},
		{"5 > 10", false},
		{"10 > 5", true},
		{"5 <= 5", true},
		{"5 <= 4", false},
		{"5 >= 5", true},
		{"5 >= 6", false},
		{"5 == 5", true},
		{"5 == 6", false},
		{"5 != 5", false},
		{"5 != 6", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertBoolValue(t, val, test.expected)
		})
	}
}

// TestLogicalOps tests logical operators.
func TestLogicalOps(t *testing.T) {
	tests := []struct {
		input    string
		expected bool
	}{
		{"true && true", true},
		{"true && false", false},
		{"false && true", false},
		{"false && false", false},
		{"true || true", true},
		{"true || false", true},
		{"false || true", true},
		{"false || false", false},
		{"!true", false},
		{"!false", true},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertBoolValue(t, val, test.expected)
		})
	}
}

// TestStringOps tests string operations.
func TestStringOps(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`"hello" + " " + "world"`, "hello world"},
		{`"glace"`, "glace"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertStringValue(t, val, test.expected)
		})
	}
}

// TestLetStatement tests let bindings.
func TestLetStatement(t *testing.T) {
	input := `
		let x = 42
		x
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 42)
}

// TestMutStatement tests mutable bindings.
func TestMutStatement(t *testing.T) {
	input := `
		mut x = 5
		x = 10
		x
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 10)
}

// TestImmutability tests that immutable variables cannot be reassigned.
func TestImmutability(t *testing.T) {
	input := `
		let x = 5
		x = 10
	`
	_, err := testEval(input)
	if err == nil {
		t.Fatalf("expected error for reassigning immutable variable")
	}
}

// TestIfElse tests if/elif/else statements.
func TestIfElse(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`if true { 1 } else { 2 }`, 1},
		{`if false { 1 } else { 2 }`, 2},
		{`if false { 1 } elif true { 2 } else { 3 }`, 2},
		{`if false { 1 } elif false { 2 } else { 3 }`, 3},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertIntValue(t, val, test.expected)
		})
	}
}

// TestLoopConditional tests conditional loops.
func TestLoopConditional(t *testing.T) {
	input := `
		mut x = 0
		loop x < 3 {
			x = x + 1
		}
		x
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 3)
}

// TestLoopForIn tests for-in loops with ranges.
func TestLoopForIn(t *testing.T) {
	input := `
		mut sum = 0
		loop i in 0..5 {
			sum = sum + i
		}
		sum
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// 0 + 1 + 2 + 3 + 4 = 10
	assertIntValue(t, val, 10)
}

// TestLoopForInArray tests for-in loops with arrays.
func TestLoopForInArray(t *testing.T) {
	input := `
		let arr = [1, 2, 3]
		mut sum = 0
		loop x in arr {
			sum = sum + x
		}
		sum
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// 1 + 2 + 3 = 6
	assertIntValue(t, val, 6)
}

// TestBreak tests break in loops.
func TestBreak(t *testing.T) {
	input := `
		mut x = 0
		loop x < 100 {
			x = x + 1
			if x == 5 {
				break
			}
		}
		x
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 5)
}

// TestContinue tests continue in loops.
func TestContinue(t *testing.T) {
	input := `
		mut sum = 0
		loop i in 0..10 {
			if i % 2 == 0 {
				continue
			}
			sum = sum + i
		}
		sum
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// 1 + 3 + 5 + 7 + 9 = 25
	assertIntValue(t, val, 25)
}

// TestFunctionDeclaration tests function declaration and calls.
func TestFunctionDeclaration(t *testing.T) {
	input := `
		fn add(a, b) {
			a + b
		}
		add(3, 4)
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 7)
}

// TestReturnStatement tests explicit return.
func TestReturnStatement(t *testing.T) {
	input := `
		fn test() {
			return 42
			999
		}
		test()
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 42)
}

// TestRecursion tests recursive functions.
func TestRecursion(t *testing.T) {
	input := `
		fn factorial(n) {
			if n == 0 {
				1
			} else {
				n * factorial(n - 1)
			}
		}
		factorial(5)
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// 5! = 120
	assertIntValue(t, val, 120)
}

// TestClosures tests closures.
func TestClosures(t *testing.T) {
	input := `
		fn counter() {
			mut x = 0
			fn() {
				x = x + 1
				x
			}
		}
		let c = counter()
		let a = c()
		let b = c()
		let d = c()
		d
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// Closure should maintain state, so d = 3
	assertIntValue(t, val, 3)
}

// TestArrayLiteral tests array creation.
func TestArrayLiteral(t *testing.T) {
	input := `
		let arr = [1, 2, 3, 4, 5]
		arr[2]
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 3)
}

// TestArrayIndexing tests array indexing.
func TestArrayIndexing(t *testing.T) {
	input := `
		let arr = [10, 20, 30]
		arr[0] + arr[1] + arr[2]
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 60)
}

// TestArrayOutOfBounds tests array index out of bounds.
func TestArrayOutOfBounds(t *testing.T) {
	input := `
		let arr = [1, 2, 3]
		arr[10]
	`
	_, err := testEval(input)
	assertError(t, err, "out of bounds")
}

// TestMapLiteral tests map creation.
func TestMapLiteral(t *testing.T) {
	input := `
		let m = { "a": 1, "b": 2 }
		m["a"]
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 1)
}

// TestMapDotAccess tests map dot access.
func TestMapDotAccess(t *testing.T) {
	input := `
		let m = { "name": "glace", "version": 1 }
		m.name
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertStringValue(t, val, "glace")
}

// TestRangeCreation tests range creation.
func TestRangeCreation(t *testing.T) {
	input := `
		let r = 1..5
		r
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	rv, ok := val.(*RangeValue)
	if !ok {
		t.Fatalf("expected RangeValue, got %T", val)
	}
	if rv.Start != 1 || rv.End != 5 {
		t.Fatalf("expected range 1..5, got %d..%d", rv.Start, rv.End)
	}
}

// TestRangeIndexing tests range indexing.
func TestRangeIndexing(t *testing.T) {
	input := `
		let r = 10..20
		r[3]
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 13) // 10 + 3
}

// TestMatchLiteral tests match with literals.
func TestMatchLiteral(t *testing.T) {
	input := `
		let x = 2
		match x {
			1 => "one"
			2 => "two"
			3 => "three"
			_ => "other"
		}
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertStringValue(t, val, "two")
}

// TestMatchWildcard tests match with wildcard.
func TestMatchWildcard(t *testing.T) {
	input := `
		let x = 99
		match x {
			1 => "one"
			2 => "two"
			_ => "other"
		}
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertStringValue(t, val, "other")
}

// TestMatchRange tests match with range pattern.
func TestMatchRange(t *testing.T) {
	input := `
		let x = 5
		match x {
			0..3 => "low"
			3..7 => "mid"
			7..10 => "high"
			_ => "out"
		}
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertStringValue(t, val, "mid")
}

// TestBuiltinPrint tests print builtin (returns none).
func TestBuiltinPrint(t *testing.T) {
	input := `print(42)`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertNone(t, val)
}

// TestBuiltinLen tests len builtin.
func TestBuiltinLen(t *testing.T) {
	tests := []struct {
		input    string
		expected int64
	}{
		{`len([1, 2, 3])`, 3},
		{`len("hello")`, 5},
		{`len(0..10)`, 10},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertIntValue(t, val, test.expected)
		})
	}
}

// TestBuiltinType tests type builtin.
func TestBuiltinType(t *testing.T) {
	tests := []struct {
		input    string
		expected string
	}{
		{`type(42)`, "int"},
		{`type(3.14)`, "float"},
		{`type("hello")`, "string"},
		{`type(true)`, "bool"},
		{`type([1, 2, 3])`, "array"},
		{`type(none)`, "none"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			assertStringValue(t, val, test.expected)
		})
	}
}

// TestBuiltinStr tests str conversion.
func TestBuiltinStr(t *testing.T) {
	input := `str(42)`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertStringValue(t, val, "42")
}

// TestBuiltinInt tests int conversion.
func TestBuiltinInt(t *testing.T) {
	input := `int("42")`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 42)
}

// TestDivisionByZero tests division by zero error.
func TestDivisionByZero(t *testing.T) {
	input := `10 / 0`
	_, err := testEval(input)
	assertError(t, err, "division by zero")
}

// TestUndefinedVariable tests undefined variable error.
func TestUndefinedVariable(t *testing.T) {
	input := `x`
	_, err := testEval(input)
	assertError(t, err, "undefined variable")
}

// TestWrongArgumentCount tests function call with wrong argument count.
func TestWrongArgumentCount(t *testing.T) {
	input := `
		fn add(a, b) { a + b }
		add(1)
	`
	_, err := testEval(input)
	assertError(t, err, "takes 2 arguments")
}

// TestEnvironmentScoping tests variable scoping.
func TestEnvironmentScoping(t *testing.T) {
	input := `
		let x = 5
		{
			let x = 10
			x
		}
		x
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// Outer scope should be unchanged
	assertIntValue(t, val, 5)
}

// TestShadowing tests variable shadowing.
func TestShadowing(t *testing.T) {
	input := `
		let x = 5
		fn test() {
			let x = 10
			x
		}
		test()
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 10)
}

// TestCoalesce tests coalesce operator.
func TestCoalesce(t *testing.T) {
	tests := []struct {
		input    string
		expected interface{}
	}{
		{`"value" ?? "default"`, "value"},
		{`none ?? "default"`, "default"},
	}

	for _, test := range tests {
		t.Run(test.input, func(t *testing.T) {
			val, err := testEval(test.input)
			if err != nil {
				t.Fatalf("eval error: %v", err)
			}
			if stringVal, ok := test.expected.(string); ok {
				assertStringValue(t, val, stringVal)
			}
		})
	}
}

// TestNestedBreak tests break in nested loops.
func TestNestedBreak(t *testing.T) {
	input := `
		mut outer = 0
		loop i in 0..5 {
			mut inner = 0
			loop j in 0..5 {
				inner = inner + 1
				if j == 2 {
					break
				}
			}
			outer = outer + inner
		}
		outer
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	// Each inner loop breaks at j==2, so inner=3 each time
	// 3 + 3 + 3 + 3 + 3 = 15
	assertIntValue(t, val, 15)
}

// TestStringInterpolation tests string interpolation.
func TestStringInterpolation(t *testing.T) {
	input := `
		let name = "Glace"
		let version = 1
		"Welcome to ${name}, version ${version}!"
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertStringValue(t, val, "Welcome to Glace, version 1!")
}

// TestPipelineExpression tests pipeline operator.
func TestPipelineExpression(t *testing.T) {
	input := `
		fn double(x) { x * 2 }
		5 |> double()
	`
	val, err := testEval(input)
	if err != nil {
		t.Fatalf("eval error: %v", err)
	}
	assertIntValue(t, val, 10)
}

// ---------------------------------------------------------------------------
// Integration Tests
// ---------------------------------------------------------------------------

// TestQuicksortProgram tests the quicksort example program.
func TestQuicksortProgram(t *testing.T) {
	// Skip if file doesn't exist
	filepath := filepath.Join("..", "examples", "quicksort.glace")
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Skip("quicksort.glace not found")
	}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read quicksort.glace: %v", err)
	}

	_, err = testEval(string(content))
	if err != nil {
		t.Fatalf("quicksort evaluation failed: %v", err)
	}
}

// TestBubblesortProgram tests the bubblesort example program.
func TestBubblesortProgram(t *testing.T) {
	// Skip if file doesn't exist
	filepath := filepath.Join("..", "examples", "bubblesort.glace")
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Skip("bubblesort.glace not found")
	}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read bubblesort.glace: %v", err)
	}

	_, err = testEval(string(content))
	if err != nil {
		t.Fatalf("bubblesort evaluation failed: %v", err)
	}
}

// TestFibonacciProgram tests the fibonacci example program.
func TestFibonacciProgram(t *testing.T) {
	// Skip if file doesn't exist
	filepath := filepath.Join("..", "examples", "fibonacci.glace")
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Skip("fibonacci.glace not found")
	}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read fibonacci.glace: %v", err)
	}

	_, err = testEval(string(content))
	if err != nil {
		t.Fatalf("fibonacci evaluation failed: %v", err)
	}
}

// TestBreakContinueProgram tests break/continue example programs.
func TestBreakContinueProgram(t *testing.T) {
	// Skip if file doesn't exist
	filepath := filepath.Join("..", "examples", "break_continue.glace")
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		t.Skip("break_continue.glace not found")
	}

	content, err := ioutil.ReadFile(filepath)
	if err != nil {
		t.Fatalf("failed to read break_continue.glace: %v", err)
	}

	_, err = testEval(string(content))
	if err != nil {
		t.Fatalf("break_continue evaluation failed: %v", err)
	}
}

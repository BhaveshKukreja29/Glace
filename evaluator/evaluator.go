package evaluator

import (
	"fmt"

	"github.com/glace-lang/glace/ast"
)

// ---------------------------------------------------------------------------
// Signal types for control flow (break, continue, return)
// ---------------------------------------------------------------------------

// ReturnSignal is used to unwind the call stack on return.
type ReturnSignal struct {
	Values []Value
}

func (r *ReturnSignal) Error() string { return "return signal" }

// BreakSignal is used to break out of loops.
type BreakSignal struct{}

func (b *BreakSignal) Error() string { return "break signal" }

// ContinueSignal is used to skip to next loop iteration.
type ContinueSignal struct{}

func (c *ContinueSignal) Error() string { return "continue signal" }

// RuntimeError represents a user-facing runtime error.
type RuntimeError struct {
	Message string
	Pos     string
}

func (e *RuntimeError) Error() string {
	if e.Pos != "" {
		return fmt.Sprintf("runtime error at %s: %s", e.Pos, e.Message)
	}
	return fmt.Sprintf("runtime error: %s", e.Message)
}

// ---------------------------------------------------------------------------
// Evaluator
// ---------------------------------------------------------------------------

// Eval evaluates an AST node in the given environment and returns a Value.
//
// This is the main dispatch function. It should switch on the node type
// and delegate to the appropriate eval function.
//
// TODO: Implement each case. The structure is laid out below.
func Eval(node ast.Node, env *Environment) (Value, error) {
	switch n := node.(type) {
	// --- Program ---
	case *ast.Program:
		return evalProgram(n, env)

	// --- Statements ---
	case *ast.LetStatement:
		return evalLetStatement(n, env)
	case *ast.MutStatement:
		return evalMutStatement(n, env)
	case *ast.AssignStatement:
		return evalAssignStatement(n, env)
	case *ast.IndexAssignStatement:
		return evalIndexAssignStatement(n, env)
	case *ast.ExpressionStatement:
		return Eval(n.Expression, env)
	case *ast.BlockStatement:
		return evalBlockStatement(n, env)
	case *ast.ReturnStatement:
		return evalReturnStatement(n, env)
	case *ast.IfStatement:
		return evalIfStatement(n, env)
	case *ast.LoopStatement:
		return evalLoopStatement(n, env)
	case *ast.BreakStatement:
		return nil, &BreakSignal{}
	case *ast.ContinueStatement:
		return nil, &ContinueSignal{}
	case *ast.FnDeclaration:
		return evalFnDeclaration(n, env)
	case *ast.MatchStatement:
		return evalMatchStatement(n, env)
	case *ast.TestBlock:
		// Test blocks are skipped in normal execution
		return NONE, nil

	// --- Expressions ---
	case *ast.IntegerLiteral:
		return NewInt(n.Value), nil
	case *ast.FloatLiteral:
		return NewFloat(n.Value), nil
	case *ast.StringLiteral:
		return NewString(n.Value), nil
	case *ast.BooleanLiteral:
		return NewBool(n.Value), nil
	case *ast.NoneLiteral:
		return NONE, nil
	case *ast.Identifier:
		return evalIdentifier(n, env)
	case *ast.BinaryExpression:
		return evalBinaryExpression(n, env)
	case *ast.UnaryExpression:
		return evalUnaryExpression(n, env)
	case *ast.CallExpression:
		return evalCallExpression(n, env)
	case *ast.IndexExpression:
		return evalIndexExpression(n, env)
	case *ast.DotExpression:
		return evalDotExpression(n, env)
	case *ast.ArrayLiteral:
		return evalArrayLiteral(n, env)
	case *ast.MapLiteral:
		return evalMapLiteral(n, env)
	case *ast.FnLiteral:
		return evalFnLiteral(n, env)
	case *ast.RangeExpression:
		return evalRangeExpression(n, env)
	case *ast.PipelineExpression:
		return evalPipelineExpression(n, env)
	case *ast.CoalesceExpression:
		return evalCoalesceExpression(n, env)
	case *ast.SafeAccessExpression:
		return evalSafeAccessExpression(n, env)
	case *ast.StringInterpolation:
		return evalStringInterpolation(n, env)

	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown node type: %T", node)}
	}
}

// ---------------------------------------------------------------------------
// Statement Evaluation
// ---------------------------------------------------------------------------

func evalProgram(program *ast.Program, env *Environment) (Value, error) {
	var result Value = NONE
	for _, stmt := range program.Statements {
		val, err := Eval(stmt, env)
		if err != nil {
			return nil, err
		}
		result = val
	}
	return result, nil
}

func evalLetStatement(stmt *ast.LetStatement, env *Environment) (Value, error) {
	val, err := Eval(stmt.Value, env)
	if err != nil {
		return nil, err
	}
	if err := env.Define(stmt.Name, val, false); err != nil {
		return nil, &RuntimeError{Message: err.Error(), Pos: stmt.Pos.String()}
	}
	return NONE, nil
}

func evalMutStatement(stmt *ast.MutStatement, env *Environment) (Value, error) {
	val, err := Eval(stmt.Value, env)
	if err != nil {
		return nil, err
	}
	if err := env.Define(stmt.Name, val, true); err != nil {
		return nil, &RuntimeError{Message: err.Error(), Pos: stmt.Pos.String()}
	}
	return NONE, nil
}

func evalAssignStatement(stmt *ast.AssignStatement, env *Environment) (Value, error) {
	val, err := Eval(stmt.Value, env)
	if err != nil {
		return nil, err
	}
	if err := env.Set(stmt.Name, val); err != nil {
		return nil, &RuntimeError{Message: err.Error(), Pos: stmt.Pos.String()}
	}
	return NONE, nil
}

func evalIndexAssignStatement(stmt *ast.IndexAssignStatement, env *Environment) (Value, error) {
	left, err := Eval(stmt.Left, env)
	if err != nil {
		return nil, err
	}
	index, err := Eval(stmt.Index, env)
	if err != nil {
		return nil, err
	}
	val, err := Eval(stmt.Value, env)
	if err != nil {
		return nil, err
	}

	switch target := left.(type) {
	case *ArrayValue:
		idx, ok := index.(*IntValue)
		if !ok {
			return nil, &RuntimeError{Message: "array index must be an integer", Pos: stmt.Pos.String()}
		}
		i := int(idx.Value)
		if i < 0 || i >= len(target.Elements) {
			return nil, &RuntimeError{Message: fmt.Sprintf("index %d out of bounds (len %d)", i, len(target.Elements)), Pos: stmt.Pos.String()}
		}
		target.Elements[i] = val
	case *MapValue:
		key, ok := index.(*StringValue)
		if !ok {
			return nil, &RuntimeError{Message: "map key must be a string", Pos: stmt.Pos.String()}
		}
		target.Pairs[key.Value] = val
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("cannot index into '%s'", left.Type()), Pos: stmt.Pos.String()}
	}

	return NONE, nil
}

func evalBlockStatement(block *ast.BlockStatement, env *Environment) (Value, error) {
	var result Value = NONE
	for _, stmt := range block.Statements {
		val, err := Eval(stmt, env)
		if err != nil {
			return nil, err // propagate return/break/continue signals
		}
		result = val
	}
	return result, nil
}

func evalReturnStatement(stmt *ast.ReturnStatement, env *Environment) (Value, error) {
	values := make([]Value, len(stmt.Values))
	for i, expr := range stmt.Values {
		val, err := Eval(expr, env)
		if err != nil {
			return nil, err
		}
		values[i] = val
	}
	return nil, &ReturnSignal{Values: values}
}

func evalIfStatement(stmt *ast.IfStatement, env *Environment) (Value, error) {
	cond, err := Eval(stmt.Condition, env)
	if err != nil {
		return nil, err
	}

	if IsTruthy(cond) {
		return Eval(stmt.Consequence, env)
	}

	for _, elif := range stmt.ElifClauses {
		cond, err := Eval(elif.Condition, env)
		if err != nil {
			return nil, err
		}
		if IsTruthy(cond) {
			return Eval(elif.Consequence, env)
		}
	}

	if stmt.Alternative != nil {
		return Eval(stmt.Alternative, env)
	}

	return NONE, nil
}

func evalLoopStatement(stmt *ast.LoopStatement, env *Environment) (Value, error) {
	// --- for-in loop ---
	if stmt.Iterator != "" {
		iterable, err := Eval(stmt.Iterable, env)
		if err != nil {
			return nil, err
		}
		return evalForInLoop(stmt, iterable, env)
	}

	// --- conditional or infinite loop ---
	for {
		if stmt.Condition != nil {
			cond, err := Eval(stmt.Condition, env)
			if err != nil {
				return nil, err
			}
			if !IsTruthy(cond) {
				break
			}
		}

		loopEnv := NewEnclosedEnvironment(env)
		_, err := Eval(stmt.Body, loopEnv)
		if err != nil {
			if _, ok := err.(*BreakSignal); ok {
				break
			}
			if _, ok := err.(*ContinueSignal); ok {
				continue
			}
			return nil, err
		}
	}
	return NONE, nil
}

func evalForInLoop(stmt *ast.LoopStatement, iterable Value, env *Environment) (Value, error) {
	switch iter := iterable.(type) {
	case *ArrayValue:
		for _, elem := range iter.Elements {
			loopEnv := NewEnclosedEnvironment(env)
			loopEnv.Define(stmt.Iterator, elem, false)
			_, err := Eval(stmt.Body, loopEnv)
			if err != nil {
				if _, ok := err.(*BreakSignal); ok {
					break
				}
				if _, ok := err.(*ContinueSignal); ok {
					continue
				}
				return nil, err
			}
		}
	case *RangeValue:
		for i := iter.Start; i < iter.End; i += iter.Step {
			loopEnv := NewEnclosedEnvironment(env)
			loopEnv.Define(stmt.Iterator, NewInt(i), false)
			_, err := Eval(stmt.Body, loopEnv)
			if err != nil {
				if _, ok := err.(*BreakSignal); ok {
					break
				}
				if _, ok := err.(*ContinueSignal); ok {
					continue
				}
				return nil, err
			}
		}
	default:
		return nil, &RuntimeError{
			Message: fmt.Sprintf("cannot iterate over '%s'", iterable.Type()),
			Pos:     stmt.Pos.String(),
		}
	}
	return NONE, nil
}

func evalFnDeclaration(stmt *ast.FnDeclaration, env *Environment) (Value, error) {
	fn := &FnValue{
		Name:   stmt.Name,
		Params: stmt.Params,
		Body:   stmt.Body,
		Env:    env,
	}
	env.Define(stmt.Name, fn, false)
	return NONE, nil
}

func evalMatchStatement(stmt *ast.MatchStatement, env *Environment) (Value, error) {
	subject, err := Eval(stmt.Subject, env)
	if err != nil {
		return nil, err
	}

	for _, arm := range stmt.Arms {
		matched, err := matchPattern(arm.Pattern, subject, env)
		if err != nil {
			return nil, err
		}
		if !matched {
			continue
		}

		// Check guard
		if arm.Guard != nil {
			guardVal, err := Eval(arm.Guard, env)
			if err != nil {
				return nil, err
			}
			if !IsTruthy(guardVal) {
				continue
			}
		}

		return Eval(arm.Body, env)
	}

	return NONE, nil
}

func matchPattern(pattern ast.Expression, subject Value, env *Environment) (bool, error) {
	switch p := pattern.(type) {
	case *ast.WildcardExpression:
		return true, nil
	case *ast.IntegerLiteral:
		if iv, ok := subject.(*IntValue); ok {
			return iv.Value == p.Value, nil
		}
		return false, nil
	case *ast.StringLiteral:
		if sv, ok := subject.(*StringValue); ok {
			return sv.Value == p.Value, nil
		}
		return false, nil
	case *ast.BooleanLiteral:
		if bv, ok := subject.(*BoolValue); ok {
			return bv.Value == p.Value, nil
		}
		return false, nil
	case *ast.NoneLiteral:
		_, ok := subject.(*NoneValue)
		return ok, nil
	case *ast.RangeExpression:
		// Check if subject is within range
		startVal, err := Eval(p.Start, env)
		if err != nil {
			return false, err
		}
		endVal, err := Eval(p.End, env)
		if err != nil {
			return false, err
		}
		sv, ok := subject.(*IntValue)
		if !ok {
			return false, nil
		}
		start, ok1 := startVal.(*IntValue)
		end, ok2 := endVal.(*IntValue)
		if !ok1 || !ok2 {
			return false, nil
		}
		return sv.Value >= start.Value && sv.Value < end.Value, nil
	case *ast.Identifier:
		// Binding pattern: bind subject to identifier name in env
		env.Define(p.Name, subject, false)
		return true, nil
	default:
		return false, nil
	}
}

// ---------------------------------------------------------------------------
// Expression Evaluation
// ---------------------------------------------------------------------------

func evalIdentifier(node *ast.Identifier, env *Environment) (Value, error) {
	val, ok := env.Get(node.Name)
	if !ok {
		return nil, &RuntimeError{
			Message: fmt.Sprintf("undefined variable '%s'", node.Name),
			Pos:     node.Pos.String(),
		}
	}
	return val, nil
}

func evalBinaryExpression(node *ast.BinaryExpression, env *Environment) (Value, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}

	// Short-circuit for logical operators
	if node.Operator == "&&" {
		if !IsTruthy(left) {
			return left, nil
		}
		return Eval(node.Right, env)
	}
	if node.Operator == "||" {
		if IsTruthy(left) {
			return left, nil
		}
		return Eval(node.Right, env)
	}

	right, err := Eval(node.Right, env)
	if err != nil {
		return nil, err
	}

	// Integer arithmetic
	if lv, ok := left.(*IntValue); ok {
		if rv, ok := right.(*IntValue); ok {
			return evalIntBinaryOp(node.Operator, lv.Value, rv.Value, node.Pos.String())
		}
		// Int + Float → promote to Float
		if rv, ok := right.(*FloatValue); ok {
			return evalFloatBinaryOp(node.Operator, float64(lv.Value), rv.Value, node.Pos.String())
		}
	}

	// Float arithmetic
	if lv, ok := left.(*FloatValue); ok {
		rv_f := float64(0)
		switch rv := right.(type) {
		case *FloatValue:
			rv_f = rv.Value
		case *IntValue:
			rv_f = float64(rv.Value)
		default:
			return nil, &RuntimeError{Message: fmt.Sprintf("cannot apply '%s' to float and %s", node.Operator, right.Type()), Pos: node.Pos.String()}
		}
		return evalFloatBinaryOp(node.Operator, lv.Value, rv_f, node.Pos.String())
	}

	// String concatenation
	if lv, ok := left.(*StringValue); ok {
		if rv, ok := right.(*StringValue); ok {
			if node.Operator == "+" {
				return NewString(lv.Value + rv.Value), nil
			}
		}
	}

	// Equality for any types
	if node.Operator == "==" {
		return NewBool(left.Equals(right)), nil
	}
	if node.Operator == "!=" {
		return NewBool(!left.Equals(right)), nil
	}

	return nil, &RuntimeError{
		Message: fmt.Sprintf("unsupported operator '%s' for types '%s' and '%s'", node.Operator, left.Type(), right.Type()),
		Pos:     node.Pos.String(),
	}
}

func evalIntBinaryOp(op string, left, right int64, pos string) (Value, error) {
	switch op {
	case "+":
		return NewInt(left + right), nil
	case "-":
		return NewInt(left - right), nil
	case "*":
		return NewInt(left * right), nil
	case "/":
		if right == 0 {
			return nil, &RuntimeError{Message: "division by zero", Pos: pos}
		}
		return NewInt(left / right), nil
	case "%":
		if right == 0 {
			return nil, &RuntimeError{Message: "modulo by zero", Pos: pos}
		}
		return NewInt(left % right), nil
	case "<":
		return NewBool(left < right), nil
	case ">":
		return NewBool(left > right), nil
	case "<=":
		return NewBool(left <= right), nil
	case ">=":
		return NewBool(left >= right), nil
	case "==":
		return NewBool(left == right), nil
	case "!=":
		return NewBool(left != right), nil
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown operator '%s' for int", op), Pos: pos}
	}
}

func evalFloatBinaryOp(op string, left, right float64, pos string) (Value, error) {
	switch op {
	case "+":
		return NewFloat(left + right), nil
	case "-":
		return NewFloat(left - right), nil
	case "*":
		return NewFloat(left * right), nil
	case "/":
		if right == 0 {
			return nil, &RuntimeError{Message: "division by zero", Pos: pos}
		}
		return NewFloat(left / right), nil
	case "<":
		return NewBool(left < right), nil
	case ">":
		return NewBool(left > right), nil
	case "<=":
		return NewBool(left <= right), nil
	case ">=":
		return NewBool(left >= right), nil
	case "==":
		return NewBool(left == right), nil
	case "!=":
		return NewBool(left != right), nil
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown operator '%s' for float", op), Pos: pos}
	}
}

func evalUnaryExpression(node *ast.UnaryExpression, env *Environment) (Value, error) {
	operand, err := Eval(node.Operand, env)
	if err != nil {
		return nil, err
	}

	switch node.Operator {
	case "-":
		switch v := operand.(type) {
		case *IntValue:
			return NewInt(-v.Value), nil
		case *FloatValue:
			return NewFloat(-v.Value), nil
		default:
			return nil, &RuntimeError{Message: fmt.Sprintf("cannot negate '%s'", operand.Type()), Pos: node.Pos.String()}
		}
	case "!":
		return NewBool(!IsTruthy(operand)), nil
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("unknown unary operator '%s'", node.Operator), Pos: node.Pos.String()}
	}
}

func evalCallExpression(node *ast.CallExpression, env *Environment) (Value, error) {
	fn, err := Eval(node.Function, env)
	if err != nil {
		return nil, err
	}

	args := make([]Value, len(node.Arguments))
	for i, arg := range node.Arguments {
		val, err := Eval(arg, env)
		if err != nil {
			return nil, err
		}
		args[i] = val
	}

	return callFunction(fn, args, node.Pos.String())
}

func callFunction(fn Value, args []Value, pos string) (Value, error) {
	switch f := fn.(type) {
	case *FnValue:
		if len(args) != len(f.Params) {
			return nil, &RuntimeError{
				Message: fmt.Sprintf("%s() takes %d arguments, got %d", f.Name, len(f.Params), len(args)),
				Pos:     pos,
			}
		}
		fnEnv := NewEnclosedEnvironment(f.Env)
		for i, param := range f.Params {
			fnEnv.Define(param, args[i], false)
		}
		body, ok := f.Body.(*ast.BlockStatement)
		if !ok {
			return nil, &RuntimeError{Message: "invalid function body", Pos: pos}
		}
		result, err := Eval(body, fnEnv)
		if err != nil {
			if rs, ok := err.(*ReturnSignal); ok {
				if len(rs.Values) == 0 {
					return NONE, nil
				}
				if len(rs.Values) == 1 {
					return rs.Values[0], nil
				}
				// Multiple return values → return as array
				return NewArray(rs.Values), nil
			}
			return nil, err
		}
		return result, nil
	case *BuiltinFn:
		return f.Fn(args)
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("'%s' is not callable", fn.Type()), Pos: pos}
	}
}

func evalIndexExpression(node *ast.IndexExpression, env *Environment) (Value, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}
	index, err := Eval(node.Index, env)
	if err != nil {
		return nil, err
	}

	switch target := left.(type) {
	case *ArrayValue:
		idx, ok := index.(*IntValue)
		if !ok {
			return nil, &RuntimeError{Message: "array index must be an integer", Pos: node.Pos.String()}
		}
		i := int(idx.Value)
		if i < 0 || i >= len(target.Elements) {
			return nil, &RuntimeError{Message: fmt.Sprintf("index %d out of bounds (len %d)", i, len(target.Elements)), Pos: node.Pos.String()}
		}
		return target.Elements[i], nil
	case *MapValue:
		key, ok := index.(*StringValue)
		if !ok {
			return nil, &RuntimeError{Message: "map key must be a string", Pos: node.Pos.String()}
		}
		val, exists := target.Pairs[key.Value]
		if !exists {
			return NONE, nil
		}
		return val, nil
	case *RangeValue:
		idx, ok := index.(*IntValue)
		if !ok {
			return nil, &RuntimeError{Message: "range index must be an integer", Pos: node.Pos.String()}
		}
		i := idx.Value
		if i < 0 || i >= target.Len() {
			return nil, &RuntimeError{Message: fmt.Sprintf("range index %d out of bounds", i), Pos: node.Pos.String()}
		}
		return NewInt(target.At(i)), nil
	case *StringValue:
		idx, ok := index.(*IntValue)
		if !ok {
			return nil, &RuntimeError{Message: "string index must be an integer", Pos: node.Pos.String()}
		}
		i := int(idx.Value)
		if i < 0 || i >= len(target.Value) {
			return nil, &RuntimeError{Message: fmt.Sprintf("string index %d out of bounds", i), Pos: node.Pos.String()}
		}
		return NewString(string(target.Value[i])), nil
	default:
		return nil, &RuntimeError{Message: fmt.Sprintf("cannot index into '%s'", left.Type()), Pos: node.Pos.String()}
	}
}

func evalDotExpression(node *ast.DotExpression, env *Environment) (Value, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}

	if m, ok := left.(*MapValue); ok {
		val, exists := m.Pairs[node.Field]
		if !exists {
			return NONE, nil
		}
		return val, nil
	}

	return nil, &RuntimeError{
		Message: fmt.Sprintf("cannot access field '%s' on type '%s'", node.Field, left.Type()),
		Pos:     node.Pos.String(),
	}
}

func evalSafeAccessExpression(node *ast.SafeAccessExpression, env *Environment) (Value, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}

	if _, ok := left.(*NoneValue); ok {
		return NONE, nil
	}

	if m, ok := left.(*MapValue); ok {
		val, exists := m.Pairs[node.Field]
		if !exists {
			return NONE, nil
		}
		return val, nil
	}

	return nil, &RuntimeError{
		Message: fmt.Sprintf("cannot safe-access field '%s' on type '%s'", node.Field, left.Type()),
		Pos:     node.Pos.String(),
	}
}

func evalArrayLiteral(node *ast.ArrayLiteral, env *Environment) (Value, error) {
	elements := make([]Value, len(node.Elements))
	for i, el := range node.Elements {
		val, err := Eval(el, env)
		if err != nil {
			return nil, err
		}
		elements[i] = val
	}
	return NewArray(elements), nil
}

func evalMapLiteral(node *ast.MapLiteral, env *Environment) (Value, error) {
	pairs := make(map[string]Value)
	for i, keyExpr := range node.Keys {
		keyVal, err := Eval(keyExpr, env)
		if err != nil {
			return nil, err
		}
		key, ok := keyVal.(*StringValue)
		if !ok {
			return nil, &RuntimeError{Message: "map key must be a string", Pos: node.Pos.String()}
		}
		val, err := Eval(node.Values[i], env)
		if err != nil {
			return nil, err
		}
		pairs[key.Value] = val
	}
	return NewMap(pairs), nil
}

func evalFnLiteral(node *ast.FnLiteral, env *Environment) (Value, error) {
	return &FnValue{
		Params: node.Params,
		Body:   node.Body,
		Env:    env,
	}, nil
}

func evalRangeExpression(node *ast.RangeExpression, env *Environment) (Value, error) {
	startVal, err := Eval(node.Start, env)
	if err != nil {
		return nil, err
	}
	endVal, err := Eval(node.End, env)
	if err != nil {
		return nil, err
	}

	start, ok1 := startVal.(*IntValue)
	end, ok2 := endVal.(*IntValue)
	if !ok1 || !ok2 {
		return nil, &RuntimeError{Message: "range bounds must be integers", Pos: node.Pos.String()}
	}

	step := int64(1)
	if node.Step != nil {
		stepVal, err := Eval(node.Step, env)
		if err != nil {
			return nil, err
		}
		s, ok := stepVal.(*IntValue)
		if !ok {
			return nil, &RuntimeError{Message: "range step must be an integer", Pos: node.Pos.String()}
		}
		step = s.Value
	}

	return &RangeValue{Start: start.Value, End: end.Value, Step: step}, nil
}

func evalPipelineExpression(node *ast.PipelineExpression, env *Environment) (Value, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}

	// Evaluate the right-hand call's function
	fn, err := Eval(node.Right.Function, env)
	if err != nil {
		return nil, err
	}

	// Evaluate remaining args (the ones explicitly written)
	args := make([]Value, 0, len(node.Right.Arguments)+1)
	args = append(args, left) // pipe value is first arg
	for _, arg := range node.Right.Arguments {
		val, err := Eval(arg, env)
		if err != nil {
			return nil, err
		}
		args = append(args, val)
	}

	return callFunction(fn, args, node.Pos.String())
}

func evalCoalesceExpression(node *ast.CoalesceExpression, env *Environment) (Value, error) {
	left, err := Eval(node.Left, env)
	if err != nil {
		return nil, err
	}
	if _, isNone := left.(*NoneValue); !isNone {
		return left, nil
	}
	return Eval(node.Right, env)
}

func evalStringInterpolation(node *ast.StringInterpolation, env *Environment) (Value, error) {
	var result string
	for _, part := range node.Parts {
		val, err := Eval(part, env)
		if err != nil {
			return nil, err
		}
		result += val.String()
	}
	return NewString(result), nil
}

// ---------------------------------------------------------------------------
// Test Runner (used by `glace test` command)
// ---------------------------------------------------------------------------

// TestResult holds the outcome of a single test block.
type TestResult struct {
	Name   string
	Passed bool
	Error  string
}

// RunTests evaluates a program and runs only its test blocks.
func RunTests(program *ast.Program, env *Environment) []TestResult {
	results := make([]TestResult, 0)

	// First evaluate all non-test statements (to define functions etc.)
	for _, stmt := range program.Statements {
		if _, isTest := stmt.(*ast.TestBlock); isTest {
			continue
		}
		Eval(stmt, env)
	}

	// Then run test blocks
	for _, stmt := range program.Statements {
		tb, ok := stmt.(*ast.TestBlock)
		if !ok {
			continue
		}
		testEnv := NewEnclosedEnvironment(env)
		_, err := Eval(tb.Body, testEnv)
		result := TestResult{Name: tb.Description, Passed: err == nil}
		if err != nil {
			result.Error = err.Error()
		}
		results = append(results, result)
	}

	return results
}

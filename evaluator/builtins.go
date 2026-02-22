package evaluator

import (
	"fmt"
	"strings"
)

// RegisterBuiltins populates the given environment with all built-in functions.
func RegisterBuiltins(env *Environment) {
	builtins := []*BuiltinFn{
		builtinPrint(),
		builtinLen(),
		builtinPush(),
		builtinPop(),
		builtinTypeOf(),
		builtinStr(),
		builtinInt(),
		builtinFloat(),
		builtinInput(),
		builtinAssert(),
		builtinArray(),
	}

	for _, b := range builtins {
		env.Define(b.Name, b, false)
	}
}

// ---------------- Built-in Implementations ----------------

func builtinPrint() *BuiltinFn {
	return &BuiltinFn{
		Name: "print",
		Fn: func(args []Value) (Value, error) {
			parts := make([]string, len(args))
			for i, a := range args {
				parts[i] = a.String()
			}
			fmt.Println(strings.Join(parts, " "))
			return NONE, nil
		},
	}
}

func builtinLen() *BuiltinFn {
	return &BuiltinFn{
		Name: "len",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("len() takes exactly 1 argument, got %d", len(args))
			}
			switch v := args[0].(type) {
			case *StringValue:
				return NewInt(int64(len(v.Value))), nil
			case *ArrayValue:
				return NewInt(int64(len(v.Elements))), nil
			case *MapValue:
				return NewInt(int64(len(v.Pairs))), nil
			case *RangeValue:
				return NewInt(v.Len()), nil
			default:
				return nil, fmt.Errorf("len() not supported for type '%s'", v.Type())
			}
		},
	}
}

func builtinPush() *BuiltinFn {
	return &BuiltinFn{
		Name: "push",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 2 {
				return nil, fmt.Errorf("push() takes 2 arguments (array, value), got %d", len(args))
			}
			arr, ok := args[0].(*ArrayValue)
			if !ok {
				return nil, fmt.Errorf("push() first argument must be an array, got '%s'", args[0].Type())
			}
			arr.Elements = append(arr.Elements, args[1])
			return arr, nil
		},
	}
}

func builtinPop() *BuiltinFn {
	return &BuiltinFn{
		Name: "pop",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("pop() takes 1 argument, got %d", len(args))
			}
			arr, ok := args[0].(*ArrayValue)
			if !ok {
				return nil, fmt.Errorf("pop() argument must be an array, got '%s'", args[0].Type())
			}
			if len(arr.Elements) == 0 {
				return nil, fmt.Errorf("pop() on empty array")
			}
			last := arr.Elements[len(arr.Elements)-1]
			arr.Elements = arr.Elements[:len(arr.Elements)-1]
			return last, nil
		},
	}
}

func builtinTypeOf() *BuiltinFn {
	return &BuiltinFn{
		Name: "type",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("type() takes 1 argument, got %d", len(args))
			}
			return NewString(args[0].Type()), nil
		},
	}
}

func builtinStr() *BuiltinFn {
	return &BuiltinFn{
		Name: "str",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("str() takes 1 argument, got %d", len(args))
			}
			return NewString(args[0].String()), nil
		},
	}
}

func builtinInt() *BuiltinFn {
	return &BuiltinFn{
		Name: "int",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("int() takes 1 argument, got %d", len(args))
			}
			switch v := args[0].(type) {
			case *IntValue:
				return v, nil
			case *FloatValue:
				return NewInt(int64(v.Value)), nil
			case *StringValue:
				var n int64
				_, err := fmt.Sscanf(v.Value, "%d", &n)
				if err != nil {
					return nil, fmt.Errorf("cannot convert '%s' to int", v.Value)
				}
				return NewInt(n), nil
			case *BoolValue:
				if v.Value {
					return NewInt(1), nil
				}
				return NewInt(0), nil
			default:
				return nil, fmt.Errorf("cannot convert '%s' to int", v.Type())
			}
		},
	}
}

func builtinFloat() *BuiltinFn {
	return &BuiltinFn{
		Name: "float",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("float() takes 1 argument, got %d", len(args))
			}
			switch v := args[0].(type) {
			case *FloatValue:
				return v, nil
			case *IntValue:
				return NewFloat(float64(v.Value)), nil
			case *StringValue:
				var f float64
				_, err := fmt.Sscanf(v.Value, "%f", &f)
				if err != nil {
					return nil, fmt.Errorf("cannot convert '%s' to float", v.Value)
				}
				return NewFloat(f), nil
			default:
				return nil, fmt.Errorf("cannot convert '%s' to float", v.Type())
			}
		},
	}
}

func builtinInput() *BuiltinFn {
	return &BuiltinFn{
		Name: "input",
		Fn: func(args []Value) (Value, error) {
			// Optional prompt
			if len(args) > 0 {
				fmt.Print(args[0].String())
			}
			var line string
			fmt.Scanln(&line)
			return NewString(line), nil
		},
	}
}

func builtinAssert() *BuiltinFn {
	return &BuiltinFn{
		Name: "assert",
		Fn: func(args []Value) (Value, error) {
			if len(args) < 1 {
				return nil, fmt.Errorf("assert() takes at least 1 argument")
			}
			if !IsTruthy(args[0]) {
				msg := "assertion failed"
				if len(args) >= 2 {
					msg = args[1].String()
				}
				return nil, fmt.Errorf("%s", msg)
			}
			return NONE, nil
		},
	}
}

func builtinArray() *BuiltinFn {
	return &BuiltinFn{
		Name: "array",
		Fn: func(args []Value) (Value, error) {
			if len(args) != 1 {
				return nil, fmt.Errorf("array() takes 1 argument, got %d", len(args))
			}
			r, ok := args[0].(*RangeValue)
			if !ok {
				return nil, fmt.Errorf("array() argument must be a range, got '%s'", args[0].Type())
			}
			elements := make([]Value, 0, r.Len())
			for i := r.Start; i < r.End; i += r.Step {
				elements = append(elements, NewInt(i))
			}
			return NewArray(elements), nil
		},
	}
}

package evaluator

import "fmt"

// ---------------------------------------------------------------------------
// Value Interface
// ---------------------------------------------------------------------------

// Value is the interface for all runtime values in Glace.
type Value interface {
	Type() string  // "int", "float", "bool", "string", "array", "map", "fn", "range", "none"
	String() string // human-readable representation
	Equals(other Value) bool
}

// IsTruthy returns the boolean truthiness of a value.
// none and false are falsy; everything else is truthy.
func IsTruthy(v Value) bool {
	switch val := v.(type) {
	case *NoneValue:
		return false
	case *BoolValue:
		return val.Value
	case *IntValue:
		return val.Value != 0
	case *StringValue:
		return val.Value != ""
	case *ArrayValue:
		return len(val.Elements) > 0
	default:
		return true
	}
}

// ---------------------------------------------------------------------------
// Concrete Value Types
// ---------------------------------------------------------------------------

// IntValue represents an integer.
type IntValue struct {
	Value int64
}

func (v *IntValue) Type() string   { return "int" }
func (v *IntValue) String() string { return fmt.Sprintf("%d", v.Value) }
func (v *IntValue) Equals(other Value) bool {
	if o, ok := other.(*IntValue); ok {
		return v.Value == o.Value
	}
	return false
}

// FloatValue represents a floating-point number.
type FloatValue struct {
	Value float64
}

func (v *FloatValue) Type() string   { return "float" }
func (v *FloatValue) String() string { return fmt.Sprintf("%g", v.Value) }
func (v *FloatValue) Equals(other Value) bool {
	if o, ok := other.(*FloatValue); ok {
		return v.Value == o.Value
	}
	return false
}

// BoolValue represents a boolean.
type BoolValue struct {
	Value bool
}

func (v *BoolValue) Type() string   { return "bool" }
func (v *BoolValue) String() string { return fmt.Sprintf("%t", v.Value) }
func (v *BoolValue) Equals(other Value) bool {
	if o, ok := other.(*BoolValue); ok {
		return v.Value == o.Value
	}
	return false
}

// StringValue represents a string.
type StringValue struct {
	Value string
}

func (v *StringValue) Type() string   { return "string" }
func (v *StringValue) String() string { return v.Value }
func (v *StringValue) Equals(other Value) bool {
	if o, ok := other.(*StringValue); ok {
		return v.Value == o.Value
	}
	return false
}

// NoneValue represents the absence of a value.
type NoneValue struct{}

func (v *NoneValue) Type() string          { return "none" }
func (v *NoneValue) String() string        { return "none" }
func (v *NoneValue) Equals(other Value) bool {
	_, ok := other.(*NoneValue)
	return ok
}

// ArrayValue represents a dynamically-sized array.
type ArrayValue struct {
	Elements []Value
}

func (v *ArrayValue) Type() string { return "array" }
func (v *ArrayValue) String() string {
	s := "["
	for i, e := range v.Elements {
		if i > 0 {
			s += ", "
		}
		s += e.String()
	}
	return s + "]"
}
func (v *ArrayValue) Equals(other Value) bool {
	o, ok := other.(*ArrayValue)
	if !ok || len(v.Elements) != len(o.Elements) {
		return false
	}
	for i, e := range v.Elements {
		if !e.Equals(o.Elements[i]) {
			return false
		}
	}
	return true
}

// MapValue represents a key-value map.
type MapValue struct {
	Pairs map[string]Value
}

func (v *MapValue) Type() string { return "map" }
func (v *MapValue) String() string {
	s := "{"
	i := 0
	for k, val := range v.Pairs {
		if i > 0 {
			s += ", "
		}
		s += fmt.Sprintf("%q: %s", k, val.String())
		i++
	}
	return s + "}"
}
func (v *MapValue) Equals(other Value) bool {
	o, ok := other.(*MapValue)
	if !ok || len(v.Pairs) != len(o.Pairs) {
		return false
	}
	for k, val := range v.Pairs {
		oval, exists := o.Pairs[k]
		if !exists || !val.Equals(oval) {
			return false
		}
	}
	return true
}

// FnValue represents a function (named or anonymous).
type FnValue struct {
	Name   string   // "" for anonymous functions
	Params []string
	Body   interface{} // *ast.BlockStatement — kept as interface to avoid import cycle
	Env    *Environment
}

func (v *FnValue) Type() string   { return "fn" }
func (v *FnValue) String() string {
	if v.Name != "" {
		return fmt.Sprintf("<fn %s>", v.Name)
	}
	return "<fn>"
}
func (v *FnValue) Equals(other Value) bool { return v == other } // identity comparison

// RangeValue represents a lazy integer range.
type RangeValue struct {
	Start int64
	End   int64
	Step  int64
}

func (v *RangeValue) Type() string   { return "range" }
func (v *RangeValue) String() string {
	if v.Step != 1 {
		return fmt.Sprintf("%d..%d step %d", v.Start, v.End, v.Step)
	}
	return fmt.Sprintf("%d..%d", v.Start, v.End)
}
func (v *RangeValue) Equals(other Value) bool {
	if o, ok := other.(*RangeValue); ok {
		return v.Start == o.Start && v.End == o.End && v.Step == o.Step
	}
	return false
}

// Len returns the number of elements in the range.
func (v *RangeValue) Len() int64 {
	if v.Step > 0 && v.End > v.Start {
		return (v.End - v.Start + v.Step - 1) / v.Step
	}
	if v.Step < 0 && v.Start > v.End {
		return (v.Start - v.End - v.Step - 1) / (-v.Step)
	}
	return 0
}

// At returns the element at position i.
func (v *RangeValue) At(i int64) int64 {
	return v.Start + i*v.Step
}

// BuiltinFn represents a built-in function implemented in Go.
type BuiltinFn struct {
	Name string
	Fn   func(args []Value) (Value, error)
}

func (v *BuiltinFn) Type() string          { return "builtin" }
func (v *BuiltinFn) String() string        { return fmt.Sprintf("<builtin %s>", v.Name) }
func (v *BuiltinFn) Equals(other Value) bool { return v == other }

// ---------------------------------------------------------------------------
// Convenience Constructors
// ---------------------------------------------------------------------------

var (
	TRUE  = &BoolValue{Value: true}
	FALSE = &BoolValue{Value: false}
	NONE  = &NoneValue{}
)

func NewInt(val int64) *IntValue       { return &IntValue{Value: val} }
func NewFloat(val float64) *FloatValue { return &FloatValue{Value: val} }
func NewString(val string) *StringValue { return &StringValue{Value: val} }
func NewBool(val bool) *BoolValue {
	if val {
		return TRUE
	}
	return FALSE
}
func NewArray(elements []Value) *ArrayValue { return &ArrayValue{Elements: elements} }
func NewMap(pairs map[string]Value) *MapValue { return &MapValue{Pairs: pairs} }

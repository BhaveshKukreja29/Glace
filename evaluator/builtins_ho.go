package evaluator

import (
    "fmt"
    "sort"
)

// builtinFilter returns a new array containing only elements for which fn(elem) is truthy.
func builtinFilter() *BuiltinFn {
    return &BuiltinFn{
        Name: "filter",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 2 {
                return nil, fmt.Errorf("filter expects 2 arguments (array, fn), got %d", len(args))
            }
            arr, ok := args[0].(*ArrayValue)
            if !ok {
                return nil, fmt.Errorf("filter: first argument must be an array, got %s", args[0].Type())
            }
            fn := args[1]
            if _, ok1 := fn.(*FnValue); !ok1 {
                if _, ok2 := fn.(*BuiltinFn); !ok2 {
                    return nil, fmt.Errorf("filter: second argument must be a function, got %s", fn.Type())
                }
            }
            result := make([]Value, 0)
            for _, elem := range arr.Elements {
                res, err := callFunction(fn, []Value{elem}, "filter")
                if err != nil {
                    return nil, err
                }
                if IsTruthy(res) {
                    result = append(result, elem)
                }
            }
            return NewArray(result), nil
        },
    }
}

func builtinMap() *BuiltinFn {
    return &BuiltinFn{
        Name: "map",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 2 {
                return nil, fmt.Errorf("map expects 2 arguments (array, fn), got %d", len(args))
            }
            arr, ok := args[0].(*ArrayValue)
            if !ok {
                return nil, fmt.Errorf("map: first argument must be an array, got %s", args[0].Type())
            }
            fn := args[1]
            if _, ok1 := fn.(*FnValue); !ok1 {
                if _, ok2 := fn.(*BuiltinFn); !ok2 {
                    return nil, fmt.Errorf("map: second argument must be a function, got %s", fn.Type())
                }
            }
            result := make([]Value, len(arr.Elements))
            for i, elem := range arr.Elements {
                res, err := callFunction(fn, []Value{elem}, "map")
                if err != nil {
                    return nil, err
                }
                result[i] = res
            }
            return NewArray(result), nil
        },
    }
}

func builtinReduce() *BuiltinFn {
    return &BuiltinFn{
        Name: "reduce",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 3 {
                return nil, fmt.Errorf("reduce expects 3 arguments (array, initial, fn), got %d", len(args))
            }
            arr, ok := args[0].(*ArrayValue)
            if !ok {
                return nil, fmt.Errorf("reduce: first argument must be an array, got %s", args[0].Type())
            }
            acc := args[1]
            fn := args[2]
            if _, ok1 := fn.(*FnValue); !ok1 {
                if _, ok2 := fn.(*BuiltinFn); !ok2 {
                    return nil, fmt.Errorf("reduce: third argument must be a function, got %s", fn.Type())
                }
            }
            var err error
            for _, elem := range arr.Elements {
                acc, err = callFunction(fn, []Value{acc, elem}, "reduce")
                if err != nil {
                    return nil, err
                }
            }
            return acc, nil
        },
    }
}

func builtinSort() *BuiltinFn {
    return &BuiltinFn{
        Name: "sort",
        Fn: func(args []Value) (Value, error) {
            if len(args) < 1 || len(args) > 2 {
                return nil, fmt.Errorf("sort expects 1 or 2 arguments (array [, fn]), got %d", len(args))
            }
            arr, ok := args[0].(*ArrayValue)
            if !ok {
                return nil, fmt.Errorf("sort: first argument must be an array, got %s", args[0].Type())
            }
            copied := make([]Value, len(arr.Elements))
            copy(copied, arr.Elements)

            var sortErr error

            if len(args) == 2 {
                fn := args[1]
                if _, ok1 := fn.(*FnValue); !ok1 {
                    if _, ok2 := fn.(*BuiltinFn); !ok2 {
                        return nil, fmt.Errorf("sort: second argument must be a function, got %s", fn.Type())
                    }
                }
                sort.SliceStable(copied, func(i, j int) bool {
                    if sortErr != nil {
                        return false
                    }
                    res, err := callFunction(fn, []Value{copied[i], copied[j]}, "sort")
                    if err != nil {
                        sortErr = err
                        return false
                    }
                    switch v := res.(type) {
                    case *IntValue:
                        return v.Value < 0
                    case *FloatValue:
                        return v.Value < 0
                    default:
                        sortErr = fmt.Errorf("sort comparator must return a number, got %s", res.Type())
                        return false
                    }
                })
            } else {
                sort.SliceStable(copied, func(i, j int) bool {
                    if sortErr != nil {
                        return false
                    }
                    a, b := copied[i], copied[j]
                    switch av := a.(type) {
                    case *IntValue:
                        if bv, ok := b.(*IntValue); ok {
                            return av.Value < bv.Value
                        }
                        if bv, ok := b.(*FloatValue); ok {
                            return float64(av.Value) < bv.Value
                        }
                    case *FloatValue:
                        if bv, ok := b.(*FloatValue); ok {
                            return av.Value < bv.Value
                        }
                        if bv, ok := b.(*IntValue); ok {
                            return av.Value < float64(bv.Value)
                        }
                    case *StringValue:
                        if bv, ok := b.(*StringValue); ok {
                            return av.Value < bv.Value
                        }
                    }
                    sortErr = fmt.Errorf("sort: cannot compare %s and %s", a.Type(), b.Type())
                    return false
                })
            }
            if sortErr != nil {
                return nil, sortErr
            }
            return NewArray(copied), nil
        },
    }
}

func builtinKeys() *BuiltinFn {
    return &BuiltinFn{
        Name: "keys",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("keys expects 1 argument (map), got %d", len(args))
            }
            m, ok := args[0].(*MapValue)
            if !ok {
                return nil, fmt.Errorf("keys: argument must be a map, got %s", args[0].Type())
            }
            keys := make([]Value, 0, len(m.Pairs))
            for k := range m.Pairs {
                keys = append(keys, NewString(k))
            }
            sort.Slice(keys, func(i, j int) bool {
                return keys[i].(*StringValue).Value < keys[j].(*StringValue).Value
            })
            return NewArray(keys), nil
        },
    }
}

func builtinValues() *BuiltinFn {
    return &BuiltinFn{
        Name: "values",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("values expects 1 argument (map), got %d", len(args))
            }
            m, ok := args[0].(*MapValue)
            if !ok {
                return nil, fmt.Errorf("values: argument must be a map, got %s", args[0].Type())
            }
            keys := make([]string, 0, len(m.Pairs))
            for k := range m.Pairs {
                keys = append(keys, k)
            }
            sort.Strings(keys)
            vals := make([]Value, 0, len(m.Pairs))
            for _, k := range keys {
                vals = append(vals, m.Pairs[k])
            }
            return NewArray(vals), nil
        },
    }
}

func builtinHas() *BuiltinFn {
    return &BuiltinFn{
        Name: "has",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 2 {
                return nil, fmt.Errorf("has expects 2 arguments (map, key), got %d", len(args))
            }
            m, ok := args[0].(*MapValue)
            if !ok {
                return nil, fmt.Errorf("has: first argument must be a map, got %s", args[0].Type())
            }
            key, ok := args[1].(*StringValue)
            if !ok {
                return nil, fmt.Errorf("has: second argument must be a string key, got %s", args[1].Type())
            }
            _, exists := m.Pairs[key.Value]
            return NewBool(exists), nil
        },
    }
}

func builtinReverse() *BuiltinFn {
    return &BuiltinFn{
        Name: "reverse",
        Fn: func(args []Value) (Value, error) {
            if len(args) != 1 {
                return nil, fmt.Errorf("reverse expects 1 argument (array), got %d", len(args))
            }
            arr, ok := args[0].(*ArrayValue)
            if !ok {
                return nil, fmt.Errorf("reverse: argument must be an array, got %s", args[0].Type())
            }
            n := len(arr.Elements)
            result := make([]Value, n)
            for i, v := range arr.Elements {
                result[n-1-i] = v
            }
            return NewArray(result), nil
        },
    }
}

// RegisterHOBuiltins registers all higher-order built-in functions into the environment.
func RegisterHOBuiltins(env *Environment) {
    env.Define("filter", builtinFilter(), false)
    env.Define("map", builtinMap(), false)
    env.Define("reduce", builtinReduce(), false)
    env.Define("sort", builtinSort(), false)
    env.Define("keys", builtinKeys(), false)
    env.Define("values", builtinValues(), false)
    env.Define("has", builtinHas(), false)
    env.Define("reverse", builtinReverse(), false)
}
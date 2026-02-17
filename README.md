# Glace

A lightweight interpreted language built in Go, prioritizing syntax cleanliness and developer ergonomics.

## Features

- **Immutable by default** — `let` for constants, `mut` for mutable variables
- **Unified `loop`** — one keyword replaces `for`, `while`, `do-while`
- **Pipeline operator `|>`** — chain function calls left-to-right
- **Pattern matching** — `match` expressions with literal, range, and wildcard patterns
- **First-class ranges** — `0..10 step 2` as values, not just syntax
- **No semicolons** — newline-based statement termination
- **Built-in testing** — `test` blocks with `assert`
- **String interpolation** — `"hello ${name}"`

## Quick Start

```bash
# Build
make build

# Start the REPL
./glace

# Run a file
./glace run examples/hello.glace

# Run test blocks in a file
./glace test examples/test_demo.glace
```

## Example — Quicksort

```
fn quicksort(arr) {
    if len(arr) <= 1 {
        return arr
    }

    let pivot = arr[0]
    mut left = []
    mut right = []

    loop i in 1..len(arr) {
        if arr[i] < pivot {
            push(left, arr[i])
        } else {
            push(right, arr[i])
        }
    }

    let sl = quicksort(left)
    let sr = quicksort(right)

    mut result = []
    loop item in sl { push(result, item) }
    push(result, pivot)
    loop item in sr { push(result, item) }
    return result
}

let data = [38, 27, 43, 3, 9, 82, 10]
print("Before: " + str(data))

let sorted = quicksort(data)
print("After:  " + str(sorted))
```

## Project Structure

```
.
├── main.go              # CLI entry point (REPL, run, test)
├── lexer/               # Tokenizer
│   ├── token.go         # Token types and definitions
│   └── lexer.go         # Scanner
├── ast/                 
│   └── ast.go           # AST node definitions
├── parser/              
│   ├── parser.go        # Recursive descent parser (Pratt)
│   └── precedence.go    # Operator precedence levels
├── evaluator/           
│   ├── value.go         # Runtime value types
│   ├── environment.go   # Scope chain
│   ├── evaluator.go     # Tree-walk interpreter
│   └── builtins.go      # Built-in functions
├── repl/                
│   └── repl.go          # Interactive REPL
└── examples/            # Example programs
```

## Built-in Functions

| Function | Description |
|---|---|
| `print(args...)` | Print values to stdout (newline-terminated) |
| `len(v)` | Length of string, array, map, or range |
| `push(arr, val)` | Append value to array |
| `pop(arr)` | Remove and return last element |
| `type(v)` | Return type name as string |
| `str(v)` | Convert to string |
| `int(v)` | Convert to integer |
| `float(v)` | Convert to float |
| `input(prompt?)` | Read line from stdin |
| `assert(cond, msg?)` | Assert condition is truthy |
| `array(range)` | Convert range to array |

## Requirements

- Go 1.21+

## License

MIT

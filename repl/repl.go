package repl

import (
	"bufio"
	"fmt"
	"io"
	"strings"

	"github.com/glace-lang/glace/evaluator"
	"github.com/glace-lang/glace/lexer"
	"github.com/glace-lang/glace/parser"
)

const PROMPT = "glace> "
const VERSION = "0.1.0"

func Start(in io.Reader, out io.Writer) {
	scanner := bufio.NewScanner(in)
	env := evaluator.NewEnvironment()
	evaluator.RegisterBuiltins(env)
	evaluator.RegisterHOBuiltins(env)

	fmt.Fprintf(out, "Glace v%s — type 'exit' to quit\n", VERSION)

	for {
		fmt.Fprint(out, PROMPT)
		if !scanner.Scan() {
			break
		}
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if line == "exit" || line == "quit" {
			fmt.Fprintln(out, "bye!")
			break
		}

		tokens := lexer.New(line, "<repl>").Tokenize()
		program, errors := parser.Parse(tokens)
		if len(errors) > 0 {
			for _, e := range errors {
				fmt.Fprintf(out, "  parse error: %s\n", e)
			}
			continue
		}

		result, err := evaluator.Eval(program, env)
		if err != nil {
			fmt.Fprintf(out, "  error: %s\n", err)
			continue
		}
		if result != nil && result.Type() != "none" {
			fmt.Fprintf(out, "=> %s\n", result.String())
		}
	}
}

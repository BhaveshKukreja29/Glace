package main

import (
	"fmt"
	"os"

	"github.com/glace-lang/glace/evaluator"
	"github.com/glace-lang/glace/lexer"
	"github.com/glace-lang/glace/parser"
	"github.com/glace-lang/glace/repl"
)

func main() {
	args := os.Args[1:]

	if len(args) == 0 {
		repl.Start(os.Stdin, os.Stdout)
		return
	}

	switch args[0] {
	case "run":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: glace run <file.glace>")
			os.Exit(1)
		}
		runFile(args[1])

	case "test":
		if len(args) < 2 {
			fmt.Fprintln(os.Stderr, "usage: glace test <file.glace>")
			os.Exit(1)
		}
		testFile(args[1])

	case "--version", "-v":
		fmt.Printf("Glace v%s\n", repl.VERSION)

	case "--help", "-h":
		printHelp()

	default:
		// Treat as filename
		runFile(args[0])
	}
}

func runFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	tokens := lexer.New(string(source), path).Tokenize()
	program, errors := parser.Parse(tokens)
	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
		}
		os.Exit(1)
	}

	env := evaluator.NewEnvironment()
	evaluator.RegisterBuiltins(env)

	_, evalErr := evaluator.Eval(program, env)
	if evalErr != nil {
		fmt.Fprintf(os.Stderr, "%s\n", evalErr)
		os.Exit(1)
	}
}

func testFile(path string) {
	source, err := os.ReadFile(path)
	if err != nil {
		fmt.Fprintf(os.Stderr, "error: %s\n", err)
		os.Exit(1)
	}

	tokens := lexer.New(string(source), path).Tokenize()
	program, errors := parser.Parse(tokens)
	if len(errors) > 0 {
		for _, e := range errors {
			fmt.Fprintf(os.Stderr, "parse error: %s\n", e)
		}
		os.Exit(1)
	}

	env := evaluator.NewEnvironment()
	evaluator.RegisterBuiltins(env)

	results := evaluator.RunTests(program, env)
	passed, failed := 0, 0
	for _, r := range results {
		if r.Passed {
			fmt.Printf("  PASS: %s\n", r.Name)
			passed++
		} else {
			fmt.Printf("  FAIL: %s — %s\n", r.Name, r.Error)
			failed++
		}
	}
	fmt.Printf("\n%d passed, %d failed\n", passed, failed)
	if failed > 0 {
		os.Exit(1)
	}
}

func printHelp() {
	fmt.Println(`Glace — A lightweight interpreted language

Usage:
  glace                   Start the REPL
  glace run <file>        Execute a .glace file
  glace test <file>       Run test blocks in a .glace file
  glace --version         Print version
  glace --help            Print this help`)
}

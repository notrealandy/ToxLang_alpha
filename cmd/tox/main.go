package main

import (
	"fmt"
	"os"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/evaluator"
	"github.com/notrealandy/tox/lexer"
	"github.com/notrealandy/tox/parser"
	"github.com/notrealandy/tox/typechecker"
)

func main() {
	// Usage instructions
	if len(os.Args) < 2 || os.Args[1] != "run" {
		fmt.Println("Usage: tox run <path>")
		os.Exit(1)
	}

	// Determine the path
	var path string
	if len(os.Args) < 3 || os.Args[2] == "." {
		path = "main.tox"
	} else {
		path = os.Args[2]
	}

	// Read the file contents
	content, err := os.ReadFile(path)
	if err != nil {
		fmt.Printf("Error reading file %s: %v\n", path, err)
		os.Exit(1)
	}

	// Create lexer -> parser
	l := lexer.New(string(content))
	p := parser.New(l)
	program := p.ParseProgram()
	fmt.Printf("%#v\n", program)

	// Print parser errors
	if len(p.Errors) > 0 {
		fmt.Println("Parser errors:")
		for _, err := range p.Errors {
			fmt.Println("  -", err)
		}
		os.Exit(1)
	}

	// Run type typechecker
	errors := typechecker.Check(program)
	if len(errors) > 0 {
		fmt.Println("Type errors:")
		for _, err := range errors {
			fmt.Println("  -", err)
		}
		os.Exit(1)
	}
	fmt.Println("Program passed type checking ✅\n")

	env := evaluator.NewEnvironment()

	// Evaluate all top-level statements to populate env
	evaluator.Eval(program, env)

	// Now run main if it exists
	if mainFn, ok := env.Get("main"); ok {
		if fnStmt, ok := mainFn.(*ast.FunctionStatement); ok {
			mainEnv := evaluator.NewEnclosedEnvironment(env)
			evaluator.Eval(fnStmt.Body, mainEnv)
		}
	}
}

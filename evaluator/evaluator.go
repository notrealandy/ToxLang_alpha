package evaluator

import (
	"fmt"

	"github.com/notrealandy/tox/ast"
)

// Environment stores variable values
// Only string/int/bool for now
// map variable name to value (interface{})
type Environment struct {
	store map[string]interface{}
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]interface{})}
}

func (env *Environment) GetFunction(name string) (*ast.FunctionStatement, bool) {
    fn, ok := env.store[name].(*ast.FunctionStatement)
    return fn, ok
}

// Eval evaluates a program (list of statements)
func Eval(stmts []ast.Statement, env *Environment) {
	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.LetStatement:
			// Evaluate the value and store in env
			switch v := stmt.Value.(type) {
			case *ast.StringLiteral:
				env.store[stmt.Name] = v.Value
			case *ast.IntegerLiteral:
				env.store[stmt.Name] = v.Value
			case *ast.BoolLiteral:
				env.store[stmt.Name] = v.Value
			}

		case *ast.FunctionStatement:
			// Only store the function definition in the environment
			env.store[stmt.Name] = stmt

		case *ast.LogFunction:
			// Print the value (look up if identifier)
			if stmt.Value.Type == "IDENT" {
				val, ok := env.store[stmt.Value.Value]
				if ok {
					fmt.Println(val)
				} else {
					fmt.Printf("undefined variable: %s\n", stmt.Value.Value)
				}
			} else {
				fmt.Println(stmt.Value.Value)
			}
		}
	}
}

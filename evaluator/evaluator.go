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
			case *ast.BinaryExpression:
				// Evaluate left and right
				left := evalExpr(v.Left, env)
				right := evalExpr(v.Right, env)
				var result int64
				l, lok := left.(int64)
				r, rok := right.(int64)
				if lok && rok {
					switch v.Operator {
					case "+":
						result = l + r
					case "-":
						result = l - r
					case "*":
						result = l * r
					case "/":
						result = l / r
					case "%":
						result = l % r
					}
				}
				env.store[stmt.Name] = result
			}

		case *ast.FunctionStatement:
			// Only store the function definition in the environment
			env.store[stmt.Name] = stmt

		case *ast.LogFunction:
			val := evalExpr(stmt.Value, env)
			fmt.Println(val)
		}
	}
}


func evalExpr(expr ast.Expression, env *Environment) interface{} {
    switch v := expr.(type) {
    case *ast.StringLiteral:
        return v.Value
    case *ast.IntegerLiteral:
        return v.Value
    case *ast.BoolLiteral:
        return v.Value
    case *ast.Identifier:
        val, _ := env.store[v.Value]
        return val
    case *ast.BinaryExpression:
        left := evalExpr(v.Left, env)
        right := evalExpr(v.Right, env)
        l, lok := left.(int64)
        r, rok := right.(int64)
        if lok && rok {
            switch v.Operator {
            case "+":
                return l + r
            case "-":
                return l - r
			case "*":
				return l * r
			case "/":
				return l / r
			case "%":
				return l % r
            }
        }
    }
    return nil
}
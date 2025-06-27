package evaluator

import (
	"fmt"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/token"
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
			case *ast.CallExpression:
				result := evalExpr(v, env)
				env.store[stmt.Name] = result
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

		case *ast.ExpressionStatement:
			evalExpr(stmt.Expr, env)

		case *ast.IfStatement:
			// Evaluate the if condition
			if isTruthy(evalExpr(stmt.IfCond, env)) {
				Eval(stmt.IfBody, env)
				return
			}
			// Check elifs
			for i, elifCond := range stmt.ElifConds {
				if isTruthy(evalExpr(elifCond, env)) {
					Eval(stmt.ElifBodies[i], env)
					return
				}
			}
			// Else
			if stmt.ElseBody != nil && len(stmt.ElseBody) > 0 {
				Eval(stmt.ElseBody, env)
			}
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

		switch v.Operator {
		case token.PLUS:
			if lok && rok {
				return l + r
			}
		case token.MINUS:
			if lok && rok {
				return l - r
			}
		case token.ASTERISK:
			if lok && rok {
				return l * r
			}
		case token.SLASH:
			if lok && rok {
				return l / r
			}
		case token.MODULUS:
			if lok && rok {
				return l % r
			}
		case token.EQ:
			return left == right
		case token.NEQ:
			return left != right
		case token.LT:
			if lok && rok {
				return l < r
			}
		case token.LTE:
			if lok && rok {
				return l <= r
			}
		case token.GT:
			if lok && rok {
				return l > r
			}
		case token.GTE:
			if lok && rok {
				return l >= r
			}
		case token.AND:
			return isTruthy(left) && isTruthy(right)
		case token.OR:
			return isTruthy(left) || isTruthy(right)
		case token.NOT:
			return !isTruthy(right)
		}
		return nil
	case *ast.CallExpression:
		if ident, ok := v.Function.(*ast.Identifier); ok {
			fnObj, ok := env.GetFunction(ident.Value)
			if !ok {
				return nil // or error
			}
			// Evaluate arguments (not used if you don't support params yet)
			// Evaluate the function body and capture the return value
			return evalFunctionBody(fnObj.Body, env)
		}
		return nil
	}
	return nil
}

func evalFunctionBody(stmts []ast.Statement, env *Environment) interface{} {
	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.ReturnStatement:
			return evalExpr(stmt.Value, env)
		default:
			Eval([]ast.Statement{stmt}, env)
		}
	}
	return nil
}

func isTruthy(val interface{}) bool {
	switch v := val.(type) {
	case bool:
		return v
	case int64:
		return v != 0
	case string:
		return v != ""
	default:
		return v != nil
	}
}

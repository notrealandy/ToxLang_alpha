package evaluator

import (
	"fmt"
	"strings"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/token"
)

// Environment stores variable values and supports lexical scoping
type Environment struct {
	store  map[string]interface{}
	parent *Environment
}

func NewEnvironment() *Environment {
	return &Environment{store: make(map[string]interface{}), parent: nil}
}

func NewEnclosedEnvironment(outer *Environment) *Environment {
	return &Environment{store: make(map[string]interface{}), parent: outer}
}

func (env *Environment) Get(name string) (interface{}, bool) {
	val, ok := env.store[name]
	if !ok && env.parent != nil {
		return env.parent.Get(name)
	}
	return val, ok
}

func (env *Environment) Set(name string, val interface{}) {
	env.store[name] = val
}

func (env *Environment) SetExisting(name string, val interface{}) bool {
	if _, ok := env.store[name]; ok {
		env.store[name] = val
		return true
	}
	if env.parent != nil {
		return env.parent.SetExisting(name, val)
	}
	return false
}

func getGlobalEnv(env *Environment) *Environment {
	for env.parent != nil {
		env = env.parent
	}
	return env
}

// Eval evaluates a program (list of statements)
func Eval(stmts []ast.Statement, env *Environment) {
	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.LetStatement:
			val := evalExpr(stmt.Value, env)
			env.Set(stmt.Name, val)
		case *ast.FunctionStatement:
			env.Set(stmt.Name, stmt)
		case *ast.LogFunction:
			val := evalExpr(stmt.Value, env)
			printValue(val)
		case *ast.ExpressionStatement:
			evalExpr(stmt.Expr, env)
		case *ast.IfStatement:
			handled := false
			if isTruthy(evalExpr(stmt.IfCond, env)) {
				Eval(stmt.IfBody, env)
				handled = true
			}
			if !handled {
				for i, elifCond := range stmt.ElifConds {
					if isTruthy(evalExpr(elifCond, env)) {
						Eval(stmt.ElifBodies[i], env)
						handled = true
						break
					}
				}
			}
			if !handled && stmt.ElseBody != nil && len(stmt.ElseBody) > 0 {
				Eval(stmt.ElseBody, env)
			}
		case *ast.AssignmentStatement:
			if idxExpr, ok := stmt.Left.(*ast.IndexExpression); ok {
				// Array mutation: xs[0] >> v
				arr := evalExpr(idxExpr.Left, env)
				idx := evalExpr(idxExpr.Index, env)
				val := evalExpr(stmt.Value, env)
				arrSlice, ok := arr.([]interface{})
				idxInt, ok2 := idx.(int64)
				if ok && ok2 && int(idxInt) >= 0 && int(idxInt) < len(arrSlice) {
					arrSlice[int(idxInt)] = val
				} else {
					// Optionally: print error for out-of-bounds or wrong type
				}
			} else {
				// Normal variable assignment
				val := evalExpr(stmt.Value, env)
				if !env.SetExisting(stmt.Name, val) {
					env.Set(stmt.Name, val)
				}
			}
		case *ast.WhileStatement:
			for isTruthy(evalExpr(stmt.Condition, env)) {
				Eval(stmt.Body, env)
			}
		case *ast.ForStatement:
			// New scope for the for loop
			forEnv := NewEnclosedEnvironment(env)
			// Run init
			if stmt.Init != nil {
				Eval([]ast.Statement{stmt.Init}, forEnv)
			}
			// Loop
			for isTruthy(evalExpr(stmt.Condition, forEnv)) {
				Eval(stmt.Body, forEnv)
				if stmt.Post != nil {
					Eval([]ast.Statement{stmt.Post}, forEnv)
				}
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
		val, _ := env.Get(v.Value)
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
			// Built-in: len(xs)
			if ident.Value == "len" && len(v.Arguments) == 1 {
				arg := evalExpr(v.Arguments[0], env)
				if arr, ok := arg.([]interface{}); ok {
					return int64(len(arr))
				}
				return int64(0) // or error
			}
			// User-defined function
			fnObj, ok := env.Get(ident.Value)
			fnStmt, isFn := fnObj.(*ast.FunctionStatement)
			if !ok || !isFn {
				return nil // or error
			}
			// Evaluate arguments
			args := []interface{}{}
			for _, argExpr := range v.Arguments {
				args = append(args, evalExpr(argExpr, env))
			}
			localEnv := NewEnclosedEnvironment(getGlobalEnv(env))
			// Bind parameters to arguments
			for i, param := range fnStmt.Params {
				if i < len(args) {
					localEnv.Set(param, args[i])
				}
			}
			return evalFunctionBody(fnStmt.Body, localEnv)
		}
		return nil
	case *ast.UnaryExpression:
		right := evalExpr(v.Right, env)
		switch v.Operator {
		case token.MINUS:
			if val, ok := right.(int64); ok {
				return -val
			}
		case token.NOT:
			return !isTruthy(right)
		}
		return nil
	case *ast.ArrayLiteral:
		arr := []interface{}{}
		for _, el := range v.Elements {
			arr = append(arr, evalExpr(el, env))
		}
		return arr

	case *ast.IndexExpression:
		arr := evalExpr(v.Left, env)
		idx := evalExpr(v.Index, env)
		arrSlice, ok := arr.([]interface{})
		idxInt, ok2 := idx.(int64)
		if ok && ok2 && int(idxInt) >= 0 && int(idxInt) < len(arrSlice) {
			return arrSlice[int(idxInt)]
		}
		return nil // or error
	case *ast.SliceExpression:
		arr := evalExpr(v.Left, env)
		arrSlice, ok := arr.([]interface{})
		if !ok {
			return nil // or error
		}
		var start, end int64
		if v.Start != nil {
			if s, ok := evalExpr(v.Start, env).(int64); ok {
				start = s
			}
		}
		if v.End != nil {
			if e, ok := evalExpr(v.End, env).(int64); ok {
				end = e
			}
		} else {
			end = int64(len(arrSlice))
		}
		if start < 0 {
			start = 0
		}
		if end > int64(len(arrSlice)) {
			end = int64(len(arrSlice))
		}
		if start > end {
			start = end
		}
		return arrSlice[start:end]
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

func printValue(val interface{}) {
    switch v := val.(type) {
    case []interface{}:
        elems := make([]string, len(v))
        for i, e := range v {
            elems[i] = fmt.Sprint(e)
        }
        fmt.Printf("[%s]\n", strings.Join(elems, ", "))
    default:
        fmt.Println(v)
    }
}
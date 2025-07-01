package evaluator

import (
	"bufio"
	"fmt"
	"os"
	"regexp"
	"strings"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/token"
)

// Environment stores variable values and supports lexical scoping
type Environment struct {
	store  map[string]interface{}
	parent *Environment
}

type breakSignal struct{}
type continueSignal struct{}

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
func Eval(stmts []ast.Statement, env *Environment) interface{} {
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
			// Field assignment: e.g., u.name >> "NewValue"
			if ident, ok := stmt.Left.(*ast.Identifier); ok && strings.Contains(ident.Value, ".") {
				parts := strings.SplitN(ident.Value, ".", 2)
				baseName := parts[0]
				fieldName := parts[1]
				base, ok := env.Get(baseName)
				if !ok || base == nil {
					fmt.Printf("Error: variable '%s' is not public or does not exist\n", baseName)
					break
				}
				if obj, ok := base.(map[string]interface{}); ok {
					val := evalExpr(stmt.Value, env)
					obj[fieldName] = val
					// Optionally, update the base variable in the environment:
					env.Set(baseName, obj)
				} else {
					fmt.Printf("Error: variable '%s' is not a struct\n", baseName)
				}
			} else if idxExpr, ok := stmt.Left.(*ast.IndexExpression); ok {
				// Evaluate the collection and index
				coll := evalExpr(idxExpr.Left, env)
				idx := evalExpr(idxExpr.Index, env)
				val := evalExpr(stmt.Value, env)

				// Array mutation: xs[0] >> v
				if arrSlice, ok := coll.([]interface{}); ok {
					if idxInt, ok2 := idx.(int64); ok2 && int(idxInt) >= 0 && int(idxInt) < len(arrSlice) {
						arrSlice[int(idxInt)] = val
					} else {
						// Optionally: print error for out-of-bounds or wrong type.
					}
				}

				// Map mutation: m["key"] >> v
				if m, ok := coll.(map[interface{}]interface{}); ok {
					m[idx] = val
				}
			} else if ident, ok := stmt.Left.(*ast.Identifier); ok {
				// Normal variable assignment.
				val := evalExpr(stmt.Value, env)
				if !env.SetExisting(ident.Value, val) {
					env.Set(ident.Value, val)
				}
			}
		case *ast.BreakStatement:
			return breakSignal{}
		case *ast.ContinueStatement:
			return continueSignal{}
		case *ast.WhileStatement:
			for isTruthy(evalExpr(stmt.Condition, env)) {
				res := Eval(stmt.Body, env)
				if _, ok := res.(breakSignal); ok {
					break
				}
				if _, ok := res.(continueSignal); ok {
					continue
				}
			}
		case *ast.ForStatement:
			forEnv := NewEnclosedEnvironment(env)
			if stmt.Init != nil {
				Eval([]ast.Statement{stmt.Init}, forEnv)
			}
			for isTruthy(evalExpr(stmt.Condition, forEnv)) {
				res := Eval(stmt.Body, forEnv)
				if _, ok := res.(breakSignal); ok {
					break
				}
				if _, ok := res.(continueSignal); ok {
					if stmt.Post != nil {
						Eval([]ast.Statement{stmt.Post}, forEnv)
					}
					continue
				}
				if stmt.Post != nil {
					Eval([]ast.Statement{stmt.Post}, forEnv)
				}
			}
		case *ast.CImportStatement:
			// TODO: Actually load the C header and expose functions/types.
			fmt.Printf("[CIMPORT] Would import C header: %s\n", stmt.Header)
		}
	}
	return nil
}

func evalExpr(expr ast.Expression, env *Environment) interface{} {
	switch v := expr.(type) {
	case *ast.StringLiteral:
		return interpolateString(v.Value, env)
	case *ast.IntegerLiteral:
		return v.Value
	case *ast.BoolLiteral:
		return v.Value
	case *ast.Identifier:
		// First, try to look up the full identifier.
		if val, ok := env.Get(v.Value); ok && val != nil {
			return val
		}
		// If full identifier lookup fails and the identifier is qualified, try field access.
		if strings.Contains(v.Value, ".") {
			parts := strings.SplitN(v.Value, ".", 2)
			baseName := parts[0]
			fieldName := parts[1]
			base, ok := env.Get(baseName)
			if !ok || base == nil {
				return fmt.Sprintf("Error: variable '%s' is not public or does not exist", baseName)
			}
			if obj, ok := base.(map[string]interface{}); ok {
				if fieldVal, exists := obj[fieldName]; exists {
					return fieldVal
				}
				return fmt.Sprintf("Error: field '%s' not found in '%s'", fieldName, baseName)
			}
			return fmt.Sprintf("Error: variable '%s' is not a struct", baseName)
		}
		// Otherwise, return an error.
		return fmt.Sprintf("Error: variable '%s' is not public or does not exist", v.Value)
	case *ast.BinaryExpression:
		left := evalExpr(v.Left, env)
		right := evalExpr(v.Right, env)
		l, lok := left.(int64)
		r, rok := right.(int64)
		switch v.Operator {
		case token.PLUS:
			// Support int + int and string + string
			switch lval := left.(type) {
			case int64:
				if rval, ok := right.(int64); ok {
					return lval + rval
				}
			case string:
				if rval, ok := right.(string); ok {
					return lval + rval
				}
			}
			return nil
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

			// Built-in functions
			if fn, ok := Builtins[ident.Value]; ok {
				args := []interface{}{}
				for _, argExpr := range v.Arguments {
					args = append(args, evalExpr(argExpr, env))
				}
				return fn(args)
			}

			// --- Method call support ---
			if strings.Contains(ident.Value, ".") {
				parts := strings.SplitN(ident.Value, ".", 2)
				baseName, methodName := parts[0], parts[1]
				baseVal, ok := env.Get(baseName)
				// Only treat as a struct method if baseVal is a struct instance
				if ok {
					if obj, ok := baseVal.(map[string]interface{}); ok {
						structType, _ := obj["_struct"].(string)
						methodFullName := structType + "." + methodName
						fnObj, ok := env.Get(methodFullName)
						fnStmt, isFn := fnObj.(*ast.FunctionStatement)
						if ok && isFn {
							args := []interface{}{baseVal}
							for _, argExpr := range v.Arguments {
								args = append(args, evalExpr(argExpr, env))
							}
							localEnv := NewEnclosedEnvironment(getGlobalEnv(env))
							localEnv.Set("this", baseVal)
							for i, param := range fnStmt.Params {
								if i < len(args) {
									localEnv.Set(param, args[i])
								}
							}
							return evalFunctionBody(fnStmt.Body, localEnv)
						}
					}
				}
			}

			// Built-in: len(xs)
			if ident.Value == "len" && len(v.Arguments) == 1 {
				arg := evalExpr(v.Arguments[0], env)
				if arr, ok := arg.([]interface{}); ok {
					return int64(len(arr))
				}
				return int64(0) // or error
			}
			// Built-in: input()
			if ident.Value == "input" && (len(v.Arguments) == 0 || len(v.Arguments) == 1) {
				if len(v.Arguments) == 1 {
					prompt := evalExpr(v.Arguments[0], env)
					if s, ok := prompt.(string); ok {
						fmt.Print(s)
					}
				}
				reader := bufio.NewReader(os.Stdin)
				text, _ := reader.ReadString('\n')
				return strings.TrimRight(text, "\r\n")
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
		// Array indexing
		if arrSlice, ok := arr.([]interface{}); ok {
			if idxInt, ok2 := idx.(int64); ok2 && int(idxInt) >= 0 && int(idxInt) < len(arrSlice) {
				return arrSlice[int(idxInt)]
			}
			return nil // or error
		}
		// Map indexing (Go built-in returns map[string]interface{})
		if m, ok := arr.(map[string]interface{}); ok {
			if key, ok := idx.(string); ok {
				return m[key]
			}
		}
		// Map indexing (user map)
		if m, ok := arr.(map[interface{}]interface{}); ok {
			return m[idx]
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
	case *ast.StructLiteral:
		// Evaluate each field and return a map representing the struct instance.
		obj := make(map[string]interface{})
		for key, exp := range v.Fields {
			obj[key] = evalExpr(exp, env)
		}
		// Optionally store the struct type name (if needed later)
		obj["_struct"] = v.StructName
		return obj
	case *ast.MapLiteral:
		m := make(map[interface{}]interface{})
		for k, v := range v.Pairs {
			key := evalExpr(k, env)
			val := evalExpr(v, env)
			m[key] = val
		}
		return m
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

// Interpolates <%var%> or <%var.field%> in a string using the current environment
func interpolateString(s string, env *Environment) string {
	re := regexp.MustCompile(`<%([^%>]+)%>`)
	return re.ReplaceAllStringFunc(s, func(match string) string {
		inner := re.FindStringSubmatch(match)
		if len(inner) == 2 {
			expr := strings.TrimSpace(inner[1])
			// Support dot notation for struct fields
			if strings.Contains(expr, ".") {
				parts := strings.Split(expr, ".")
				val, ok := env.Get(parts[0])
				if !ok {
					return match
				}
				// Traverse fields
				for _, field := range parts[1:] {
					obj, ok := val.(map[string]interface{})
					if !ok {
						return match
					}
					val, ok = obj[field]
					if !ok {
						return match
					}
				}
				return fmt.Sprint(val)
			} else {
				val, ok := env.Get(expr)
				if ok {
					return fmt.Sprint(val)
				}
			}
		}
		return match // leave as-is if not found
	})
}

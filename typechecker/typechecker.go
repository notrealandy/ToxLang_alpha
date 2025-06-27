package typechecker

import (
	"fmt"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/token"
)

// Helper to infer the type of an expression as a string
func inferExprType(expr ast.Expression, funcTypes map[string]string, varTypes map[string]string) string {
	switch v := expr.(type) {
	case *ast.StringLiteral:
		return "string"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.BoolLiteral:
		return "bool"
	case *ast.Identifier:
		if t, ok := varTypes[v.Value]; ok {
			return t
		}
		return ""
	case *ast.BinaryExpression:
		switch v.Operator {
		case token.EQ, token.NEQ, token.LT, token.LTE, token.GT, token.GTE, token.AND, token.OR:
			return "bool"
		case token.PLUS, token.MINUS, token.ASTERISK, token.SLASH, token.MODULUS:
			return "int"
		default:
			return ""
		}
	case *ast.CallExpression:
		if v.Function != nil {
			if ident, ok := v.Function.(*ast.Identifier); ok {
				if ret, ok := funcTypes[ident.Value]; ok {
					return ret
				}
			}
		}
		return ""
	case *ast.UnaryExpression:
		switch v.Operator {
		case token.MINUS:
			return "int"
		case token.NOT:
			return "bool"
		default:
			return ""
		}
	default:
		return ""
	}
}

func Check(stmts []ast.Statement) []error {
	funcTypes := map[string]string{}
	funcDefs := map[string]*ast.FunctionStatement{}
	for _, s := range stmts {
		if fn, ok := s.(*ast.FunctionStatement); ok {
			funcTypes[fn.Name] = fn.ReturnType
			funcDefs[fn.Name] = fn
		}
	}
	return checkWithReturnType(stmts, "", funcTypes, funcDefs, map[string]string{})
}

func checkWithReturnType(
	stmts []ast.Statement,
	currentReturnType string,
	funcTypes map[string]string,
	funcDefs map[string]*ast.FunctionStatement,
	varTypes map[string]string,
) []error {
	var errs []error

	// Register nested functions
	for _, s := range stmts {
		if fn, ok := s.(*ast.FunctionStatement); ok {
			funcTypes[fn.Name] = fn.ReturnType
			funcDefs[fn.Name] = fn
		}
	}

	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.LetStatement:
			valType := inferExprType(stmt.Value, funcTypes, varTypes)
			varTypes[stmt.Name] = stmt.Type
			if valType != stmt.Type {
				errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign %s to %s (variable '%s')", stmt.Line, stmt.Col, valType, stmt.Type, stmt.Name))
			}
			// If it's a function call, check argument types
			if call, ok := stmt.Value.(*ast.CallExpression); ok {
				errs = append(errs, checkCallExpr(call, funcDefs, funcTypes, varTypes, stmt.Line, stmt.Col)...)
			}
		case *ast.FunctionStatement:
			// New scope for function body
			funcVarTypes := make(map[string]string)
			for k, v := range varTypes {
				funcVarTypes[k] = v
			}
			// Add parameters to local scope
			for i, param := range stmt.Params {
				funcVarTypes[param] = stmt.ParamTypes[i]
			}
			errs = append(errs, checkWithReturnType(stmt.Body, stmt.ReturnType, funcTypes, funcDefs, funcVarTypes)...)
		case *ast.ReturnStatement:
			if currentReturnType == "void" {
				if stmt.Value != nil {
					if _, ok := stmt.Value.(*ast.NilLiteral); !ok {
						errs = append(errs, fmt.Errorf("Cannot return a value from a void function (line %d:%d)", stmt.Line, stmt.Col))
					}
				}
			} else {
				if stmt.Value == nil {
					errs = append(errs, fmt.Errorf("Must return a value from non-void function (line %d:%d)", stmt.Line, stmt.Col))
				} else {
					valType := inferExprType(stmt.Value, funcTypes, varTypes)
					if valType != currentReturnType {
						errs = append(errs, fmt.Errorf("Return type mismatch on line %d:%d: expected %s, got %s", stmt.Line, stmt.Col, currentReturnType, valType))
					}
				}
			}
		case *ast.AssignmentStatement:
			expectedType, ok := varTypes[stmt.Name]
			if !ok {
				errs = append(errs, fmt.Errorf("Assignment to undeclared variable '%s' on line %d:%d", stmt.Name, stmt.Line, stmt.Col))
			} else {
				valType := inferExprType(stmt.Value, funcTypes, varTypes)
				if valType != expectedType {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign %s to %s (variable '%s')", stmt.Line, stmt.Col, valType, expectedType, stmt.Name))
				}
			}
		case *ast.ExpressionStatement:
			// Check if the expression is a function call
			if call, ok := stmt.Expr.(*ast.CallExpression); ok {
				errs = append(errs, checkCallExpr(call, funcDefs, funcTypes, varTypes, stmt.Line, stmt.Col)...)
			}
		case *ast.WhileStatement:
			condType := inferExprType(stmt.Condition, funcTypes, varTypes)
			if condType != "bool" {
				errs = append(errs, fmt.Errorf("While condition must be boolean, got %s on line %d:%d", condType, stmt.Line, stmt.Col))
			}
			// Typecheck the body
			errs = append(errs, checkWithReturnType(stmt.Body, currentReturnType, funcTypes, funcDefs, copyVarTypes(varTypes))...)

		case *ast.ForStatement:
			// New scope for the for loop
			forVarTypes := copyVarTypes(varTypes)
			// Typecheck the init statement
			if stmt.Init != nil {
				errs = append(errs, checkWithReturnType([]ast.Statement{stmt.Init}, currentReturnType, funcTypes, funcDefs, forVarTypes)...)
			}
			// Typecheck the condition
			condType := inferExprType(stmt.Condition, funcTypes, forVarTypes)
			if condType != "bool" {
				errs = append(errs, fmt.Errorf("For condition must be boolean, got %s on line %d:%d", condType, stmt.Line, stmt.Col))
			}
			// Typecheck the body
			errs = append(errs, checkWithReturnType(stmt.Body, currentReturnType, funcTypes, funcDefs, forVarTypes)...)
			// Typecheck the post statement
			if stmt.Post != nil {
				errs = append(errs, checkWithReturnType([]ast.Statement{stmt.Post}, currentReturnType, funcTypes, funcDefs, forVarTypes)...)
			}
		}
	}

	return errs
}

// Helper to check function call argument types
func checkCallExpr(
	call *ast.CallExpression,
	funcDefs map[string]*ast.FunctionStatement,
	funcTypes map[string]string,
	varTypes map[string]string,
	line, col int,
) []error {
	var errs []error
	ident, ok := call.Function.(*ast.Identifier)
	if !ok {
		return errs
	}
	fn, ok := funcDefs[ident.Value]
	if !ok {
		errs = append(errs, fmt.Errorf("Unknown function '%s' on line %d:%d", ident.Value, line, col))
		return errs
	}
	if len(call.Arguments) != len(fn.Params) {
		errs = append(errs, fmt.Errorf("Function '%s' expects %d arguments, got %d on line %d:%d", ident.Value, len(fn.Params), len(call.Arguments), line, col))
		return errs
	}
	for i, arg := range call.Arguments {
		argType := inferExprType(arg, funcTypes, varTypes)
		paramType := fn.ParamTypes[i]
		if argType != paramType {
			errs = append(errs, fmt.Errorf("Type error: argument %d to '%s' expects %s, got %s on line %d:%d", i+1, ident.Value, paramType, argType, line, col))
		}
	}
	return errs
}

// Helper to copy variable types map for new scopes
func copyVarTypes(src map[string]string) map[string]string {
	dst := make(map[string]string)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

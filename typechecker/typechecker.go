package typechecker

import (
	"fmt"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/token"
)

// Helper to infer the type of an expression as a string
func inferExprType(expr ast.Expression, funcTypes map[string]string) string {
	switch v := expr.(type) {
	case *ast.StringLiteral:
		return "string"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.BoolLiteral:
		return "bool"
	case *ast.Identifier:
		// For simplicity, treat identifiers as int (expand as needed)
		return "int"
	case *ast.BinaryExpression:
		switch v.Operator {
		case token.EQ, token.NEQ, token.LT, token.LTE, token.GT, token.GTE, token.AND, token.OR, token.NOT:
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
	for _, s := range stmts {
		if fn, ok := s.(*ast.FunctionStatement); ok {
			funcTypes[fn.Name] = fn.ReturnType
		}
	}
	return checkWithReturnType(stmts, "", funcTypes, map[string]string{})
}

func checkWithReturnType(stmts []ast.Statement, currentReturnType string, funcTypes map[string]string, varTypes map[string]string) []error {
	var errs []error

	// Build funcTypes for nested functions
	for _, s := range stmts {
		if fn, ok := s.(*ast.FunctionStatement); ok {
			funcTypes[fn.Name] = fn.ReturnType
		}
	}

	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.LetStatement:
			varTypes[stmt.Name] = stmt.Type
			switch v := stmt.Value.(type) {
			case *ast.StringLiteral:
				if stmt.Type != "string" {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: expected string, got %s", stmt.Line, stmt.Col, stmt.Type))
				}
			case *ast.IntegerLiteral:
				if stmt.Type != "int" {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign int to %s", stmt.Line, stmt.Col, stmt.Type))
				}
			case *ast.BoolLiteral:
				if stmt.Type != "bool" {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign bool to %s", stmt.Line, stmt.Col, stmt.Type))
				}
			case *ast.BinaryExpression:
				valType := inferExprType(v, funcTypes)
				if valType != stmt.Type {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign binary expression of type %s to %s", stmt.Line, stmt.Col, valType, stmt.Type))
				}
			case *ast.CallExpression:
				valType := inferExprType(v, funcTypes)
				if valType == "" {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign unknown function call result to %s", stmt.Line, stmt.Col, stmt.Type))
				} else if valType != stmt.Type {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign function call returning %s to %s", stmt.Line, stmt.Col, valType, stmt.Type))
				}
			default:
				errs = append(errs, fmt.Errorf("Unknown value type for variable '%s' on line %d:%d", stmt.Name, stmt.Line, stmt.Col))
			}
		case *ast.FunctionStatement:
			// Pass a copy of the current varTypes to the function body
			funcVarTypes := make(map[string]string)
			for k, v := range varTypes {
				funcVarTypes[k] = v
			}
			errs = append(errs, checkWithReturnType(stmt.Body, stmt.ReturnType, funcTypes, funcVarTypes)...)
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
					valType := inferExprType(stmt.Value, funcTypes)
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
				valType := inferExprType(stmt.Value, funcTypes)
				if valType != expectedType {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign %s to %s (variable '%s')", stmt.Line, stmt.Col, valType, expectedType, stmt.Name))
				}
			}
		}
	}

	return errs
}

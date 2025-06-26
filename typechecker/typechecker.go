package typechecker

import (
	"fmt"
	"github.com/notrealandy/tox/ast"
)

func Check(stmts []ast.Statement) []error {
	var errs []error

	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.LetStatement:
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
				// Recursively check left and right
				// Only allow int for now
				left, right := v.Left, v.Right
				// Check both sides are IntegerLiteral or BinaryExpression
				isInt := func(e ast.Expression) bool {
					switch e.(type) {
					case *ast.IntegerLiteral, *ast.BinaryExpression, *ast.Identifier:
						return true
					}
					return false
				}
				if stmt.Type != "int" {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign binary expression to %s", stmt.Line, stmt.Col, stmt.Type))
				} else if !isInt(left) || !isInt(right) {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: binary expression must be int", stmt.Line, stmt.Col))
				}
				// Recursively check left and right
				if leftStmt, ok := left.(*ast.BinaryExpression); ok {
					errs = append(errs, Check([]ast.Statement{&ast.LetStatement{Type: "int", Value: leftStmt, Line: stmt.Line, Col: stmt.Col}})...)
				}
				if rightStmt, ok := right.(*ast.BinaryExpression); ok {
					errs = append(errs, Check([]ast.Statement{&ast.LetStatement{Type: "int", Value: rightStmt, Line: stmt.Line, Col: stmt.Col}})...)
				}
			default:
				errs = append(errs, fmt.Errorf("Unknown value type for variable '%s' on line %d:%d", stmt.Name, stmt.Line, stmt.Col))
			}
		case *ast.FunctionStatement:
			// Recursively check the function body
			errs = append(errs, Check(stmt.Body)...)
		}
	}

	return errs
}

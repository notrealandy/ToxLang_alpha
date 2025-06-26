package typechecker

import (
    "fmt"

    "github.com/notrealandy/tox/ast"
)

// Helper to infer the type of an expression as a string
func inferExprType(expr ast.Expression) string {
    switch expr.(type) {
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
        // Only int binary expressions supported
        return "int"
    default:
        return ""
    }
}

func Check(stmts []ast.Statement) []error {
    return checkWithReturnType(stmts, "")
}

func checkWithReturnType(stmts []ast.Statement, currentReturnType string) []error {
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
                left, right := v.Left, v.Right
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
                if leftStmt, ok := left.(*ast.BinaryExpression); ok {
                    errs = append(errs, checkWithReturnType([]ast.Statement{&ast.LetStatement{Type: "int", Value: leftStmt, Line: stmt.Line, Col: stmt.Col}}, currentReturnType)...)
                }
                if rightStmt, ok := right.(*ast.BinaryExpression); ok {
                    errs = append(errs, checkWithReturnType([]ast.Statement{&ast.LetStatement{Type: "int", Value: rightStmt, Line: stmt.Line, Col: stmt.Col}}, currentReturnType)...)
                }
            default:
                errs = append(errs, fmt.Errorf("Unknown value type for variable '%s' on line %d:%d", stmt.Name, stmt.Line, stmt.Col))
            }
        case *ast.FunctionStatement:
            // Pass the function's declared return type to its body
            errs = append(errs, checkWithReturnType(stmt.Body, stmt.ReturnType)...)
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
                    valType := inferExprType(stmt.Value)
                    if valType != currentReturnType {
                        errs = append(errs, fmt.Errorf("Return type mismatch on line %d:%d: expected %s, got %s", stmt.Line, stmt.Col, currentReturnType, valType))
                    }
                }
            }
        }
    }

    return errs
}
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
            switch stmt.Value.(type) {
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
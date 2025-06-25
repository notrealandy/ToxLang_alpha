package typechecker

import (
    "fmt"
    "github.com/notrealandy/tox/ast"
)

func Check(stmts []ast.Statement) []error {
    var errs []error
    // TODO: Implement a symbol table to track variable types and function signatures

    for _, s := range stmts {
        errs = append(errs, checkStatement(s, "")...) // No expected return type at top level
    }

    return errs
}

func checkStatement(s ast.Statement, expectedReturnType string) []error {
	var errs []error
	switch stmt := s.(type) {
	case *ast.LetStatement:
		valType, valErrs := checkExpression(stmt.Value)
		if len(valErrs) > 0 {
			errs = append(errs, valErrs...)
			// Potentially return early if value expression itself is fundamentally broken
		}
		// If valType is determinable even with errors, proceed to check assignment compatibility
		if valType != "" && stmt.Type != valType {
			errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign %s to variable '%s' of type %s", stmt.Line, stmt.Col, valType, stmt.Name, stmt.Type))
		}
		// TODO: Add variable to symbol table with its type stmt.Type

	case *ast.FunctionStatement:
		// TODO: Add function signature to symbol table (params, return type)
		// Recursively check the function body, passing the function's return type
		for _, bodyStmt := range stmt.Body {
			errs = append(errs, checkStatement(bodyStmt, stmt.ReturnType)...)
		}

	case *ast.ReturnStatement:
		if expectedReturnType == "" && stmt.ReturnValue != nil { // Allow `return;` in void if ReturnValue is nil
			errs = append(errs, fmt.Errorf("Type error on line %d:%d: return statement with value in function with no return type", stmt.Line, stmt.Col))
			// If there's no return value (e.g. `return;`), it might be allowed for void functions.
			// Current parser requires an expression, so this path might not be hit if ReturnValue is always non-nil.
			// This check needs to be robust for `return;` vs `return <expr>;`
		}

		if stmt.ReturnValue != nil {
			retValType, retValErrs := checkExpression(stmt.ReturnValue)
			if len(retValErrs) > 0 {
				errs = append(errs, retValErrs...)
			}

			if expectedReturnType == "" && retValType != "" { // Returning a value from void function
				// This is subtly different from the above check; this means expected is void but got a type
				errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot return value of type %s from function with no return type", stmt.Line, stmt.Col, retValType))
			} else if retValType != "" && expectedReturnType != "" && expectedReturnType != retValType {
				errs = append(errs, fmt.Errorf("Type error on line %d:%d: return type mismatch, expected %s, got %s", stmt.Line, stmt.Col, expectedReturnType, retValType))
			}
		} else if expectedReturnType != "" { // No return value but function expects one
			 errs = append(errs, fmt.Errorf("Type error on line %d:%d: function expects return type %s but got no return value", stmt.Line, stmt.Col, expectedReturnType))
		}


	// TODO: Add cases for other statements like expression statements (e.g. `foo();`), if, for, etc.
	default:
		// This case should ideally not be reached if all statement types are handled
		errs = append(errs, fmt.Errorf("Type error: unknown statement type encountered: %T", stmt))
	}
	return errs
}


func checkExpression(expr ast.Expression) (string, []error) {
	var errs []error
	switch e := expr.(type) {
	case *ast.IntegerLiteral:
		return "int", nil
	case *ast.StringLiteral:
		return "string", nil
	case *ast.BoolLiteral:
		return "bool", nil
	case *ast.Identifier:
		// TODO: Symbol Table: Lookup identifier type.
		// For now, return a placeholder "unknown" or error.
		// To allow basic tests, we might temporarily assume a type or return "any"
		// For this phase, let's be strict: if it's an identifier, we can't know its type yet.
		// However, to allow `let x int >> y;` to not immediately fail if y's type is unknown,
		// we need a way for the calling context (checkStatement for Let) to handle this.
		// For now, let's return "identifier_placeholder" and let the statement checker decide.
		// A better approach for now might be to return an error.
		errs = append(errs, fmt.Errorf("Type error on line %d:%d: type of identifier '%s' cannot be determined without symbol table", e.Line, e.Col, e.Value))
		return "unknown_identifier", errs // Special type indicating it's an unresolved identifier
	case *ast.PrefixExpression:
		rightType, rightErrs := checkExpression(e.Right)
		if len(rightErrs) > 0 {
			errs = append(errs, rightErrs...)
		}
		if rightType == "unknown_identifier" { // Propagate identifier uncertainty
			return "unknown", errs
		}

		switch e.Operator {
		case "-":
			if rightType != "int" {
				errs = append(errs, fmt.Errorf("Type error on line %d:%d: unary '-' operator can only be applied to int, got %s", e.Line, e.Col, rightType))
				return "unknown", errs
			}
			return "int", errs
		case "!":
			if rightType != "bool" {
				errs = append(errs, fmt.Errorf("Type error on line %d:%d: unary '!' operator can only be applied to bool, got %s", e.Line, e.Col, rightType))
				return "unknown", errs
			}
			return "bool", errs
		default:
			errs = append(errs, fmt.Errorf("Type error on line %d:%d: unknown prefix operator '%s'", e.Line, e.Col, e.Operator))
			return "unknown", errs
		}
	case *ast.InfixExpression:
		leftType, leftErrs := checkExpression(e.Left)
		if len(leftErrs) > 0 {
			errs = append(errs, leftErrs...)
		}
		rightType, rightErrs := checkExpression(e.Right)
		if len(rightErrs) > 0 {
			errs = append(errs, rightErrs...)
		}

		if leftType == "unknown_identifier" || rightType == "unknown_identifier" {
			// If any operand's type is unknown due to it being an unresolved identifier,
			// we can't reliably check the infix operation type yet.
			// We could add an error or return "unknown".
			// For now, let's try to proceed if one is known, but specific ops might still fail.
		}


		switch e.Operator {
		case "+", "-", "*", "/":
			if leftType != "int" || rightType != "int" {
				// Allow unknown_identifier to prevent cascading errors if one side is known to be int
				if !( (leftType == "int" && rightType == "unknown_identifier") || (leftType == "unknown_identifier" && rightType == "int") || (leftType == "unknown_identifier" && rightType == "unknown_identifier") ) {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: arithmetic operator '%s' requires int operands, got %s and %s", e.Line, e.Col, e.Operator, leftType, rightType))
					return "unknown", errs
				}
				// If one is unknown_identifier and other is int, result is tentatively int. If both unknown, result unknown.
                if leftType == "unknown_identifier" && rightType == "unknown_identifier" {
                    return "unknown", errs
                }
                 return "int", errs // Tentatively int, errors might have been added
			}
			return "int", errs
		case "==", "!=", "<", ">", "<=", ">=":
			// For comparisons, allow int-int, string-string, bool-bool for ==, !=
			// Allow int-int, string-string for <, >, <=, >=
			// For now, simplify: require types to be the same and not "unknown" unless both are "unknown_identifier"
			if leftType != rightType {
				if !((leftType == "unknown_identifier" && rightType != "unknown_identifier") || (leftType != "unknown_identifier" && rightType == "unknown_identifier") || (leftType == "unknown_identifier" && rightType == "unknown_identifier") ) {
					errs = append(errs, fmt.Errorf("Type error on line %d:%d: comparison operator '%s' requires operands of the same type, got %s and %s", e.Line, e.Col, e.Operator, leftType, rightType))
					return "bool", errs // comparison result is bool, but types are incompatible
				}
			}
			if leftType == "unknown_identifier" && rightType == "unknown_identifier" {
                 // If both are unknown, we can't check, but result of comparison is bool
                 return "bool", errs
            }
			// Add more specific checks for <, >, <=, >= if they shouldn't apply to bools
			if (e.Operator == "<" || e.Operator == ">" || e.Operator == "<=" || e.Operator == ">=") && (leftType == "bool" || rightType == "bool") {
				 actualLType := leftType
                if leftType == "unknown_identifier" && rightType == "bool" { actualLType = "bool" }
                actualRType := rightType
                if rightType == "unknown_identifier" && leftType == "bool" { actualRType = "bool" }

                if actualLType == "bool" && actualRType == "bool" {
				    errs = append(errs, fmt.Errorf("Type error on line %d:%d: comparison operator '%s' cannot be applied to booleans", e.Line, e.Col, e.Operator))
				    return "bool", errs
                }
			}
			return "bool", errs
		default:
			errs = append(errs, fmt.Errorf("Type error on line %d:%d: unknown infix operator '%s'", e.Line, e.Col, e.Operator))
			return "unknown", errs
		}
	case *ast.GroupedExpression:
		return checkExpression(e.Expression) // Type of a grouped expression is the type of its inner expression
	default:
		errs = append(errs, fmt.Errorf("Type error: unknown expression type encountered: %T", expr))
		return "unknown", errs
	}
}
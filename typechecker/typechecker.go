package typechecker

import (
	"fmt"
	"strings"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/token"
)

// inferExprType returns the type (as a string) of an expression.
func inferExprType(expr ast.Expression, funcTypes map[string]string, varTypes map[string]string, structDefs map[string]*ast.StructStatement) string {
	switch v := expr.(type) {
	case *ast.StringLiteral:
		return "string"
	case *ast.IntegerLiteral:
		return "int"
	case *ast.BoolLiteral:
		return "bool"
	case *ast.Identifier:
		// Direct lookup.
		if t, ok := varTypes[v.Value]; ok {
			return t
		}
		// Fallback: if the identifier is qualified (e.g. "u.name")
		if strings.Contains(v.Value, ".") {
			parts := strings.SplitN(v.Value, ".", 2)
			baseName, fieldName := parts[0], parts[1]
			if baseType, ok := varTypes[baseName]; ok {
				if def, ok := structDefs[baseType]; ok {
					for _, fld := range def.Fields {
						if fld.Name == fieldName {
							return fld.Type
						}
					}
				}
			}
			// Optionally try an unqualified lookup.
			if t, ok := varTypes[fieldName]; ok {
				return t
			}
		}
		return ""
	case *ast.BinaryExpression:
		leftType := inferExprType(v.Left, funcTypes, varTypes, structDefs)
		rightType := inferExprType(v.Right, funcTypes, varTypes, structDefs)
		switch v.Operator {
		case token.EQ, token.NEQ, token.LT, token.LTE, token.GT, token.GTE, token.AND, token.OR:
			return "bool"
		case token.PLUS:
			if leftType == "string" && rightType == "string" {
				return "string"
			}
			if leftType == "int" && rightType == "int" {
				return "int"
			}
			return ""
		case token.MINUS, token.ASTERISK, token.SLASH, token.MODULUS:
			if leftType == "int" && rightType == "int" {
				return "int"
			}
			return ""
		default:
			return ""
		}
	case *ast.CallExpression:
		if v.Function != nil {
			if ident, ok := v.Function.(*ast.Identifier); ok {
				// --- Method call support ---
				if strings.Contains(ident.Value, ".") {
					parts := strings.SplitN(ident.Value, ".", 2)
					baseName, methodName := parts[0], parts[1]
					baseType, ok := varTypes[baseName]
					if ok {
						methodFullName := baseType + "." + methodName
						if ret, ok := funcTypes[methodFullName]; ok {
							return ret
						}
					}
				}
				// Normal function
				if ret, ok := funcTypes[ident.Value]; ok {
					return ret
				}
				if ident.Value == "len" {
					return "int"
				}
				if ident.Value == "input" {
					return "string"
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
	case *ast.ArrayLiteral:
		if len(v.Elements) == 0 {
			return "unknown[]" // Or trigger an error.
		}
		elemType := inferExprType(v.Elements[0], funcTypes, varTypes, structDefs)
		for _, el := range v.Elements[1:] {
			if inferExprType(el, funcTypes, varTypes, structDefs) != elemType {
				return "" // Mixed types error.
			}
		}
		return elemType + "[]"
	case *ast.IndexExpression:
		leftType := inferExprType(v.Left, funcTypes, varTypes, structDefs)
		if len(leftType) > 2 && leftType[len(leftType)-2:] == "[]" {
			return leftType[:len(leftType)-2]
		}
		return ""
	case *ast.SliceExpression:
		leftType := inferExprType(v.Left, funcTypes, varTypes, structDefs)
		if len(leftType) > 2 && leftType[len(leftType)-2:] == "[]" {
			return leftType
		}
		return ""
	case *ast.StructLiteral:
		// Return the struct name as its type.
		return v.StructName
	default:
		return ""
	}
}

// Check is the entry point for typechecking a program.
func Check(stmts []ast.Statement) []error {
	funcTypes := map[string]string{}
	funcDefs := map[string]*ast.FunctionStatement{}
	structDefs := map[string]*ast.StructStatement{}
	globalVars := map[string]string{}

	// First pass: register public functions, structs, and global let statements.
	for _, s := range stmts {
		switch st := s.(type) {
		case *ast.FunctionStatement:
			funcTypes[st.Name] = st.ReturnType
			funcDefs[st.Name] = st
		case *ast.StructStatement:
			structDefs[st.Name] = st
		case *ast.LetStatement:
			globalVars[st.Name] = st.Type
		}
	}

	// Merge global variables into varTypes and start typechecking the full AST.
	return checkWithReturnType(stmts, "", funcTypes, funcDefs, globalVars, structDefs)
}

// checkWithReturnType recursively typechecks statements with the current expected return type.
func checkWithReturnType(
	stmts []ast.Statement,
	currentReturnType string,
	funcTypes map[string]string,
	funcDefs map[string]*ast.FunctionStatement,
	varTypes map[string]string,
	structDefs map[string]*ast.StructStatement,
) []error {
	var errs []error

	// Register nested functions.
	for _, s := range stmts {
		if fn, ok := s.(*ast.FunctionStatement); ok {
			funcTypes[fn.Name] = fn.ReturnType
			funcDefs[fn.Name] = fn
		}
	}

	for _, s := range stmts {
		switch stmt := s.(type) {
		case *ast.LetStatement:
			valType := inferExprType(stmt.Value, funcTypes, varTypes, structDefs)
			if valType == "" {
				errs = append(errs, fmt.Errorf("Error on line %d:%d: initialization of variable '%s' uses an undeclared or non‑public variable", stmt.Line, stmt.Col, stmt.Name))
			}
			varTypes[stmt.Name] = stmt.Type
			if valType != stmt.Type {
				errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign %s to %s (variable '%s')", stmt.Line, stmt.Col, valType, stmt.Type, stmt.Name))
			}
			// --- Struct literal field validation ---
			if structLit, ok := stmt.Value.(*ast.StructLiteral); ok {
				if def, ok := structDefs[structLit.StructName]; ok {
					// Check for missing fields
					for _, field := range def.Fields {
						if _, exists := structLit.Fields[field.Name]; !exists {
							errs = append(errs, fmt.Errorf("Missing field '%s' in struct literal for '%s' on line %d:%d", field.Name, structLit.StructName, stmt.Line, stmt.Col))
						}
					}
					// Check for extra fields
					for fieldName := range structLit.Fields {
						found := false
						for _, field := range def.Fields {
							if field.Name == fieldName {
								found = true
								break
							}
						}
						if !found {
							errs = append(errs, fmt.Errorf("Unknown field '%s' in struct literal for '%s' on line %d:%d", fieldName, structLit.StructName, stmt.Line, stmt.Col))
						}
					}
				}
			}
		case *ast.ExpressionStatement:
			// If the expression is a CallExpression, typecheck its arguments via checkCallExpr.
			if call, ok := stmt.Expr.(*ast.CallExpression); ok {
				errs = append(errs, checkCallExpr(call, funcDefs, funcTypes, varTypes, structDefs, stmt.Line, stmt.Col)...)
			}
			exprType := inferExprType(stmt.Expr, funcTypes, varTypes, structDefs)
			if exprType == "" {
				errs = append(errs, fmt.Errorf("Error on line %d:%d: expression uses an undeclared or non‑public variable", stmt.Line, stmt.Col))
			}
		case *ast.LogFunction:
			exprType := inferExprType(stmt.Value, funcTypes, varTypes, structDefs)
			if exprType == "" {
				errs = append(errs, fmt.Errorf("Error on line %d:%d: log expression uses an undeclared or non‑public variable", stmt.Line, stmt.Col))
			}
		case *ast.FunctionStatement:
			// Check that the return type is valid (built-in or declared struct)
			builtin := stmt.ReturnType == "int" || stmt.ReturnType == "string" || stmt.ReturnType == "bool" || stmt.ReturnType == "void"
			if !builtin {
				if _, ok := structDefs[stmt.ReturnType]; !ok {
					errs = append(errs, fmt.Errorf("Unknown return type '%s' for function '%s' on line %d:%d", stmt.ReturnType, stmt.Name, stmt.Line, stmt.Col))
				}
			}
			// Create a new scope for the function body.
			funcVarTypes := make(map[string]string)
			for k, v := range varTypes {
				funcVarTypes[k] = v
			}
			for i, param := range stmt.Params {
				funcVarTypes[param] = stmt.ParamTypes[i]
			}
			errs = append(errs, checkWithReturnType(stmt.Body, stmt.ReturnType, funcTypes, funcDefs, funcVarTypes, structDefs)...)
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
					valType := inferExprType(stmt.Value, funcTypes, varTypes, structDefs)
					if valType != currentReturnType {
						errs = append(errs, fmt.Errorf("Return type mismatch on line %d:%d: expected %s, got %s", stmt.Line, stmt.Col, currentReturnType, valType))
					}
				}
			}
		case *ast.AssignmentStatement:
			// For field assignments (qualified identifiers like "u.name")
			if ident, ok := stmt.Left.(*ast.Identifier); ok && strings.Contains(ident.Value, ".") {
				parts := strings.SplitN(ident.Value, ".", 2)
				baseName := parts[0]
				fieldName := parts[1]
				baseType, ok := varTypes[baseName]
				if !ok {
					errs = append(errs, fmt.Errorf("Assignment to undeclared variable '%s' on line %d:%d", baseName, stmt.Line, stmt.Col))
				} else {
					if def, ok := structDefs[baseType]; ok {
						found := false
						for _, fld := range def.Fields {
							if fld.Name == fieldName {
								found = true
								break
							}
						}
						if !found {
							errs = append(errs, fmt.Errorf("Field '%s' not found in struct '%s' on line %d:%d", fieldName, baseType, stmt.Line, stmt.Col))
						}
					} else {
						errs = append(errs, fmt.Errorf("Variable '%s' is not a struct on line %d:%d", baseName, stmt.Line, stmt.Col))
					}
				}
			} else {
				// Normal assignment: variable must be declared.
				if _, ok := varTypes[stmt.Name]; !ok {
					errs = append(errs, fmt.Errorf("Assignment to undeclared variable '%s' on line %d:%d", stmt.Name, stmt.Line, stmt.Col))
				} else {
					expectedType := varTypes[stmt.Name]
					valType := inferExprType(stmt.Value, funcTypes, varTypes, structDefs)
					if valType == "" {
						errs = append(errs, fmt.Errorf("Error on line %d:%d: assignment of variable '%s' uses an undeclared or non‑public variable", stmt.Line, stmt.Col, stmt.Name))
					} else if valType != expectedType {
						errs = append(errs, fmt.Errorf("Type error on line %d:%d: cannot assign %s to %s (variable '%s')", stmt.Line, stmt.Col, valType, expectedType, stmt.Name))
					}
				}
			}
		case *ast.WhileStatement:
			condType := inferExprType(stmt.Condition, funcTypes, varTypes, structDefs)
			if condType != "bool" {
				errs = append(errs, fmt.Errorf("While condition must be boolean, got %s on line %d:%d", condType, stmt.Line, stmt.Col))
			}
			errs = append(errs, checkWithReturnType(stmt.Body, currentReturnType, funcTypes, funcDefs, copyVarTypes(varTypes), structDefs)...)
		case *ast.ForStatement:
			forVarTypes := copyVarTypes(varTypes)
			if stmt.Init != nil {
				errs = append(errs, checkWithReturnType([]ast.Statement{stmt.Init}, currentReturnType, funcTypes, funcDefs, forVarTypes, structDefs)...)
			}
			condType := inferExprType(stmt.Condition, funcTypes, forVarTypes, structDefs)
			if condType != "bool" {
				errs = append(errs, fmt.Errorf("For condition must be boolean, got %s on line %d:%d", condType, stmt.Line, stmt.Col))
			}
			errs = append(errs, checkWithReturnType(stmt.Body, currentReturnType, funcTypes, funcDefs, forVarTypes, structDefs)...)
			if stmt.Post != nil {
				errs = append(errs, checkWithReturnType([]ast.Statement{stmt.Post}, currentReturnType, funcTypes, funcDefs, forVarTypes, structDefs)...)
			}
		}
	}

	return errs
}

// checkCallExpr verifies that a call expression has the correct number and types of arguments.
func checkCallExpr(
	call *ast.CallExpression,
	funcDefs map[string]*ast.FunctionStatement,
	funcTypes map[string]string,
	varTypes map[string]string,
	structDefs map[string]*ast.StructStatement,
	line, col int,
) []error {
	var errs []error
	ident, ok := call.Function.(*ast.Identifier)
	if !ok {
		return errs
	}

	// --- Method call support ---
	if strings.Contains(ident.Value, ".") {
		parts := strings.SplitN(ident.Value, ".", 2)
		baseName, methodName := parts[0], parts[1]
		baseType, ok := varTypes[baseName]
		if ok {
			methodFullName := baseType + "." + methodName
			fn, ok := funcDefs[methodFullName]
			if ok {
				// Insert the base as the first argument (for 'this')
				args := append([]ast.Expression{&ast.Identifier{Value: baseName}}, call.Arguments...)
				if len(args) != len(fn.Params) {
					errs = append(errs, fmt.Errorf("Method '%s' expects %d arguments, got %d on line %d:%d", methodFullName, len(fn.Params), len(args), line, col))
					return errs
				}
				for i, arg := range args {
					argType := inferExprType(arg, funcTypes, varTypes, structDefs)
					paramType := fn.ParamTypes[i]
					if argType != paramType {
						errs = append(errs, fmt.Errorf("Type error: argument %d to '%s' expects %s, got %s on line %d:%d", i+1, methodFullName, paramType, argType, line, col))
					}
				}
				return errs
			}
		}
	}

	// Built-in len function.
	if ident.Value == "len" {
		if len(call.Arguments) != 1 {
			errs = append(errs, fmt.Errorf("Built-in 'len' expects 1 argument, got %d on line %d:%d", len(call.Arguments), line, col))
		}
		argType := inferExprType(call.Arguments[0], funcTypes, varTypes, structDefs)
		if len(argType) < 3 || argType[len(argType)-2:] != "[]" {
			errs = append(errs, fmt.Errorf("Built-in 'len' expects an array argument, got %s on line %d:%d", argType, line, col))
		}
		return errs
	}
	// Built-in input function.
	if ident.Value == "input" {
		if len(call.Arguments) > 1 {
			errs = append(errs, fmt.Errorf("Built-in 'input' expects 0 or 1 argument, got %d on line %d:%d", len(call.Arguments), line, col))
		}
		if len(call.Arguments) == 1 {
			argType := inferExprType(call.Arguments[0], funcTypes, varTypes, structDefs)
			if argType != "string" {
				errs = append(errs, fmt.Errorf("Built-in 'input' expects a string argument, got %s on line %d:%d", argType, line, col))
			}
		}
		return errs
	}
	// Look up user-defined function.
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
		argType := inferExprType(arg, funcTypes, varTypes, structDefs)
		paramType := fn.ParamTypes[i]
		if argType != paramType {
			errs = append(errs, fmt.Errorf("Type error: argument %d to '%s' expects %s, got %s on line %d:%d", i+1, ident.Value, paramType, argType, line, col))
		}
	}
	return errs
}

// copyVarTypes makes a shallow copy of a map of variable types.
func copyVarTypes(src map[string]string) map[string]string {
	dst := make(map[string]string)
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

package parser

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/notrealandy/tox/ast"
	"github.com/notrealandy/tox/lexer"
	"github.com/notrealandy/tox/token"
)

type Parser struct {
	l         *lexer.Lexer
	curToken  token.Token
	peekToken token.Token
	Errors    []string
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:      l,
		Errors: []string{},
	}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) ParseProgram() []ast.Statement {
	var statements []ast.Statement

	for p.curToken.Type != token.EOF {
		// Check for optional pub modifier for functions or let statements
		if p.curToken.Type == token.PUB {
			vis := "pub"
			p.nextToken() // consume 'pub'
			if p.curToken.Type == token.FNC {
				fn := p.parseFunctionStatement()
				fn.Visibility = vis
				statements = append(statements, fn)
				continue
			} else if p.curToken.Type == token.LET {
				letStmt := p.parseLetStatement()
				letStmt.Visibility = vis
				statements = append(statements, letStmt)
				continue
			} else {
				p.Errors = append(p.Errors, fmt.Sprintf("unexpected token '%s' after pub on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
				p.nextToken()
				continue
			}
		}
		var stmt ast.Statement
		if p.curToken.Type == token.LET {
			stmt = p.parseLetStatement()
		} else if p.curToken.Type == token.FNC {
			stmt = p.parseFunctionStatement()
		} else if p.curToken.Type == token.LOG {
			stmt = p.parseLogFunctionStatement()
		} else if p.curToken.Type == token.RETURN {
			stmt = p.parseReturnStatement()
		} else if p.curToken.Type == token.IF {
			stmt = p.parseIfStatement()
		} else if p.curToken.Type == token.IDENT && (p.peekToken.Type == token.ASSIGN_OP || p.peekToken.Type == token.LBRACKET) {
			stmt = p.parseAssignmentStatement()
		} else if p.curToken.Type == token.WHILE {
			stmt = p.parseWhileStatement()
		} else if p.curToken.Type == token.FOR {
			stmt = p.parseForStatement()
		} else if p.curToken.Type == token.PACKAGE {
			stmt = p.parsePackageStatement()
		} else if p.curToken.Type == token.IMPORT {
			stmt = p.parseImportStatement()
		} else {
			p.Errors = append(p.Errors, fmt.Sprintf("[PARSE PROGRAM] unexpected token '%s' on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
			p.nextToken()
			continue
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}

	}

	return statements
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	if p.curToken.Type != token.LET {
		p.Errors = append(p.Errors, fmt.Sprintf("expected 'let' on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	p.nextToken()

	if p.curToken.Type != token.IDENT {
		p.Errors = append(p.Errors, fmt.Sprintf("expected identifier on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	name := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != token.TYPE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected type on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	typ := p.curToken.Literal
	p.nextToken()

	if p.curToken.Type != token.ASSIGN_OP {
		p.Errors = append(p.Errors, fmt.Sprintf("[PARSE LET STATEMENT] expected assignment operator '>>' on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	p.nextToken()

	value := p.parseExpression()

	return &ast.LetStatement{
		Name:  name,
		Type:  typ,
		Value: value,
		Line:  p.curToken.Line,
		Col:   p.curToken.Col,
	}
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	// Assume current token is FNC
	fn := &ast.FunctionStatement{Line: p.curToken.Line, Col: p.curToken.Col}

	p.nextToken() // move to function name
	if p.curToken.Type != token.IDENT {
		p.Errors = append(p.Errors, fmt.Sprintf("expected function name on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	fn.Name = p.curToken.Literal

	p.nextToken() // move to (
	if p.curToken.Type != token.LPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '(' after function name on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	params := []string{}
	paramTypes := []string{}
	p.nextToken() // move to first param or ')'
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		if p.curToken.Type == token.IDENT {
			paramName := p.curToken.Literal
			params = append(params, paramName)
			p.nextToken() // move to type

			if p.curToken.Type != token.TYPE {
				p.Errors = append(p.Errors, fmt.Sprintf("expected type after parameter '%s' on line %d:%d", paramName, p.curToken.Line, p.curToken.Col))
				return nil
			}

			paramTypes = append(paramTypes, p.curToken.Literal)
			p.nextToken()
			if p.curToken.Type == token.COMMA {
				p.nextToken() // skip comma and continue to next param
			}

		} else {
			p.Errors = append(p.Errors, fmt.Sprintf("expected parameter identifier on line %d:%d", p.curToken.Line, p.curToken.Col))
			return nil
		}
	}
	if p.curToken.Type != token.RPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ')' after parameters on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	fn.Params = params
	fn.ParamTypes = paramTypes

	p.nextToken() // move to >>
	if p.curToken.Type != token.ASSIGN_OP {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '>>' after ')' on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	p.nextToken() // move to return type (e.g. string, int, bool, void)
	if p.curToken.Type != token.TYPE && p.curToken.Type != token.FNCVOID {
		p.Errors = append(p.Errors, fmt.Sprintf("expected return type after '>>' on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	fn.ReturnType = p.curToken.Literal

	p.nextToken() // move to {
	if p.curToken.Type != token.LBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after return type on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	// Parse body
	fn.Body = p.parseBlock()

	return fn
}

func (p *Parser) parseLogFunctionStatement() *ast.LogFunction {
	lg := &ast.LogFunction{Line: p.curToken.Line, Col: p.curToken.Col}

	p.nextToken() // move to (
	if p.curToken.Type != token.LPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '(' after 'log' on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	p.nextToken() // move to the start of the expression
	lg.Value = p.parseExpression()

	if p.curToken.Type != token.RPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ')' after log argument on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	p.nextToken() // move past ')'

	return lg
}

func (p *Parser) parseExpression() ast.Expression {
	return p.parseUnary()
}

// parseAdditive parses left-associative chains of + and -
func (p *Parser) parseAdditive() ast.Expression {
	left := p.parseMultiplicitave()
	for p.curToken.Type == token.PLUS || p.curToken.Type == token.MINUS {
		op := p.curToken.Type
		line, col := p.curToken.Line, p.curToken.Col
		p.nextToken()
		right := p.parseMultiplicitave()
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
			Line:     line,
			Col:      col,
		}
	}
	return left
}

func (p *Parser) parseMultiplicitave() ast.Expression {
	left := p.parsePrimary()
	for p.curToken.Type == token.SLASH || p.curToken.Type == token.ASTERISK || p.curToken.Type == token.MODULUS {
		op := p.curToken.Type
		line, col := p.curToken.Line, p.curToken.Col
		p.nextToken()
		right := p.parsePrimary()
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
			Line:     line,
			Col:      col,
		}
	}
	return left
}

// parsePrimary parses literals and identifiers
func (p *Parser) parsePrimary() ast.Expression {
	switch p.curToken.Type {
	case token.STRING:
		lit := &ast.StringLiteral{Value: p.curToken.Literal}
		p.nextToken()
		return lit
	case token.INT:
		intVal, err := strconv.ParseInt(p.curToken.Literal, 10, 64)
		if err != nil {
			p.Errors = append(p.Errors, fmt.Sprintf("invalid int literal '%s' on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
			p.nextToken()
			return nil
		}
		lit := &ast.IntegerLiteral{Value: intVal}
		p.nextToken()
		return lit
	case token.BOOL:
		boolVal := p.curToken.Literal == "true"
		lit := &ast.BoolLiteral{Value: boolVal}
		p.nextToken()
		return lit
	case token.IDENT, token.LEN, token.INPUT:
		var expr ast.Expression = &ast.Identifier{Value: p.curToken.Literal, Line: p.curToken.Line, Col: p.curToken.Col}
		p.nextToken()
		// Handle dot notation: App.run or App.foo.bar
		for p.curToken.Type == token.DOT {
			p.nextToken()
			if p.curToken.Type != token.IDENT {
				p.Errors = append(p.Errors, fmt.Sprintf("expected identifier after '.' on line %d:%d", p.curToken.Line, p.curToken.Col))
				return nil
			}
			// Combine previous and current identifier
			if id, ok := expr.(*ast.Identifier); ok {
				expr = &ast.Identifier{
					Value: id.Value + "." + p.curToken.Literal,
					Line:  id.Line,
					Col:   id.Col,
				}
			}
			p.nextToken()
		}
		// Support function calls: foo(), len(), input(), etc.
		for p.curToken.Type == token.LPAREN {
			p.nextToken()
			args := []ast.Expression{}
			if p.curToken.Type != token.RPAREN {
				args = append(args, p.parseExpression())
				for p.curToken.Type == token.COMMA {
					p.nextToken()
					args = append(args, p.parseExpression())
				}
			}
			if p.curToken.Type != token.RPAREN {
				p.Errors = append(p.Errors, fmt.Sprintf("expected ')' after function call on line %d:%d", p.curToken.Line, p.curToken.Col))
				return nil
			}
			p.nextToken()
			expr = &ast.CallExpression{Function: expr, Arguments: args}
		}
		// Support arr[0] and chaining
		for p.curToken.Type == token.LBRACKET {
			p.nextToken()
			var start, end ast.Expression
			// xs[1:4], xs[:4], xs[1:], xs[:]
			if p.curToken.Type != token.COLON && p.curToken.Type != token.RBRACKET {
				start = p.parseExpression()
			}
			if p.curToken.Type == token.COLON {
				p.nextToken()
				if p.curToken.Type != token.RBRACKET {
					end = p.parseExpression()
				}
				if p.curToken.Type != token.RBRACKET {
					p.Errors = append(p.Errors, fmt.Sprintf("expected ']' after slice on line %d:%d", p.curToken.Line, p.curToken.Col))
					return nil
				}
				p.nextToken()
				expr = &ast.SliceExpression{Left: expr, Start: start, End: end}
			} else {
				if p.curToken.Type != token.RBRACKET {
					p.Errors = append(p.Errors, fmt.Sprintf("expected ']' after index on line %d:%d", p.curToken.Line, p.curToken.Col))
					return nil
				}
				p.nextToken()
				expr = &ast.IndexExpression{Left: expr, Index: start}
			}
		}
		return expr
	case token.LPAREN:
		p.nextToken()
		expr := p.parseExpression()
		if p.curToken.Type != token.RPAREN {
			p.Errors = append(p.Errors, fmt.Sprintf("expected ')' after expression on line %d:%d", p.curToken.Line, p.curToken.Col))
			return nil
		}
		p.nextToken()
		return expr
	case token.NIL:
		expr := &ast.NilLiteral{}
		p.nextToken()
		return expr
	case token.LBRACKET:
		elements := []ast.Expression{}
		p.nextToken()
		for p.curToken.Type != token.RBRACKET && p.curToken.Type != token.EOF {
			elements = append(elements, p.parseExpression())
			if p.curToken.Type == token.COMMA {
				p.nextToken()
			}
		}
		p.nextToken() // skip ']'
		return &ast.ArrayLiteral{Elements: elements}
	default:
		p.Errors = append(p.Errors, fmt.Sprintf("[PARSE PRIMARY] unexpected token '%s' in expression on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
		return nil
	}
}

func (p *Parser) parseComparison() ast.Expression {
	left := p.parseAdditive()
	for p.curToken.Type == token.EQ || p.curToken.Type == token.NEQ ||
		p.curToken.Type == token.LT || p.curToken.Type == token.GT ||
		p.curToken.Type == token.LTE || p.curToken.Type == token.GTE {
		op := p.curToken.Type
		line, col := p.curToken.Line, p.curToken.Col
		p.nextToken()
		right := p.parseAdditive()
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
			Line:     line,
			Col:      col,
		}
	}
	return left
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	line, col := p.curToken.Line, p.curToken.Col
	p.nextToken()
	value := p.parseExpression()
	return &ast.ReturnStatement{
		Value: value,
		Line:  line,
		Col:   col,
	}
}

func (p *Parser) parseIfStatement() *ast.IfStatement {
	is := &ast.IfStatement{Line: p.curToken.Line, Col: p.curToken.Col}

	p.nextToken() // move to condition
	// Parse the condition expression until '{'
	cond := p.parseExpression()
	is.IfCond = cond

	if p.curToken.Type != token.LBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after if condition on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	// Parse if body
	ifBody := p.parseBlock()

	// Parse elif blocks
	var elifConds []ast.Expression
	var elifBodies [][]ast.Statement
	for p.curToken.Type == token.ELIF {
		p.nextToken() // move to elif condition
		elifCond := p.parseExpression()
		elifConds = append(elifConds, elifCond)
		if p.curToken.Type != token.LBRACE {
			p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after elif condition on line %d:%d", p.curToken.Line, p.curToken.Col))
			return is
		}
		elifBody := p.parseBlock()
		elifBodies = append(elifBodies, elifBody)
	}

	// Parse else block
	var elseBody []ast.Statement
	if p.curToken.Type == token.ELSE {
		p.nextToken()
		if p.curToken.Type != token.LBRACE {
			p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after else on line %d:%d", p.curToken.Line, p.curToken.Col))
			return is
		}
		elseBody = p.parseBlock()
	}

	// Store bodies in the AST node (expand IfStatement struct as needed)
	is.IfBody = ifBody
	is.ElifConds = elifConds
	is.ElifBodies = elifBodies
	is.ElseBody = elseBody

	return is
}

func (p *Parser) parseBlock() []ast.Statement {
	stmts := []ast.Statement{}
	p.nextToken() // move past '{'
	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		var stmt ast.Statement
		switch p.curToken.Type {
		case token.LET:
			stmt = p.parseLetStatement()
		case token.FNC:
			stmt = p.parseFunctionStatement()
		case token.LOG:
			stmt = p.parseLogFunctionStatement()
		case token.RETURN:
			stmt = p.parseReturnStatement()
		case token.IF:
			stmt = p.parseIfStatement()
		case token.WHILE:
			stmt = p.parseWhileStatement()
		case token.FOR:
			stmt = p.parseForStatement()
		default:
			if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN_OP {
				stmt = p.parseAssignmentStatement()
			} else {
				expr := p.parseExpression()
				if expr != nil {
					stmt = &ast.ExpressionStatement{
						Expr: expr,
						Line: p.curToken.Line,
						Col:  p.curToken.Col,
					}
				} else {
					p.Errors = append(p.Errors, fmt.Sprintf("[PARSE BLOCK] unexpected token '%s' on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
					p.nextToken()
					continue
				}
			}
		}
		if stmt != nil {
			stmts = append(stmts, stmt)
		}
	}
	p.nextToken() // skip '}'
	return stmts
}

func (p *Parser) parseLogical() ast.Expression {
	left := p.parseComparison()
	for p.curToken.Type == token.AND || p.curToken.Type == token.OR {
		op := p.curToken.Type
		line, col := p.curToken.Line, p.curToken.Col
		p.nextToken()
		right := p.parseComparison()
		left = &ast.BinaryExpression{
			Left:     left,
			Operator: op,
			Right:    right,
			Line:     line,
			Col:      col,
		}
	}
	return left
}

func (p *Parser) parseUnary() ast.Expression {
	if p.curToken.Type == token.NOT || p.curToken.Type == token.MINUS {
		op := p.curToken.Type
		line, col := p.curToken.Line, p.curToken.Col
		p.nextToken()
		right := p.parseUnary()
		return &ast.UnaryExpression{
			Operator: op,
			Right:    right,
			Line:     line,
			Col:      col,
		}
	}
	return p.parseLogical()
}

func (p *Parser) parseAssignmentStatement() *ast.AssignmentStatement {
	line, col := p.curToken.Line, p.curToken.Col

	// Parse left side: could be identifier or index expression
	var left ast.Expression
	if p.curToken.Type == token.IDENT {
		left = &ast.Identifier{Value: p.curToken.Literal, Line: line, Col: col}
		p.nextToken()
		// Support xs[0] on left side
		for p.curToken.Type == token.LBRACKET {
			p.nextToken()
			index := p.parseExpression()
			if p.curToken.Type != token.RBRACKET {
				p.Errors = append(p.Errors, fmt.Sprintf("expected ']' after index on line %d:%d", p.curToken.Line, p.curToken.Col))
				return nil
			}
			p.nextToken()
			left = &ast.IndexExpression{Left: left, Index: index}
		}
	} else {
		p.Errors = append(p.Errors, fmt.Sprintf("expected identifier or index expression on line %d:%d", line, col))
		return nil
	}

	if p.curToken.Type != token.ASSIGN_OP {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '>>' after assignment target on line %d:%d", line, col))
		return nil
	}
	p.nextToken()
	value := p.parseExpression()

	// If left is identifier, set Name; if index, set Left
	name := ""
	if ident, ok := left.(*ast.Identifier); ok {
		name = ident.Value
	}

	return &ast.AssignmentStatement{
		Name:  name,
		Left:  left,
		Value: value,
		Line:  line,
		Col:   col,
	}
}

func (p *Parser) parseWhileStatement() *ast.WhileStatement {
	ws := &ast.WhileStatement{Line: p.curToken.Line, Col: p.curToken.Col}
	p.nextToken() // move to condition
	ws.Condition = p.parseExpression()
	if p.curToken.Type != token.LBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after while condition on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	ws.Body = p.parseBlock()
	return ws
}

func (p *Parser) parseForStatement() *ast.ForStatement {
	fs := &ast.ForStatement{Line: p.curToken.Line, Col: p.curToken.Col}
	p.nextToken() // move to init

	// Parse init statement (let or assignment)
	var init ast.Statement
	if p.curToken.Type == token.LET {
		init = p.parseLetStatement()
	} else if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN_OP {
		init = p.parseAssignmentStatement()
	} else {
		p.Errors = append(p.Errors, fmt.Sprintf("expected init statement in for loop on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	fs.Init = init

	if p.curToken.Type != token.SEMICOLON {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ';' after for-init on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	p.nextToken()

	// Parse condition
	fs.Condition = p.parseExpression()
	if p.curToken.Type != token.SEMICOLON {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ';' after for-condition on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	p.nextToken()

	// Parse post statement (assignment)
	if p.curToken.Type == token.IDENT && p.peekToken.Type == token.ASSIGN_OP {
		fs.Post = p.parseAssignmentStatement()
	} else {
		p.Errors = append(p.Errors, fmt.Sprintf("expected post statement in for loop on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	if p.curToken.Type != token.LBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after for-post on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}
	fs.Body = p.parseBlock()
	return fs
}

func (p *Parser) parsePackageStatement() *ast.PackageStatement {
	p.nextToken()

	if p.curToken.Type != token.IDENT {
		msg := "expected package name after 'package'"
		p.Errors = append(p.Errors, msg)
		return nil
	}

	parts := []string{p.curToken.Literal}

	// Keep parsing dot-separated identifiers
	for p.peekToken.Type == token.DOT {
		p.nextToken() // consume '.'
		p.nextToken() // move to next IDENT
		if p.curToken.Type != token.IDENT {
			msg := "expected identifier after '.' in package path"
			p.Errors = append(p.Errors, msg)
			return nil
		}
		parts = append(parts, p.curToken.Literal)
	}
	p.nextToken()

	pkg := &ast.PackageStatement{Name: strings.Join(parts, ".")}

	return pkg
}

func (p *Parser) parseImportStatement() *ast.ImportStatement {
	p.nextToken()

	if p.curToken.Type != token.IDENT {
		msg := "expected import path after 'import'"
		p.Errors = append(p.Errors, msg)
		return nil
	}

	parts := []string{p.curToken.Literal}

	// Keep parsing dot-separated identifiers
	for p.peekToken.Type == token.DOT {
		p.nextToken() // consume '.'
		p.nextToken() // move to next IDENT
		if p.curToken.Type != token.IDENT {
			msg := "expected identifier after '.' in import path"
			p.Errors = append(p.Errors, msg)
			return nil
		}
		parts = append(parts, p.curToken.Literal)
	}
	p.nextToken()

	ipt := &ast.ImportStatement{Path: strings.Join(parts, ".")}

	return ipt
}

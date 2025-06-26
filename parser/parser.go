package parser

import (
	"fmt"
	"strconv"

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
		if p.curToken.Type == token.RBRACE {
			p.nextToken()
			continue
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

	// (skip parameter parsing for now)
	for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
		p.nextToken()
	}
	if p.curToken.Type != token.RPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ')' after parameters on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

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
	fn.Body = []ast.Statement{}
	p.nextToken()
	for {
		if p.curToken.Type == token.RBRACE || p.curToken.Type == token.EOF {
			break
		}

		// Use the main statement parsing logic
		var stmt ast.Statement
		if p.curToken.Type == token.LET {
			stmt = p.parseLetStatement()
		} else if p.curToken.Type == token.LOG {
			stmt = p.parseLogFunctionStatement()
		} else if p.curToken.Type == token.FNC {
			stmt = p.parseFunctionStatement()
		} else if p.curToken.Type == token.RETURN {
			stmt = p.parseReturnStatement()
		} else {
			// Try to parse as an expression (e.g., function call)
			expr := p.parseExpression()
			if expr != nil {
				stmt = &ast.ExpressionStatement{
					Expr: expr,
					Line: p.curToken.Line,
					Col:  p.curToken.Col,
				}
			} else {
				p.Errors = append(p.Errors, fmt.Sprintf("[PARSE FNC STATEMENT] unexpected token '%s' in function body on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
				p.nextToken()
				continue
			}
		}

		if stmt != nil {
			fn.Body = append(fn.Body, stmt)
		}
	}

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
	return p.parseAdditive()
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
	case token.IDENT:
		var expr ast.Expression = &ast.Identifier{Value: p.curToken.Literal, Line: p.curToken.Line, Col: p.curToken.Col}
		p.nextToken()
		// Support chaining: foo(), foo()(), etc.
		for p.curToken.Type == token.LPAREN {
			p.nextToken() // move to first arg or RPAREN
			args := []ast.Expression{}
			for p.curToken.Type != token.RPAREN && p.curToken.Type != token.EOF {
				args = append(args, p.parseExpression())
				if p.curToken.Type == token.COMMA {
					p.nextToken()
				}
			}
			p.nextToken() // skip ')'
			expr = &ast.CallExpression{Function: expr, Arguments: args}
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
	default:
		p.Errors = append(p.Errors, fmt.Sprintf("[PARSE PRIMARY] unexpected token '%s' in expression on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
		return nil
	}
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

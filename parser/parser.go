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
		var stmt ast.Statement
		if p.curToken.Type == token.LET {
			stmt = p.parseLetStatement()
		} else if p.curToken.Type == token.FNC {
			stmt = p.parseFunctionStatement()
		} else if p.curToken.Type == token.LOG {
			stmt = p.parseLogFunctionStatement()
		} else {
			p.Errors = append(p.Errors, fmt.Sprintf("unexpected token '%s' on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
			p.nextToken()
			continue
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}

		p.nextToken()
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
        p.Errors = append(p.Errors, fmt.Sprintf("expected assignment operator '>>' on line %d:%d", p.curToken.Line, p.curToken.Col))
        return nil
    }
    p.nextToken()

    var value ast.Expression
    switch p.curToken.Type {
    case token.STRING:
        value = &ast.StringLiteral{Value: p.curToken.Literal}
    case token.INT:
        intVal, _ := strconv.ParseInt(p.curToken.Literal, 0, 64)
        value = &ast.IntegerLiteral{Value: intVal}
    case token.BOOL:
        boolVal := p.curToken.Literal == "true"
        value = &ast.BoolLiteral{Value: boolVal}
    default:
        p.Errors = append(p.Errors, fmt.Sprintf("unexpected literal type on line %d:%d", p.curToken.Line, p.curToken.Col))
        return nil
    }

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

	p.nextToken() // move to {
	if p.curToken.Type != token.LBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '{' after return type on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	// Parse body
	fn.Body = []ast.Statement{}
	for {
		p.nextToken()
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
		} else {
			p.Errors = append(p.Errors, fmt.Sprintf("unexpected token '%s' in function body on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
			continue
		}

		if stmt != nil {
			fn.Body = append(fn.Body, stmt)
		}
	}

	return fn
}

func (p *Parser) parseLogFunctionStatement() *ast.LogFunction {
	lg := &ast.LogFunction{ Line: p.curToken.Line, Col: p.curToken.Col }

	p.nextToken() // move to (
	if p.curToken.Type != token.LPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '(' after 'log' on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	p.nextToken()
	switch p.curToken.Type {
	case token.STRING, token.INT, token.BOOL, token.IDENT:
		lg.Value = &ast.Identifier{
			Value: p.curToken.Literal,
			Type:  p.curToken.Type,
			Line:  p.curToken.Line,
			Col:   p.curToken.Col,
		}
	default:
		p.Errors = append(p.Errors, fmt.Sprintf("invalid log argument on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	p.nextToken() // move past the argument

		if p.curToken.Type != token.RPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ')' after log argument on line %d:%d", p.curToken.Line, p.curToken.Col))
		return nil
	}

	return lg
}

// func (p *Parser) parseExpression() ast.Expression {
// 	left := 
// }
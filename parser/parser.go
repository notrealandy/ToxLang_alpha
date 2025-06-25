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

	prefixParseFns map[token.TokenType]prefixParseFn
	infixParseFns  map[token.TokenType]infixParseFn
}

// Pratt parser function types
type (
	prefixParseFn func() ast.Expression
	infixParseFn  func(ast.Expression) ast.Expression
)

// Precedence constants
const (
	_ int = iota
	LOWEST
	EQUALS      // == or !=
	LESSGREATER // > or < or <= or >=
	SUM         // + or -
	PRODUCT     // * or /
	PREFIX      // -X or !X
	CALL        // myFunction(X) (Future)
	INDEX       // array[index] (Future)
)

// Precedence map for token types
var precedences = map[token.TokenType]int{
	token.EQ:       EQUALS,
	token.NEQ:      EQUALS,
	token.LT:       LESSGREATER,
	token.GT:       LESSGREATER,
	token.LTE:      LESSGREATER,
	token.GTE:      LESSGREATER,
	token.PLUS:     SUM,
	token.MINUS:    SUM,
	token.SLASH:    PRODUCT,
	token.ASTERISK: PRODUCT,
	// token.LPAREN: CALL, // For function calls
	// token.LBRACKET: INDEX, // For array indexing
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{
		l:              l,
		Errors:         []string{},
		prefixParseFns: make(map[token.TokenType]prefixParseFn),
		infixParseFns:  make(map[token.TokenType]infixParseFn),
	}

	// Register prefix parsing functions
	p.registerPrefix(token.IDENT, p.parseIdentifier)
	p.registerPrefix(token.INT, p.parseIntegerLiteral)
	p.registerPrefix(token.STRING, p.parseStringLiteral)
	p.registerPrefix(token.BOOL, p.parseBooleanLiteral)
	p.registerPrefix(token.BANG, p.parsePrefixExpression)
	p.registerPrefix(token.MINUS, p.parsePrefixExpression)
	p.registerPrefix(token.LPAREN, p.parseGroupedExpression)

	// Register infix parsing functions for binary operators
	p.registerInfix(token.PLUS, p.parseInfixExpression)
	p.registerInfix(token.MINUS, p.parseInfixExpression)
	p.registerInfix(token.SLASH, p.parseInfixExpression)
	p.registerInfix(token.ASTERISK, p.parseInfixExpression)
	p.registerInfix(token.EQ, p.parseInfixExpression)
	p.registerInfix(token.NEQ, p.parseInfixExpression)
	p.registerInfix(token.LT, p.parseInfixExpression)
	p.registerInfix(token.GT, p.parseInfixExpression)
	p.registerInfix(token.LTE, p.parseInfixExpression)
	p.registerInfix(token.GTE, p.parseInfixExpression)

	// Initialize curToken and peekToken
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) registerPrefix(tokenType token.TokenType, fn prefixParseFn) {
	p.prefixParseFns[tokenType] = fn
}

func (p *Parser) registerInfix(tokenType token.TokenType, fn infixParseFn) {
	p.infixParseFns[tokenType] = fn
}

func (p *Parser) peekPrecedence() int {
	if pr, ok := precedences[p.peekToken.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *Parser) curPrecedence() int {
	if pr, ok := precedences[p.curToken.Type]; ok {
		return pr
	}
	return LOWEST
}

func (p *Parser) parseExpression(precedence int) ast.Expression {
	prefix := p.prefixParseFns[p.curToken.Type]
	if prefix == nil {
		p.Errors = append(p.Errors, fmt.Sprintf("no prefix parse function for %s found on line %d:%d", p.curToken.Type, p.curToken.Line, p.curToken.Col))
		return nil
	}
	leftExp := prefix()

	for precedence < p.peekPrecedence() { // No semicolon at the end of the condition for `for`
		infix := p.infixParseFns[p.peekToken.Type]
		if infix == nil {
			return leftExp
		}
		p.nextToken() // Consume the infix operator or the token that starts the next part of expression
		leftExp = infix(leftExp)
	}
	return leftExp
}

func (p *Parser) nextToken() {
	p.curToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

// Prefix parsing functions
func (p *Parser) parseIdentifier() ast.Expression {
	return &ast.Identifier{Value: p.curToken.Literal, Line: p.curToken.Line, Col: p.curToken.Col}
}

func (p *Parser) parseIntegerLiteral() ast.Expression {
	lit := &ast.IntegerLiteral{}
	value, err := strconv.ParseInt(p.curToken.Literal, 0, 64)
	if err != nil {
		p.Errors = append(p.Errors, fmt.Sprintf("could not parse %q as integer on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
		return nil
	}
	lit.Value = value
	return lit
}

func (p *Parser) parseStringLiteral() ast.Expression {
	return &ast.StringLiteral{Value: p.curToken.Literal} // Assuming lexer handles unquoting
}

func (p *Parser) parseBooleanLiteral() ast.Expression {
	val := p.curToken.Literal == "true" // Lexer ensures it's "true" or "false" if token.BOOL
	return &ast.BoolLiteral{Value: val}
}

func (p *Parser) parsePrefixExpression() ast.Expression {
	expression := &ast.PrefixExpression{
		Operator: p.curToken.Literal,
		Line:     p.curToken.Line,
		Col:      p.curToken.Col,
	}
	p.nextToken() // Consume the prefix operator (e.g., "-" or "!")
	expression.Right = p.parseExpression(PREFIX)
	return expression
}

func (p *Parser) parseGroupedExpression() ast.Expression {
	line := p.curToken.Line
	col := p.curToken.Col
	p.nextToken() // Consume '('

	exp := p.parseExpression(LOWEST)

	if p.peekToken.Type != token.RPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected ')' to close grouped expression on line %d:%d, got %s instead of RPAREN", p.peekToken.Line, p.peekToken.Col, p.peekToken.Type))
		return nil
	}
	p.nextToken() // Consume ')'

	return &ast.GroupedExpression{Expression: exp, Line: line, Col: col}
}

func (p *Parser) parseInfixExpression(left ast.Expression) ast.Expression {
	expression := &ast.InfixExpression{
		Left:     left,
		Operator: p.curToken.Literal,
		Line:     p.curToken.Line,
		Col:      p.curToken.Col,
	}
	precedence := p.curPrecedence()
	p.nextToken() // Consume the infix operator
	expression.Right = p.parseExpression(precedence)
	return expression
}


func (p *Parser) ParseProgram() []ast.Statement {
	var statements []ast.Statement

	for p.curToken.Type != token.EOF {
		var stmt ast.Statement
		switch p.curToken.Type {
		case token.LET:
			stmt = p.parseLetStatement()
		case token.FNC:
			stmt = p.parseFunctionStatement()
		case token.RETURN: // Should only be allowed inside functions, handled by parseBlockStatement
			p.Errors = append(p.Errors, fmt.Sprintf("unexpected 'return' statement outside function on line %d:%d", p.curToken.Line, p.curToken.Col))
			p.nextToken() // consume return
			// Consume until a common statement starter or EOF to allow further parsing
			for p.curToken.Type != token.LET && p.curToken.Type != token.FNC && p.curToken.Type != token.EOF {
				p.nextToken()
			}
			continue
		default:
			p.Errors = append(p.Errors, fmt.Sprintf("unexpected token '%s' at top level on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
			p.nextToken()
			continue
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken() // Crucial: advance token after successfully parsing a statement
	}

	return statements
}

func (p *Parser) parseLetStatement() *ast.LetStatement {
	stmt := &ast.LetStatement{Line: p.curToken.Line, Col: p.curToken.Col}
	// curToken is LET
	p.nextToken() // consume LET

	if p.curToken.Type != token.IDENT {
		p.Errors = append(p.Errors, fmt.Sprintf("expected identifier after 'let' on line %d:%d, got %s", p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	stmt.Name = p.curToken.Literal
	p.nextToken() // consume IDENT

	if p.curToken.Type != token.TYPE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected type for 'let %s' on line %d:%d, got %s", stmt.Name, p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	stmt.Type = p.curToken.Literal
	p.nextToken() // consume TYPE

	if p.curToken.Type != token.ASSIGN_OP {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '>>' for 'let %s' on line %d:%d, got %s", stmt.Name, p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	p.nextToken() // consume >>

	// Parse the expression for the value
	stmt.Value = p.parseExpression(LOWEST)
	if stmt.Value == nil { // Check if parseExpression returned an error indicator
		// Error already recorded by parseExpression or its children
		return nil
	}

	// The p.nextToken() that used to be here to consume the literal is now implicitly
	// handled by the parseExpression consuming all its necessary tokens.
	// The main ParseProgram loop (or block statement loop) will call p.nextToken()
	// after this statement is fully parsed.

	// If the expression parsing expects a semicolon or newline to terminate,
	// that would be handled after the call to p.parseExpression here.
	// For now, we assume expressions are consumed correctly and the next token is
	// either EOF or the start of a new statement.

	return stmt
}

func (p *Parser) parseFunctionParameters() []*ast.ParameterLiteral {
	params := []*ast.ParameterLiteral{}

	if p.peekToken.Type == token.RPAREN { // No parameters
		p.nextToken() // consume ( to move to )
		return params
	}

	p.nextToken() // consume (

	for {
		if p.curToken.Type != token.IDENT {
			p.Errors = append(p.Errors, fmt.Sprintf("expected parameter name (identifier) on line %d:%d, got %s", p.curToken.Line, p.curToken.Col, p.curToken.Literal))
			return nil // Error, stop parsing params
		}
		paramName := p.curToken.Literal
		paramLine := p.curToken.Line
		paramCol := p.curToken.Col
		p.nextToken() // consume param name

		if p.curToken.Type != token.TYPE {
			p.Errors = append(p.Errors, fmt.Sprintf("expected type for parameter '%s' on line %d:%d, got %s", paramName, p.curToken.Line, p.curToken.Col, p.curToken.Literal))
			return nil // Error
		}
		paramType := p.curToken.Literal
		p.nextToken() // consume param type

		params = append(params, &ast.ParameterLiteral{Name: paramName, Type: paramType, Line: paramLine, Col: paramCol})

		if p.curToken.Type == token.RPAREN {
			break
		} else if p.curToken.Type == token.COMMA {
			p.nextToken() // consume comma, expect another param
		} else {
			p.Errors = append(p.Errors, fmt.Sprintf("expected ',' or ')' after parameter on line %d:%d, got %s", p.curToken.Line, p.curToken.Col, p.curToken.Literal))
			return nil // Error
		}
	}
	// curToken is now RPAREN
	return params
}

func (p *Parser) parseReturnStatement() *ast.ReturnStatement {
	stmt := &ast.ReturnStatement{Line: p.curToken.Line, Col: p.curToken.Col}
	// curToken is RETURN
	p.nextToken() // consume RETURN

	stmt.ReturnValue = p.parseExpression(LOWEST)
	if stmt.ReturnValue == nil { // Check if parseExpression returned an error indicator
		// Error already recorded by parseExpression or its children
		return nil
	}

	// Similar to LetStatement, parseExpression consumes its tokens.
	// The loop in parseBlockStatement will call p.nextToken() after this.
	return stmt
}

func (p *Parser) parseBlockStatement() []ast.Statement {
	var statements []ast.Statement
	// curToken is LBRACE
	p.nextToken() // consume LBRACE

	for p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
		var stmt ast.Statement
		switch p.curToken.Type {
		case token.LET:
			stmt = p.parseLetStatement()
		case token.RETURN:
			stmt = p.parseReturnStatement()
		// TODO: Add other statement types like if, for, expression statements
		default:
			p.Errors = append(p.Errors, fmt.Sprintf("unexpected token '%s' inside block on line %d:%d", p.curToken.Literal, p.curToken.Line, p.curToken.Col))
			// Skip to next potential statement to allow finding more errors
			for p.curToken.Type != token.LET && p.curToken.Type != token.RETURN && p.curToken.Type != token.RBRACE && p.curToken.Type != token.EOF {
				p.nextToken()
			}
			if p.curToken.Type == token.RBRACE || p.curToken.Type == token.EOF { // Avoid infinite loop if error is at the end
				continue
			}
			// If we skipped tokens, we need to advance to the next token to attempt parsing it.
			// However, if parseLetStatement or parseReturnStatement is called, they expect current token to be LET/RETURN
			// This part needs careful handling of token advancement after an error.
			// For now, if we hit default, we consume one token and continue the loop.
			// The next iteration might then parse it or error again.
			// A more robust approach might be to sync to the next known statement keyword or '}'.
			p.nextToken()
			continue
		}

		if stmt != nil {
			statements = append(statements, stmt)
		}
		p.nextToken() // Consume the last token of the statement (e.g. literal for let/return)
	}

	if p.curToken.Type != token.RBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '}' to close block on line %d:%d, got %s", p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		// We don't return nil here because we might have parsed some valid statements.
	}
	// curToken is RBRACE (or EOF if error)
	return statements
}

func (p *Parser) parseFunctionStatement() *ast.FunctionStatement {
	fn := &ast.FunctionStatement{Line: p.curToken.Line, Col: p.curToken.Col}
	// curToken is FNC
	p.nextToken() // consume FNC

	if p.curToken.Type != token.IDENT {
		p.Errors = append(p.Errors, fmt.Sprintf("expected function name (identifier) after 'fnc' on line %d:%d, got %s", p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	fn.Name = p.curToken.Literal
	p.nextToken() // consume function name

	if p.curToken.Type != token.LPAREN {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '(' after function name '%s' on line %d:%d, got %s", fn.Name, p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	// p.nextToken() // consume LPAREN is done by parseFunctionParameters
	fn.Parameters = p.parseFunctionParameters()
	if fn.Parameters == nil && len(p.Errors) > 0 { // Check if parseFunctionParameters failed
		// Error already recorded by parseFunctionParameters
		return nil
	}
	// After parseFunctionParameters, curToken is RPAREN

	p.nextToken() // consume RPAREN

	// Expect '>>'
	if p.curToken.Type != token.ASSIGN_OP {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '>>' after function parameters for '%s' on line %d:%d, got %s", fn.Name, p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	p.nextToken() // consume '>>'

	// Optional Return Type
	if p.curToken.Type == token.TYPE { // e.g. string, int, bool
		fn.ReturnType = p.curToken.Literal
		p.nextToken() // consume return type
	} else {
		fn.ReturnType = "" // Or a specific "void" type if you add one
	}

	// Expect '{'
	if p.curToken.Type != token.LBRACE {
		p.Errors = append(p.Errors, fmt.Sprintf("expected '{' to start function body for '%s' on line %d:%d, got %s", fn.Name, p.curToken.Line, p.curToken.Col, p.curToken.Literal))
		return nil
	}
	fn.Body = p.parseBlockStatement()
	// After parseBlockStatement, curToken is RBRACE (or EOF if malformed)
	// The main ParseProgram loop will call nextToken after this function returns.
	return fn
}

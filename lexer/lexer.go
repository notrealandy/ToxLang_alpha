package lexer

import (
	"strings"

	"github.com/notrealandy/tox/token"
)

type Lexer struct {
	input        string
	position     int  // current char position
	readPosition int  // next char position
	ch           byte // current char under examination
	line         int  // track line number
	col          int  // track column number
}

// prepares the string for tokenization
func New(input string) *Lexer {
	l := &Lexer{input: input, line: 1, col: 0}
	l.readChar()
	return l
}

// a function to read characters in string
func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}

	if l.ch == '\n' {
		l.col = 0
	} else {
		l.col++
	}

	l.position = l.readPosition
	l.readPosition++
}

// a function to skip whitespaces and other non-important characters in code
func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' || (l.ch == '/' && l.peekChar() == '/') {
		if l.ch == '/' && l.peekChar() == '/' {
			// Skip the comment
			for l.ch != '\n' && l.ch != 0 {
				l.readChar()
			}
		} else {
			if l.ch == '\n' {
				l.line++
			}
			l.readChar()
		}
	}
}

// a function that allows you to look which character is next
func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

// a function that reads strings
func (l *Lexer) readString() string {
	position := l.position + 1
	for {
		l.readChar()
		if l.ch == '"' || l.ch == 0 {
			break
		}
	}
	str := l.input[position:l.position]
	l.readChar()
	return str
}

// a function that reads numbers
func (l *Lexer) readNumber() string {
	pos := l.position
	for isDigit(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func isIdentChar(ch byte) bool {
	return isLetter(ch) || isDigit(ch)
}

// a function to read identifiers like `let x string`, each word is identifier
// for example:
// 1. let
// 2. x
// 3. string
func (l *Lexer) readIdentifier() string {
	pos := l.position
	if !isLetter(l.ch) && l.ch != '_' {
		return ""
	}
	l.readChar()
	for isIdentChar(l.ch) {
		l.readChar()
	}
	// If the next two characters are '[' and ']', include them
	if l.ch == '[' && l.peekChar() == ']' {
		l.readChar() // skip '['
		l.readChar() // skip ']'
	}
	return l.input[pos:l.position]
}

// a function that checks if the character is letter or not
func isLetter(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ch == '_'
}

// a function that checks if the character is a digit
func isDigit(ch byte) bool {
	return '0' <= ch && ch <= '9'
}

// a function to validate token types
func lookupIdent(ident string) token.TokenType {
	switch strings.ToLower(ident) {
	case "let":
		return token.LET
	case "fnc":
		return token.FNC
	case "log":
		return token.LOG
	case "string", "int", "bool", "void", "int[]", "string[]", "bool[]":
		return token.TYPE
	case "true", "false":
		return token.BOOL
	case "return":
		return token.RETURN
	case "nil":
		return token.NIL
	case "if":
		return token.IF
	case "elif":
		return token.ELIF
	case "else":
		return token.ELSE
	case "while":
		return token.WHILE
	case "for":
		return token.FOR
	case "len":
		return token.LEN
	default:
		return token.IDENT
	}
}

// Tokenizer
func (l *Lexer) NextToken() token.Token {
	l.skipWhitespace()

	var tok token.Token
	startCol := l.col

	switch l.ch {
	case '>':
		if l.peekChar() == '>' {
			ch := l.ch
			l.readChar()
			tok = token.Token{Type: token.ASSIGN_OP, Literal: string(ch) + string(l.ch), Line: l.line, Col: startCol}
		} else if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.GTE, Literal: ">=", Line: l.line, Col: startCol}
		} else {
			tok = token.Token{Type: token.GT, Literal: ">", Line: l.line, Col: startCol} // <-- FIXED HERE
		}
	case '<':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.LTE, Literal: "<=", Line: l.line, Col: startCol}
		} else {
			tok = token.Token{Type: token.LT, Literal: "<", Line: l.line, Col: startCol}
		}
	case '=':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.EQ, Literal: "==", Line: l.line, Col: startCol}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: l.line, Col: startCol}
		}
	case '!':
		if l.peekChar() == '=' {
			l.readChar()
			tok = token.Token{Type: token.NEQ, Literal: "!=", Line: l.line, Col: startCol}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: l.line, Col: startCol}
		}
	case '"':
		tok.Type = token.STRING
		tok.Literal = l.readString()
		tok.Line = l.line
		tok.Col = startCol
		return tok
	case '(':
		tok = token.Token{Type: token.LPAREN, Literal: "(", Line: l.line, Col: startCol}
	case ')':
		tok = token.Token{Type: token.RPAREN, Literal: ")", Line: l.line, Col: startCol}
	case '{':
		tok = token.Token{Type: token.LBRACE, Literal: "{", Line: l.line, Col: startCol}
	case '}':
		tok = token.Token{Type: token.RBRACE, Literal: "}", Line: l.line, Col: startCol}
	case '+':
		tok = token.Token{Type: token.PLUS, Literal: "+", Line: l.line, Col: startCol}
	case '-':
		tok = token.Token{Type: token.MINUS, Literal: "-", Line: l.line, Col: startCol}
	case '*':
		tok = token.Token{Type: token.ASTERISK, Literal: "*", Line: l.line, Col: startCol}
	case '/':
		tok = token.Token{Type: token.SLASH, Literal: "/", Line: l.line, Col: startCol}
	case '%':
		tok = token.Token{Type: token.MODULUS, Literal: "%", Line: l.line, Col: startCol}
	case '&':
		if l.peekChar() == '&' {
			l.readChar()
			tok = token.Token{Type: token.AND, Literal: "&&", Line: l.line, Col: startCol}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: l.line, Col: startCol}
		}
	case '|':
		if l.peekChar() == '|' {
			l.readChar()
			tok = token.Token{Type: token.OR, Literal: "||", Line: l.line, Col: startCol}
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: l.line, Col: startCol}
		}
	case ',':
		tok = token.Token{Type: token.COMMA, Literal: ",", Line: l.line, Col: startCol}
	case ';':
		tok = token.Token{Type: token.SEMICOLON, Literal: ";", Line: l.line, Col: startCol}
	case '[':
		tok = token.Token{Type: token.LBRACKET, Literal: "[", Line: l.line, Col: startCol}
	case ']':
		tok = token.Token{Type: token.RBRACKET, Literal: "]", Line: l.line, Col: startCol}
	case ':':
		tok = token.Token{Type: token.COLON, Literal: ":", Line: l.line, Col: startCol}
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) {
			literal := l.readIdentifier()
			tok.Type = lookupIdent(literal)
			tok.Literal = literal
			tok.Line = l.line
			tok.Col = startCol
			return tok
		} else if isDigit(l.ch) {
			tok.Type = token.INT
			tok.Literal = l.readNumber()
			tok.Line = l.line
			tok.Col = startCol
			return tok
		} else {
			tok = token.Token{Type: token.ILLEGAL, Literal: string(l.ch), Line: l.line, Col: startCol}
		}
	}
	l.readChar()
	return tok
}

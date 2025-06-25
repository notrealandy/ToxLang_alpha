package token

type TokenType string

const (
	LET = "LET" // reserved keyword
	FNC = "FNC" // function keyword
	LPAREN = "LPAREN" // (
	RPAREN = "RPAREN" // )
	LBRACE = "LBRACE" // {
	RBRACE = "RBRACE" // }
	IDENT = "IDENT" // variables/functions
	TYPE = "TYPE" // string, int, bool
	ASSIGN_OP = "ASSIGN_OP" // >>
	STRING = "STRING" // string literal, e.g. "test"
	INT = "INT" // int literal, e.g. 3
	BOOL = "BOOL" // bool literal, e.g. true/false
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"
)

type Token struct {
	Type TokenType
	Literal string
	Line int
	Col int
}
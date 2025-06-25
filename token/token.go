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
	RETURN = "RETURN" // return keyword
	COMMA  = "COMMA"   // ,

	// Operators
	PLUS     = "+"
	MINUS    = "-"
	ASTERISK = "*"
	SLASH    = "/"

	LT  = "<"
	GT  = ">"
	EQ  = "=="
	NEQ = "!=" // Using NEQ for "Not Equal" to avoid confusion with BANG
	LTE = "<=" // Less Than or Equal
	GTE = ">=" // Greater Than or Equal

	// Logical
	BANG = "!" // For logical NOT, distinct from NEQ
	// AND = "&&" // For logical AND
	// OR  = "||" // For logical OR

	ILLEGAL = "ILLEGAL"
	EOF     = "EOF"
)

type Token struct {
	Type TokenType
	Literal string
	Line int
	Col int
}
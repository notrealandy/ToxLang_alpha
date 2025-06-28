package token

type TokenType string

const (
	LET = "LET" // reserved keyword
	FNC = "FNC" // function keyword
	LOG = "LOG" // native function keyword
	LEN = "LEN" // native function keyword
	INPUT = "INPUT" // native function keyword
	IF = "IF" // if statement
	ELIF = "ELIF" // else if statement
	ELSE = "ELSE" // else statement
	WHILE = "WHILE" // while loop
	FOR = "FOR" // for loop
	RETURN = "RETURN"
	LPAREN = "LPAREN" // (
	RPAREN = "RPAREN" // )
	LBRACE = "LBRACE" // {
	RBRACE = "RBRACE" // }
	LBRACKET = "LBRACKET" // [
	RBRACKET = "RBRACKET" // ]
	IDENT = "IDENT" // variables/functions
	TYPE = "TYPE" // string, int, bool
	ASSIGN_OP = "ASSIGN_OP" // >>
	STRING = "STRING" // string literal, e.g. "test"
	INT = "INT" // int literal, e.g. 3
	BOOL = "BOOL" // bool literal, e.g. true/false
	FNCVOID = "FNCVOID" // function return type void
	PLUS = "+"
	MINUS = "-"
	ASTERISK = "*"
	SLASH = "/"
	MODULUS = "%"
	NIL = "NIL"
	COMMA = "COMMA" // ,
	EQ = "EQ" // ==
	NEQ = "NEQ" // !=
	LT = "LT" // <
	LTE = "LTE" // <=
	GT = "GT" // >
	GTE = "GTE" // >=
	AND = "AND" // &&
	OR = "OR" // ||
	NOT = "NOT" // !
	SEMICOLON = "SEMICOLON" // ;
	COLON = "COLON" // :
	ILLEGAL = "ILLEGAL"
	EOF = "EOF"
)

type Token struct {
	Type TokenType
	Literal string
	Line int
	Col int
}
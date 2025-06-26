package ast

import "github.com/notrealandy/tox/token"

type Statement interface{}

type LetStatement struct {
	Name  string     // variable name
	Type  string     // type as declared
	Value Expression // the value assigned
	Line  int
	Col   int
}

type FunctionStatement struct {
	Name   string // function name
	Params []string
	Body   []Statement
	ReturnType string
	Line   int
	Col    int
}

type LogFunction struct {
	Line  int
	Col   int
	Value Expression
}

type ReturnStatement struct {
	Value Expression
	Line  int
	Col   int
}

type Identifier struct {
	Value string
	Type  token.TokenType
	Line  int
	Col   int
}

type NilLiteral struct{}

type Expression interface {
	expressionNode()
}

type BinaryExpression struct {
	Left     Expression
	Operator token.TokenType
	Right    Expression
	Line     int
	Col      int
}

// Define type check string value
type StringLiteral struct {
	Value string
}

// Define type check int value
type IntegerLiteral struct {
	Value int64
}

// Define type check bool value
type BoolLiteral struct {
	Value bool
}

func (lf *LogFunction) statementNode()     {}
func (rs *ReturnStatement) statementNode() {}

func (id *Identifier) expressionNode()       {}
func (il *IntegerLiteral) expressionNode()   {}
func (sl *StringLiteral) expressionNode()    {}
func (bl *BoolLiteral) expressionNode()      {}
func (be *BinaryExpression) expressionNode() {}
func (nl *NilLiteral) expressionNode() {}
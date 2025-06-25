package ast

type Statement interface{}

type LetStatement struct {
	Name string // variable name
	Type string // type as declared
	Value Expression // the value assigned
	Line int
	Col int
}

type FunctionStatement struct {
	Name string // function name
	Params []string
	Body []Statement
	Line int
	Col int
}

type Expression interface {
	expressionNode()
}

// Define type check string value
type StringLiteral struct {
	Value string
}

func (sl *StringLiteral) expressionNode() {}

// Define type check int value
type IntegerLiteral struct {
	Value int64
}

func (il *IntegerLiteral) expressionNode() {}

// Define type check bool value
type BoolLiteral struct {
	Value bool
}

func (bl *BoolLiteral) expressionNode() {}
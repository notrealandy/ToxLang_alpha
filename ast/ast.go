package ast

type Statement interface{}

type LetStatement struct {
	Name string // variable name
	Type string // type as declared
	Value Expression // the value assigned
	Line int
	Col int
}

type ParameterLiteral struct {
	Name string
	Type string
	Line int
	Col  int
}

func (pl *ParameterLiteral) expressionNode() {} // Parameters can be part of expressions in some contexts, or used in type checking
func (pl *ParameterLiteral) statementNode()  {} // Can also be considered a form of declaration

type FunctionStatement struct {
	Name       string // function name
	Parameters []*ParameterLiteral
	ReturnType string // e.g., "string", "int", "void" (if void is empty string)
	Body       []Statement
	Line       int
	Col        int
}

type ReturnStatement struct {
	ReturnValue Expression // The value to return
	Line        int
	Col         int
}

func (rs *ReturnStatement) statementNode() {}

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

// Identifier (for variables, function names in expressions, etc.)
type Identifier struct {
	Value string // e.g., variable name "x"
	Line  int
	Col   int
}

func (i *Identifier) expressionNode() {}

// PrefixExpression (e.g., -5, !true)
type PrefixExpression struct {
	Operator string // e.g., "-", "!"
	Right    Expression
	Line     int
	Col      int
}

func (pe *PrefixExpression) expressionNode() {}

// InfixExpression (e.g., 5 + 5, x == y)
type InfixExpression struct {
	Left     Expression
	Operator string // e.g., "+", "==", "<"
	Right    Expression
	Line     int // Line/Col typically of the operator token
	Col      int
}

func (ie *InfixExpression) expressionNode() {}

// GroupedExpression (e.g., (5 + 5))
// This helps in parsing to override precedence, and might be useful for AST clarity
// or can sometimes be optimized away by the parser directly building the correct precedence tree.
// For a Pratt parser, this is often explicitly parsed.
type GroupedExpression struct {
	Expression Expression
	Line       int // Line/Col of the opening parenthesis
	Col        int
}

func (ge *GroupedExpression) expressionNode() {}
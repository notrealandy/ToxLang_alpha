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
	Name       string // function name
	Params     []string
	ParamTypes []string
	Body       []Statement
	ReturnType string
	Visibility string
	Line       int
	Col        int
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

type IfStatement struct {
	IfCond     Expression    // condition for the if
	IfBody     []Statement   // body for the if
	ElifConds  []Expression  // conditions for each elif
	ElifBodies [][]Statement // bodies for each elif
	ElseBody   []Statement   // body for else, if present
	Line       int
	Col        int
}

type AssignmentStatement struct {
	Name  string
	Left  Expression
	Value Expression
	Line  int
	Col   int
}

type WhileStatement struct {
	Condition Expression
	Body      []Statement
	Line      int
	Col       int
}

type ForStatement struct {
	Init      Statement  // e.g. let i int >> 0 or i >> 0
	Condition Expression // e.g. i < 10
	Post      Statement  // e.g. i >> i + 1
	Body      []Statement
	Line      int
	Col       int
}

type PackageStatement struct {
    Name string
}

type ImportStatement struct {
    Path string
}

type ArrayLiteral struct {
	Elements []Expression
}

type IndexExpression struct {
	Left  Expression
	Index Expression
}

type Identifier struct {
	Value string
	Type  token.TokenType
	Line  int
	Col   int
}

type CallExpression struct {
	Function  Expression
	Arguments []Expression
}

type ExpressionStatement struct {
	Expr Expression
	Line int
	Col  int
}

type SliceExpression struct {
	Left  Expression
	Start Expression // can be nil
	End   Expression // can be nil
}

type UnaryExpression struct {
	Operator token.TokenType
	Right    Expression
	Line     int
	Col      int
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

func (lf *LogFunction) statementNode()         {}
func (rs *ReturnStatement) statementNode()     {}
func (es *ExpressionStatement) statementNode() {}
func (is *IfStatement) statementNode()         {}
func (as *AssignmentStatement) statementNode() {}
func (ws *WhileStatement) statementNode()      {}
func (fs *ForStatement) statementNode()        {}

func (id *Identifier) expressionNode()       {}
func (il *IntegerLiteral) expressionNode()   {}
func (sl *StringLiteral) expressionNode()    {}
func (bl *BoolLiteral) expressionNode()      {}
func (be *BinaryExpression) expressionNode() {}
func (nl *NilLiteral) expressionNode()       {}
func (ce *CallExpression) expressionNode()   {}
func (ue *UnaryExpression) expressionNode()  {}
func (al *ArrayLiteral) expressionNode()     {}
func (ie *IndexExpression) expressionNode()  {}
func (se *SliceExpression) expressionNode()  {}
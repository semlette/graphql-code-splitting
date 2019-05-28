package ast

import (
	"github.com/semlette/graphql-code-splitting/interpreter/token"
)

type Node interface {
	TokenLiteral() string
}

type Statement interface {
	Node
	statementNode()
}

type Expression interface {
	Node
	expressionNode()
}

type Query struct {
	Token        token.Token
	SelectionSet *SelectionSet
	Fragments    []*Fragment
}

func (q *Query) TokenLiteral() string {
	return q.Token.Literal
}

func (q *Query) statementNode() {}
func (q *Query) queryNode()     {}

type Fragment struct {
	Token        token.Token
	Name         *Identifier
	TargetObject *Identifier
	SelectionSet *SelectionSet
	Directives   []*Directive
}

type FragmentNode interface {
	fragmentNode()
}

func (Fragment) statementNode() {}
func (Fragment) fragmentNode()  {}

func (fmnt *Fragment) TokenLiteral() string {
	return fmnt.Token.Literal
}

type Directive struct {
	Token     token.Token
	Name      *Identifier
	Arguments []*Argument
}

type Argument struct {
	Token token.Token
	Name  *Identifier
	Value string // not going to support other value types for this expirement
}

type FieldNode interface {
	fieldNode()
}

type Field struct {
	Token        token.Token
	Name         *Identifier
	Directives   []*Directive
	SelectionSet *SelectionSet
	// Arguments map[string]interface{} // ha, don't need those
}

func (Field) fieldNode() {}
func (f *Field) TokenLiteral() string {
	return f.Token.Literal
}

type SelectionSetNode interface {
	selectionSetNode()
}

type SelectionSet struct {
	Token           token.Token
	Fields          []*Field
	FragmentSpreads []*FragmentSpread
}

func (SelectionSet) statementNode()    {}
func (SelectionSet) selectionSetNode() {}

func (ss *SelectionSet) TokenLiteral() string {
	return ss.Token.Literal
}

type FragmentSpreadNode interface {
	fragmentSpreadNode()
}

type FragmentSpread struct {
	Token        token.Token
	FragmentName *Identifier
	Directive    *Directive
}

func (FragmentSpread) fragmentSpreadNode() {}
func (fs *FragmentSpread) TokenLiteral() string {
	return fs.Token.Literal
}

type Document struct {
	Operation *Query
	Fragments []*Fragment
}

type Identifier struct {
	Token token.Token
	Value string
}

package ast

import (
	"github.com/semlette/graphql-code-splitting/interpreter/token"
)

// Node is an AST node
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
	Name      string
	Arguments map[string]string
}

type Field struct {
	Name       *Identifier
	Directives []*Directive
}

type SelectionSetNode interface {
	selectionSetNode()
}

type SelectionSet struct {
	Token         token.Token
	Fields        []*Field
	SelectionSets []*SelectionSet
}

func (SelectionSet) statementNode()    {}
func (SelectionSet) selectionSetNode() {}

func (ss *SelectionSet) TokenLiteral() string {
	return ss.Token.Literal
}

type Document struct {
	Query     *Query
	Fragments []*Fragment
}

type Identifier struct {
	Token token.Token
	Value string
}

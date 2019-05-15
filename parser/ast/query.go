package ast

import (
	"github.com/semlette/graphql-code-splitting/interpreter/token"
)

type Query struct {
	Token         token.Token
	SelectionSets []*SelectionSet
	Fragments     []*Fragment
}

func (q *Query) TokenLiteral() string {
	return q.Token.Literal
}

func (q *Query) statementNode() {}
func (q *Query) queryNode()     {}

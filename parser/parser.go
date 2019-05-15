package parser

import (
	"fmt"

	"github.com/semlette/graphql-code-splitting/interpreter/lexer"
	"github.com/semlette/graphql-code-splitting/interpreter/token"
	"github.com/semlette/graphql-code-splitting/parser/ast"
)

type Parser struct {
	l         *lexer.Lexer
	currToken token.Token
	peekToken token.Token
	err       error
}

func New(l *lexer.Lexer) *Parser {
	p := &Parser{l: l}
	p.nextToken()
	p.nextToken()
	return p
}

func (p *Parser) Error() error {
	return p.err
}

func (p *Parser) nextToken() {
	p.currToken = p.peekToken
	p.peekToken = p.l.NextToken()
}

func (p *Parser) Parse() *ast.Document {
	doc := new(ast.Document)
	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			q, ok := stmt.(*ast.Query)
			if ok {
				doc.Query = q
				p.nextToken()
				continue
			}
			fmnt, ok := stmt.(*ast.Fragment)
			if ok {
				q.Fragments = append(q.Fragments, fmnt)
				p.nextToken()
				continue
			}
		}
		p.nextToken()
	}
	return doc
}

func (p *Parser) parseStatement() ast.Statement {
	switch p.currToken.Type {
	case token.QUERY:
		return p.parseQuery()
	case token.FRAGMENT:
		return p.parseFragment()
	default:
		return nil
	}
}

func (p *Parser) parseQuery() *ast.Query {
	stmt := &ast.Query{
		Token:         p.currToken,
		SelectionSets: []*ast.SelectionSet{},
		Fragments:     []*ast.Fragment{},
	}
	// SKIP QUERY NAME
	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}
	ss := p.parseSelectionSet()
	if ss != nil {
		stmt.SelectionSets = append(stmt.SelectionSets, ss)
	}

	return stmt
}

func (p *Parser) parseFragment() *ast.Fragment {
	stmt := &ast.Fragment{Token: p.currToken}
	if !p.expectPeek(token.IDENT) {
		p.peekError(token.IDENT)
		return nil
	}
	stmt.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	if !p.expectPeek(token.ON) {
		p.peekError(token.ON)
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		p.peekError(token.IDENT)
		return nil
	}
	stmt.TargetObject = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}

	ss := p.parseSelectionSet()
	stmt.SelectionSet = ss

	return stmt
}

func (p *Parser) parseSelectionSet() *ast.SelectionSet {
	ss := new(ast.SelectionSet)
	ss.Fields = []*ast.Field{}
	ss.SelectionSets = []*ast.SelectionSet{}
	switch {
	case p.peekTokenIs(token.IDENT):
		p.nextToken()
		switch {
		case p.peekTokenIs(token.COMMA):
			ss.Fields = append(ss.Fields, &ast.Field{
				Name: &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal},
				// Directives TODO
			})
			p.nextToken()
		case p.peekTokenIs(token.LBRACE):
			p.nextToken()
			nextSS := p.parseSelectionSet()
			ss.SelectionSets = append(ss.SelectionSets, nextSS)
		default:
			p.peekError(token.COMMA)
			return nil
		}

	case p.peekTokenIs(token.SPREAD):
		// Doesn't do anything to the query TODO
		if !p.expectPeek(token.IDENT) {
			p.peekError(token.IDENT)
			return nil
		}
		if !p.expectPeek(token.COMMA) {
			p.peekError(token.COMMA)
			return nil
		}

	case p.peekTokenIs(token.RBRACE):
		return ss

	default:
		p.peekError(token.IDENT)
		return nil
	}
	return ss
}

func (p *Parser) currTokenIs(t token.Type) bool {
	return p.currToken.Type == t
}

func (p *Parser) peekTokenIs(t token.Type) bool {
	return p.peekToken.Type == t
}

func (p *Parser) expectPeek(t token.Type) bool {
	if p.peekTokenIs(t) {
		p.nextToken()
		return true
	}
	p.peekError(t)
	return false
}

func (p *Parser) peekError(t token.Type) {
	err := fmt.Errorf("expected token to be %s but got %s instead (%s)", t, p.peekToken.Type, p.peekToken.Literal)
	p.err = err
}

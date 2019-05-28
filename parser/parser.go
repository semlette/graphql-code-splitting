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
	fragments := []*ast.Fragment{}
	for p.currToken.Type != token.EOF {
		stmt := p.parseStatement()
		if stmt != nil {
			q, ok := stmt.(*ast.Query)
			if ok {
				doc.Operation = q
				p.nextToken()
				continue
			}
			fmnt, ok := stmt.(*ast.Fragment)
			if ok {
				fragments = append(fragments, fmnt)
				p.nextToken()
				continue
			}
		}
		p.nextToken()
	}
	if doc.Operation == nil {
		p.err = fmt.Errorf("no operation")
		return nil
	}
	doc.Operation.Fragments = fragments
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
		Token:        p.currToken,
		SelectionSet: &ast.SelectionSet{Token: p.currToken},
		Fragments:    []*ast.Fragment{},
	}

	if !p.expectPeek(token.LBRACE) {
		p.peekError(token.LBRACE)
		return nil
	}
	ss := p.parseSelectionSet()
	stmt.SelectionSet = ss

	return stmt
}

func (p *Parser) parseFields() []*ast.Field {
	var fields []*ast.Field
	for {
		field := p.parseField()
		if field == nil {
			break
		}
		fields = append(fields, field)
	}
	return fields
}

func (p *Parser) parseField() *ast.Field {
	field := &ast.Field{Token: p.currToken}
	if !p.expectPeek(token.IDENT) {
		p.peekErrorD(token.IDENT, "parseField()")
		return nil
	}
	field.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	// Skip parsing arguments

	// Parse the next token.
	// If it is a comma, the field selection is done.
	// If it is a '{', a new SelectionSet has been started
	switch {
	case p.peekTokenIs(token.COMMA):
		p.nextToken()
	case p.peekTokenIs(token.LBRACE):
		p.nextToken()
		ss := p.parseSelectionSet()
		field.SelectionSet = ss
	default:
		p.peekError(token.COMMA)
		return nil
	}
	return field
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
	ss := &ast.SelectionSet{Token: p.currToken}
	ss.Fields = []*ast.Field{}
	for {
		switch {
		case p.peekTokenIs(token.IDENT):
			f := p.parseField()
			if f == nil {
				return nil
			}
			ss.Fields = append(ss.Fields, f)
		case p.peekTokenIs(token.SPREAD):
			// parse fragment spread
			fs := p.parseFragmentSpread()
			if fs == nil {
				return nil
			}
			ss.FragmentSpreads = append(ss.FragmentSpreads, fs)
		case p.peekTokenIs(token.RBRACE):
			// return selection set
			p.nextToken()
			return ss
		default:
			p.peekErrorD(token.RBRACE, "parseSelectionSet()")
			return nil
		}
	}
}

func (p *Parser) parseFragmentSpread() *ast.FragmentSpread {
	fs := &ast.FragmentSpread{Token: p.currToken}
	if !p.expectPeek(token.SPREAD) {
		p.peekErrorD(token.SPREAD, "parseFragmentSpread()")
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		p.peekErrorD(token.IDENT, "parseFragmentSpread()")
		return nil
	}
	fs.FragmentName = &ast.Identifier{
		Token: p.currToken,
		Value: p.currToken.Literal,
	}
	if p.peekTokenIs(token.AT) {
		d := p.parseDirective()
		if d == nil {
			return nil
		}
		fs.Directive = d
	}
	if !p.expectPeek(token.COMMA) {
		p.peekErrorD(token.COMMA, "parseFragmentSpread()")
		return nil
	}
	return fs
}

func (p *Parser) parseDirective() *ast.Directive {
	d := &ast.Directive{Token: p.currToken}
	if !p.expectPeek(token.AT) {
		p.peekErrorD(token.AT, "parseDirective()")
		return nil
	}
	if !p.expectPeek(token.IDENT) {
		p.peekErrorD(token.IDENT, "parseDirective()")
		return nil
	}
	d.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	// Parse the arguments
	if p.peekTokenIs(token.LPAREN) {
		p.nextToken()
		arg := p.parseArguments()
		if arg == nil {
			return nil
		}
		d.Arguments = []*ast.Argument{arg}
	}
	return d
}

func (p *Parser) parseArguments() *ast.Argument {
	arg := &ast.Argument{Token: p.currToken}
	if !p.expectPeek(token.IDENT) {
		p.peekErrorD(token.IDENT, "parseArguments()")
		return nil
	}
	arg.Name = &ast.Identifier{Token: p.currToken, Value: p.currToken.Literal}
	if !p.expectPeek(token.COLON) {
		p.peekErrorD(token.COLON, "parseArguments()")
		return nil
	}
	if !p.expectPeek(token.STRING) {
		p.peekErrorD(token.STRING, "parseArguments()")
		return nil
	}
	arg.Value = p.currToken.Literal
	if !p.expectPeek(token.RPAREN) {
		p.peekErrorD(token.RPAREN, "parseArguments()")
		return nil
	}
	return arg
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

func (p *Parser) peekErrorD(t token.Type, details string) {
	err := fmt.Errorf("%s: expected token to be %s but got %s instead", details, t, p.peekToken.Type)
	p.err = err
}

package lexer

import (
	"github.com/semlette/graphql-code-splitting/interpreter/token"
)

// Lexer is a lexer. Nice
type Lexer struct {
	input string
	// position is the current character position
	position int
	// readPosition is the position of the next character
	readPosition int
	// ch is the current character under examination
	ch byte
}

// New creates a new Lexer with input
func New(input string) *Lexer {
	l := &Lexer{input: input}
	l.readChar()
	return l
}

func (l *Lexer) readChar() {
	if l.readPosition >= len(l.input) {
		l.ch = 0
	} else {
		l.ch = l.input[l.readPosition]
	}
	l.position = l.readPosition
	l.readPosition++
}

// NextToken reads the next token
func (l *Lexer) NextToken() token.Token {
	var tok token.Token
	l.skipWhitespace()

	switch l.ch {
	case '@':
		tok = token.NewToken(token.AT, l.ch)
	case '.':
		if l.peekChars(2) == ".." {
			l.readChar()
			l.readChar()
			literal := "..."
			tok = token.Token{Type: token.SPREAD, Literal: literal}
		} else {
			tok = token.NewToken(token.ILLEGAL, l.ch)
		}
	case '{':
		tok = token.NewToken(token.LBRACE, l.ch)
	case '}':
		tok = token.NewToken(token.RBRACE, l.ch)
	case '(':
		tok = token.NewToken(token.LPAREN, l.ch)
	case ')':
		tok = token.NewToken(token.RPAREN, l.ch)
	case ':':
		tok = token.NewToken(token.COLON, l.ch)
	case '"':
		tok = token.NewToken(token.QUOTE, l.ch)
	case ',':
		tok = token.NewToken(token.COMMA, l.ch)
	case 0:
		tok.Type = token.EOF
		tok.Literal = ""
	default:
		if isLetter(l.ch) {
			tok.Literal = l.readIdentifier()
			tok.Type = token.LookupIdentifier(tok.Literal)
			return tok
		}
		tok = token.NewToken(token.ILLEGAL, l.ch)
	}

	l.readChar()
	return tok
}

func (l *Lexer) peekChar() byte {
	if l.readPosition >= len(l.input) {
		return 0
	}
	return l.input[l.readPosition]
}

func (l *Lexer) peekChars(chars int) string {
	if l.readPosition >= len(l.input) {
		return ""
	}
	return l.input[l.readPosition : l.readPosition+chars]
}

func (l *Lexer) readIdentifier() string {
	pos := l.position
	for isLetter(l.ch) {
		l.readChar()
	}
	return l.input[pos:l.position]
}

func (l *Lexer) skipWhitespace() {
	for l.ch == ' ' || l.ch == '\t' || l.ch == '\n' || l.ch == '\r' {
		l.readChar()
	}
}

func isLetter(ch byte) bool {
	return 'a' <= ch && ch <= 'z' || 'A' <= ch && ch <= 'Z' || ch == '_'
}

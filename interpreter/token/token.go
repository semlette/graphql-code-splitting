package token

// Type is the type for tokens
type Type string

// Token is lexical token
// TODO
type Token struct {
	Type    Type
	Literal string
}

const (
	ILLEGAL Type = "ILLEGAL"
	EOF     Type = "EOF"
	IDENT   Type = "IDENT"
	AT      Type = "@"
	SPREAD  Type = "..."
	LBRACE  Type = "{"
	RBRACE  Type = "}"
	LPAREN  Type = "("
	RPAREN  Type = ")"
	COLON   Type = ":"
	COMMA   Type = ","

	STRING Type = "STRING"

	QUERY    Type = "QUERY"
	ON       Type = "ON"
	FRAGMENT Type = "FRAGMENT"
)

var keywords = map[string]Type{
	"query":    QUERY,
	"on":       ON,
	"fragment": FRAGMENT,
}

// LookupIdentifier looks up if the ident is a keyword or user-defined identifier
func LookupIdentifier(ident string) Type {
	if tok, ok := keywords[ident]; ok {
		return tok
	}
	return IDENT
}

// NewToken creates a new token from a type and a literal
func NewToken(typeof Type, literal byte) Token {
	return Token{
		Type:    typeof,
		Literal: string(literal),
	}
}

package syntax

type Symbol string

func (symbol *Symbol) String() string {
	if symbol == nil {
		return ""
	}
	return string(*symbol)
}

const (
	Bang      Symbol = "!"
	Colon     Symbol = ":"
	SemiColon Symbol = ";"
	LBrace    Symbol = "{"
	RBrace    Symbol = "}"
	LPara     Symbol = "("
	RPara     Symbol = ")"
	LBracket  Symbol = "["
	RBracket  Symbol = "]"
)

type Keyword string

func (keyword *Keyword) String() string {
	if keyword == nil {
		return ""
	}
	return string(*keyword)
}

const (
	// TypeScript
	Export    Keyword = "export"
	Interface Keyword = "interface"
	Enum      Keyword = "enum"

	// GraphQL
	Input      Keyword = "input"
	ObjectType Keyword = "type"
	Mutation   Keyword = "Mutation"
	Queries    Keyword = "Queries"
)

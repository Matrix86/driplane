package core

import (
	"io/ioutil"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

// A custom regexp lexer
var ruleLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(^[#].*$)` +
		`|(?P<Keyword>(?i)FEEDER|SET)` +
		`|(?P<Ident>[a-zA-Z][a-zA-Z_\d]*)` +
		`|(?P<String>"(?:[^"]*("")?)*")` +
		`|(?P<Float>\d+(?:\.\d+)?)` +
		`|(?P<Punct>[]["|,:;()=<>@"])` +
		`|(?P<Operators>!)`,
))

type AST struct {
	Rules []*RuleNode `@@*`
}

type RuleNode struct {
	Identifier string      `@Ident "="">"`
	Feeder     *FeederNode `( @@`
	First      *Node       `| @@ ) ";"`
}

type Node struct {
	//Action   *ActionNode `( @@ `
	Filter   *FilterNode `( @@`
	RuleCall *RuleCall   `| @@)`
}

type FeederNode struct {
	Name   string   `"<" @Ident`
	Params []*Param `(":" @@ ("," @@)*)? ">"`
	Next   *Node    `("|" @@)?`
}

type FilterNode struct {
	Name   string   `@Ident`
	Params []*Param `"(" ( @@ ("," @@)* )? ")"`
	Next   *Node    `("|" @@)?`
}

//type ActionNode struct {
//	Name   string   `@Ident`
//	Params []*Value `"(" ( @@ ("," @@)* )? ")"`
//	Next   *Node    `("|" @@)?`
//}

type RuleCall struct {
	Name   string `"@" @Ident`
	Next   *Node  `("|" @@)?`
}

type Param struct {
	Name  string `@Ident "="`
	Value *Value `@@`
}

type Value struct {
	String *string  `  @String`
	Number *float64 `| @Float`
}

type Parser struct {
	handle *participle.Parser
}

func NewParser() (*Parser, error) {
	var err error
	parser := &Parser{}
	parser.handle, err = participle.Build(&AST{},
		participle.Lexer(ruleLexer),
		participle.Unquote("String"),
		participle.CaseInsensitive("Keyword"),
	)
	if err != nil {
		return nil, err
	}

	return parser, nil
}

func (p *Parser) ParseFile(filename string) (*AST, error) {
	ast := &AST{}
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, err
	}
	err = p.handle.ParseBytes(content, ast)
	if err != nil {
		return nil, err
	}
	return ast, nil
}

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
		`|(?P<Ident>[a-zA-Z][a-zA-Z_\d]*)` +
		`|(?P<String>(?:(?:"(?:\\.|[^\"])*")|(?:'(?:\\.|[^'])*')))` +
		`|(?P<Float>\d+(?:\.\d+)?)` +
		`|(?P<Punct>[]["|,:;()=<>@"])` +
		`|(?P<Operators>!)`,
))

// AST defines a set of Rules
type AST struct {
	Rules []*RuleNode `@@*`
}

// RuleNode defines the first part of the Rule
type RuleNode struct {
	Identifier string      `@Ident "="">"`
	Feeder     *FeederNode `( @@`
	First      *Node       `| @@ ) ";"`
}

// Node identifies a Filter or a RuleCall
type Node struct {
	//Action   *ActionNode `( @@ `
	Filter   *FilterNode `( @@`
	RuleCall *RuleCall   `| @@)`
}

// FeederNode identifies the Feeder in the rule
type FeederNode struct {
	Name   string   `"<" @Ident`
	Params []*Param `(":" @@ ("," @@)*)? ">"`
	Next   *Node    `("|" @@)?`
}

// FilterNode identifies the Filter in the rule
type FilterNode struct {
	Neg    bool     `@("!")?`
	Name   string   `@Ident`
	Params []*Param `"(" ( @@ ("," @@)* )? ")"`
	Next   *Node    `("|" @@)?`
}

//type ActionNode struct {
//	Name   string   `@Ident`
//	Params []*Value `"(" ( @@ ("," @@)* )? ")"`
//	Next   *Node    `("|" @@)?`
//}

// RuleCall identifies the Call nodes in the rule
type RuleCall struct {
	Name string `"@" @Ident`
	Next *Node  `("|" @@)?`
}

// Param identifies the parameters accepted by nodes
type Param struct {
	Name  string `@Ident "="`
	Value *Value `@@`
}

// Value identifies a String or a Number
type Value struct {
	String *string  `  @String`
	Number *float64 `| @Float`
}

// Parser handles the parsing of the rules
type Parser struct {
	handle *participle.Parser
}

// NewParser creates a new Parser struct
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

// ParseFile returns the AST of the input file
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

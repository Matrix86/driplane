package core

import (
	"fmt"
	"github.com/evilsocket/islazy/log"
	"io/ioutil"
	"path/filepath"
	"regexp"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

var importRegexp = regexp.MustCompile(`(?m)^#import\s+"([^"]+)"\s*$`)

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
	Dependencies map[string]*AST
	Rules        []*RuleNode `@@*`
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
	handle     *participle.Parser
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

func (p *Parser) extractImports(content string, relativeTo string) []string {
	imports := make([]string, 0)
	matches := importRegexp.FindAllStringSubmatch(content, -1)
	if matches != nil {
		for _, m := range matches {
			f := filepath.Join(relativeTo, m[1])
			imports = append(imports, f)
		}
	}
	return imports
}

func (p *Parser) parseFile(filename string, deps []string) (*AST, error){
	for _, i := range deps {
		if i == filename {
			return nil, fmt.Errorf("cyclic dependency on %s", filename)
		}
	}

	ast := &AST{}
	log.Debug("parsing %s", filename)
	content, err := ioutil.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("parsing '%s': %s", filename, err)
	}

	// Parsing current file
	err = p.handle.ParseBytes(content, ast)
	if err != nil {
		return nil, err
	}

	deps = append(deps, filename)

	// init the map after ParseBytes to avoid overwriting
	ast.Dependencies = make(map[string]*AST)
	// preprocessing phase for imports
	imports := p.extractImports(string(content), filepath.Dir(filename))
	for _, f := range imports {
		if !filepath.IsAbs(f) {
			// path should be relative respect filename
			abs, err := filepath.Abs(f)
			if err != nil {
				log.Error("getting abs path for '%s': %s", filename, abs)
			}
			f = abs
		}

		// avoid to parse imported file twice in the same file
		if _, ok := ast.Dependencies[f]; ok {
			return nil, fmt.Errorf("file '%s' has been imported twice", f)
		}
		i, err := p.parseFile(f, deps)
		if err != nil {
			return nil, fmt.Errorf("can't parse import file '%s': %s", f, err)
		}
		ast.Dependencies[f] = i
	}

	return ast, nil
}

// ParseFile fills the map with all the ASTs parsed from the input file
func (p *Parser) ParseFile(filename string) (*AST, error) {
	deps := make([]string, 0)
	return p.parseFile(filename, deps)
}

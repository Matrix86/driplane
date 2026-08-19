package core

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"sync/atomic"

	"github.com/evilsocket/islazy/log"

	"github.com/alecthomas/participle"
	"github.com/alecthomas/participle/lexer"
)

var importRegexp = regexp.MustCompile(`(?m)^#import\s+"([^"]+)"\s*$`)

// maxChainLinks bounds the number of '|'-separated links (feeder|filter|...)
// that a rule chain may contain, counted outside of quoted strings. The
// participle grammar recurses once per link, so an extremely long chain
// (tens of thousands of links, still well under any reasonable body size
// limit) exhausts the goroutine stack with an unrecoverable Go runtime
// "stack overflow" fatal error that crashes the whole process, not just the
// request. Real rules chain a handful of nodes; 1000 is far beyond any
// legitimate use and far below the depth that triggers the crash.
const maxChainLinks = 1000

// bufferSeq generates unique synthetic filenames for ParseBytes so that a
// real on-disk file that happens to share the placeholder name can never be
// misreported as a cyclic self-import.
var bufferSeq atomic.Uint64

// A custom regexp lexer
var ruleLexer = lexer.Must(lexer.Regexp(
	`(?m)` +
		`(\s+)` +
		`|(^[#].*$)` +
		`|(?P<Ident>[a-zA-Z][a-zA-Z_\d-]*)` +
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
	handle *participle.Parser

	// ImportRoot, when non-empty, confines #import resolution to this
	// directory: any import resolving outside it, or to a non-regular
	// file, is rejected. It is empty for ParseFile (trusted, operator
	// authored files already on disk) and is set automatically by
	// ParseBytes to the directory it is given, since that content may come
	// from an untrusted HTTP request.
	ImportRoot string
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

func (p *Parser) parseFile(filename string, deps []string) (*AST, error) {
	for _, i := range deps {
		if i == filename {
			return nil, fmt.Errorf("cyclic dependency on %s", filename)
		}
	}

	log.Debug("parsing %s", filename)
	content, err := os.ReadFile(filename)
	if err != nil {
		return nil, fmt.Errorf("parsing '%s': %s", filename, err)
	}

	return p.parseContent(content, filename, deps)
}

// parseContent parses a rule source. filename is only used to resolve the
// #import directives and to report cyclic dependencies.
func (p *Parser) parseContent(content []byte, filename string, deps []string) (*AST, error) {
	if n := countLinksOutsideStrings(content); n > maxChainLinks {
		return nil, fmt.Errorf("rule chain too long (%d links, max %d)", n, maxChainLinks)
	}

	ast := &AST{}
	if err := p.handle.ParseBytes(content, ast); err != nil {
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

		if err := p.confineImport(f); err != nil {
			return nil, err
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

// confineImport rejects an import path that resolves outside p.ImportRoot,
// or that is not a regular file (an operator could otherwise point an
// import at a device such as /dev/zero and hang the parsing goroutine
// forever). It is a no-op when ImportRoot is empty, which preserves the
// existing trusted behaviour of ParseFile for on-disk rule files. The
// returned error never includes the target file's contents, only its name.
func (p *Parser) confineImport(path string) error {
	if p.ImportRoot == "" {
		return nil
	}

	rootAbs, err := filepath.Abs(p.ImportRoot)
	if err != nil {
		return fmt.Errorf("resolving the rules directory: %s", err)
	}
	rootReal, err := filepath.EvalSymlinks(rootAbs)
	if err != nil {
		return fmt.Errorf("resolving the rules directory: %s", err)
	}

	pathAbs, err := filepath.Abs(path)
	if err != nil {
		return fmt.Errorf("import '%s' could not be resolved", filepath.Base(path))
	}
	pathReal, err := filepath.EvalSymlinks(pathAbs)
	if err != nil {
		return fmt.Errorf("import '%s' could not be resolved", filepath.Base(path))
	}

	rel, err := filepath.Rel(rootReal, pathReal)
	if err != nil || rel == ".." || strings.HasPrefix(rel, ".."+string(filepath.Separator)) {
		return fmt.Errorf("import '%s' is outside the rules directory", filepath.Base(path))
	}

	info, err := os.Stat(pathReal)
	if err != nil {
		return fmt.Errorf("import '%s' could not be resolved", filepath.Base(path))
	}
	if !info.Mode().IsRegular() {
		return fmt.Errorf("import '%s' is not a regular file", filepath.Base(path))
	}

	return nil
}

// countLinksOutsideStrings counts the '|' bytes in content that ruleLexer
// would tokenize as chain-separator Punct tokens, by mirroring ruleLexer's
// own alternation order (whitespace, then comment, then Ident, then String,
// ...) at each position:
//
//  1. A '#' at the start of a line is a whole-line comment -- the lexer's
//     comment alternative is anchored "^[#].*$", so only a '#' in the very
//     first column of a line starts one; a '#' after leading whitespace, or
//     anywhere mid-line, is not a comment and must not be treated as one.
//     Everything up to (not including) the next newline is skipped without
//     inspecting it further: a quote inside a comment is not a string
//     delimiter, and a '|' inside a comment is not a chain link.
//  2. Otherwise, a double or single quote character opens a quoted string
//     that runs to its matching, non-escaped closing quote of the same
//     kind. Strings may legitimately span multiple lines -- the lexer's own
//     [^"] class does not exclude newlines -- so a line inside a string
//     that happens to start with '#' is string content, not a comment, and
//     is not reinspected by rule 1 until the string closes.
//  3. Otherwise, a '|' is a real chain link and is counted.
//
// This precisely avoids the failure mode of a simpler "any quote is a
// delimiter" scan: two unrelated quote characters sitting in two different
// comments would otherwise pair up and swallow everything between them,
// undercounting an arbitrarily long real chain down to zero.
//
// If a string is left unterminated (no matching closing quote before the
// end of content), the remainder is not silently swallowed as "probably not
// a string after all" -- every '|' from the opening quote to the end is
// counted instead, deliberately erring towards an over-count. Malformed
// input being over-counted only costs a slightly confusing "too long" error;
// under-counting it could let an oversized chain slip past this guard and
// crash the daemon, which is exactly the bug this function exists to
// prevent.
func countLinksOutsideStrings(content []byte) int {
	count := 0
	i := 0
	n := len(content)
	for i < n {
		atLineStart := i == 0 || content[i-1] == '\n'
		b := content[i]

		if atLineStart && b == '#' {
			for i < n && content[i] != '\n' {
				i++
			}
			continue
		}

		if b == '"' || b == '\'' {
			end := findStringEnd(content, i, b)
			if end < 0 {
				// Fail closed: count every remaining '|' rather than guess
				// where the "string" that never closes actually ends.
				for ; i < n; i++ {
					if content[i] == '|' {
						count++
					}
				}
				return count
			}
			i = end + 1
			continue
		}

		if b == '|' {
			count++
		}
		i++
	}
	return count
}

// findStringEnd returns the index of the closing quote matching the opening
// quote at content[start], honouring backslash escapes, or -1 if content
// has no matching closing quote before the end.
func findStringEnd(content []byte, start int, quote byte) int {
	for i := start + 1; i < len(content); i++ {
		switch content[i] {
		case '\\':
			i++
		case quote:
			return i
		}
	}
	return -1
}

// ParseFile fills the map with all the ASTs parsed from the input file
func (p *Parser) ParseFile(filename string) (*AST, error) {
	deps := make([]string, 0)
	return p.parseFile(filename, deps)
}

// ParseBytes parses a rule source that is not on disk yet. relativeTo is the
// directory used to resolve the #import directives; imports are confined to
// it, since this content may come from an untrusted request (see
// ImportRoot).
func (p *Parser) ParseBytes(content []byte, relativeTo string) (*AST, error) {
	p.ImportRoot = relativeTo
	sentinel := fmt.Sprintf("<buffer-%d>.rule", bufferSeq.Add(1))
	return p.parseContent(content, filepath.Join(relativeTo, sentinel), nil)
}

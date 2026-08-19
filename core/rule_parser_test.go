package core

import (
	"fmt"
	"os"
	"path"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewParser(t *testing.T) {
	parser, err := NewParser()
	if err != nil {
		t.Errorf("wrong error: expected=%#v had=%#v", nil, err)
	}
	if parser.handle == nil {
		t.Errorf("wrong handle: expected=%#v had=%#v", "!nil", parser)
	}
}

func TestParser_ParseFile(t *testing.T) {
	type Test struct {
		Name          string
		Filename1     string
		CreateFile1   bool
		FileContent1  string
		Filename2     string
		CreateFile2   bool
		FileContent2  string
		ExpectedAST   *AST
		ExpectedError string
	}

	v1, v2, v3 := "value1", "value2", "{'k':'v'}"
	notExistFile := path.Join(os.TempDir(), "notexist")
	cyclicFile1 := path.Join(os.TempDir(), "test1")
	cyclicFile2 := path.Join(os.TempDir(), "test2")

	tests := []Test{
		{"FileNotExist", notExistFile, false, "", "", false, "", nil, fmt.Sprintf("parsing '%s': open %s: no such file or directory", notExistFile, notExistFile)},
		{"EmptyFile", path.Join(os.TempDir(), "test"), true, "", "", false, "", &AST{Dependencies: map[string]*AST{}, Rules: []*RuleNode(nil)}, ""},
		{"UnexpectedEOF", path.Join(os.TempDir(), "test"), true, "ident =>", "", false, "", nil, "1:9: unexpected token \"<EOF>\" (expected <ident>)"},
		{"CyclicDep", cyclicFile1, true, "#import \"test1\"", cyclicFile2, true, "#import \"test2\"", nil, fmt.Sprintf("can't parse import file '%s': cyclic dependency on %s", cyclicFile1, cyclicFile1)},
		{
			"ParseOk",
			path.Join(os.TempDir(), "test"),
			true,
			"rule1 => <identifier: param1='value1', param2=\"value2\", param3=\"{'k':'v'}\">;\n" +
				"# comment ignored\n" +
				"rule2 => @rule1 | filter1(p1='value1',p2=\"value2\") | @anotherrule | ok();",
			"", false, "",
			&AST{
				Rules: []*RuleNode{
					&RuleNode{
						Identifier: "rule1",
						Feeder: &FeederNode{
							Name: "identifier",
							Params: []*Param{
								&Param{
									Name: "param1",
									Value: &Value{
										String: &v1,
										Number: (*float64)(nil),
									},
								},
								&Param{
									Name: "param2",
									Value: &Value{
										String: &v2,
										Number: (*float64)(nil),
									},
								},
								&Param{
									Name: "param3",
									Value: &Value{
										String: &v3,
										Number: (*float64)(nil),
									},
								},
							},
							Next: (*Node)(nil),
						},
						First: (*Node)(nil),
					},
					&RuleNode{
						Identifier: "rule2",
						Feeder:     (*FeederNode)(nil),
						First: &Node{
							Filter: (*FilterNode)(nil),
							RuleCall: &RuleCall{
								Name: "rule1",
								Next: &Node{
									Filter: &FilterNode{
										Name: "filter1",
										Params: []*Param{
											&Param{
												Name: "p1",
												Value: &Value{
													String: &v1,
													Number: (*float64)(nil),
												},
											},
											&Param{
												Name: "p2",
												Value: &Value{
													String: &v2,
													Number: (*float64)(nil),
												},
											},
										},
										Next: &Node{
											Filter: (*FilterNode)(nil),
											RuleCall: &RuleCall{
												Name: "anotherrule",
												Next: &Node{
													Filter: &FilterNode{
														Name:   "ok",
														Params: nil,
														Next:   (*Node)(nil),
													},
													RuleCall: (*RuleCall)(nil),
												},
											},
										},
									},
									RuleCall: (*RuleCall)(nil),
								},
							},
						},
					},
				},
				Dependencies: map[string]*AST{},
			},
			"",
		},
	}

	for _, v := range tests {
		if v.CreateFile1 {
			file, err := os.Create(v.Filename1)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename1)

			if _, err = file.Write([]byte(v.FileContent1)); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		if v.CreateFile2 {
			file, err := os.Create(v.Filename2)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename2)

			if _, err = file.Write([]byte(v.FileContent2)); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		parser, _ := NewParser()
		had, err := parser.ParseFile(v.Filename1)

		if v.ExpectedError == "" && err != nil {
			t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
		}

		if assert.Equal(t, v.ExpectedAST, had) == false {
			t.Errorf("%s: wrong AST: expected=%#v had=%#v", v.Name, v.ExpectedAST, had)
		}
	}
}

// TestParseContentBoundsChainLength proves the fix for the stack-overflow
// DoS: a rule chain far beyond any legitimate use is rejected with a plain
// error instead of crashing the process, while both a legitimate chain
// length and a legitimate quoted regexp alternation (which must not be
// mistaken for chain links) keep working.
func TestParseContentBoundsChainLength(t *testing.T) {
	underLimit := "R => a()" + strings.Repeat("|a()", maxChainLinks-1) + ";\n"
	overLimit := "R => a()" + strings.Repeat("|a()", 130000) + ";\n"
	quotedAlternation := "R => text(regexp=\"" + strings.Repeat("a|", 2000) + "a\");\n"

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser: %s", err)
	}
	if _, err := parser.ParseBytes([]byte(underLimit), os.TempDir()); err != nil {
		t.Errorf("a chain just under the limit should parse, got: %s", err)
	}

	parser2, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser: %s", err)
	}
	_, err = parser2.ParseBytes([]byte(overLimit), os.TempDir())
	if err == nil {
		t.Fatal("a chain far over the limit should be rejected with an error, not crash the process")
	}
	if !strings.Contains(err.Error(), "rule chain too long") {
		t.Errorf("unexpected error: %s", err)
	}

	parser3, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser: %s", err)
	}
	if _, err := parser3.ParseBytes([]byte(quotedAlternation), os.TempDir()); err != nil {
		t.Errorf("a regexp alternation inside a quoted parameter must not be mistaken for chain links, got: %s", err)
	}
}

// TestCountLinksOutsideStrings asserts the exact number of links
// countLinksOutsideStrings reports, not merely whether a rule parses or is
// rejected -- fix round 1 shipped a version of this scan that treated any
// quote character as a string delimiter, including one sitting inside a
// '#' comment. Two such quotes in two different comments would then pair up
// and swallow everything between them -- including every real '|' -- so a
// chain of 130,000 links was silently counted as 0 and the stack-overflow
// guard never fired. This table pins down the fix: a '#' only starts a
// comment at the very start of a line (mirroring ruleLexer's own
// "^[#].*$" anchoring), a comment is skipped whole without inspecting any
// quotes inside it, a real quoted string still consumes its content
// (including any '#' at the start of an internal line) without being
// treated as a comment, and an unterminated string fails closed by
// counting -- never silently discarding -- every '|' in the remainder.
func TestCountLinksOutsideStrings(t *testing.T) {
	long := 130000
	longChain := "a()" + strings.Repeat("|a()", long)

	tests := []struct {
		name    string
		content string
		want    int
	}{
		{
			name:    "plain chain, no comments or strings",
			content: "R => a()" + strings.Repeat("|a()", 9) + ";\n",
			want:    9,
		},
		{
			name:    "quote inside a comment before a long chain",
			content: "# \"\nR => " + longChain + ";\n",
			want:    long,
		},
		{
			name:    "quote inside a comment after a long chain",
			content: "R => " + longChain + ";\n# \"\n",
			want:    long,
		},
		{
			name:    "quote inside comments both before and after a long chain",
			content: "# \"\nR => " + longChain + ";\n# \"\n",
			want:    long,
		},
		{
			name:    "an innocent apostrophe in a comment before a long chain",
			content: "# don't do this\nR => " + longChain + ";\n",
			want:    long,
		},
		{
			name: "a comment quote after a legitimate quoted string",
			content: "R => text(regexp=\"a|b\")|echo();\n" +
				"# it's fine\n",
			want: 1, // only the real link between text() and echo()
		},
		{
			name: "a legitimate multi-line string containing a line starting with #",
			content: "R => text(regexp=\"line one\n" +
				"#not a comment, still inside the string\n" +
				"line three|line four\");\n",
			want: 0, // the '|' is string content, not a link
		},
		{
			name:    "an unterminated string followed by many links fails closed",
			content: "R => text(regexp=\"never closed" + strings.Repeat("|a()", 50) + ";\n",
			want:    50,
		},
		{
			name:    "a real regexp alternation inside a quoted parameter",
			content: "R => text(regexp=\"a|b|c\");\n",
			want:    0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := countLinksOutsideStrings([]byte(tt.content)); got != tt.want {
				t.Errorf("countLinksOutsideStrings() = %d, want %d", got, tt.want)
			}
		})
	}
}

// TestParseBytesConfinesImports proves the fix for the arbitrary file read:
// an import that resolves outside the given root is rejected, a sibling
// import inside the root still works, and a refused import's error never
// carries the target file's contents.
func TestParseBytesConfinesImports(t *testing.T) {
	root, err := os.MkdirTemp("", "rules-root")
	if err != nil {
		t.Fatalf("MkdirTemp: %s", err)
	}
	defer os.RemoveAll(root)

	secret := "top-secret-content-should-never-leak-into-an-error-message"
	outside := filepath.Join(os.TempDir(), fmt.Sprintf("outside-%d.rule", os.Getpid()))
	if err := os.WriteFile(outside, []byte(secret), 0644); err != nil {
		t.Fatalf("writing outside file: %s", err)
	}
	defer os.Remove(outside)

	sibling := filepath.Join(root, "sibling.rule")
	if err := os.WriteFile(sibling, []byte("S => echo();\n"), 0644); err != nil {
		t.Fatalf("writing sibling file: %s", err)
	}

	rel, err := filepath.Rel(root, outside)
	if err != nil {
		t.Fatalf("computing relative path: %s", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser: %s", err)
	}
	_, err = parser.ParseBytes([]byte(fmt.Sprintf("#import \"%s\"\nR => echo();\n", rel)), root)
	if err == nil {
		t.Fatal("an import escaping the rules root should be rejected")
	}
	if !strings.Contains(err.Error(), "outside") {
		t.Errorf("expected an 'outside the rules directory' error, got: %s", err)
	}
	if strings.Contains(err.Error(), secret) {
		t.Errorf("the error must not leak the refused file's contents: %s", err)
	}

	parser2, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser: %s", err)
	}
	ast, err := parser2.ParseBytes([]byte("#import \"sibling.rule\"\nR => echo();\n"), root)
	if err != nil {
		t.Fatalf("a sibling import inside the root should work, got: %s", err)
	}
	if len(ast.Dependencies) != 1 {
		t.Errorf("expected 1 dependency, got %d", len(ast.Dependencies))
	}
}

// TestParseBytesUsesUniqueSyntheticFilename proves that importing a real
// on-disk file that happens to share the old hardcoded placeholder name
// ("<buffer>.rule") is not mistaken for a cyclic self-import.
func TestParseBytesUsesUniqueSyntheticFilename(t *testing.T) {
	root, err := os.MkdirTemp("", "rules-root-collision")
	if err != nil {
		t.Fatalf("MkdirTemp: %s", err)
	}
	defer os.RemoveAll(root)

	collider := filepath.Join(root, "<buffer>.rule")
	if err := os.WriteFile(collider, []byte("C => echo();\n"), 0644); err != nil {
		t.Fatalf("writing collider file: %s", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("NewParser: %s", err)
	}
	if _, err := parser.ParseBytes([]byte("#import \"<buffer>.rule\"\nR => echo();\n"), root); err != nil {
		t.Errorf("importing a real file sharing the old placeholder name should not be mistaken for a self-import: %s", err)
	}
}

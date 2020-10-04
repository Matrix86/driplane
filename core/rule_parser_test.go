package core

import (
	"github.com/stretchr/testify/assert"
	"os"
	"path"
	"testing"
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
		Filename      string
		CreateFile    bool
		FileContent   string
		ExpectedAST   *AST
		ExpectedError string
	}

	v1, v2, v3 := "value1", "value2", "{'k':'v'}"

	tests := []Test{
		{"FileNotExist", path.Join(os.TempDir(), "notexist"), false, "", nil, "open /tmp/notexist: no such file or directory"},
		{"EmptyFile", path.Join(os.TempDir(), "test"), true, "", &AST{Rules: nil}, ""},
		{"UnexpectedEOF", path.Join(os.TempDir(), "test"), true, "ident =>", nil, "<source>:0:0: unexpected \"<EOF>\" (expected <ident> ...)"},
		{
			"ParseOk",
			path.Join(os.TempDir(), "test"),
			true,
			"rule1 => <identifier: param1='value1', param2=\"value2\", param3=\"{'k':'v'}\">;\n" +
				"# comment ignored\n" +
				"rule2 => @rule1 | filter1(p1='value1',p2=\"value2\") | @anotherrule | ok();",
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
			},
			"",
		},
	}

	for _, v := range tests {
		if v.CreateFile {
			file, err := os.Create(v.Filename)
			if err != nil {
				t.Errorf("%s: cannot create a temporary file", v.Name)
			}
			defer os.Remove(v.Filename)

			if _, err = file.Write([]byte(v.FileContent)); err != nil {
				t.Errorf("%s: can't write on file", v.Name)
			}
		}

		parser, _ := NewParser()
		had, err := parser.ParseFile(v.Filename)

		if v.ExpectedError == "" && err != nil {
			t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
		}

		if assert.Equal(t, had, v.ExpectedAST) == false {
			t.Errorf("%s: wrong AST: expected=%#v had=%#v", v.Name, v.ExpectedAST, had)
		}
	}
}

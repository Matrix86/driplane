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
		Filename1      string
		CreateFile1    bool
		FileContent1   string
		Filename2      string
		CreateFile2    bool
		FileContent2   string
		ExpectedAST   *AST
		ExpectedError string
	}

	v1, v2, v3 := "value1", "value2", "{'k':'v'}"

	tests := []Test{
		{"FileNotExist", path.Join(os.TempDir(), "notexist"), false, "", "", false, "",nil, "parsing '/tmp/notexist': open /tmp/notexist: no such file or directory"},
		{"EmptyFile", path.Join(os.TempDir(), "test"), true, "", "", false, "", &AST{Dependencies:map[string]*AST{}, Rules:[]*RuleNode(nil)}, ""},
		{"UnexpectedEOF", path.Join(os.TempDir(), "test"), true, "ident =>", "", false, "", nil, "<source>:0:0: unexpected \"<EOF>\" (expected <ident> ...)"},
		{"CyclicDep", path.Join(os.TempDir(), "test1"), true, "#import \"test1\"", path.Join(os.TempDir(), "test2"), true, "#import \"test2\"", nil, "can't parse import file '/tmp/test1': cyclic dependency on /tmp/test1"},
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
				Dependencies:map[string]*AST{},
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

package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestRuleSetInstanceSingleton(t *testing.T) {
	rs1 := RuleSetInstance()
	rs2 := RuleSetInstance()
	if rs1 != rs2 {
		t.Error("RuleSetInstance should return the same instance (singleton)")
	}
}

func TestRuleSetInstanceNotNil(t *testing.T) {
	rs := RuleSetInstance()
	if rs == nil {
		t.Fatal("RuleSetInstance should not return nil")
	}
	if rs.rules == nil {
		t.Error("rules map should be initialized")
	}
	if rs.compiledDeps == nil {
		t.Error("compiledDeps map should be initialized")
	}
	if rs.bus == nil {
		t.Error("bus should be initialized")
	}
}

func TestAddRuleNilNode(t *testing.T) {
	rs := RuleSetInstance()
	err := rs.AddRule("testfile", nil, &Configuration{flat: map[string]string{}}, nil)
	if err == nil {
		t.Error("AddRule with nil node should return error")
	}
}

func TestAddRuleEmptyIdentifier(t *testing.T) {
	rs := RuleSetInstance()
	node := &RuleNode{Identifier: ""}
	err := rs.AddRule("testfile", node, &Configuration{flat: map[string]string{}}, nil)
	if err == nil {
		t.Error("AddRule with empty identifier should return error")
	}
}

func TestAddRuleValid(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	node := &RuleNode{
		Identifier: "ruleset_test_rule_valid",
		First: &Node{
			Filter: &FilterNode{
				Name:   "echo",
				Params: []*Param{},
			},
		},
	}

	err := rs.AddRule("ruleset_test.rule", node, config, nil)
	if err != nil {
		t.Fatalf("AddRule returned error: %s", err)
	}

	// Verify the rule was added
	name := "ruleset_test.rule:ruleset_test_rule_valid"
	if _, ok := rs.rules[name]; !ok {
		t.Errorf("rule '%s' not found after AddRule", name)
	}
}

func TestAddRuleDuplicate(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	node := &RuleNode{
		Identifier: "ruleset_dup_rule",
		First: &Node{
			Filter: &FilterNode{
				Name:   "echo",
				Params: []*Param{},
			},
		},
	}

	err := rs.AddRule("ruleset_dup_test.rule", node, config, nil)
	if err != nil {
		t.Fatalf("first AddRule returned error: %s", err)
	}

	// Adding the same rule again should fail
	err = rs.AddRule("ruleset_dup_test.rule", node, config, nil)
	if err == nil {
		t.Error("AddRule with duplicate rule name should return error")
	}
}

func TestAddRuleWithFeeder(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}
	freqVal := "1s"

	node := &RuleNode{
		Identifier: "ruleset_feeder_rule",
		Feeder: &FeederNode{
			Name: "timer",
			Params: []*Param{
				{Name: "freq", Value: &Value{String: &freqVal}},
			},
		},
	}

	err := rs.AddRule("ruleset_feeder.rule", node, config, nil)
	if err != nil {
		t.Fatalf("AddRule with feeder returned error: %s", err)
	}

	name := "ruleset_feeder.rule:ruleset_feeder_rule"
	pr, ok := rs.rules[name]
	if !ok {
		t.Fatalf("rule '%s' not found", name)
	}
	if !pr.HasFeeder {
		t.Error("rule should have HasFeeder=true")
	}

	// Check it was added to feedRules
	found := false
	for _, fr := range rs.feedRules {
		if fr == name {
			found = true
			break
		}
	}
	if !found {
		t.Error("rule should be in feedRules list")
	}
}

func TestAddRuleInvalidFilter(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	node := &RuleNode{
		Identifier: "ruleset_bad_filter_rule",
		First: &Node{
			Filter: &FilterNode{
				Name:   "nonexistent_filter_xyz",
				Params: []*Param{},
			},
		},
	}

	err := rs.AddRule("ruleset_bad.rule", node, config, nil)
	if err == nil {
		t.Error("AddRule with nonexistent filter should return error")
	}
}

func TestCompileAstEmpty(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	ast := &AST{
		Dependencies: map[string]*AST{},
		Rules:        []*RuleNode{},
	}

	deps, err := rs.CompileAst("compile_empty.rule", ast, config)
	if err != nil {
		t.Fatalf("CompileAst returned error: %s", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected 0 deps, got %d", len(deps))
	}
}

func TestCompileAstWithRules(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	ast := &AST{
		Dependencies: map[string]*AST{},
		Rules: []*RuleNode{
			{
				Identifier: "compile_rule1",
				First: &Node{
					Filter: &FilterNode{
						Name:   "echo",
						Params: []*Param{},
					},
				},
			},
		},
	}

	deps, err := rs.CompileAst("compile_rules.rule", ast, config)
	if err != nil {
		t.Fatalf("CompileAst returned error: %s", err)
	}
	if len(deps) != 0 {
		t.Errorf("expected 0 deps (no dependencies), got %d", len(deps))
	}

	// The rule should now be registered
	name := "compile_rules.rule:compile_rule1"
	if _, ok := rs.rules[name]; !ok {
		t.Errorf("rule '%s' not found after CompileAst", name)
	}
}

func TestCompileAstAlreadyCompiled(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	ast := &AST{
		Dependencies: map[string]*AST{},
		Rules: []*RuleNode{
			{
				Identifier: "already_compiled_rule",
				First: &Node{
					Filter: &FilterNode{
						Name:   "echo",
						Params: []*Param{},
					},
				},
			},
		},
	}

	// First compilation
	_, err := rs.CompileAst("already_compiled.rule", ast, config)
	if err != nil {
		t.Fatalf("first CompileAst returned error: %s", err)
	}

	// Second compilation of the same file should return early
	deps, err := rs.CompileAst("already_compiled.rule", ast, config)
	if err != nil {
		t.Fatalf("second CompileAst returned error: %s", err)
	}
	// Should return the compiled deps (which were empty)
	_ = deps
}

func TestCompileAstWithDependencies(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	childAST := &AST{
		Dependencies: map[string]*AST{},
		Rules: []*RuleNode{
			{
				Identifier: "child_dep_rule",
				First: &Node{
					Filter: &FilterNode{
						Name:   "echo",
						Params: []*Param{},
					},
				},
			},
		},
	}

	parentAST := &AST{
		Dependencies: map[string]*AST{
			"dep_child.rule": childAST,
		},
		Rules: []*RuleNode{
			{
				Identifier: "parent_dep_rule",
				First: &Node{
					Filter: &FilterNode{
						Name:   "echo",
						Params: []*Param{},
					},
				},
			},
		},
	}

	deps, err := rs.CompileAst("dep_parent.rule", parentAST, config)
	if err != nil {
		t.Fatalf("CompileAst with dependencies returned error: %s", err)
	}

	// Should include the dependency file
	found := false
	for _, d := range deps {
		if d == "dep_child.rule" {
			found = true
			break
		}
	}
	if !found {
		t.Error("deps should include 'dep_child.rule'")
	}
}

func TestCompileAstWithInvalidRule(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	ast := &AST{
		Dependencies: map[string]*AST{},
		Rules: []*RuleNode{
			{
				Identifier: "compile_invalid_filter_rule",
				First: &Node{
					Filter: &FilterNode{
						Name:   "nonexistent_xyz_filter",
						Params: []*Param{},
					},
				},
			},
		},
	}

	_, err := rs.CompileAst("compile_invalid.rule", ast, config)
	if err == nil {
		t.Error("CompileAst with invalid filter should return error")
	}
}

func TestCompileAstWithInvalidDependency(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	childAST := &AST{
		Dependencies: map[string]*AST{},
		Rules: []*RuleNode{
			{
				Identifier: "bad_dep_child_rule",
				First: &Node{
					Filter: &FilterNode{
						Name:   "nonexistent_dep_filter",
						Params: []*Param{},
					},
				},
			},
		},
	}

	parentAST := &AST{
		Dependencies: map[string]*AST{
			"bad_dep_child.rule": childAST,
		},
		Rules: []*RuleNode{},
	}

	_, err := rs.CompileAst("bad_dep_parent.rule", parentAST, config)
	if err == nil {
		t.Error("CompileAst with invalid dependency should return error")
	}
}

func TestCompileAstFromParsedFile(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{flat: map[string]string{}}

	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "parsed.rule")
	content := "parsed_ast_rule => echo();"
	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rule file: %s", err)
	}

	parser, err := NewParser()
	if err != nil {
		t.Fatalf("failed to create parser: %s", err)
	}
	ast, err := parser.ParseFile(ruleFile)
	if err != nil {
		t.Fatalf("failed to parse rule file: %s", err)
	}

	deps, err := rs.CompileAst(ruleFile, ast, config)
	if err != nil {
		t.Fatalf("CompileAst from parsed file returned error: %s", err)
	}
	_ = deps
}

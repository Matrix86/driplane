package core

import (
	"testing"

	"github.com/Matrix86/driplane/filters"
)

func TestGetFirstNodeEmpty(t *testing.T) {
	p := &PipeRule{
		nodes: make([]INode, 0),
	}
	if p.getFirstNode() != nil {
		t.Error("getFirstNode on empty nodes should return nil")
	}
}

func TestGetLastNodeEmpty(t *testing.T) {
	p := &PipeRule{
		nodes: make([]INode, 0),
	}
	if p.getLastNode() != nil {
		t.Error("getLastNode on empty nodes should return nil")
	}
}

func TestGetFirstNode(t *testing.T) {
	first := "first_node"
	p := &PipeRule{
		nodes: []INode{first, "second", "third"},
	}
	if p.getFirstNode() != first {
		t.Errorf("getFirstNode should return the first node")
	}
}

func TestGetLastNode(t *testing.T) {
	last := "last_node"
	p := &PipeRule{
		nodes: []INode{"first", "second", last},
	}
	if p.getLastNode() != last {
		t.Errorf("getLastNode should return the last node")
	}
}

func TestGetFirstAndLastNodeSingleElement(t *testing.T) {
	only := "only_node"
	p := &PipeRule{
		nodes: []INode{only},
	}
	if p.getFirstNode() != only {
		t.Error("getFirstNode should return the only node")
	}
	if p.getLastNode() != only {
		t.Error("getLastNode should return the only node")
	}
}

func TestGetIdentifier(t *testing.T) {
	p := &PipeRule{
		Name: "myrule",
		file: "/path/to/rules.rule",
	}
	expected := "/path/to/rules.rule:myrule"
	if got := p.GetIdentifier(); got != expected {
		t.Errorf("expected identifier '%s', got '%s'", expected, got)
	}
}

func TestGetIdentifierEmptyFile(t *testing.T) {
	p := &PipeRule{
		Name: "test",
		file: "",
	}
	expected := ":test"
	if got := p.GetIdentifier(); got != expected {
		t.Errorf("expected identifier '%s', got '%s'", expected, got)
	}
}

func TestNewFilterCreation(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{},
	}
	p := &PipeRule{
		Name:   "test_rule",
		config: config,
		file:   "testfile",
	}

	fn := &FilterNode{
		Name:   "echo",
		Params: []*Param{},
	}

	f, err := p.newFilter(fn)
	if err != nil {
		t.Fatalf("newFilter returned error: %s", err)
	}
	if f == nil {
		t.Fatal("newFilter should return a non-nil filter")
	}
}

func TestNewFilterWithParams(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{},
	}
	p := &PipeRule{
		Name:   "test_rule_params",
		config: config,
		file:   "testfile",
	}

	extraVal := "true"
	fn := &FilterNode{
		Name: "echo",
		Params: []*Param{
			{Name: "extra", Value: &Value{String: &extraVal}},
		},
	}

	f, err := p.newFilter(fn)
	if err != nil {
		t.Fatalf("newFilter with params returned error: %s", err)
	}
	if f == nil {
		t.Fatal("newFilter should return a non-nil filter")
	}
}

func TestNewFilterWithNumberParam(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{},
	}
	p := &PipeRule{
		Name:   "test_rule_numpar",
		config: config,
		file:   "testfile",
	}

	num := 42.0
	fn := &FilterNode{
		Name: "echo",
		Params: []*Param{
			{Name: "count", Value: &Value{Number: &num}},
		},
	}

	f, err := p.newFilter(fn)
	if err != nil {
		t.Fatalf("newFilter with number param returned error: %s", err)
	}
	if f == nil {
		t.Fatal("newFilter should return a non-nil filter")
	}
}

func TestNewFilterWithConfigOverride(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{
			"echo.extra":     "true",
			"general.debug":  "true",
			"custom.setting": "val",
		},
	}
	p := &PipeRule{
		Name:   "test_rule_cfg",
		config: config,
		file:   "testfile",
	}

	fn := &FilterNode{
		Name:   "echo",
		Params: []*Param{},
	}

	f, err := p.newFilter(fn)
	if err != nil {
		t.Fatalf("newFilter with config returned error: %s", err)
	}
	if f == nil {
		t.Fatal("newFilter should return a non-nil filter")
	}
}

func TestNewFilterNegation(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{},
	}
	p := &PipeRule{
		Name:   "test_rule_neg",
		config: config,
		file:   "testfile",
	}

	fn := &FilterNode{
		Name:   "echo",
		Neg:    true,
		Params: []*Param{},
	}

	f, err := p.newFilter(fn)
	if err != nil {
		t.Fatalf("newFilter with negation returned error: %s", err)
	}
	if f == nil {
		t.Fatal("newFilter should return a non-nil filter")
	}
}

func TestNewFilterUnknownFilter(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{},
	}
	p := &PipeRule{
		Name:   "test_rule_unknown",
		config: config,
		file:   "testfile",
	}

	fn := &FilterNode{
		Name:   "this_filter_does_not_exist",
		Params: []*Param{},
	}

	_, err := p.newFilter(fn)
	if err == nil {
		t.Error("expected error for unknown filter type")
	}
}

func TestNewPipeRuleWithFeeder(t *testing.T) {
	freqVal := "1s"
	node := &RuleNode{
		Identifier: "timer_rule",
		Feeder: &FeederNode{
			Name: "timer",
			Params: []*Param{
				{Name: "freq", Value: &Value{String: &freqVal}},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	pr, err := NewPipeRule(node, config, "test.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule returned error: %s", err)
	}
	if pr.Name != "timer_rule" {
		t.Errorf("expected Name 'timer_rule', got '%s'", pr.Name)
	}
	if !pr.HasFeeder {
		t.Error("HasFeeder should be true")
	}
	if len(pr.nodes) == 0 {
		t.Error("should have at least one node (feeder)")
	}
}

func TestNewPipeRuleWithFeederAndFilter(t *testing.T) {
	freqVal := "1s"
	node := &RuleNode{
		Identifier: "timer_echo_rule",
		Feeder: &FeederNode{
			Name: "timer",
			Params: []*Param{
				{Name: "freq", Value: &Value{String: &freqVal}},
			},
			Next: &Node{
				Filter: &FilterNode{
					Name:   "echo",
					Params: []*Param{},
				},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	pr, err := NewPipeRule(node, config, "test2.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule returned error: %s", err)
	}
	if !pr.HasFeeder {
		t.Error("HasFeeder should be true")
	}
	if len(pr.nodes) < 2 {
		t.Errorf("expected at least 2 nodes (feeder + filter), got %d", len(pr.nodes))
	}
}

func TestNewPipeRuleWithFiltersOnly(t *testing.T) {
	node := &RuleNode{
		Identifier: "filter_only_rule",
		First: &Node{
			Filter: &FilterNode{
				Name:   "echo",
				Params: []*Param{},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	pr, err := NewPipeRule(node, config, "test3.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule returned error: %s", err)
	}
	if pr.HasFeeder {
		t.Error("HasFeeder should be false for filter-only rule")
	}
	if len(pr.nodes) != 1 {
		t.Errorf("expected 1 node, got %d", len(pr.nodes))
	}
}

func TestNewPipeRuleChainedFilters(t *testing.T) {
	node := &RuleNode{
		Identifier: "chained_filters_rule",
		First: &Node{
			Filter: &FilterNode{
				Name:   "echo",
				Params: []*Param{},
				Next: &Node{
					Filter: &FilterNode{
						Name:   "echo",
						Params: []*Param{},
					},
				},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	pr, err := NewPipeRule(node, config, "test4.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule returned error: %s", err)
	}
	if len(pr.nodes) != 2 {
		t.Errorf("expected 2 nodes, got %d", len(pr.nodes))
	}
}

func TestNewPipeRuleUnknownFeeder(t *testing.T) {
	node := &RuleNode{
		Identifier: "bad_feeder_rule",
		Feeder: &FeederNode{
			Name:   "nonexistent_feeder",
			Params: []*Param{},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	_, err := NewPipeRule(node, config, "test5.rule", nil)
	if err == nil {
		t.Error("expected error for unknown feeder type")
	}
}

func TestNewPipeRuleUnknownFilter(t *testing.T) {
	node := &RuleNode{
		Identifier: "bad_filter_rule",
		First: &Node{
			Filter: &FilterNode{
				Name:   "nonexistent_filter",
				Params: []*Param{},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	_, err := NewPipeRule(node, config, "test6.rule", nil)
	if err == nil {
		t.Error("expected error for unknown filter type")
	}
}

func TestNewPipeRuleFeederWithNumberParam(t *testing.T) {
	num := 5.0
	node := &RuleNode{
		Identifier: "timer_num_rule",
		Feeder: &FeederNode{
			Name: "timer",
			Params: []*Param{
				{Name: "numval", Value: &Value{Number: &num}},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{},
	}

	pr, err := NewPipeRule(node, config, "test7.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule returned error: %s", err)
	}
	if !pr.HasFeeder {
		t.Error("should have feeder")
	}
}

func TestNewPipeRuleFeederWithConfigParams(t *testing.T) {
	freqVal := "2s"
	node := &RuleNode{
		Identifier: "timer_cfg_rule",
		Feeder: &FeederNode{
			Name: "timer",
			Params: []*Param{
				{Name: "freq", Value: &Value{String: &freqVal}},
			},
		},
	}

	config := &Configuration{
		flat: map[string]string{
			"timer.freq":    "10s",
			"general.debug": "true",
			"custom.key":    "val",
		},
	}

	pr, err := NewPipeRule(node, config, "test8.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule returned error: %s", err)
	}
	if !pr.HasFeeder {
		t.Error("should have feeder")
	}
}

func TestAddNodeNil(t *testing.T) {
	p := &PipeRule{
		Name:  "test",
		nodes: make([]INode, 0),
	}
	err := p.addNode(nil, "")
	if err != nil {
		t.Errorf("addNode(nil) should return nil, got: %s", err)
	}
}

func TestGetRuleCallNotFound(t *testing.T) {
	p := &PipeRule{
		Name:         "test",
		file:         "testfile",
		dependencies: []string{},
	}
	rc := &RuleCall{Name: "nonexistent_rule"}
	_, err := p.getRuleCall(rc)
	if err == nil {
		t.Error("expected error for rule call not found")
	}
}

func TestAddNodeWithRuleCall(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{
		flat: map[string]string{},
	}

	// First register a rule that can be referenced via AddRule
	referencedNode := &RuleNode{
		Identifier: "referenced_rule",
		First: &Node{
			Filter: &FilterNode{
				Name:   "echo",
				Params: []*Param{},
			},
		},
	}
	err := rs.AddRule("rulecall_test.rule", referencedNode, config, nil)
	if err != nil {
		t.Fatalf("failed to add referenced rule: %s", err)
	}

	// Now create a rule that calls the referenced one (same file)
	callerNode := &RuleNode{
		Identifier: "caller_rule",
		First: &Node{
			RuleCall: &RuleCall{
				Name: "referenced_rule",
			},
		},
	}
	pr, err := NewPipeRule(callerNode, config, "rulecall_test.rule", nil)
	if err != nil {
		t.Fatalf("NewPipeRule with rule call returned error: %s", err)
	}
	if pr == nil {
		t.Fatal("should return a PipeRule")
	}
}

func TestAddNodeRuleCallWithFeederError(t *testing.T) {
	rs := RuleSetInstance()
	config := &Configuration{
		flat: map[string]string{},
	}

	// Create and register a rule with a feeder via AddRule
	freqVal := "1s"
	feederRule := &RuleNode{
		Identifier: "feeder_ref_rule",
		Feeder: &FeederNode{
			Name: "timer",
			Params: []*Param{
				{Name: "freq", Value: &Value{String: &freqVal}},
			},
		},
	}
	err := rs.AddRule("rulecall_feeder_test.rule", feederRule, config, nil)
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// Now try to add a rule call to a feeder rule from a context with prev set
	// This should fail: "rule 'feeder_ref_rule' contains a feeder and cannot be here"
	p := &PipeRule{
		Name:   "caller_with_prev",
		config: config,
		file:   "rulecall_feeder_test.rule",
		nodes:  make([]INode, 0),
	}

	node := &Node{
		RuleCall: &RuleCall{
			Name: "feeder_ref_rule",
		},
	}
	err = p.addNode(node, "some_prev_identifier")
	if err == nil {
		t.Error("expected error when rule call references a feeder rule with prev set")
	}
}

// Ensure filter package init runs (factories available)
var _ filters.Filter

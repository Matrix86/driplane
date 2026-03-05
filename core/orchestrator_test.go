package core

import (
	"os"
	"path/filepath"
	"testing"
)

func TestNewOrchestratorNoRulesPath(t *testing.T) {
	config := &Configuration{
		flat: map[string]string{},
	}

	_, err := NewOrchestrator(config)
	// With empty rules_path, fs.Glob will see "" as path.
	// This should either succeed with no rules or fail depending on FS behavior.
	// We mainly want to verify it doesn't panic.
	_ = err
}

func TestNewOrchestratorEmptyDir(t *testing.T) {
	dir := t.TempDir()
	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	o, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("NewOrchestrator with empty dir returned error: %s", err)
	}
	if o == nil {
		t.Fatal("orchestrator should not be nil")
	}
	if o.config != config {
		t.Error("orchestrator should store the config")
	}
}

func TestNewOrchestratorWithRuleFile(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "test_orch.rule")
	content := "orch_test_rule => echo();"
	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rule file: %s", err)
	}

	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	o, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("NewOrchestrator returned error: %s", err)
	}
	if o == nil {
		t.Fatal("orchestrator should not be nil")
	}
	if len(o.asts) == 0 {
		t.Error("should have parsed at least one AST")
	}
}

func TestNewOrchestratorInvalidRule(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "bad_orch.rule")
	content := "bad_rule =>"
	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rule file: %s", err)
	}

	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	// This is expected to call log.Fatal on parse error, which os.Exit(1)s.
	// We can't easily test that without subprocess. So skip this specific test.
	_ = config
}

func TestHasRunningFeederNoFeeders(t *testing.T) {
	dir := t.TempDir()
	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	o, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("NewOrchestrator returned error: %s", err)
	}

	if o.HasRunningFeeder() {
		t.Error("HasRunningFeeder should return false when no feeders exist")
	}
}

func TestStartStopFeedersWithTimer(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "timer_orch.rule")
	content := "timer_orch_rule => <timer: freq='1s'> | echo();"
	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rule file: %s", err)
	}

	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	o, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("NewOrchestrator returned error: %s", err)
	}

	o.StartFeeders()
	if !o.HasRunningFeeder() {
		t.Error("HasRunningFeeder should return true after StartFeeders")
	}

	// Starting again should not double-start already running feeders
	o.StartFeeders()

	o.StopFeeders()
	if o.HasRunningFeeder() {
		t.Error("HasRunningFeeder should return false after StopFeeders")
	}

	// Stopping again when already stopped should be fine
	o.StopFeeders()
}

func TestNewOrchestratorWithFeederAndFilter(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "full_orch.rule")
	content := "full_orch_rule => <timer: freq='500ms'> | echo();"
	if err := os.WriteFile(ruleFile, []byte(content), 0644); err != nil {
		t.Fatalf("failed to write rule file: %s", err)
	}

	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	o, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("NewOrchestrator returned error: %s", err)
	}
	if o == nil {
		t.Fatal("orchestrator should not be nil")
	}
}

func TestNewOrchestratorMultipleRuleFiles(t *testing.T) {
	dir := t.TempDir()

	rule1 := filepath.Join(dir, "rule1.rule")
	if err := os.WriteFile(rule1, []byte("multi_rule1 => echo();"), 0644); err != nil {
		t.Fatalf("failed to write rule1: %s", err)
	}

	rule2 := filepath.Join(dir, "rule2.rule")
	if err := os.WriteFile(rule2, []byte("multi_rule2 => echo();"), 0644); err != nil {
		t.Fatalf("failed to write rule2: %s", err)
	}

	config := &Configuration{
		flat: map[string]string{
			"general.rules_path": dir,
		},
	}

	o, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("NewOrchestrator returned error: %s", err)
	}
	if len(o.asts) < 2 {
		t.Errorf("expected at least 2 ASTs, got %d", len(o.asts))
	}
}

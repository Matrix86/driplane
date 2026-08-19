package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Matrix86/driplane/feeders"
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

// TestWaitFeedersReturnsForSelfStoppedFeeder proves the fix for the
// WaitGroup imbalance: StartFeeders always banks a waitFeeder.Add(1) when it
// starts a feeder, but several real feeders (feeders/folder.go,
// feeders/telegram.go, feeders/twitter.go) clear isRunning from their own
// goroutine instead of going through StopFeeders. Before the fix,
// StopFeeders only called Done() for feeders whose IsRunning() was still
// true at that moment, so a feeder that had already stopped itself left the
// counter permanently above zero and WaitFeeders() never returned.
//
// This test drives that scenario without depending on a specific feeder's
// internal failure path: it starts a real timer feeder through the
// orchestrator (banking the Add(1)), then calls Stop() on it directly,
// bypassing StopFeeders entirely -- exactly what a feeder clearing its own
// isRunning without the orchestrator's involvement looks like from the
// orchestrator's point of view. StopFeeders must still release the
// WaitGroup for it.
func TestWaitFeedersReturnsForSelfStoppedFeeder(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "selfstop_orch.rule")
	content := "selfstop_orch_rule => <timer: freq='1s'> | echo();"
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

	// Simulate a feeder that stops itself: Stop() is called directly on the
	// feeder, never through Orchestrator.StopFeeders, so the WaitGroup
	// accounting has not been settled when StopFeeders runs below.
	rs := o.ruleset
	for _, rulename := range rs.feedRules {
		f := rs.rules[rulename].getFirstNode().(feeders.Feeder)
		f.Stop()
	}

	o.StopFeeders()

	done := make(chan struct{})
	go func() {
		o.WaitFeeders()
		close(done)
	}()

	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("WaitFeeders() did not return: a feeder that clears isRunning on its own leaked the waitFeeder counter")
	}
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

func TestNewOrchestratorReturnsErrorOnBrokenRule(t *testing.T) {
	dir := t.TempDir()
	broken := filepath.Join(dir, "broken.rule")
	if err := os.WriteFile(broken, []byte("this is not a rule at all ||| ;"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	config := &Configuration{flat: map[string]string{"general.rules_path": dir}}

	// it must not terminate the process: it must return an error
	if _, err := NewOrchestrator(config); err == nil {
		t.Fatal("NewOrchestrator should return an error on a broken rule file")
	}
}

func TestOrchestratorReloadPicksUpChanges(t *testing.T) {
	dir := t.TempDir()
	rulePath := filepath.Join(dir, "r.rule")
	if err := os.WriteFile(rulePath, []byte("First => echo();\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	config := &Configuration{flat: map[string]string{"general.rules_path": dir}}

	o1, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("first NewOrchestrator: %s", err)
	}
	if len(o1.Rules()) != 1 || o1.Rules()[0].Name != "First" {
		t.Fatalf("expected rule 'First', got %+v", o1.Rules())
	}

	if err := os.WriteFile(rulePath, []byte("Second => echo();\n"), 0644); err != nil {
		t.Fatalf("rewriting rule: %s", err)
	}

	o2, err := NewOrchestrator(config)
	if err != nil {
		t.Fatalf("second NewOrchestrator: %s", err)
	}
	if len(o2.Rules()) != 1 || o2.Rules()[0].Name != "Second" {
		t.Fatalf("reload should expose rule 'Second', got %+v", o2.Rules())
	}
}

// TestStopFeedersConcurrentCallsDoNotDoubleStop is a regression test for a
// Critical found in review of the item-9 fix: releasing the Orchestrator's
// embedded lock while calling each feeder's Stop() (so a blocking Stop()
// cannot also stall Rules(), which /api/status calls every few seconds)
// reopened a window where two concurrent StopFeeders calls could both
// observe the same feeder as running and both call Stop() on it. Feeder
// Stop() implementations are not idempotent: most (including timer, used
// here) send on an unbuffered stopChan whose only reader returns after the
// first receive, so a second Stop() blocks forever; twitter's closes a
// channel a second time, which panics. This happens on the ordinary SIGTERM
// path (Supervisor.Stop() and the tail of Supervisor.Run()'s own loop can
// both call StopFeeders on the same Orchestrator) and from the web UI (a
// double-click on Reload, or two open tabs, produce two concurrent
// Reload() calls).
//
// This is a logical double-send, not a data race, so a plain `-race` run
// does not catch it -- the test instead drives two goroutines through
// StopFeeders concurrently, across many iterations, and fails on an explicit
// timeout naming the cause rather than hanging the suite if the mutual
// exclusion regresses.
func TestStopFeedersConcurrentCallsDoNotDoubleStop(t *testing.T) {
	dir := t.TempDir()
	ruleFile := filepath.Join(dir, "concurrent_stop.rule")
	content := "concurrent_stop_rule => <timer: freq='1s'> | echo();"
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

	const iterations = 100
	for i := 0; i < iterations; i++ {
		o.StartFeeders()

		done := make(chan struct{}, 2)
		for g := 0; g < 2; g++ {
			go func() {
				o.StopFeeders()
				done <- struct{}{}
			}()
		}

		for received := 0; received < 2; received++ {
			select {
			case <-done:
			case <-time.After(5 * time.Second):
				t.Fatalf("iteration %d: a second Stop() blocked forever -- two concurrent StopFeeders calls raced to stop the same feeder", i)
			}
		}
	}
}

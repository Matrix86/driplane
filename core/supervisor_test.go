package core

import (
	"os"
	"path/filepath"
	"testing"
	"time"
)

// waitState polls the supervisor until it reaches the wanted state or times out.
func waitState(t *testing.T, s *Supervisor, want string, timeout time.Duration) Status {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last Status
	for time.Now().Before(deadline) {
		last = s.Status()
		if last.State == want {
			return last
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for state %q, last status: %+v", want, last)
	return last
}

func TestSupervisorStatusBeforeRun(t *testing.T) {
	s := NewSupervisor(&Configuration{flat: map[string]string{"general.rules_path": t.TempDir()}})
	if got := s.Status().State; got != "stopped" {
		t.Errorf("expected state 'stopped', got %q", got)
	}
}

func TestSupervisorRunsAndReloads(t *testing.T) {
	dir := t.TempDir()
	rulePath := filepath.Join(dir, "r.rule")
	if err := os.WriteFile(rulePath, []byte("First => echo();\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	s := NewSupervisor(&Configuration{flat: map[string]string{"general.rules_path": dir}})
	done := make(chan error, 1)
	go func() { done <- s.Run() }()

	st := waitState(t, s, "running", 5*time.Second)
	if len(st.Rules) != 1 || st.Rules[0].Name != "First" {
		t.Fatalf("expected rule 'First', got %+v", st.Rules)
	}

	// a broken rule must not kill the supervisor
	if err := os.WriteFile(rulePath, []byte("this is broken ||| ;"), 0644); err != nil {
		t.Fatalf("rewriting rule: %s", err)
	}
	s.Reload()
	st = waitState(t, s, "error", 5*time.Second)
	if st.Error == "" {
		t.Error("expected a non-empty error in status")
	}

	// and a valid rule should load again
	if err := os.WriteFile(rulePath, []byte("Second => echo();\n"), 0644); err != nil {
		t.Fatalf("rewriting rule: %s", err)
	}
	s.Reload()
	st = waitState(t, s, "running", 5*time.Second)
	if len(st.Rules) != 1 || st.Rules[0].Name != "Second" {
		t.Fatalf("expected rule 'Second' after reload, got %+v", st.Rules)
	}

	s.Stop()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %s", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// waitRuleName polls the supervisor until it reports state "running" with
// exactly one rule named want, or times out.
func waitRuleName(t *testing.T, s *Supervisor, want string, timeout time.Duration) Status {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last Status
	for time.Now().Before(deadline) {
		last = s.Status()
		if last.State == "running" && len(last.Rules) == 1 && last.Rules[0].Name == want {
			return last
		}
		time.Sleep(10 * time.Millisecond)
	}
	t.Fatalf("timeout waiting for rule %q, last status: %+v", want, last)
	return last
}

// TestSupervisorReloadWithRunningFeeder covers C1: a Reload() issued while a
// real feeder is running must not be lost, and the rebuilt orchestrator must
// actually reflect the edited rule file.
func TestSupervisorReloadWithRunningFeeder(t *testing.T) {
	dir := t.TempDir()
	rulePath := filepath.Join(dir, "ticker.rule")
	if err := os.WriteFile(rulePath, []byte(`Ticker => <timer: freq="100ms"> | echo();`+"\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	s := NewSupervisor(&Configuration{flat: map[string]string{"general.rules_path": dir}})
	done := make(chan error, 1)
	go func() { done <- s.Run() }()

	waitRuleName(t, s, "Ticker", 5*time.Second)

	if err := os.WriteFile(rulePath, []byte(`TickerV2 => <timer: freq="100ms"> | echo();`+"\n"), 0644); err != nil {
		t.Fatalf("rewriting rule: %s", err)
	}
	s.Reload()
	waitRuleName(t, s, "TickerV2", 5*time.Second)

	s.Stop()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %s", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// TestSupervisorStopWithRunningFeeder covers C2: Stop() must make Run()
// return even while a genuine feeder goroutine is running, not just when the
// rule set has no feeder.
func TestSupervisorStopWithRunningFeeder(t *testing.T) {
	dir := t.TempDir()
	rulePath := filepath.Join(dir, "ticker.rule")
	if err := os.WriteFile(rulePath, []byte(`Ticker => <timer: freq="100ms"> | echo();`+"\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	s := NewSupervisor(&Configuration{flat: map[string]string{"general.rules_path": dir}})
	done := make(chan error, 1)
	go func() { done <- s.Run() }()

	waitRuleName(t, s, "Ticker", 5*time.Second)

	s.Stop()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %s", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop while a feeder was running")
	}
}

func TestSupervisorPauseAndResume(t *testing.T) {
	dir := t.TempDir()
	if err := os.WriteFile(filepath.Join(dir, "r.rule"), []byte("First => echo();\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	s := NewSupervisor(&Configuration{flat: map[string]string{"general.rules_path": dir}})
	done := make(chan error, 1)
	go func() { done <- s.Run() }()

	waitState(t, s, "running", 5*time.Second)

	s.PauseFeeders()
	waitState(t, s, "stopped", 5*time.Second)

	s.ResumeFeeders()
	waitState(t, s, "running", 5*time.Second)

	s.Stop()
	select {
	case <-done:
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop")
	}
}

// TestSupervisorStopWhilePaused covers a regression shape an earlier review
// round found in this file: Stop() returning cleanly to its caller while
// Run() never actually unblocks, because it is parked somewhere Stop()'s
// signal does not reach. PauseFeeders leaves Run() parked in waitSignal()
// from the isPaused branch with a real feeder's goroutine having already
// exited (StopFeeders was called), which is a different resting point than
// the other Stop tests exercise (no feeder at all, or a feeder still
// running via waitFeedersOrSignal). Stop() is called here with no
// intervening ResumeFeeders, so this is the only test that proves Stop can
// reach Run() while paused.
func TestSupervisorStopWhilePaused(t *testing.T) {
	dir := t.TempDir()
	rulePath := filepath.Join(dir, "ticker.rule")
	if err := os.WriteFile(rulePath, []byte(`Ticker => <timer: freq="100ms"> | echo();`+"\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	s := NewSupervisor(&Configuration{flat: map[string]string{"general.rules_path": dir}})
	done := make(chan error, 1)
	go func() { done <- s.Run() }()

	waitState(t, s, "running", 5*time.Second)

	s.PauseFeeders()
	waitState(t, s, "stopped", 5*time.Second)

	s.Stop()
	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Run returned an error: %s", err)
		}
	case <-time.After(5 * time.Second):
		t.Fatal("Run did not return after Stop while paused: it is parked somewhere Stop's signal does not reach")
	}
}

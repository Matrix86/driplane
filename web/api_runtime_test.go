package web

import (
	"bufio"
	"context"
	"encoding/json"
	"net"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/Matrix86/driplane/core"
)

func TestStatusReportsStoppedWithoutSupervisor(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodGet, "/api/status", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		State string `json:"state"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.State != "stopped" {
		t.Errorf("expected state 'stopped', got %q", resp.State)
	}
}

func TestReloadWithoutSupervisorReportsError(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/reload", map[string]string{})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("reload without a supervisor should report ok=false")
	}
}

// TestRuntimeActionsAcceptEmptyBody makes sure POST /api/reload, /api/start
// and /api/stop all succeed with no body at all (not even "{}"): decodeBody's
// errEmptyBody sentinel exists precisely so handlers that legitimately accept
// an empty body (these three) do not 400 on it. The runtime handlers here
// never call decodeBody at all, which is the simplest way to satisfy that.
func TestRuntimeActionsAcceptEmptyBody(t *testing.T) {
	s := newTestServer(t)

	for _, path := range []string{"/api/reload", "/api/start", "/api/stop"} {
		rec := do(t, s, http.MethodPost, path, nil)
		if rec.Code != http.StatusOK {
			t.Fatalf("%s with empty body: expected 200, got %d (%s)", path, rec.Code, rec.Body.String())
		}
		var resp actionResponse
		if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
			t.Fatalf("%s: decoding: %s", path, err)
		}
		if resp.OK {
			t.Errorf("%s: expected ok=false (no supervisor attached), got ok=true", path)
		}
	}
}

// TestRuntimeEndpointsWireToRealSupervisor proves handleStart and handleStop
// are actually wired to the right Supervisor methods. Every other test in
// this file builds a Server with NO supervisor attached at all, so swapping
// the bodies of handleStart and handleStop (calling ResumeFeeders from
// handleStop and PauseFeeders from handleStart) would not fail a single
// existing test -- both handlers would still just report ok=false for the
// same "no supervisor attached" reason. This test drives a real
// *core.Supervisor, running a timer feeder over a temp rules directory,
// through the HTTP handlers themselves: POST /api/stop must actually stop
// the feeder and POST /api/start must actually bring it back, observed via
// GET /api/status, which also gives the supervisor-attached branch of
// handleStatus its first coverage.
func TestRuntimeEndpointsWireToRealSupervisor(t *testing.T) {
	dir := t.TempDir()
	rulePath := filepath.Join(dir, "r.rule")
	if err := os.WriteFile(rulePath, []byte("Ticker => <timer: freq='50ms'> | echo();\n"), 0644); err != nil {
		t.Fatalf("writing rule: %s", err)
	}

	cfg := &core.Configuration{}
	cfg.SetAll(map[string]string{"general.rules_path": dir})
	sup := core.NewSupervisor(cfg)

	runDone := make(chan error, 1)
	go func() { runDone <- sup.Run() }()
	defer func() {
		sup.Stop()
		select {
		case <-runDone:
		case <-time.After(5 * time.Second):
			t.Fatal("Supervisor.Run did not return after Stop")
		}
	}()

	store, _ := newTestStore(t)
	s, err := New(Options{
		Address:    "127.0.0.1:0",
		Token:      "secret-token",
		Store:      store,
		Ring:       NewLogRing(10),
		Supervisor: sup,
	})
	if err != nil {
		t.Fatalf("New: %s", err)
	}

	if !waitRuleRunning(t, s, true, 5*time.Second) {
		t.Fatal("expected the feeder to be running before the first /api/stop")
	}

	rec := do(t, s, http.MethodPost, "/api/stop", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/stop: expected 200, got %d", rec.Code)
	}
	var resp actionResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding /api/stop response: %s", err)
	}
	if !resp.OK {
		t.Fatalf("/api/stop: expected ok=true (a real supervisor is attached), got %+v", resp)
	}
	if waitRuleRunning(t, s, false, 5*time.Second) {
		t.Fatal("/api/stop did not stop the feeder")
	}

	rec = do(t, s, http.MethodPost, "/api/start", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("/api/start: expected 200, got %d", rec.Code)
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding /api/start response: %s", err)
	}
	if !resp.OK {
		t.Fatalf("/api/start: expected ok=true, got %+v", resp)
	}
	if !waitRuleRunning(t, s, true, 5*time.Second) {
		t.Fatal("/api/start did not bring the feeder back")
	}
}

// waitRuleRunning polls GET /api/status through the real HTTP handler until
// the single rule's Running flag matches want, or times out and returns the
// last observed value.
func waitRuleRunning(t *testing.T, s *Server, want bool, timeout time.Duration) bool {
	t.Helper()
	deadline := time.Now().Add(timeout)
	var last bool
	for time.Now().Before(deadline) {
		rec := do(t, s, http.MethodGet, "/api/status", nil)
		var st struct {
			State string `json:"state"`
			Rules []struct {
				Running bool `json:"running"`
			} `json:"rules"`
		}
		if err := json.Unmarshal(rec.Body.Bytes(), &st); err != nil {
			t.Fatalf("decoding /api/status: %s", err)
		}
		if len(st.Rules) == 1 {
			last = st.Rules[0].Running
			if last == want {
				return last
			}
		}
		time.Sleep(20 * time.Millisecond)
	}
	return last
}

func TestLogsStreamSendsBacklogAndNewLines(t *testing.T) {
	s := newTestServer(t)
	s.opts.Ring.Append(LogLine{Level: "info", Message: "backlog line"})

	ts := httptest.NewServer(s.Handler())
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/logs?token=secret-token", nil)
	if err != nil {
		t.Fatalf("building request: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/logs: %s", err)
	}
	defer resp.Body.Close()

	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected an SSE content type, got %q", ct)
	}

	reader := bufio.NewReader(resp.Body)
	line, err := reader.ReadString('\n')
	if err != nil {
		t.Fatalf("reading the backlog line: %s", err)
	}
	if !strings.Contains(line, "backlog line") {
		t.Errorf("expected the backlog line first, got %q", line)
	}

	go func() {
		time.Sleep(100 * time.Millisecond)
		s.opts.Ring.Append(LogLine{Level: "error", Message: "fresh line"})
	}()

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		l, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("reading the stream: %s", err)
		}
		if strings.Contains(l, "fresh line") {
			return
		}
	}
	t.Fatal("the new line never arrived on the stream")
}

// TestLogsStreamSurvivesReadTimeout proves that http.Server's ReadTimeout —
// added to s.srv in New() to stop a client trickling a request body forever —
// does not cut a long-lived SSE response. ReadTimeout only bounds how long
// the server waits to finish reading the incoming request; it says nothing
// about how long the handler may then take writing the response, which is
// exactly the SSE case (WriteTimeout, which would bound both, is left at 0).
//
// httptest.NewServer(s.Handler()) does not exercise this at all: it builds
// its own *http.Server around the handler with no timeouts configured, so it
// cannot tell us anything about the timeouts New() sets on s.srv. This test
// instead builds its own httptest server with a deliberately tiny ReadTimeout
// (well under what a real deployment would use), keeps the SSE connection
// open past that deadline, and only then appends a fresh line. If ReadTimeout
// were killing the response, the connection would already be closed by the
// time the line is appended and the read below would fail or return early;
// observing the line arrive proves it is not.
// TestShutdownClosesOpenSSEStreamPromptly proves the fix for a shutdown
// stall: http.Server.Shutdown waits for every connection to go idle, and an
// SSE response (handleLogs) never does on its own, so with a browser tab
// left open on the log view every graceful shutdown used to cost the full
// shutdown timeout. Server.Shutdown now closes s.closing first, which
// handleLogs selects on, so an open stream ends immediately and Shutdown
// returns quickly instead of stalling.
//
// This must exercise s.srv itself (the real *http.Server New() built and
// configured, with its real timeouts), not a second http.Server wrapped
// around s.Handler() by httptest.NewServer -- shutting down an unrelated,
// never-started server would trivially return fast regardless of whether the
// fix works, proving nothing.
func TestShutdownClosesOpenSSEStreamPromptly(t *testing.T) {
	s := newTestServer(t)

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %s", err)
	}
	go s.srv.Serve(ln)

	req, err := http.NewRequest(http.MethodGet, "http://"+ln.Addr().String()+"/api/logs?token=secret-token", nil)
	if err != nil {
		t.Fatalf("building request: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/logs: %s", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.HasPrefix(ct, "text/event-stream") {
		t.Fatalf("expected an SSE content type, got %q", ct)
	}

	done := make(chan error, 1)
	go func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		done <- s.Shutdown(ctx)
	}()

	select {
	case err := <-done:
		if err != nil {
			t.Fatalf("Shutdown returned an error: %s", err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("Shutdown did not return promptly with an SSE stream open: it is waiting on the stream to go idle")
	}
}

func TestLogsStreamSurvivesReadTimeout(t *testing.T) {
	s := newTestServer(t)

	const readTimeout = 150 * time.Millisecond

	ts := httptest.NewUnstartedServer(s.Handler())
	ts.Config.ReadTimeout = readTimeout
	ts.Config.WriteTimeout = 0
	ts.Start()
	defer ts.Close()

	req, err := http.NewRequest(http.MethodGet, ts.URL+"/api/logs?token=secret-token", nil)
	if err != nil {
		t.Fatalf("building request: %s", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("GET /api/logs: %s", err)
	}
	defer resp.Body.Close()

	reader := bufio.NewReader(resp.Body)

	// Let the connection sit idle well past ReadTimeout before anything new
	// is appended. If ReadTimeout applied to the response, the server would
	// have already torn the connection down by now.
	time.Sleep(readTimeout * 4)

	s.opts.Ring.Append(LogLine{Level: "info", Message: "after read timeout"})

	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		l, err := reader.ReadString('\n')
		if err != nil {
			t.Fatalf("reading the stream after ReadTimeout elapsed (%s since connect): %s", readTimeout*4, err)
		}
		if strings.Contains(l, "after read timeout") {
			return
		}
	}
	t.Fatal("a line appended well after ReadTimeout elapsed never arrived: ReadTimeout may be cutting the SSE stream")
}

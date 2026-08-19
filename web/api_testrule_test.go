package web

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"sync"
	"testing"
)

// fakeDriplane writes a script behaving like the driplane binary: it succeeds
// unless the rules directory contains a file with the word "broken" in it.
func fakeDriplane(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake binary is a shell script")
	}

	path := filepath.Join(t.TempDir(), "fake-driplane")
	script := `#!/bin/sh
rules=""
while [ $# -gt 0 ]; do
  case "$1" in
    -rules) rules="$2"; shift 2 ;;
    *) shift ;;
  esac
done
if grep -rq broken "$rules" 2>/dev/null; then
  echo "rule parsing: file 'x.rule': unexpected token" >&2
  exit 1
fi
echo "rules syntax OK"
exit 0
`
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("writing the fake binary: %s", err)
	}
	return path
}

func TestTestRuleSuccess(t *testing.T) {
	s := newTestServer(t)
	s.opts.Executable = fakeDriplane(t)

	rec := do(t, s, http.MethodPost, "/api/test", map[string]string{
		"name":   "a.rule",
		"source": "Good => echo();\n",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK     bool   `json:"ok"`
		Output string `json:"output"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got %s", rec.Body.String())
	}
}

func TestTestRuleFailureReportsOutput(t *testing.T) {
	s := newTestServer(t)
	s.opts.Executable = fakeDriplane(t)

	rec := do(t, s, http.MethodPost, "/api/test", map[string]string{
		"name":   "a.rule",
		"source": "broken rule content",
	})

	var resp struct {
		OK     bool   `json:"ok"`
		Output string `json:"output"`
		Error  string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("a broken rule should report ok=false")
	}
	if resp.Output == "" {
		t.Error("the subprocess output should be reported")
	}
}

func TestTestRuleDoesNotTouchTheRealRulesDir(t *testing.T) {
	s := newTestServer(t)
	s.opts.Executable = fakeDriplane(t)

	if rec := do(t, s, http.MethodPost, "/api/test", map[string]string{
		"name":   "ghost.rule",
		"source": "Ghost => echo();\n",
	}); rec.Code != http.StatusOK {
		t.Fatalf("test: got %d", rec.Code)
	}

	files, err := s.opts.Store.List(KindRules)
	if err != nil {
		t.Fatalf("List: %s", err)
	}
	if len(files) != 0 {
		t.Errorf("the rules directory must stay untouched, found %+v", files)
	}
}

func TestTestRuleRejectsInvalidName(t *testing.T) {
	s := newTestServer(t)
	s.opts.Executable = fakeDriplane(t)

	rec := do(t, s, http.MethodPost, "/api/test", map[string]string{
		"name":   "../evil.rule",
		"source": "x",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

// fakeDriplaneDumpArgs writes a script that echoes every argument it
// receives on its own line, prefixed with "ARG:" so a value that happens to
// be empty or to contain spaces is still unambiguous to parse. Used to pin
// down exactly what handleTestRule puts on the subprocess command line.
func fakeDriplaneDumpArgs(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake binary is a shell script")
	}

	path := filepath.Join(t.TempDir(), "fake-driplane-dumpargs")
	script := `#!/bin/sh
for a in "$@"; do
  echo "ARG:$a"
done
exit 0
`
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("writing the fake binary: %s", err)
	}
	return path
}

// TestTestRuleNameNeverBecomesAFlagArgument pins down, as a regression test,
// what the code review verified by inspection: a rule name shaped like a
// flag (e.g. "-config.rule") must never reach the subprocess as a standalone
// argv token. It only ever selects a filename inside the throwaway temp
// directory; the temp directory path itself is what follows "-rules".
func TestTestRuleNameNeverBecomesAFlagArgument(t *testing.T) {
	s := newTestServer(t)
	s.opts.Executable = fakeDriplaneDumpArgs(t)

	const flagLikeName = "-config.rule"
	if err := s.opts.Store.Create(KindRules, flagLikeName, []byte("X => echo();\n")); err != nil {
		t.Fatalf("Create: %s", err)
	}

	rec := do(t, s, http.MethodPost, "/api/test", map[string]string{
		"name":   flagLikeName,
		"source": "X => echo();\n",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var resp testRuleResponse
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}

	lines := strings.Split(strings.TrimSpace(resp.Output), "\n")
	foundRulesFlag := false
	for i, line := range lines {
		if line == "ARG:"+flagLikeName {
			t.Fatalf("the rule name leaked as a standalone argv token: %v", lines)
		}
		if line == "ARG:-rules" {
			foundRulesFlag = true
			if i+1 >= len(lines) {
				t.Fatalf("-rules flag has no value: %v", lines)
			}
			// The handler's deferred os.RemoveAll has already run by the
			// time this test observes the response, so the temp dir no
			// longer exists to os.Stat; just check its name looks like the
			// one handleTestRule generates (filepath.IsAbs plus the
			// os.MkdirTemp pattern), which is enough to confirm it is not
			// the rule name.
			dir := strings.TrimPrefix(lines[i+1], "ARG:")
			if !filepath.IsAbs(dir) || !strings.Contains(filepath.Base(dir), "driplane-test-") {
				t.Errorf("-rules value %q does not look like the generated temp dir: %v", dir, lines)
			}
		}
	}
	if !foundRulesFlag {
		t.Fatalf("-rules flag not found in subprocess args: %v", lines)
	}
}

// fakeSlowDriplane behaves like the real binary but takes a moment, so a
// batch of concurrent /api/test requests actually overlap in flight instead
// of finishing before the next one starts.
func fakeSlowDriplane(t *testing.T) string {
	t.Helper()
	if runtime.GOOS == "windows" {
		t.Skip("the fake binary is a shell script")
	}

	path := filepath.Join(t.TempDir(), "fake-slow-driplane")
	script := `#!/bin/sh
sleep 0.3
echo "rules syntax OK"
exit 0
`
	if err := os.WriteFile(path, []byte(script), 0755); err != nil {
		t.Fatalf("writing the fake binary: %s", err)
	}
	return path
}

// TestTestRuleRejectsExcessConcurrentRequests fires more concurrent
// /api/test requests than testRuleSlots and asserts the overflow is
// rejected with 503 rather than every request being served (which would
// mean an unbounded number of forked dry-run subprocesses).
func TestTestRuleRejectsExcessConcurrentRequests(t *testing.T) {
	s := newTestServer(t)
	s.opts.Executable = fakeSlowDriplane(t)

	const n = testRuleSlots + 3
	var wg sync.WaitGroup
	codes := make([]int, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			raw, err := json.Marshal(map[string]string{
				"name":   "concurrent.rule",
				"source": "X => echo();\n",
			})
			if err != nil {
				codes[i] = -1
				return
			}
			req := httptest.NewRequest(http.MethodPost, "/api/test", bytes.NewReader(raw))
			req.Header.Set("Authorization", "Bearer secret-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			codes[i] = rec.Code
		}(i)
	}
	wg.Wait()

	ok, busy := 0, 0
	for i, code := range codes {
		switch code {
		case http.StatusOK:
			ok++
		case http.StatusServiceUnavailable:
			busy++
		default:
			t.Errorf("goroutine %d: unexpected status %d", i, code)
		}
	}
	if ok == 0 {
		t.Error("expected at least one request to be served")
	}
	if busy == 0 {
		t.Error("expected at least one request to be rejected with 503 because the concurrency limit was exceeded")
	}
	if ok+busy != n {
		t.Errorf("expected all %d requests accounted for as ok or busy, got ok=%d busy=%d", n, ok, busy)
	}
}

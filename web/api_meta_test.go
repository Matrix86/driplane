package web

import (
	"encoding/json"
	"net/http"
	"strings"
	"testing"

	"github.com/Matrix86/driplane/core"
)

func TestValidateAcceptsValidRule(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": "Good => echo();\n",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if !resp.OK {
		t.Errorf("expected ok=true, got error %q", resp.Error)
	}
}

func TestValidateReportsSyntaxError(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": "this is ||| broken ;",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with ok=false, got %d", rec.Code)
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("a broken rule should not validate")
	}
	if resp.Error == "" {
		t.Error("expected a non-empty error message")
	}
}

func TestValidateReportsUnknownFilter(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": "Bad => thisfilterdoesnotexist();\n",
	})

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("an unknown filter should not validate")
	}
}

func TestMetaListsFiltersAndFeeders(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodGet, "/api/meta", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Filters []string `json:"filters"`
		Feeders []string `json:"feeders"`
		Kinds   []string `json:"kinds"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}

	if !contains(resp.Filters, "echo") {
		t.Errorf("expected 'echo' among the filters, got %v", resp.Filters)
	}
	if !contains(resp.Feeders, "file") {
		t.Errorf("expected 'file' among the feeders, got %v", resp.Feeders)
	}
	if !contains(resp.Kinds, "rules") {
		t.Errorf("expected 'rules' among the kinds, got %v", resp.Kinds)
	}
}

// TestMetaDoesNotReportParams proves the dead paramHints table (7 of its 11
// entries named parameters that do not exist for that feeder/filter, e.g.
// "text" had no "expr"/"matchcase" and "timer"/"rss" used "freq" not
// "interval") was removed rather than half-corrected: a rule written from a
// wrong hint would compile with unknown parameters silently ignored and
// never match anything, which is worse than no hint at all.
func TestMetaDoesNotReportParams(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodGet, "/api/meta", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var raw map[string]json.RawMessage
	if err := json.Unmarshal(rec.Body.Bytes(), &raw); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if _, ok := raw["params"]; ok {
		t.Error("/api/meta should no longer report a 'params' field")
	}
}

// TestMetaReportsOnlyConfiguredKinds proves the UI no longer has to hardcode
// rules/templates/js: a store built with only general.rules_path configured
// reports just that kind, matching Store.Kinds().
func TestMetaReportsOnlyConfiguredKinds(t *testing.T) {
	dir := t.TempDir()
	cfg := &core.Configuration{}
	cfg.SetAll(map[string]string{"general.rules_path": dir})

	store, err := NewStore(cfg)
	if err != nil {
		t.Fatalf("NewStore: %s", err)
	}

	s, err := New(Options{
		Address: "127.0.0.1:0",
		Token:   "secret-token",
		Store:   store,
		Ring:    NewLogRing(10),
	})
	if err != nil {
		t.Fatalf("New: %s", err)
	}

	rec := do(t, s, http.MethodGet, "/api/meta", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rec.Code)
	}

	var resp struct {
		Kinds []string `json:"kinds"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if len(resp.Kinds) != 1 || resp.Kinds[0] != "rules" {
		t.Errorf("expected kinds=[rules], got %v", resp.Kinds)
	}
}

// TestValidateReportsUnknownRuleCall proves fix round 1's Critical/High fix
// for a rule call target that does not exist: the editor must not report
// ok:true for a rule that will fail to compile the moment it is loaded.
func TestValidateReportsUnknownRuleCall(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": "Bad => @doesNotExist;\n",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with ok=false, got %d", rec.Code)
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("a call to a rule that does not exist should not validate")
	}
}

// TestValidateAcceptsKnownRuleCall proves the fix does not produce false
// positives: a rule call to a target defined earlier in the same buffer
// must still validate.
func TestValidateAcceptsKnownRuleCall(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": "Good => echo();\nCaller => @Good;\n",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if !resp.OK {
		t.Errorf("a call to a rule defined in the same buffer should validate, got error %q", resp.Error)
	}
}

// TestValidateRejectsUnsupportedKind proves a bad "kind" is a malformed
// request (400), not a rule that fails validation (200 ok:false).
func TestValidateRejectsUnsupportedKind(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "templates",
		"source": "Good => echo();\n",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400 for an unsupported kind, got %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestValidateRejectsExcessiveChainWithoutCrashing sends a rule with roughly
// 130,000 chained links -- well under the 1 MiB body cap, but enough to blow
// the goroutine stack via the participle grammar's per-link recursion if
// unbounded -- and proves the server answers with a clean ok:false instead
// of dying.
func TestValidateRejectsExcessiveChainWithoutCrashing(t *testing.T) {
	s := newTestServer(t)

	source := "Bad => echo()" + strings.Repeat("|echo()", 130000) + ";\n"

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": source,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with ok=false, got %d", rec.Code)
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("an excessively long chain should not validate")
	}
	if !strings.Contains(resp.Error, "too long") {
		t.Errorf("expected a 'too long' error, got %q", resp.Error)
	}
}

// TestValidateRejectsExcessiveChainWithLeadingCommentQuoteWithoutCrashing is
// fix round 2's end-to-end regression test: round 1's chain-length counter
// treated any quote as a string delimiter, including one inside a '#'
// comment, so a quote in a leading comment line would pair with a later one
// and swallow the entire chain from the count, letting it straight through
// to the parser and crashing the process with a stack overflow. This is the
// coordinator's own reproduction, exercised through the real HTTP handler.
func TestValidateRejectsExcessiveChainWithLeadingCommentQuoteWithoutCrashing(t *testing.T) {
	s := newTestServer(t)

	source := "# \"\nBad => echo()" + strings.Repeat("|echo()", 130000) + ";\n"

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": source,
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with ok=false, got %d", rec.Code)
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("an excessively long chain preceded by a comment containing a quote should not validate")
	}
	if !strings.Contains(resp.Error, "too long") {
		t.Errorf("expected a 'too long' error, got %q", resp.Error)
	}
}

// TestValidateRejectsImportEscapingRoot proves the #import confinement is
// wired through the HTTP endpoint: an import climbing out of the rules
// directory is a rule-content problem (ok:false), not a crash and not a
// successful arbitrary file read.
func TestValidateRejectsImportEscapingRoot(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/validate", map[string]string{
		"kind":   "rules",
		"source": "#import \"../../../../../../../../etc/hostname\"\nBad => echo();\n",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200 with ok=false, got %d", rec.Code)
	}

	var resp struct {
		OK    bool   `json:"ok"`
		Error string `json:"error"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &resp); err != nil {
		t.Fatalf("decoding: %s", err)
	}
	if resp.OK {
		t.Error("an import escaping the rules directory should not validate")
	}
}

func contains(list []string, want string) bool {
	for _, v := range list {
		if v == want {
			return true
		}
	}
	return false
}

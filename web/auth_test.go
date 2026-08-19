package web

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func newTestServer(t *testing.T) *Server {
	t.Helper()
	store, _ := newTestStore(t)
	s, err := New(Options{
		Address: "127.0.0.1:0",
		Token:   "secret-token",
		Store:   store,
		Ring:    NewLogRing(10),
	})
	if err != nil {
		t.Fatalf("New: %s", err)
	}
	return s
}

func TestGenerateTokenIsRandom(t *testing.T) {
	a, err := GenerateToken()
	if err != nil {
		t.Fatalf("GenerateToken: %s", err)
	}
	b, _ := GenerateToken()
	if a == b {
		t.Error("GenerateToken should not return the same token twice")
	}
	if len(a) < 32 {
		t.Errorf("token too short: %d chars", len(a))
	}
}

func TestAuthRejectsMissingOrWrongToken(t *testing.T) {
	s := newTestServer(t)

	cases := []struct {
		name   string
		header string
	}{
		{"missing", ""},
		{"wrong", "Bearer nope"},
		{"malformed", "secret-token"},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
			if tc.header != "" {
				req.Header.Set("Authorization", tc.header)
			}
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)

			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestAuthAcceptsBearerToken(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/status", nil)
	req.Header.Set("Authorization", "Bearer secret-token")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code == http.StatusUnauthorized {
		t.Fatal("a valid bearer token should be accepted")
	}
}

func TestAuthAcceptsQueryTokenOnlyOnGet(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodGet, "/api/status?token=secret-token", nil)
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code == http.StatusUnauthorized {
		t.Error("GET with a query token should be accepted (EventSource cannot set headers)")
	}

	req = httptest.NewRequest(http.MethodPost, "/api/reload?token=secret-token", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST with a query token must be rejected, got %d", rec.Code)
	}
}

func TestStaticUIAssetsAreExemptFromAuth(t *testing.T) {
	s := newTestServer(t)

	paths := []string{"/", "/app.js", "/style.css", "/mode-rule.js", "/vendor/codemirror.js"}
	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusOK {
				t.Errorf("GET %s with no token: expected 200, got %d, body: %s", p, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestAPIStillRequiresAuthWithoutToken(t *testing.T) {
	s := newTestServer(t)

	cases := []struct {
		method string
		path   string
	}{
		{http.MethodGet, "/api/status"},
		{http.MethodGet, "/api/files?kind=rules"},
	}

	for _, tc := range cases {
		t.Run(tc.method+" "+tc.path, func(t *testing.T) {
			req := httptest.NewRequest(tc.method, tc.path, nil)
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("expected 401, got %d", rec.Code)
			}
		})
	}
}

func TestReloadStillRequiresAuth(t *testing.T) {
	s := newTestServer(t)

	// No token at all.
	req := httptest.NewRequest(http.MethodPost, "/api/reload", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST /api/reload with no token: expected 401, got %d", rec.Code)
	}

	// Query token only: POST must not accept it, exemption or not.
	req = httptest.NewRequest(http.MethodPost, "/api/reload?token=secret-token", strings.NewReader("{}"))
	req.Header.Set("Content-Type", "application/json")
	rec = httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	if rec.Code != http.StatusUnauthorized {
		t.Errorf("POST /api/reload with only a query token: expected 401, got %d", rec.Code)
	}
}

// TestStaticExemptionDoesNotLeakIntoAPI proves the static-asset exemption
// cannot be defeated by a path that resolves under /api/ after cleaning,
// whether via a literal ".." segment or a percent-encoded one. Both must
// still 401 without a token: isExemptStaticPath runs path.Clean itself,
// ahead of http.ServeMux's own routing-time cleanup, specifically so these
// requests never reach the mux while still looking "static".
func TestStaticExemptionDoesNotLeakIntoAPI(t *testing.T) {
	s := newTestServer(t)

	paths := []string{
		"/api/../api/status",     // literal ".." segment
		"/api/%2e%2e/api/status", // percent-encoded ".." segment
		"/vendor/../api/status",  // looks static, cleans to an API path
		"/API/status",            // differently-cased API path
		"/Api/Files?kind=rules",  // differently-cased API path, mixed case
	}

	for _, p := range paths {
		t.Run(p, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodGet, p, nil)
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			// A 404 here would mean the request reached the file server
			// unauthenticated -- exactly the leak this test exists to catch
			// -- so the assertion must pin 401 specifically, not just
			// "not 200".
			if rec.Code != http.StatusUnauthorized {
				t.Errorf("GET %s with no token: expected 401 (not exempt), got %d, body: %s", p, rec.Code, rec.Body.String())
			}
		})
	}
}

func TestMutatingRequestsRequireJSONContentType(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/reload", strings.NewReader("x=1"))
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code != http.StatusUnsupportedMediaType {
		t.Errorf("expected 415, got %d", rec.Code)
	}
}

// TestMutatingRequestsAcceptCaseInsensitiveJSONContentType proves the
// content-type gate matches RFC 9110, which treats "Application/JSON" as
// equivalent to "application/json": a case-sensitive prefix check would
// reject a perfectly valid request from a client that happens to send a
// different casing.
func TestMutatingRequestsAcceptCaseInsensitiveJSONContentType(t *testing.T) {
	s := newTestServer(t)

	req := httptest.NewRequest(http.MethodPost, "/api/reload", strings.NewReader("{}"))
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Content-Type", "Application/JSON")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)

	if rec.Code == http.StatusUnsupportedMediaType {
		t.Errorf("a differently-cased 'Application/JSON' content type should be accepted, got 415")
	}
}

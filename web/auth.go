package web

import (
	"crypto/rand"
	"crypto/subtle"
	"encoding/base64"
	"net/http"
	"path"
	"strings"
)

// GenerateToken returns a new random API token
func GenerateToken() (string, error) {
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	return base64.RawURLEncoding.EncodeToString(raw), nil
}

// tokenFromRequest extracts the token from the Authorization header. The query
// string is accepted on GET requests only: EventSource cannot set headers, and
// keeping it out of the mutating methods is what stops cross-origin forms.
func tokenFromRequest(r *http.Request) string {
	if auth := r.Header.Get("Authorization"); auth != "" {
		if after, found := strings.CutPrefix(auth, "Bearer "); found {
			return strings.TrimSpace(after)
		}
		return ""
	}
	if r.Method == http.MethodGet {
		return r.URL.Query().Get("token")
	}
	return ""
}

// isExemptStaticPath reports whether r is a GET for something other than the
// JSON API, i.e. the UI shell itself: index.html, style.css, app.js,
// mode-rule.js and the vendored CodeMirror assets under vendor/.
//
// This exists because a browser cannot authenticate these requests: it
// fetches every <link>/<script src> referenced by index.html as its own,
// separate request, before app.js (which is one of those sub-resources) has
// had a chance to run and there is no way to attach the query-string token
// or an Authorization header to a tag-driven fetch. Without this exemption
// the page loads its HTML shell and then 401s on every CSS/JS sub-resource,
// so app.js never executes and the UI is permanently blank.
//
// The exemption is safe to grant because none of these files are secret:
// they are vendored third-party code plus this repository's own HTML/CSS/JS,
// already public in source control, containing no per-installation data.
// GET is the only method exempted, and only for paths that do not start with
// "/api/" — every JSON endpoint (including GET ones like /api/files and
// /api/status, which do return operator data) stays fully gated behind the
// existing token check below, with the query-token-on-GET-only rule
// unchanged for them.
//
// An alternative considered and rejected: issue a session cookie on the
// first authenticated navigation so the browser attaches it automatically to
// sub-resource requests. That would reintroduce a CSRF surface (cookies are
// sent automatically by the browser on cross-origin requests too, unlike the
// header-only bearer token this server otherwise relies on) purely to
// protect files that hold no secrets. Not worth the trade — do not "fix"
// this by switching to cookies without re-deriving that trade-off.
//
// The path is cleaned with path.Clean before the prefix test, since this
// check runs in front of http.ServeMux (which only cleans paths for its own
// routing, after this middleware has already run), so a raw ../ segment
// could otherwise walk a path that looks static onto something under /api/.
// Anything this function is not certain is exempt must fall through to
// requiring the token: it returns true only for the exact GET-and-not-/api/
// case and false for everything else, including any path that fails to
// resolve to a clean absolute form.
func isExemptStaticPath(r *http.Request) bool {
	if r.Method != http.MethodGet {
		return false
	}
	clean := path.Clean(r.URL.Path)
	if !strings.HasPrefix(clean, "/") {
		// path.Clean always returns a rooted path for a rooted input, and
		// r.URL.Path always starts with "/" for a server-side request; this
		// is just a defensive belt-and-suspenders check against something
		// this function did not anticipate.
		return false
	}
	// Lower-cased deliberately: this check must stay safe on its own, not
	// because http.ServeMux and embed.FS also happen to be case-sensitive.
	// Without this, "/API/status" would slip past a case-sensitive "/api/"
	// prefix test and only fail to expose anything because the router
	// downstream also treats "/API" as a 404 rather than routing it to the
	// status handler -- i.e. the auth boundary would be relying on an
	// unstated coupling to routing behavior in another file. Comparing
	// lower-cased makes any casing of "/api/..." require the token
	// regardless of how the mux or embed.FS happen to resolve it.
	clean = strings.ToLower(clean)
	return clean != "/api" && !strings.HasPrefix(clean, "/api/")
}

// authMiddleware enforces the token on every request and, for the mutating
// methods, a JSON content type. It wraps the whole mux instead of the single
// handlers so that unknown paths and wrong methods answer 401 too, without
// leaking which routes exist. GET requests for the static UI shell are
// exempt from the token check; see isExemptStaticPath.
func (s *Server) authMiddleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if isExemptStaticPath(r) {
			next.ServeHTTP(w, r)
			return
		}

		got := tokenFromRequest(r)
		if subtle.ConstantTimeCompare([]byte(got), []byte(s.opts.Token)) != 1 {
			w.Header().Set("WWW-Authenticate", "Bearer")
			writeError(w, http.StatusUnauthorized, errUnauthorized)
			return
		}

		switch r.Method {
		case http.MethodPost, http.MethodPut, http.MethodDelete:
			ct := r.Header.Get("Content-Type")
			// Lower-cased deliberately, for the same reason as
			// isExemptStaticPath above: RFC 9110 treats "Application/JSON"
			// as equivalent to "application/json", and a case-sensitive
			// prefix test would reject a perfectly valid request.
			if ct != "" && !strings.HasPrefix(strings.ToLower(ct), "application/json") {
				writeError(w, http.StatusUnsupportedMediaType, errNotJSON)
				return
			}
			if ct == "" && r.Method != http.MethodDelete {
				writeError(w, http.StatusUnsupportedMediaType, errNotJSON)
				return
			}
		}

		next.ServeHTTP(w, r)
	})
}

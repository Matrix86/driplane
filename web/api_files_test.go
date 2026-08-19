package web

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
)

func do(t *testing.T, s *Server, method, path string, body any) *httptest.ResponseRecorder {
	t.Helper()

	var reader *bytes.Reader
	if body != nil {
		raw, err := json.Marshal(body)
		if err != nil {
			t.Fatalf("marshalling body: %s", err)
		}
		reader = bytes.NewReader(raw)
	} else {
		reader = bytes.NewReader(nil)
	}

	req := httptest.NewRequest(method, path, reader)
	req.Header.Set("Authorization", "Bearer secret-token")
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	s.Handler().ServeHTTP(rec, req)
	return rec
}

func TestFilesCreateListReadDelete(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/files/rules", map[string]string{
		"name":    "a.rule",
		"content": "A => echo();\n",
	})
	if rec.Code != http.StatusCreated {
		t.Fatalf("create: expected 201, got %d (%s)", rec.Code, rec.Body.String())
	}

	rec = do(t, s, http.MethodGet, "/api/files?kind=rules", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("list: expected 200, got %d", rec.Code)
	}
	var list []FileInfo
	if err := json.Unmarshal(rec.Body.Bytes(), &list); err != nil {
		t.Fatalf("decoding list: %s", err)
	}
	if len(list) != 1 || list[0].Name != "a.rule" {
		t.Fatalf("expected [a.rule], got %+v", list)
	}

	rec = do(t, s, http.MethodGet, "/api/files/rules/a.rule", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("read: expected 200, got %d", rec.Code)
	}
	var read struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &read); err != nil {
		t.Fatalf("decoding read: %s", err)
	}
	if read.Content != "A => echo();\n" {
		t.Errorf("unexpected content: %q", read.Content)
	}

	rec = do(t, s, http.MethodDelete, "/api/files/rules/a.rule", nil)
	if rec.Code != http.StatusNoContent {
		t.Fatalf("delete: expected 204, got %d", rec.Code)
	}

	rec = do(t, s, http.MethodGet, "/api/files/rules/a.rule", nil)
	if rec.Code != http.StatusNotFound {
		t.Errorf("read after delete: expected 404, got %d", rec.Code)
	}
}

func TestFilesRejectTraversalWithStatus400(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodPost, "/api/files/rules", map[string]string{
		"name":    "../evil.rule",
		"content": "x",
	})
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestFilesUnknownKind(t *testing.T) {
	s := newTestServer(t)

	rec := do(t, s, http.MethodGet, "/api/files?kind=config", nil)
	if rec.Code != http.StatusBadRequest {
		t.Errorf("expected 400, got %d", rec.Code)
	}
}

func TestFilesWriteConflict(t *testing.T) {
	s := newTestServer(t)

	if rec := do(t, s, http.MethodPost, "/api/files/rules", map[string]string{
		"name": "a.rule", "content": "A => echo();\n",
	}); rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d", rec.Code)
	}

	rec := do(t, s, http.MethodPut, "/api/files/rules/a.rule", map[string]string{
		"content": "B => echo();\n",
		"mtime":   "2000-01-01T00:00:00Z",
	})
	if rec.Code != http.StatusConflict {
		t.Errorf("expected 409, got %d (%s)", rec.Code, rec.Body.String())
	}
}

// TestFilesCreateConcurrentIsAtomic fires N concurrent POSTs for the same
// name (20, which is what exposed the check-then-act race in the previous
// Read-then-Write implementation of handleFileCreate) and asserts exactly one
// wins with 201 while every other one observes 409, with the surviving file's
// content matching whichever writer actually won. It intentionally does not
// call *testing.T from inside the goroutines (only testing.T.FailNow-family
// calls from the test's own goroutine are safe); results are collected and
// asserted after every goroutine has finished.
func TestFilesCreateConcurrentIsAtomic(t *testing.T) {
	s := newTestServer(t)

	const n = 20
	var wg sync.WaitGroup
	codes := make([]int, n)
	contents := make([]string, n)

	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			content := fmt.Sprintf("W%d => echo();\n", i)
			contents[i] = content

			raw, err := json.Marshal(map[string]string{
				"name":    "race.rule",
				"content": content,
			})
			if err != nil {
				codes[i] = -1
				return
			}
			req := httptest.NewRequest(http.MethodPost, "/api/files/rules", bytes.NewReader(raw))
			req.Header.Set("Authorization", "Bearer secret-token")
			req.Header.Set("Content-Type", "application/json")
			rec := httptest.NewRecorder()
			s.Handler().ServeHTTP(rec, req)
			codes[i] = rec.Code
		}(i)
	}
	wg.Wait()

	created, conflict, winner := 0, 0, -1
	for i, code := range codes {
		switch code {
		case http.StatusCreated:
			created++
			winner = i
		case http.StatusConflict:
			conflict++
		default:
			t.Errorf("goroutine %d: unexpected status %d", i, code)
		}
	}
	if created != 1 {
		t.Fatalf("expected exactly 1 created, got %d created and %d conflicts (codes=%v)", created, conflict, codes)
	}
	if conflict != n-1 {
		t.Fatalf("expected %d conflicts, got %d", n-1, conflict)
	}

	rec := do(t, s, http.MethodGet, "/api/files/rules/race.rule", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("read winner: expected 200, got %d", rec.Code)
	}
	var read struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &read); err != nil {
		t.Fatalf("decoding read: %s", err)
	}
	if read.Content != contents[winner] {
		t.Errorf("file content does not match the winning writer: got %q, want %q", read.Content, contents[winner])
	}
}

// TestFilesWriteEmptyBodyRejected makes sure a bodyless PUT is rejected with
// 400 and, crucially, that it does not truncate the existing file: a 400
// response with the file already wiped would still be a data-loss bug.
func TestFilesWriteEmptyBodyRejected(t *testing.T) {
	s := newTestServer(t)

	if rec := do(t, s, http.MethodPost, "/api/files/rules", map[string]string{
		"name": "keep.rule", "content": "KEEP => echo();\n",
	}); rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d", rec.Code)
	}

	rec := do(t, s, http.MethodPut, "/api/files/rules/keep.rule", nil)
	if rec.Code != http.StatusBadRequest {
		t.Fatalf("expected 400, got %d (%s)", rec.Code, rec.Body.String())
	}

	rec = do(t, s, http.MethodGet, "/api/files/rules/keep.rule", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("read: expected 200, got %d", rec.Code)
	}
	var read struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &read); err != nil {
		t.Fatalf("decoding read: %s", err)
	}
	if read.Content != "KEEP => echo();\n" {
		t.Errorf("file content was modified by the rejected write: got %q", read.Content)
	}
}

// TestFilesWriteExplicitEmptyContentSucceeds proves the empty-body rejection
// above does not over-reject: an operator explicitly asking for an empty file
// via {"content":""} must still succeed.
func TestFilesWriteExplicitEmptyContentSucceeds(t *testing.T) {
	s := newTestServer(t)

	if rec := do(t, s, http.MethodPost, "/api/files/rules", map[string]string{
		"name": "empty.rule", "content": "X => echo();\n",
	}); rec.Code != http.StatusCreated {
		t.Fatalf("create: got %d", rec.Code)
	}

	rec := do(t, s, http.MethodPut, "/api/files/rules/empty.rule", map[string]string{
		"content": "",
	})
	if rec.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d (%s)", rec.Code, rec.Body.String())
	}

	rec = do(t, s, http.MethodGet, "/api/files/rules/empty.rule", nil)
	if rec.Code != http.StatusOK {
		t.Fatalf("read: expected 200, got %d", rec.Code)
	}
	var read struct {
		Content string `json:"content"`
	}
	if err := json.Unmarshal(rec.Body.Bytes(), &read); err != nil {
		t.Fatalf("decoding read: %s", err)
	}
	if read.Content != "" {
		t.Errorf("expected empty content, got %q", read.Content)
	}
}

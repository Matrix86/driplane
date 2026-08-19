package web

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/evilsocket/islazy/log"
)

// errEmptyBody signals "no body was sent at all", as distinct from a body
// that explicitly says the content is empty (e.g. {"content":""}). Do not
// collapse this back into a plain nil return from decodeBody: some handlers
// (the runtime action endpoints, called with "{}") legitimately accept an
// empty body and must be able to tell the difference from a client that
// forgot to send one, which for handleFileWrite/handleFileCreate must be a
// 400 rather than silently truncating a file.
var errEmptyBody = errors.New("request body is required")

type fileWriteRequest struct {
	Content string `json:"content"`
	MTime   string `json:"mtime"`
}

type fileCreateRequest struct {
	Name    string `json:"name"`
	Content string `json:"content"`
}

type fileReadResponse struct {
	Name    string    `json:"name"`
	Content string    `json:"content"`
	MTime   time.Time `json:"mtime"`
}

// statusForStoreError maps the store errors to HTTP status codes
func statusForStoreError(err error) int {
	switch {
	case errors.Is(err, ErrNotFound):
		return http.StatusNotFound
	case errors.Is(err, ErrConflict):
		return http.StatusConflict
	case errors.Is(err, ErrInvalidKind), errors.Is(err, ErrInvalidPath), errors.Is(err, ErrInvalidExt):
		return http.StatusBadRequest
	default:
		return http.StatusInternalServerError
	}
}

func decodeBody(r *http.Request, v any) error {
	body := http.MaxBytesReader(nil, r.Body, maxBodySize)
	defer body.Close()

	raw, err := io.ReadAll(body)
	if err != nil {
		return fmt.Errorf("reading body: %s", err)
	}
	if len(raw) == 0 {
		return errEmptyBody
	}
	return json.Unmarshal(raw, v)
}

// writeStoreError maps a Store error to its HTTP status and writes it. On a
// 500 the underlying error (which can contain absolute filesystem paths) is
// logged server-side only; the client gets a generic message. The 400/404/409
// messages are left as-is: those are actionable and the editor shows them to
// the operator.
func writeStoreError(w http.ResponseWriter, err error) {
	status := statusForStoreError(err)
	if status == http.StatusInternalServerError {
		log.Error("web: file store error: %s", err)
		writeError(w, status, errors.New("internal server error"))
		return
	}
	writeError(w, status, err)
}

func (s *Server) handleFileList(w http.ResponseWriter, r *http.Request) {
	kind := Kind(r.URL.Query().Get("kind"))
	files, err := s.opts.Store.List(kind)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, files)
}

func (s *Server) handleFileRead(w http.ResponseWriter, r *http.Request) {
	kind := Kind(r.PathValue("kind"))
	name := r.PathValue("name")

	content, mtime, err := s.opts.Store.Read(kind, name)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, fileReadResponse{Name: name, Content: string(content), MTime: mtime})
}

func (s *Server) handleFileWrite(w http.ResponseWriter, r *http.Request) {
	kind := Kind(r.PathValue("kind"))
	name := r.PathValue("name")

	var req fileWriteRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	var ifUnmodified time.Time
	if req.MTime != "" {
		parsed, err := time.Parse(time.RFC3339Nano, req.MTime)
		if err != nil {
			writeError(w, http.StatusBadRequest, fmt.Errorf("invalid mtime: %s", err))
			return
		}
		ifUnmodified = parsed
	}

	if err := s.opts.Store.Write(kind, name, []byte(req.Content), ifUnmodified); err != nil {
		writeStoreError(w, err)
		return
	}

	_, mtime, err := s.opts.Store.Read(kind, name)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusOK, map[string]any{"name": name, "mtime": mtime})
}

func (s *Server) handleFileCreate(w http.ResponseWriter, r *http.Request) {
	kind := Kind(r.PathValue("kind"))

	var req fileCreateRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}

	// Store.Create resolves the name (the security boundary) and then opens
	// the file with O_EXCL, so "does not already exist" is an atomic property
	// of the create itself: concurrent callers racing on the same name cannot
	// both win, unlike a preliminary Read-then-Write check.
	if err := s.opts.Store.Create(kind, req.Name, []byte(req.Content)); err != nil {
		writeStoreError(w, err)
		return
	}

	_, mtime, err := s.opts.Store.Read(kind, req.Name)
	if err != nil {
		writeStoreError(w, err)
		return
	}
	writeJSON(w, http.StatusCreated, map[string]any{"name": req.Name, "mtime": mtime})
}

func (s *Server) handleFileDelete(w http.ResponseWriter, r *http.Request) {
	kind := Kind(r.PathValue("kind"))
	name := r.PathValue("name")

	if err := s.opts.Store.Delete(kind, name); err != nil {
		writeStoreError(w, err)
		return
	}
	w.WriteHeader(http.StatusNoContent)
}

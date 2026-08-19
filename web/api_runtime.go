package web

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

type actionResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

// runtimeAction runs fn on the supervisor and always answers with 200 and
// ok:true once fn has been called: fn (Supervisor.Reload/ResumeFeeders/
// PauseFeeders) only requests the action and returns immediately -- the
// actual rebuild happens later, asynchronously, in the Supervisor's Run
// loop. So ok:true here means only "the request was accepted", never "the
// reload succeeded"; this handler has no way to know that outcome without
// blocking on the Run loop, which an HTTP handler must not do. The only
// ok:false case left is "no supervisor attached", which cannot happen in
// production (web.Start always wires one in) but keeps this endpoint usable
// in tests that build a Server without one. Callers must poll /api/status to
// learn whether a reload actually succeeded.
func (s *Server) runtimeAction(w http.ResponseWriter, fn func()) {
	if s.opts.Supervisor == nil {
		writeJSON(w, http.StatusOK, actionResponse{OK: false, Error: "no supervisor attached"})
		return
	}
	fn()
	writeJSON(w, http.StatusOK, actionResponse{OK: true})
}

func (s *Server) handleReload(w http.ResponseWriter, r *http.Request) {
	s.runtimeAction(w, func() { s.opts.Supervisor.Reload() })
}

func (s *Server) handleStart(w http.ResponseWriter, r *http.Request) {
	s.runtimeAction(w, func() { s.opts.Supervisor.ResumeFeeders() })
}

func (s *Server) handleStop(w http.ResponseWriter, r *http.Request) {
	s.runtimeAction(w, func() { s.opts.Supervisor.PauseFeeders() })
}

// handleLogs streams the log lines as Server-Sent Events. A line appended
// between the subscription and the backlog dump can be delivered twice: the UI
// tolerates it, and losing lines would be worse than repeating one.
func (s *Server) handleLogs(w http.ResponseWriter, r *http.Request) {
	flusher, ok := w.(http.Flusher)
	if !ok {
		writeError(w, http.StatusInternalServerError, fmt.Errorf("streaming unsupported"))
		return
	}

	ch, cancel := s.opts.Ring.Subscribe(256)
	defer cancel()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)

	for _, line := range s.opts.Ring.Backlog() {
		writeSSE(w, line)
	}
	flusher.Flush()

	keepalive := time.NewTicker(20 * time.Second)
	defer keepalive.Stop()

	for {
		select {
		case <-r.Context().Done():
			return
		case <-s.closing:
			// Server.Shutdown is in progress: end the stream promptly instead
			// of leaving http.Server.Shutdown waiting for this connection to
			// go idle, which it never does on its own.
			return
		case line, open := <-ch:
			if !open {
				return
			}
			writeSSE(w, line)
			flusher.Flush()
		case <-keepalive.C:
			fmt.Fprint(w, ": keepalive\n\n")
			flusher.Flush()
		}
	}
}

func writeSSE(w http.ResponseWriter, line LogLine) {
	raw, err := json.Marshal(line)
	if err != nil {
		return
	}
	fmt.Fprintf(w, "data: %s\n\n", raw)
}

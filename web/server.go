package web

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"fmt"
	"io/fs"
	"net"
	"net/http"
	"os"
	"strings"
	"sync"
	"time"

	"github.com/Matrix86/driplane/core"

	"github.com/evilsocket/islazy/log"
)

//go:embed ui
var uiFS embed.FS

const maxBodySize = 1 << 20 // 1 MiB

var (
	errUnauthorized = errors.New("unauthorized")
	errNotJSON      = errors.New("content type must be application/json")
)

// Options configures the web Server
type Options struct {
	Address    string
	Token      string
	Supervisor *core.Supervisor
	Store      *Store
	Ring       *LogRing
	Executable string // binary used for the dry-run subprocess
}

// testRuleSlots bounds how many dry-run subprocesses (see handleTestRule)
// may be forked at once. It is not a general request-concurrency limit: it
// exists only because each /api/test call copies the whole rules directory
// and forks the full driplane binary, which re-reads the config and
// re-compiles every rule on the same host as the live daemon. A double
// click, a few open editor tabs or a trivial script must not be able to
// drive an unbounded number of those.
const testRuleSlots = 2

// Server serves the driplane web interface and its JSON API
type Server struct {
	opts      Options
	mux       *http.ServeMux
	handler   http.Handler
	srv       *http.Server
	testSlots chan struct{} // bounds concurrent dry-run subprocesses, see testRuleSlots

	closing     chan struct{} // closed by Shutdown, see handleLogs in api_runtime.go
	closingOnce sync.Once
}

// New creates a Server. It does not start listening.
func New(opts Options) (*Server, error) {
	if opts.Token == "" {
		return nil, fmt.Errorf("a token is required")
	}
	if opts.Store == nil {
		return nil, fmt.Errorf("a store is required")
	}
	if opts.Ring == nil {
		opts.Ring = NewLogRing(1000)
	}
	if opts.Executable == "" {
		exe, err := os.Executable()
		if err != nil {
			return nil, fmt.Errorf("resolving the executable path: %s", err)
		}
		opts.Executable = exe
	}

	s := &Server{opts: opts, mux: http.NewServeMux(), testSlots: make(chan struct{}, testRuleSlots), closing: make(chan struct{})}
	s.routes()
	s.handler = s.authMiddleware(s.mux)
	s.srv = &http.Server{
		Addr:              opts.Address,
		Handler:           s.handler,
		ReadHeaderTimeout: 10 * time.Second,
		ReadTimeout:       30 * time.Second, // caps reading the request only; SSE responses are unaffected, see api_runtime_test.go
		WriteTimeout:      0,                // SSE streams must not be cut
		IdleTimeout:       120 * time.Second,
	}
	return s, nil
}

func (s *Server) routes() {
	static, err := fs.Sub(uiFS, "ui")
	if err != nil {
		panic(fmt.Sprintf("embedded ui: %s", err))
	}
	s.mux.HandleFunc("GET /api/status", s.handleStatus)
	s.mux.HandleFunc("POST /api/reload", s.handleReload)
	s.mux.HandleFunc("POST /api/start", s.handleStart)
	s.mux.HandleFunc("POST /api/stop", s.handleStop)
	s.mux.HandleFunc("GET /api/logs", s.handleLogs)

	s.mux.HandleFunc("GET /api/files", s.handleFileList)
	s.mux.HandleFunc("GET /api/files/{kind}/{name}", s.handleFileRead)
	s.mux.HandleFunc("PUT /api/files/{kind}/{name}", s.handleFileWrite)
	s.mux.HandleFunc("DELETE /api/files/{kind}/{name}", s.handleFileDelete)
	s.mux.HandleFunc("POST /api/files/{kind}", s.handleFileCreate)

	s.mux.HandleFunc("GET /api/meta", s.handleMeta)
	s.mux.HandleFunc("POST /api/validate", s.handleValidate)
	s.mux.HandleFunc("POST /api/test", s.handleTestRule)

	s.mux.Handle("GET /", http.FileServer(http.FS(static)))
}

// Handler returns the authenticated http.Handler serving the whole interface
func (s *Server) Handler() http.Handler {
	return s.handler
}

// ListenAndServe starts serving and blocks
func (s *Server) ListenAndServe() error {
	return s.srv.ListenAndServe()
}

// Shutdown stops the server gracefully. It closes a channel that handleLogs
// selects on alongside its other cases, before deferring to
// http.Server.Shutdown: that method otherwise waits for every connection to
// go idle, and an SSE stream (handleLogs) never does on its own -- a browser
// tab left open on the log view would make every SIGTERM cost the full
// shutdown timeout. Closing this channel first ends any open SSE stream
// promptly, so Shutdown can return as soon as everything else is idle.
func (s *Server) Shutdown(ctx context.Context) error {
	s.closingOnce.Do(func() { close(s.closing) })
	return s.srv.Shutdown(ctx)
}

// Start builds and starts the web server from the driplane configuration.
// It returns nil, nil when the web interface is disabled.
func Start(sup *core.Supervisor) (*Server, error) {
	cfg := sup.Config()
	if cfg.Get("web.enable") != "true" {
		return nil, nil
	}

	address := cfg.Get("web.address")
	if address == "" {
		address = "127.0.0.1:8080"
	}

	token := cfg.Get("web.token")
	generated := false
	if token == "" {
		var err error
		if token, err = GenerateToken(); err != nil {
			return nil, err
		}
		generated = true
	}

	store, err := NewStore(cfg)
	if err != nil {
		return nil, err
	}

	ring := NewLogRing(1000)
	ring.Attach()

	srv, err := New(Options{
		Address:    address,
		Token:      token,
		Supervisor: sup,
		Store:      store,
		Ring:       ring,
	})
	if err != nil {
		return nil, err
	}

	if !isLoopback(address) {
		log.Warning("the web interface is bound to %s, which is not loopback: anyone able to reach it can write rules that this daemon will execute", address)
	}

	go func() {
		if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Error("web server: %s", err)
		}
	}()

	if generated {
		log.Important("web interface on http://%s/?token=%s", address, token)
	} else {
		log.Important("web interface on http://%s/", address)
	}
	return srv, nil
}

func isLoopback(address string) bool {
	host, _, err := net.SplitHostPort(address)
	if err != nil {
		host = address
	}
	if host == "" || strings.EqualFold(host, "localhost") {
		return host != "" // an empty host means "all interfaces"
	}
	ip := net.ParseIP(host)
	return ip != nil && ip.IsLoopback()
}

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json")
	w.Header().Set("Cache-Control", "no-store")
	w.WriteHeader(status)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Debug("web: writing response: %s", err)
	}
}

func writeError(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]string{"error": err.Error()})
}

// handleStatus reports the supervisor state.
func (s *Server) handleStatus(w http.ResponseWriter, r *http.Request) {
	if s.opts.Supervisor == nil {
		writeJSON(w, http.StatusOK, core.Status{State: "stopped", Rules: []core.RuleInfo{}})
		return
	}
	writeJSON(w, http.StatusOK, s.opts.Supervisor.Status())
}

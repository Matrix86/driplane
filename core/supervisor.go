package core

import (
	"sync"

	"github.com/evilsocket/islazy/log"
)

// State represents the lifecycle state of the Supervisor
type State int

// The possible Supervisor states
const (
	StateStopped State = iota
	StateRunning
	StateReloading
	StateError
)

// String returns the lowercase name of the state
func (s State) String() string {
	switch s {
	case StateRunning:
		return "running"
	case StateReloading:
		return "reloading"
	case StateError:
		return "error"
	default:
		return "stopped"
	}
}

// Status is a snapshot of the Supervisor state, safe to serialize
type Status struct {
	State string     `json:"state"`
	Error string     `json:"error,omitempty"`
	Rules []RuleInfo `json:"rules"`
}

// Supervisor owns the Orchestrator lifecycle: build, start, reload, stop
type Supervisor struct {
	config *Configuration

	mu      sync.RWMutex
	orch    *Orchestrator
	state   State
	lastErr error
	paused  bool

	reload   chan struct{}
	quit     chan struct{}
	quitOnce sync.Once
}

// NewSupervisor creates a Supervisor for the given configuration
func NewSupervisor(cfg *Configuration) *Supervisor {
	return &Supervisor{
		config: cfg,
		state:  StateStopped,
		reload: make(chan struct{}, 1),
		quit:   make(chan struct{}),
	}
}

// Config returns the configuration the Supervisor was built with
func (s *Supervisor) Config() *Configuration {
	return s.config
}

// Status returns a snapshot of the current state
func (s *Supervisor) Status() Status {
	s.mu.RLock()
	defer s.mu.RUnlock()

	st := Status{State: s.state.String(), Rules: []RuleInfo{}}
	if s.lastErr != nil {
		st.Error = s.lastErr.Error()
	}
	if s.orch != nil {
		// on StateError the previous rules are still reported: they are the
		// set the daemon is holding on to until a valid reload arrives
		st.Rules = s.orch.Rules()
	}
	return st
}

func (s *Supervisor) setState(state State, err error, orch *Orchestrator) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.state = state
	s.lastErr = err
	if orch != nil {
		s.orch = orch
	}
}

func (s *Supervisor) currentOrchestrator() *Orchestrator {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.orch
}

func (s *Supervisor) isPaused() bool {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.paused
}

// PauseFeeders stops the feeders and keeps them stopped until ResumeFeeders
func (s *Supervisor) PauseFeeders() {
	s.mu.Lock()
	s.paused = true
	s.mu.Unlock()

	if orch := s.currentOrchestrator(); orch != nil {
		orch.StopFeeders()
	}
	s.signalReload()
}

// ResumeFeeders rebuilds the orchestrator and starts the feeders again
func (s *Supervisor) ResumeFeeders() {
	s.mu.Lock()
	s.paused = false
	s.mu.Unlock()

	s.signalReload()
}

func (s *Supervisor) stopping() bool {
	select {
	case <-s.quit:
		return true
	default:
		return false
	}
}

// waitSignal blocks until a reload is requested or the Supervisor is stopped.
// It returns true when the Supervisor must quit.
func (s *Supervisor) waitSignal() bool {
	select {
	case <-s.reload:
		return false
	case <-s.quit:
		return true
	}
}

// signalReload sends a non-blocking reload token. A pending token is enough
// to trigger the next reload, so a second one arriving before it is
// consumed is simply dropped.
func (s *Supervisor) signalReload() {
	select {
	case s.reload <- struct{}{}:
	default: // a reload is already pending
	}
}

// drainReload discards a pending reload token without blocking. Called at
// the top of each Run loop iteration: a token banked by a reload that is
// about to be served by the rebuild we are about to do would otherwise fire
// spuriously the next time this loop parks.
func (s *Supervisor) drainReload() {
	select {
	case <-s.reload:
	default:
	}
}

// waitFeedersOrSignal blocks until every feeder has stopped, a reload is
// requested, or the Supervisor is stopped. It returns true when the
// Supervisor must quit. Waiting on a channel rather than on WaitFeeders
// directly means a feeder that dies without going through StopFeeders can
// never wedge Run: Stop and Reload still get through.
func (s *Supervisor) waitFeedersOrSignal(orch *Orchestrator) bool {
	done := make(chan struct{})
	go func() {
		orch.WaitFeeders()
		close(done)
	}()

	select {
	case <-done:
		return false
	case <-s.reload:
		return false
	case <-s.quit:
		return true
	}
}

// Run drives the orchestrator lifecycle until Stop is called. It blocks.
func (s *Supervisor) Run() error {
	for {
		if s.stopping() {
			s.setState(StateStopped, nil, nil)
			return nil
		}

		if s.isPaused() {
			s.setState(StateStopped, nil, nil)
			if s.waitSignal() {
				s.setState(StateStopped, nil, nil)
				return nil
			}
			continue
		}

		// a token banked by a reload we are already about to serve would
		// otherwise fire spuriously on the next park
		s.drainReload()

		orch, err := NewOrchestrator(s.config)
		if err != nil {
			log.Error("orchestrator: %s", err)
			s.setState(StateError, err, nil)
			if s.waitSignal() {
				s.setState(StateStopped, nil, nil)
				return nil
			}
			continue
		}

		s.setState(StateRunning, nil, orch)
		orch.StartFeeders()

		if orch.HasRunningFeeder() {
			if s.waitFeedersOrSignal(orch) {
				orch.StopFeeders()
				s.setState(StateStopped, nil, nil)
				return nil
			}
		} else {
			// no feeder to wait for: block until a reload or a stop, otherwise
			// this loop would spin at full speed
			log.Info("no running feeder: waiting for a reload")
			if s.waitSignal() {
				orch.StopFeeders()
				s.setState(StateStopped, nil, nil)
				return nil
			}
		}

		orch.StopFeeders()
		if s.stopping() {
			s.setState(StateStopped, nil, nil)
			return nil
		}
		s.setState(StateReloading, nil, orch)
	}
}

// Reload rebuilds the orchestrator picking up the rules currently on disk.
// A no-op while paused: the loop just cycles back through the isPaused
// branch in Run() and re-parks in StateStopped instead of rebuilding, since
// Reload, PauseFeeders and ResumeFeeders all share the same paused flag and
// reload channel and only ResumeFeeders clears the flag.
func (s *Supervisor) Reload() {
	// both actions, always: StopFeeders unblocks a loop waiting on the
	// feeders, the signal unblocks one parked on waitSignal, and which of
	// the two Run is in cannot be known from here without racing it
	if orch := s.currentOrchestrator(); orch != nil {
		orch.StopFeeders()
	}
	s.signalReload()
}

// Stop terminates the Supervisor and its feeders
func (s *Supervisor) Stop() {
	s.quitOnce.Do(func() { close(s.quit) })
	if orch := s.currentOrchestrator(); orch != nil {
		orch.StopFeeders()
	}
}

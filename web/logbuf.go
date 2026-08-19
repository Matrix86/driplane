// Package web serves the driplane embedded web interface and its JSON API.
package web

import (
	"sync"

	"github.com/evilsocket/islazy/log"
)

// LogLine is a single log entry exposed to the UI
type LogLine struct {
	Level   string `json:"level"`
	Message string `json:"message"`
}

// LogRing keeps the last N log lines in memory and fans them out to subscribers
type LogRing struct {
	mu     sync.Mutex
	buf    []LogLine
	size   int
	subs   map[int]chan LogLine
	nextID int
}

// NewLogRing creates a ring buffer holding at most size lines
func NewLogRing(size int) *LogRing {
	if size <= 0 {
		size = 1
	}
	return &LogRing{
		buf:  make([]LogLine, 0, size),
		size: size,
		subs: make(map[int]chan LogLine),
	}
}

// Append stores a line and delivers it to the subscribers. It never blocks:
// lines are dropped for subscribers that are not keeping up.
func (r *LogRing) Append(l LogLine) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buf = append(r.buf, l)
	if len(r.buf) > r.size {
		r.buf = r.buf[len(r.buf)-r.size:]
	}

	for _, ch := range r.subs {
		select {
		case ch <- l:
		default: // subscriber too slow: drop
		}
	}
}

// Backlog returns a copy of the buffered lines, oldest first
func (r *LogRing) Backlog() []LogLine {
	r.mu.Lock()
	defer r.mu.Unlock()

	out := make([]LogLine, len(r.buf))
	copy(out, r.buf)
	return out
}

// Subscribe registers a new consumer and returns its channel plus an
// idempotent cancel function that unregisters and closes it.
func (r *LogRing) Subscribe(buffer int) (<-chan LogLine, func()) {
	if buffer <= 0 {
		buffer = 1
	}

	r.mu.Lock()
	id := r.nextID
	r.nextID++
	ch := make(chan LogLine, buffer)
	r.subs[id] = ch
	r.mu.Unlock()

	var once sync.Once
	cancel := func() {
		once.Do(func() {
			r.mu.Lock()
			delete(r.subs, id)
			r.mu.Unlock()
			close(ch)
		})
	}
	return ch, cancel
}

func (r *LogRing) subscriberCount() int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return len(r.subs)
}

// Attach installs the ring as the islazy log callback, chaining the previous
// one if it was already set. The callback runs while the logger holds its
// global lock, so it must never block: Append is non-blocking by design.
func (r *LogRing) Attach() {
	prev := log.Callback
	log.Callback = func(v log.Verbosity, message string) {
		if prev != nil {
			prev(v, message)
		}
		r.Append(LogLine{Level: log.LevelName(v), Message: message})
	}
}

package core

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"path/filepath"
	"sort"
	"sync"

	"github.com/Matrix86/driplane/feeders"
	"github.com/Matrix86/driplane/filters"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
)

// Orchestrator handles the pipelines and rules
type Orchestrator struct {
	asts    map[string]*AST
	config  *Configuration
	ruleset *Ruleset

	waitFeeder     sync.WaitGroup
	startedFeeders int // how many pending Add(1)s waitFeeder currently owes a Done(); see StartFeeders/StopFeeders
	sync.Mutex

	// stopMu serializes StopFeeders end to end. It is deliberately a SEPARATE
	// mutex from the embedded sync.Mutex above (which Rules()/StartFeeders
	// also take): StopFeeders releases the embedded Mutex while it calls each
	// feeder's Stop() (see StopFeeders), specifically so a feeder whose
	// Stop() blocks cannot also stall /api/status, which takes the embedded
	// Mutex every few seconds via Rules(). But releasing that lock reopened a
	// different hole: two goroutines can call StopFeeders concurrently in
	// the ordinary course of things (Supervisor.Stop() and the tail of
	// Supervisor.Run()'s own loop both call it on the same Orchestrator on
	// an ordinary SIGTERM; a double-click on the UI's Reload button, or two
	// open tabs, produce two concurrent Reload() calls the same way), and
	// without something serializing the whole function, both could snapshot
	// the same running feeder and call Stop() on it twice. Feeder Stop()
	// implementations are not idempotent: most send on an unbuffered
	// stopChan whose only reader returns after the first receive, so a
	// second Stop() blocks forever; twitter's closes a channel a second
	// time, which panics. stopMu closes that window by making sure only one
	// StopFeeders is ever in flight, while leaving the embedded Mutex free
	// for Rules() to take during the (potentially slow) Stop() calls. Do
	// NOT merge these two mutexes back into one.
	stopMu sync.Mutex
}

// NodeInfo describes a single feeder or filter inside a rule.
//
// In means different things depending on Kind: for a filter it is the number
// of messages the filter received (Stats().In); for a feeder there is no
// "received" count, so it is instead filled from Stats().Out, i.e. the
// number of messages the feeder emitted into the pipeline. The UI sums In
// across all nodes of a rule, where this reads sensibly (total traffic
// in/out of the pipeline), but an API client consuming this field directly
// should not assume it always means "received".
type NodeInfo struct {
	Name    string `json:"name"`
	Kind    string `json:"kind"` // "feeder" or "filter"
	In      uint64 `json:"in"`
	Matched uint64 `json:"matched"`
	Errors  uint64 `json:"errors"`
}

// RuleInfo describes a compiled rule and its runtime state
type RuleInfo struct {
	Name      string     `json:"name"`
	File      string     `json:"file"`
	HasFeeder bool       `json:"has_feeder"`
	Running   bool       `json:"running"`
	Nodes     []NodeInfo `json:"nodes"`
}

// NewOrchestrator create a new instance of the Orchestrator
func NewOrchestrator(config *Configuration) (*Orchestrator, error) {
	o := &Orchestrator{
		config:  config,
		asts:    make(map[string]*AST),
		ruleset: NewRuleset(),
	}

	parser, err := NewParser()
	if err != nil {
		return nil, fmt.Errorf("parser creation: %s", err)
	}

	err = fs.Glob(config.Get("general.rules_path"), "*.rule", func(file string) error {
		abs, err := filepath.Abs(file)
		if err != nil {
			return fmt.Errorf("cannot get absolute path of %s: %s", file, err)
		}
		file = abs
		log.Info("parsing rule file: %s", file)
		ast, err := parser.ParseFile(file)
		if err != nil {
			return fmt.Errorf("rule parsing: file '%s': %s", file, err)
		}
		o.asts[file] = ast

		if _, err := o.ruleset.CompileAst(file, ast, o.config); err != nil {
			return fmt.Errorf("compilation of '%s': %s", file, err)
		}
		return nil
	})
	if err != nil {
		return nil, fmt.Errorf("%s", err)
	}
	return o, nil
}

// Rules returns a snapshot of the compiled rules and their runtime state
func (o *Orchestrator) Rules() []RuleInfo {
	o.Lock()
	defer o.Unlock()

	rules := o.ruleset.Rules()
	out := make([]RuleInfo, 0, len(rules))
	for _, r := range rules {
		info := RuleInfo{
			Name:      r.Name,
			File:      r.file,
			HasFeeder: r.HasFeeder,
			Nodes:     make([]NodeInfo, 0, len(r.nodes)),
		}
		for _, n := range r.Nodes() {
			switch node := n.(type) {
			case feeders.Feeder:
				info.Running = node.IsRunning()
				info.Nodes = append(info.Nodes, NodeInfo{
					Name: node.Name(),
					Kind: "feeder",
					In:   node.Stats().Out,
				})
			case filters.Filter:
				st := node.Stats()
				info.Nodes = append(info.Nodes, NodeInfo{
					Name:    node.Name(),
					Kind:    "filter",
					In:      st.In,
					Matched: st.Matched,
					Errors:  st.Errors,
				})
			}
		}
		out = append(out, info)
	}
	sort.Slice(out, func(i, j int) bool {
		if out[i].File == out[j].File {
			return out[i].Name < out[j].Name
		}
		return out[i].File < out[j].File
	})
	return out
}

// StartFeeders opens the gates
func (o *Orchestrator) StartFeeders() {
	o.Lock()
	defer o.Unlock()
	rs := o.ruleset
	for _, rulename := range rs.feedRules {
		f := rs.rules[rulename].getFirstNode().(feeders.Feeder)
		if f.IsRunning() == false {
			log.Debug("[%s] Starting %s", rulename, f.Name())
			o.waitFeeder.Add(1)
			o.startedFeeders++
			f.Start()
		}
	}
}

// HasRunningFeeder return true if one or more feeders are running
func (o *Orchestrator) HasRunningFeeder() bool {
	rs := o.ruleset
	for _, rulename := range rs.feedRules {
		f := rs.rules[rulename].getFirstNode().(feeders.Feeder)
		if f.IsRunning() {
			return true
		}
	}
	return false
}

// WaitFeeders waits until all the feeders are stopped
func (o *Orchestrator) WaitFeeders() {
	log.Debug("Waiting")
	o.waitFeeder.Wait()
	log.Debug("Stop waiting")
}

// StopFeeders closes the gates
func (o *Orchestrator) StopFeeders() {
	// Serializes the whole function against a second concurrent StopFeeders
	// call: see the stopMu field comment for why this cannot just be the
	// embedded Mutex.
	o.stopMu.Lock()
	defer o.stopMu.Unlock()

	o.Lock()
	rs := o.ruleset
	type running struct {
		rulename string
		feeder   feeders.Feeder
	}
	var toStop []running
	for _, rulename := range rs.feedRules {
		f := rs.rules[rulename].getFirstNode().(feeders.Feeder)
		if f.IsRunning() {
			toStop = append(toStop, running{rulename, f})
		}
	}
	o.Unlock()

	// Stop() is called with the lock released: several feeders' Stop()
	// pushes to a channel and can block, and holding the lock across that
	// would stall every other lock holder -- including /api/status, which
	// takes this same lock every few seconds to serve Rules().
	for _, r := range toStop {
		log.Debug("[%s] Stopping %s", r.rulename, r.feeder.Name())
		r.feeder.Stop()
		log.Debug("[%s] Stopped %s", r.rulename, r.feeder.Name())
	}

	o.Lock()
	defer o.Unlock()

	// Release exactly as many Done()s as StartFeeders banked Add(1)s for,
	// regardless of what IsRunning() reports at this point: several feeders
	// clear isRunning from their own goroutine (feeders/folder.go,
	// feeders/telegram.go, feeders/twitter.go), so gating Done() on a fresh
	// IsRunning() read would leave waitFeeder permanently above zero for any
	// feeder that had already stopped itself before this call -- the
	// WaitGroup imbalance that used to wedge WaitFeeders() forever. The
	// counter is reset to 0 so a following StartFeeders/StopFeeders cycle
	// stays balanced.
	if o.startedFeeders > 0 {
		o.waitFeeder.Add(-o.startedFeeders)
		o.startedFeeders = 0
	}

	// sending a shutdown event on the bus. This stays under the lock (unlike
	// the Stop() calls above): rs.bus.WaitAsync() waits on the EventBus's own
	// internal WaitGroup, which is not safe to drive from two StopFeeders
	// calls running concurrently -- e.g. Supervisor.Stop() and the tail end
	// of Supervisor.Run()'s loop can both call StopFeeders on the same
	// Orchestrator at nearly the same time. Serializing this part the same
	// way the whole function used to be serialized avoids that race.
	rs.bus.Publish(data.EventTopicName, &data.Event{Type: "shutdown"})
	rs.bus.WaitAsync()
}

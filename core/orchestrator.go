package core

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"path/filepath"
	"sync"

	"github.com/Matrix86/driplane/feeders"

	"github.com/evilsocket/islazy/fs"
	"github.com/evilsocket/islazy/log"
)

// Orchestrator handles the pipelines and rules
type Orchestrator struct {
	asts   map[string]*AST
	config *Configuration

	waitFeeder sync.WaitGroup
	sync.Mutex
}

// NewOrchestrator create a new instance of the Orchestrator
func NewOrchestrator(config *Configuration) (Orchestrator, error) {
	o := Orchestrator{
		config: config,
		asts:   make(map[string]*AST),
	}

	parser, _ := NewParser()

	err := fs.Glob(config.Get("general.rules_path"), "*.rule", func(file string) error {
		abs, err := filepath.Abs(file)
		if err != nil {
			log.Fatal("cannot get absolute path of %s: %s", file, err)
		}
		file = abs
		log.Info("parsing rule file: %s", file)
		ast, err := parser.ParseFile(file)
		if err != nil {
			log.Fatal("rule parsing: file '%s': %s", file, err)
		}
		o.asts[file] = ast

		_, err = RuleSetInstance().CompileAst(file, ast, o.config)
		if err != nil {
			return fmt.Errorf("compilation of '%s': %s", file, err)
		}
		return nil
	})
	if err != nil {
		return o, fmt.Errorf("%s", err)
	}
	return o, nil
}

// StartFeeders opens the gates
func (o *Orchestrator) StartFeeders() {
	o.Lock()
	defer o.Unlock()
	rs := RuleSetInstance()
	for _, rulename := range rs.feedRules {
		f := rs.rules[rulename].getFirstNode().(feeders.Feeder)
		if f.IsRunning() == false {
			log.Debug("[%s] Starting %s", rulename, f.Name())
			o.waitFeeder.Add(1)
			f.Start()
		}
	}
}

// HasRunningFeeder return true if one or more feeders are running
func (o *Orchestrator) HasRunningFeeder() bool {
	rs := RuleSetInstance()
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
	o.Lock()
	defer o.Unlock()

	rs := RuleSetInstance()
	for _, rulename := range rs.feedRules {
		f := rs.rules[rulename].getFirstNode().(feeders.Feeder)
		if f.IsRunning() {
			log.Debug("[%s] Stopping %s", rulename, f.Name())
			f.Stop()
			o.waitFeeder.Done()
			log.Debug("[%s] Stopped %s", rulename, f.Name())
		}
	}

	// sending a shutdown event on the bus
	rs.bus.Publish(data.EventTopicName, &data.Event{Type: "shutdown"})
	rs.bus.WaitAsync()
}

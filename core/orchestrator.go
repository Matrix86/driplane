package core

import (
	"fmt"
	"github.com/Matrix86/driplane/feeders"
	"github.com/evilsocket/islazy/log"
	"sync"
)

// Orchestrator handles the pipelines and rules
type Orchestrator struct {
	asts   map[string]*AST
	config *Configuration

	waitFeeder sync.WaitGroup
	sync.Mutex
}

// NewOrchestrator create a new instance of the Orchestrator
func NewOrchestrator(asts map[string]*AST, config *Configuration) (Orchestrator, error) {
	o := Orchestrator{}

	o.asts = asts
	o.config = config

	for file, ast := range asts {
		for _, rn := range ast.Rules {
			//pp.Println(rn)
			err := RuleSetInstance().AddRule(rn, o.config)
			if err != nil {
				err = fmt.Errorf("file '%s': %s", file, err)
				return o, err
			}
		}
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
}

package core

import (
	"fmt"
	"github.com/Matrix86/driplane/feeders"
	"github.com/evilsocket/islazy/log"
	"sync"
)

type Orchestrator struct {
	asts   map[string]*AST
	config *Configuration

	waitFeeder sync.WaitGroup
	sync.Mutex
}

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

func (o *Orchestrator) WaitFeeders() {
	log.Debug("Waiting")
	o.waitFeeder.Wait()
	log.Debug("Stop waiting")
}

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

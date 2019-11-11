package core

import (
	"fmt"
	"strconv"
	"sync"

	"github.com/Matrix86/driplane/com"
	"github.com/Matrix86/driplane/feeder"
	"github.com/Matrix86/driplane/filter"

	"github.com/evilsocket/islazy/log"
)

var (
	instance *Ruleset
	once     sync.Once
)

type Ruleset struct {
	rules map[string]*PipeRule

	feedRules []string
}

func RuleSetInstance() *Ruleset {
	once.Do(func() {
		instance = &Ruleset{
			rules: make(map[string]*PipeRule),
		}
	})
	return instance
}

func (r *Ruleset) AddRule(node *RuleNode, config Configuration) error {
	if node == nil || node.Identifier == "" {
		return fmt.Errorf("Ruleset.AddRule: rules without name are not supported")
	}

	if _, ok := r.rules[node.Identifier]; ok {
		return fmt.Errorf("Ruleset.AddRule: rule '%s' redefined previously", node.Identifier)
	}

	pr, err := NewPipeRule(node, config)
	if err != nil {
		return err
	}

	log.Info("Added %s to rules", pr.Name)
	r.rules[pr.Name] = pr
	if pr.HasFeeder {
		r.feedRules = append(r.feedRules, node.Identifier)
	}

	return nil
}

type PipeRule struct {
	Name      string
	HasFeeder bool

	nodes []com.Subscriber
}

func (p *PipeRule) getLastNode() *com.Subscriber {
	if len(p.nodes) == 0 {
		return nil
	}
	return &p.nodes[len(p.nodes)-1]
}

func (p *PipeRule) getFirstNode() *com.Subscriber {
	if len(p.nodes) == 0 {
		return nil
	}
	return &p.nodes[0]
}

func (p *PipeRule) newFilter(fn *FilterNode) (filter.Filter, error) {
	params := make(map[string]string)
	for _, par := range fn.Params {
		value := ""
		if par.Value.Number != nil {
			value = strconv.FormatFloat(*par.Value.Number, 'E', -1, 64)
		} else {
			value = *par.Value.String
		}
		params[par.Name] = value
	}

	filter, err := filter.NewFilter(fn.Name+"filter", params)
	if err != nil {
		return nil, err
	}

	return filter, nil
}

func (p *PipeRule) getRuleCall(node *RuleCall) (*PipeRule, error) {
	if foundrule, ok := RuleSetInstance().rules[node.Name]; ok {
		return foundrule, nil
	}
	return nil, fmt.Errorf("rule '%s' not found...you need to define it", node.Name)
}

func (p *PipeRule) addNode(node *Node, prev com.Subscriber) error {
	if node == nil {
		return nil
	}

	if node.Filter != nil {
		log.Info("['%s'] new filter found '%s'", p.Name, node.Filter.Name)

		f, err := p.newFilter(node.Filter)
		if err != nil {
			return err
		}

		if prev != nil {
			prev.SetEventMessageClb(func(msg com.DataMessage) {
				if b, _ := f.DoFilter(&msg); b {
					log.Debug("[%s] filter %s match", p.Name, node.Filter.Name)
					f.(com.Subscriber).Propagate(msg)
				}
			})
		}

		p.nodes = append(p.nodes, f.(com.Subscriber))

		return p.addNode(node.Filter.Next, f.(com.Subscriber))
	} else if node.RuleCall != nil {
		log.Info("['%s'] new rulecall found '%s'", p.Name, node.RuleCall.Name)
		var last com.Subscriber
		var err error

		r, err := p.getRuleCall(node.RuleCall)
		if err != nil {
			return err
		}

		if prev != nil {
			if r.HasFeeder {
				return fmt.Errorf("rule '%s' contains a feeder and cannot be here", node.RuleCall.Name)
			}

			first := *r.getFirstNode()
			prev.SetEventMessageClb(func(msg com.DataMessage) {
				if b, _ := first.(filter.Filter).DoFilter(&msg); b {
					log.Debug("[%s] filter %s match", p.Name, node.RuleCall.Name)
					first.Propagate(msg)
				}
			})
		}

		last = *r.getLastNode()

		return p.addNode(node.RuleCall.Next, last)
	}

	return nil
}

func NewPipeRule(node *RuleNode, config Configuration) (*PipeRule, error) {
	rule := &PipeRule{}
	rule.Name = node.Identifier

	log.Info("Rule '%s' found", rule.Name)

	var next *Node
	// The Rule has a feeder specified
	if node.Feeder != nil {
		log.Info("['%s'] new feeder found '%s'", rule.Name, node.Feeder.Name)

		// configuration override from the rule itself
		config := config.GetConfig()
		for _, par := range node.Feeder.Params {
			value := ""
			if par.Value.Number != nil {
				value = strconv.FormatFloat(*par.Value.Number, 'E', -1, 64)
			} else {
				value = *par.Value.String
			}
			config[node.Feeder.Name + "." + par.Name] = value
		}

		f, err := feeder.NewFeeder(node.Feeder.Name+"feeder", config)
		if err != nil {
			log.Error("piperule.NewRule: %s", err)
			return nil, err
		}

		rule.HasFeeder = true
		rule.nodes = append(rule.nodes, f.(com.Subscriber))
		next = node.Feeder.Next

		if err := rule.addNode(next, f.(com.Subscriber)); err != nil {
			return nil, err
		}
	} else { // It doesn't start with a feeder
		if err := rule.addNode(node.First, nil); err != nil {
			return nil, err
		}
	}

	return rule, nil
}

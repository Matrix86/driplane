package core

import (
	"fmt"
	"strconv"
	"strings"
	"sync"

	"github.com/Matrix86/driplane/feeders"
	"github.com/Matrix86/driplane/filters"

	bus "github.com/asaskevich/EventBus"
	"github.com/evilsocket/islazy/log"
)

var (
	instance *Ruleset
	once     sync.Once
)

// Ruleset identifies a set of rules
type Ruleset struct {
	rules map[string]*PipeRule

	feedRules []string
	bus       bus.Bus
	lastID    int32
}

// RuleSetInstance is the singleton for the Ruleset object
func RuleSetInstance() *Ruleset {
	once.Do(func() {
		instance = &Ruleset{
			rules:  make(map[string]*PipeRule),
			bus:    bus.New(),
			lastID: 0,
		}
	})
	return instance
}

// AddRule appends a new rule to the set
func (r *Ruleset) AddRule(node *RuleNode, config *Configuration) error {
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

// INode for the nodes generalization
type INode interface{}

// PipeRule identifies a single rule
type PipeRule struct {
	Name      string
	HasFeeder bool

	config *Configuration
	nodes  []INode
}

func (p *PipeRule) getLastNode() INode {
	if len(p.nodes) == 0 {
		return nil
	}
	return p.nodes[len(p.nodes)-1]
}

func (p *PipeRule) getFirstNode() INode {
	if len(p.nodes) == 0 {
		return nil
	}
	return p.nodes[0]
}

/*
// configuration override from the rule itself
		config := config.GetConfig()
		for _, par := range node.Feeder.Params {
			value := ""
			if par.Value.Number != nil {
				value = strconv.FormatFloat(*par.Value.Number, 'E', -1, 64)
			} else {
				value = *par.Value.String
			}
			config[node.Feeder.Name+"."+par.Name] = value
		}
*/

func (p *PipeRule) newFilter(fn *FilterNode) (filters.Filter, error) {
	params := make(map[string]string)
	config := p.config.GetConfig()
	// The filter will receive only his configuration and general config in the parameters
	prefix := strings.ToLower(fn.Name+".")
	for k, v := range config {
		if strings.HasPrefix(k, prefix) {
			params[strings.TrimPrefix(k, prefix)] = v
		} else if strings.HasPrefix(k, "general.") || strings.HasPrefix(k, "custom.") {
			params[k] = v
		}
	}

	// configurations will be overrided by the parameters defined in the rule file
	for _, par := range fn.Params {
		value := ""
		if par.Value.Number != nil {
			value = strconv.FormatFloat(*par.Value.Number, 'E', -1, 64)
		} else {
			value = *par.Value.String
		}
		params[par.Name] = value
	}

	rs := RuleSetInstance()
	f, err := filters.NewFilter(p.Name, fn.Name+"filter", params, rs.bus, rs.lastID+1, fn.Neg)
	if err != nil {
		return nil, err
	}
	rs.lastID++

	return f, nil
}

func (p *PipeRule) getRuleCall(node *RuleCall) (*PipeRule, error) {
	if foundrule, ok := RuleSetInstance().rules[node.Name]; ok {
		return foundrule, nil
	}
	return nil, fmt.Errorf("rule '%s' not found...you need to define it", node.Name)
}

func (p *PipeRule) addNode(node *Node, prev string) error {
	if node == nil {
		return nil
	}

	rs := RuleSetInstance()
	if node.Filter != nil {
		log.Info("['%s'] new filter found '%s'", p.Name, node.Filter.Name)

		f, err := p.newFilter(node.Filter)
		if err != nil {
			return err
		}

		if prev != "" {
			err := rs.bus.SubscribeAsync(prev, f.Pipe, false)
			if err != nil {
				return err
			}
		}

		p.nodes = append(p.nodes, f)

		return p.addNode(node.Filter.Next, f.GetIdentifier())
	} else if node.RuleCall != nil {
		log.Info("['%s'] new rulecall found '%s'", p.Name, node.RuleCall.Name)
		var err error

		r, err := p.getRuleCall(node.RuleCall)
		if err != nil {
			return err
		}

		if prev != "" {
			if r.HasFeeder {
				return fmt.Errorf("rule '%s' contains a feeder and cannot be here", node.RuleCall.Name)
			}

			node := r.getFirstNode()
			err := rs.bus.SubscribeAsync(prev, node.(filters.Filter).Pipe, false)
			if err != nil {
				return err
			}
		}

		// This is a filter for sure!
		last := r.getLastNode()
		if _, ok := last.(filters.Filter); ok {
			return p.addNode(node.RuleCall.Next, last.(filters.Filter).GetIdentifier())
		} else if _, ok := last.(feeders.Feeder); ok {
			return p.addNode(node.RuleCall.Next, last.(feeders.Feeder).GetIdentifier())
		} else {
			return fmt.Errorf("found an unknown node type")
		}
	}

	return nil
}

// NewPipeRule creates and returns a PipeRule struct
func NewPipeRule(node *RuleNode, config *Configuration) (*PipeRule, error) {
	rule := &PipeRule{}
	rule.Name = node.Identifier
	rule.config = config

	log.Info("Rule '%s' found", rule.Name)

	var next *Node
	// The Rule has a feeder specified
	if node.Feeder != nil {
		log.Info("['%s'] new feeder found '%s'", rule.Name, node.Feeder.Name)

		// configuration override from the rule itself
		params := make(map[string]string)

		config := config.GetConfig()
		// The filter will receive only his configuration and general config in the parameters
		prefix := strings.ToLower(node.Feeder.Name+".")
		for k, v := range config {
			if strings.HasPrefix(k, prefix) {
				params[k] = v
			} else if strings.HasPrefix(k, "general.") || strings.HasPrefix(k, "custom."){
				params[k] = v
			}
		}

		// Feeder params in the rule will overwrite that ones specified in the config file
		for _, par := range node.Feeder.Params {
			value := ""
			if par.Value.Number != nil {
				value = strconv.FormatFloat(*par.Value.Number, 'E', -1, 64)
			} else {
				value = *par.Value.String
			}
			params[node.Feeder.Name+"."+par.Name] = value
		}

		rs := RuleSetInstance()
		f, err := feeders.NewFeeder(node.Feeder.Name+"feeder", params, rs.bus, rs.lastID+1)
		if err != nil {
			log.Error("piperule.NewRule: %s", err)
			return nil, err
		}
		rs.lastID++

		rule.HasFeeder = true
		rule.nodes = append(rule.nodes, f)
		next = node.Feeder.Next

		if err := rule.addNode(next, f.GetIdentifier()); err != nil {
			return nil, err
		}
	} else { // It doesn't start with a feeder
		if err := rule.addNode(node.First, ""); err != nil {
			return nil, err
		}
	}

	return rule, nil
}

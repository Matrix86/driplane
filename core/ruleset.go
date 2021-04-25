package core

import (
	"fmt"
	"strings"
	"sync"

	bus "github.com/asaskevich/EventBus"
	"github.com/evilsocket/islazy/log"
)

var (
	instance *Ruleset
	once     sync.Once
)

// Ruleset identifies a set of rules
type Ruleset struct {
	rules        map[string]*PipeRule
	compiledDeps map[string][]string

	feedRules []string
	bus       bus.Bus
	lastID    int32
}

// RuleSetInstance is the singleton for the Ruleset object
func RuleSetInstance() *Ruleset {
	once.Do(func() {
		instance = &Ruleset{
			rules:        make(map[string]*PipeRule),
			compiledDeps: make(map[string][]string),
			bus:          bus.New(),
			lastID:       0,
		}
	})
	return instance
}

// CompileAst compiles the AST
func (r *Ruleset) CompileAst(filename string, ast *AST, config *Configuration) ([]string, error) {
	// file has been already compiled
	if c, ok := r.compiledDeps[filename]; ok {
		return c, nil
	}

	deps := make([]string, 0)
	// compile dependencies first
	for f, d := range ast.Dependencies {
		childDeps, err := r.CompileAst(f, d, config)
		if err != nil {
			return nil, fmt.Errorf("compile file '%s': %s", f, err)
		}
		deps = append(deps, f)
		deps = append(deps, childDeps...)
	}
	// track the compiled files and its dependencies
	r.compiledDeps[filename] = deps

	for _, rn := range ast.Rules {
		//pp.Println(rn)
		err := r.AddRule(filename, rn, config, deps)
		if err != nil {
			err = fmt.Errorf("adding rule '%s': %s", rn.Identifier, err)
			return nil, err
		}
	}

	return deps, nil
}

// AddRule appends a new rule to the set
func (r *Ruleset) AddRule(filename string, node *RuleNode, config *Configuration, deps []string) error {
	if node == nil || node.Identifier == "" {
		return fmt.Errorf("Ruleset.AddRule: rules without name are not supported")
	}

	// Prepend the filename to the node identifier
	name := strings.Join([]string{filename, node.Identifier}, ":")
	if _, ok := r.rules[name]; ok {
		return fmt.Errorf("Ruleset.AddRule: rule '%s' redefined previously", node.Identifier)
	}

	pr, err := NewPipeRule(node, config, filename, deps)
	if err != nil {
		return err
	}

	log.Debug("Added @%s to rules", pr.Name)
	r.rules[pr.GetIdentifier()] = pr
	if pr.HasFeeder {
		r.feedRules = append(r.feedRules, pr.GetIdentifier())
	}

	return nil
}

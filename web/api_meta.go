package web

import (
	"fmt"
	"net/http"

	"github.com/Matrix86/driplane/core"
	"github.com/Matrix86/driplane/feeders"
	"github.com/Matrix86/driplane/filters"
)

type validateRequest struct {
	Kind   string `json:"kind"`
	Source string `json:"source"`
}

type validateResponse struct {
	OK    bool   `json:"ok"`
	Error string `json:"error,omitempty"`
}

type metaResponse struct {
	Filters []string `json:"filters"`
	Feeders []string `json:"feeders"`
	// Kinds lists the file kinds this store actually has a configured
	// directory for -- general.templates_path and general.js_path are
	// optional, so this can be just ["rules"]. The UI builds its file-tree
	// sections from this instead of a hardcoded list of all three.
	Kinds []Kind `json:"kinds"`
}

func (s *Server) handleMeta(w http.ResponseWriter, r *http.Request) {
	writeJSON(w, http.StatusOK, metaResponse{
		Filters: filters.RegisteredNames(),
		Feeders: feeders.RegisteredNames(),
		Kinds:   s.opts.Store.Kinds(),
	})
}

func (s *Server) handleValidate(w http.ResponseWriter, r *http.Request) {
	var req validateRequest
	if err := decodeBody(r, &req); err != nil {
		writeError(w, http.StatusBadRequest, err)
		return
	}
	// "rules" is the only kind the editor validates; an empty kind defaults
	// to it. Anything else is a malformed request, not an invalid rule.
	if req.Kind != "" && req.Kind != "rules" {
		writeError(w, http.StatusBadRequest, fmt.Errorf("unsupported kind '%s'", req.Kind))
		return
	}

	// A Store.Root failure here means the server's own configuration is
	// broken (no rules directory configured), not that the caller sent a
	// bad request.
	root, err := s.opts.Store.Root(KindRules)
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	parser, err := core.NewParser()
	if err != nil {
		writeError(w, http.StatusInternalServerError, err)
		return
	}

	// ParseBytes confines #import resolution to root and bounds chain
	// length, so a hostile rule body can only ever produce an ok:false
	// response here, never crash the process or read files outside root.
	ast, err := parser.ParseBytes([]byte(req.Source), root)
	if err != nil {
		writeJSON(w, http.StatusOK, validateResponse{OK: false, Error: err.Error()})
		return
	}

	if err := checkKnownNodes(ast); err != nil {
		writeJSON(w, http.StatusOK, validateResponse{OK: false, Error: err.Error()})
		return
	}

	writeJSON(w, http.StatusOK, validateResponse{OK: true})
}

// checkKnownNodes walks the AST -- and, recursively, every AST reachable
// through its Dependencies (i.e. #import'd files) -- and reports the first
// unknown feeder, unknown filter, or rule call with no matching target. The
// parser only checks the grammar, so a typo in a filter or feeder name, or a
// call to a rule that does not exist, would otherwise surface only at
// compile time.
func checkKnownNodes(ast *core.AST) error {
	known := collectRuleNames(ast, make(map[string]bool), make(map[*core.AST]bool))
	return checkASTNodes(ast, known, make(map[*core.AST]bool))
}

// collectRuleNames gathers every rule identifier defined in ast and in every
// AST it (transitively) imports. getRuleCall resolves a rule call against
// the current file's dependencies first, then the current file itself, so a
// rule call target is valid if its name is defined anywhere in this set --
// mirroring that resolution without needing to reproduce its per-file
// priority order, since we only need to know whether the target exists at
// all.
func collectRuleNames(ast *core.AST, names map[string]bool, visited map[*core.AST]bool) map[string]bool {
	if ast == nil || visited[ast] {
		return names
	}
	visited[ast] = true

	for _, rule := range ast.Rules {
		names[rule.Identifier] = true
	}
	for _, dep := range ast.Dependencies {
		collectRuleNames(dep, names, visited)
	}
	return names
}

// checkASTNodes walks the rules defined directly in ast, then recurses into
// every dependency AST so a bad filter/feeder/rule-call name inside an
// imported file is reported too, not just in the top-level buffer.
func checkASTNodes(ast *core.AST, known map[string]bool, visited map[*core.AST]bool) error {
	if ast == nil || visited[ast] {
		return nil
	}
	visited[ast] = true

	for _, rule := range ast.Rules {
		if rule.Feeder != nil {
			if !feeders.Exists(rule.Feeder.Name) {
				return fmt.Errorf("rule '%s': unknown feeder '%s'", rule.Identifier, rule.Feeder.Name)
			}
			if err := checkNode(rule.Identifier, rule.Feeder.Next, known); err != nil {
				return err
			}
			continue
		}
		if err := checkNode(rule.Identifier, rule.First, known); err != nil {
			return err
		}
	}

	for _, dep := range ast.Dependencies {
		if err := checkASTNodes(dep, known, visited); err != nil {
			return err
		}
	}
	return nil
}

func checkNode(ruleName string, node *core.Node, known map[string]bool) error {
	for node != nil {
		switch {
		case node.Filter != nil:
			if !filters.Exists(node.Filter.Name) {
				return fmt.Errorf("rule '%s': unknown filter '%s'", ruleName, node.Filter.Name)
			}
			node = node.Filter.Next
		case node.RuleCall != nil:
			if !known[node.RuleCall.Name] {
				return fmt.Errorf("rule '%s': unknown rule call '@%s'", ruleName, node.RuleCall.Name)
			}
			node = node.RuleCall.Next
		default:
			return nil
		}
	}
	return nil
}

package filters

import (
	"fmt"
	"path/filepath"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/plugin"
	"github.com/robertkrimen/otto"
)

// Js is a Filter that load a plugin written in Javascript to create a custom Filter
type Js struct {
	Base

	filepath string
	function string

	p *plugin.Plugin

	params map[string]string
}

// NewJsFilter is the registered method to instantiate a JsFilter
func NewJsFilter(p map[string]string) (Filter, error) {
	f := &Js{
		params:   p,
		function: "DoFilter",
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["path"]; ok {
		f.filepath = v
	}

	// If function is not defined we call DoFilter
	if v, ok := p["function"]; ok {
		f.function = v
	}

	var err error

	// If the path specified is relative, we're resolving it with 'rules_path' config
	if !filepath.IsAbs(f.filepath) {
		if v, ok := p["general.js_path"]; !ok {
			r := ""
			if r, ok = p["general.rules_path"]; !ok {
				return nil, fmt.Errorf("NewJsFilter: rules_path or js_path configs not found")
			}
			f.filepath = filepath.Join(r, f.filepath)

		} else {
			f.filepath = filepath.Join(v, f.filepath)
		}
	}

	// load the plugin
	f.p, err = plugin.Load(f.filepath)
	if err != nil {
		return nil, fmt.Errorf("JsFilter '%s': %s", f.filepath, err)
	}

	// Check if the JS plugin contains the DoFilter method
	if !f.p.HasFunc(f.function) {
		return nil, fmt.Errorf("NewJsFilter: %s doesn't contain the %s function", f.filepath, f.function)
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Js) DoFilter(msg *data.Message) (bool, error) {
	triggered := false
	text := msg.GetMessage()
	extra := msg.GetExtra()

	res, err := f.p.Call(f.function, text, extra, f.params)
	if err != nil {
		err := err.(*otto.Error)
		return false, fmt.Errorf("DoFilter: file '%s': '%s'", f.p.Path, err.String())
	}

	if res != nil {
		result, ok := res.(map[string]interface{})
		if ok {
			if v, ok := result["filtered"]; ok {
				if t, ok := v.(bool); ok {
					triggered = t
				} else if t, ok := v.(string); ok {
					triggered = t == "true"
				}

				if triggered {
					if v, ok := result["data"]; ok {
						switch t := v.(type) {
						case string:
							msg.SetMessage(t)
							msg.SetExtra("fulltext", text)

						case map[string]interface{}:
							for key, vi := range t {
								if value, ok := vi.(string); ok {
									if key == "data" {
										msg.SetMessage(value)
									} else {
										msg.SetExtra(key, value)
									}
								}
							}
						}
					}
				}
			}
		}
	}

	return triggered, nil
}

// Set the name of the filter
func init() {
	register("js", NewJsFilter)
}

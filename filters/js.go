package filters

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/plugins"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/plugin"
)

type Js struct {
	Base

	filepath string
	function string

	plugin *plugin.Plugin

	params map[string]string
}

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

	// Thx @evilsocket for the hint =)
	// https://github.com/evilsocket/shellz/blob/master/plugins/plugin.go#L18
	plugin.Defines = map[string]interface{}{
		"log": func(s string) interface{} {
			log.Info("%s", s)
			return nil
		},
		"http": plugins.GetHTTP(),
		"file": plugins.GetFile(),
	}

	var err error
	// load the plugin
	f.plugin, err = plugin.Load(f.filepath)
	if err != nil {
		return nil, fmt.Errorf("JsFilter '%s': %s", f.filepath, err)
	}

	// Check if the JS plugin contains the DoFilter method
	if !f.plugin.HasFunc(f.function) {
		return nil, fmt.Errorf("NewJsFilter: %s doesn't contain the %s function", f.filepath, f.function)
	}

	return f, nil
}

func (f *Js) DoFilter(msg *data.Message) (bool, error) {
	triggered := false
	text := msg.GetMessage()
	extra := msg.GetExtra()

	res, err := f.plugin.Call(f.function, text, extra, f.params)
	if err != nil {
		return false, fmt.Errorf("js: DoFilter function returned '%s'", err)
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
						if s, ok := v.(string); ok {
							msg.SetMessage(s)
							msg.SetExtra("fulltext", text)
						}
						if msi, ok := v.(map[string]interface{}); ok {
							for key, vi := range msi {
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

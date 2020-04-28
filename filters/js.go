package filters

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"

	"github.com/evilsocket/islazy/plugin"
)

type Js struct {
	Base

	filepath string

	plugin *plugin.Plugin

	params map[string]string
}

func NewJsFilter(p map[string]string) (Filter, error) {
	f := &Js{
		params: p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["path"]; ok {
		f.filepath = v
	}

	var err error
	// load the plugin
	f.plugin, err = plugin.Load(f.filepath)
	if err != nil {
		return nil, fmt.Errorf("NewJsFilter: %s", err)
	}

	// Check if the JS plugin contains the DoFilter method
	if !f.plugin.HasFunc("DoFilter") {
		return nil, fmt.Errorf("NewJsFilter: %s doesn't contain the DoFilter function", f.filepath)
	}

	return f, nil
}

func (f *Js) DoFilter(msg *data.Message) (bool, error) {
	triggered := false
	text := msg.GetMessage()

	res, err := f.plugin.Call("DoFilter", text)
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

	plugin.Defines = map[string]interface{}{
		"log": func(s string) interface{} {
			log.Info("%s", s)
			return nil
		},
	}
}

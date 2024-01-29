package filters

import (
	"errors"
	"fmt"
	"strings"

	"github.com/Matrix86/driplane/data"

	"github.com/antchfx/jsonquery"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/str"
)

// JSON is a filter to parse the JSON format
type JSON struct {
	Base

	selector string
	target   string

	params map[string]string
}

// NewJSONFilter is the registered method to instantiate a JSONFilter
func NewJSONFilter(p map[string]string) (Filter, error) {
	f := &JSON{
		params:   p,
		target:   "main",
		selector: "",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["selector"]; ok {
		f.selector = v
	}

	if f.selector == "" {
		return nil, errors.New("no selector specified for JSON filter")
	}
	if v, ok := f.params["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *JSON) DoFilter(msg *data.Message) (bool, error) {
	//var err error
	var text string

	if v, ok := msg.GetTarget(f.target).(string); ok {
		text = str.Trim(v)
	} else if v, ok := msg.GetTarget(f.target).([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}

	if len(text) > 0 {
		var jsonData string

		if text[0] == '{' {
			// json text
			jsonData = str.Trim(text)
		} else {
			log.Error("'%v' is not a json document", text)
			return false, nil
		}

		if doc, err := jsonquery.Parse(strings.NewReader(jsonData)); err == nil {
			atLeastOne := false
			for _, node := range jsonquery.Find(doc, f.selector) {
				atLeastOne = true
				clone := msg.Clone()
				clone.SetMessage(node.Value())
				f.Propagate(clone)
			}

			return atLeastOne, nil

		} else {
			log.Debug("'%v' could not be parsed as JSON: %v", text, err)
			return false, nil
		}

	}

	return false, nil
}

// OnEvent is called when an event occurs
func (f *JSON) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("json", NewJSONFilter)
}

package filters

import (
	"fmt"
	"github.com/Matrix86/driplane/utils"

	"github.com/Matrix86/driplane/data"
)

// StripTag is a filter that removes HTML tags from a text.
type StripTag struct {
	Base

	target string

	params map[string]string
}

// NewStripTagFilter is the registered method to instantiate a StripTagFilter
func NewStripTagFilter(p map[string]string) (Filter, error) {
	f := &StripTag{
		params: p,
		target: "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *StripTag) DoFilter(msg *data.Message) (bool, error) {
	var text string
	if v, ok := msg.GetTarget(f.target).(string); ok {
		text = v
	} else if v, ok := msg.GetTarget(f.target).([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}
	stripped := utils.ExtractTextFromHTML(text)
	msg.SetExtra("fulltext", text)
	msg.SetMessage(stripped)
	return true, nil
}

// OnEvent is called when an event occurs
func (f *StripTag) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("striptag", NewStripTagFilter)
}

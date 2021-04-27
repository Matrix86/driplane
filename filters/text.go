package filters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Matrix86/driplane/data"
)

// Text is a Filter to search and extract strings from the input Message
type Text struct {
	Base

	regexp *regexp.Regexp

	extract      bool
	pattern      string
	enableRegexp bool
	target       string

	params map[string]string
}

// NewTextFilter is the registered method to instantiate a TextFilter
func NewTextFilter(p map[string]string) (Filter, error) {
	var err error
	f := &Text{
		params:       p,
		regexp:       nil,
		extract:      false,
		pattern:      "",
		enableRegexp: false,
		target:       "main",
	}
	f.cbFilter = f.DoFilter

	// mandatory field
	if v, ok := p["pattern"]; ok {
		f.pattern = v
	} else {
		return nil, fmt.Errorf("pattern field is required on textfilter")
	}

	if v, ok := p["regexp"]; ok && v == "true" {
		f.regexp, err = regexp.Compile(f.pattern)
		if err != nil {
			return nil, fmt.Errorf("textfilter: cannot compile the regular expression '%s'", f.pattern)
		}
	}
	if v, ok := f.params["extract"]; ok && v == "true" {
		f.extract = true
	}
	if v, ok := p["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Text) DoFilter(msg *data.Message) (bool, error) {
	var text string
	target := msg.GetTarget(f.target)
	if target == nil {
		return false, nil
	}
	if v, ok := target.(string); ok {
		text = v
	} else if v, ok := target.([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}

	found := false
	if f.regexp != nil {
		if f.extract {
			matched := make([]string, 0)
			match := f.regexp.FindAllStringSubmatch(text, -1)
			if match != nil {
				for _, m := range match {
					matched = append(matched, m[1:]...)
				}
			}
			if len(matched) == 1 {
				msg.SetMessage(matched[0])
				msg.SetExtra("fulltext", text)
				return true, nil
			} else if len(matched) > 1 {
				for _, m := range matched {
					clone := msg.Clone()
					clone.SetMessage(m)
					clone.SetExtra("fulltext", text)
					f.Propagate(clone)
				}
				return false, nil
			}
		} else if f.regexp.MatchString(text) {
			found = true
		}
	} else if f.pattern != "" && strings.Contains(text, f.pattern) {
		found = true
	}

	return found, nil
}

// OnEvent is called when an event occurs
func (f *Text) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("text", NewTextFilter)
}

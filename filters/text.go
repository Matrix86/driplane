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
	text   string

	extractText bool
	target      string

	params map[string]string
}

// NewTextFilter is the registered method to instantiate a TextFilter
func NewTextFilter(p map[string]string) (Filter, error) {
	var err error
	f := &Text{
		params:      p,
		regexp:      nil,
		extractText: false,
		text:        "",
		target: "main",
	}
	f.cbFilter = f.DoFilter

	// Regexp initialization
	if v, ok := p["regexp"]; ok {
		f.regexp, err = regexp.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("textfilter: cannot compile the regular expression in 'regexp' parameter")
		}
	}
	if v, ok := f.params["extract"]; ok && v == "true" {
		f.extractText = true
	}
	if v, ok := p["text"]; ok {
		f.text = v
	}
	if v, ok := p["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Text) DoFilter(msg *data.Message) (bool, error) {
	var text string

	if f.target == "main" {
		text = msg.GetMessage().(string)
	} else if v, ok := msg.GetExtra()[f.target].(string); ok {
		text = v
	} else {
		return false, nil
	}

	found := false
	if f.regexp != nil {
		if f.extractText {
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
	} else if f.text != "" && strings.Contains(text, f.text) {
		found = true
	}

	return found, nil
}

// Set the name of the filter
func init() {
	register("text", NewTextFilter)
}

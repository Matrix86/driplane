package filters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Matrix86/driplane/com"
)

type Text struct {
	Base

	regexp *regexp.Regexp
 	text   string

	extractText bool

	params map[string]string
}

func NewTextFilter(p map[string]string) (Filter, error) {
	f := &Text{
		params: p,
		regexp: nil,
		extractText: false,
		text: "",
	}
	f.cbFilter = f.DoFilter

	// Regexp initialization
	if v, ok := p["regexp"]; ok {
		r, err := regexp.Compile(v)
		if err != nil {
			return nil, fmt.Errorf("textfilter: cannot compile the regular expression in 'regexp' parameter")
		}
		f.regexp = r
	}
	if v, ok := f.params["extract"]; ok && v == "true" {
		f.extractText = true
	}
	if v, ok := p["text"]; ok {
		f.text = v
	}

	return f, nil
}

func (f *Text) DoFilter(msg *com.DataMessage) (bool, error) {
	text := msg.GetMessage()

	found := false
	if f.regexp != nil {
		if f.extractText {
			match := f.regexp.FindStringSubmatch(text)
			if match != nil {
				msg.SetMessage(match[0])
			}
			found = match != nil
		} else {
			if f.regexp.MatchString(text) {
				found = true
			}
		}
	}

	if f.text != "" && strings.Contains(text, f.text) {
		found = true
	}

	return found, nil
}

// Set the name of the filter
func init() {
	register("text", NewTextFilter)
}
package filters

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/Matrix86/driplane/data"
)

// URL is a Filter to search urls in the input Message
type URL struct {
	Base

	rURL *regexp.Regexp

	getHTTP  bool
	getHTTPS bool
	getFTP   bool
	target   string

	extractURL bool

	params map[string]string
}

// NewURLFilter is the registered method to instantiate a UrlFilter
func NewURLFilter(p map[string]string) (Filter, error) {
	f := &URL{
		params:     p,
		getHTTP:    true,
		getHTTPS:   true,
		getFTP:     true,
		extractURL: true,
		target:     "main",
	}
	f.cbFilter = f.DoFilter

	f.rURL = regexp.MustCompile(`(?i)((http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/|ftp:\/\/)?(([a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5})|([0-9]{1,3}\.){3}[0-9]{1,3})(:[0-9]{1,5})?(\/[^\s]+)?)`)

	if v, ok := f.params["http"]; ok && v == "false" {
		f.getHTTP = false
	}
	if v, ok := f.params["https"]; ok && v == "false" {
		f.getHTTPS = false
	}
	if v, ok := f.params["ftp"]; ok && v == "false" {
		f.getFTP = false
	}
	if v, ok := f.params["extract"]; ok && v == "false" {
		f.extractURL = false
	}
	if v, ok := f.params["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *URL) DoFilter(msg *data.Message) (bool, error) {
	var text string

	if v, ok := msg.GetTarget(f.target).(string); ok {
		text = v
	} else if v, ok := msg.GetTarget(f.target).([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}

	found := false
	match := f.rURL.FindAllStringSubmatch(text, -1)
	if match != nil {
		for _, m := range match {
			mm := m[0]
			if f.getHTTP && strings.HasPrefix(strings.ToLower(mm), "http://") {
				found = true
			} else if f.getHTTPS && strings.HasPrefix(strings.ToLower(mm), "https://") {
				found = true
			} else if f.getFTP && strings.HasPrefix(strings.ToLower(mm), "ftp://") {
				found = true
			}

			if f.extractURL && found {
				clone := msg.Clone()
				clone.SetMessage(mm)
				clone.SetExtra("fulltext", text)
				f.Propagate(clone)
				// We need to stop the propagation of the first message
				found = false
			} else if found {
				break
			}
		}
	}
	return found, nil
}

// OnEvent is called when an event occurs
func (f *URL) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("url", NewURLFilter)
}

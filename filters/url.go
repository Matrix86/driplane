package filters

import (
	"github.com/Matrix86/driplane/data"
	"regexp"
	"strings"
)

type URL struct {
	Base

	rUrl *regexp.Regexp

	getHttp  bool
	getHttps bool
	getFtp   bool

	extractUrl bool

	params map[string]string
}

func NewUrlFilter(p map[string]string) (Filter, error) {
	f := &URL{
		params:     p,
		getHttp:    true,
		getHttps:   true,
		getFtp:     true,
		extractUrl: true,
	}
	f.cbFilter = f.DoFilter

	f.rUrl = regexp.MustCompile(`(?i)((http:\/\/www\.|https:\/\/www\.|http:\/\/|https:\/\/)?(([a-z0-9]+([\-\.]{1}[a-z0-9]+)*\.[a-z]{2,5})|([0-9]{1,3}\.){3}[0-9]{1,3})(:[0-9]{1,5})?(\/[^\s]+)?)`)

	if v, ok := f.params["http"]; ok && v == "false" {
		f.getHttp = false
	}
	if v, ok := f.params["https"]; ok && v == "false" {
		f.getHttps = false
	}
	if v, ok := f.params["ftp"]; ok && v == "false" {
		f.getFtp = false
	}
	if v, ok := f.params["extract"]; ok && v == "false" {
		f.extractUrl = false
	}

	return f, nil
}

func (f *URL) DoFilter(msg *data.Message) (bool, error) {
	text := msg.GetMessage()

	found := false
	match := f.rUrl.FindAllStringSubmatch(text, -1)
	if match != nil {
		for _, m := range match {
			mm := m[0]
			if f.getHttp && strings.HasPrefix(strings.ToLower(mm), "http://") {
				found = true
			} else if f.getHttps && strings.HasPrefix(strings.ToLower(mm), "https://") {
				found = true
			} else if f.getFtp && strings.HasPrefix(strings.ToLower(mm), "ftp://") {
				found = true
			}

			if f.extractUrl {
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

// Set the name of the filter
func init() {
	register("url", NewUrlFilter)
}

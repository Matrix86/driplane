package filters

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"github.com/PuerkitoBio/goquery"
	"strings"

	"github.com/evilsocket/islazy/log"
)

// HTML is a filter to parse the HTML format
type HTML struct {
	Base

	selectors string
	getType   string
	attr      string
	target    string

	params map[string]string
}

// NewHTMLFilter is the registered method to instantiate a HtmlFilter
func NewHTMLFilter(p map[string]string) (Filter, error) {
	f := &HTML{
		params:  p,
		target:  "main",
		getType: "html",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["selector"]; ok {
		f.selectors = v
	}
	if v, ok := f.params["get"]; ok {
		switch v {
		case "attr",
			"text",
			"html":
			f.getType = v

		default:
			return nil, fmt.Errorf("get param is not valid")
		}

	}
	if v, ok := f.params["target"]; ok {
		f.target = v
	}
	if v, ok := f.params["attr"]; ok {
		f.attr = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *HTML) DoFilter(msg *data.Message) (bool, error) {
	var err error
	var text string

	if v, ok := msg.GetTarget(f.target).(string); ok {
		text = v
	} else if v, ok := msg.GetTarget(f.target).([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}

	stringReader := strings.NewReader(text)

	doc, err := goquery.NewDocumentFromReader(stringReader)
	if err != nil {
		return false, err
	}

	doc.Find(f.selectors).Each(func(i int, s *goquery.Selection) {
		message := ""

		switch f.getType {
		case "attr":
			message, _ = s.Attr(f.attr)

		case "text":
			message = s.Text()

		case "html":
			message, err = s.Html()
			if err != nil {
				log.Error("error on parsing: %s", err)
				return
			}
		}

		if message != "" {
			clone := msg.Clone()
			clone.SetMessage(message)
			clone.SetExtra("fulltext", text)
			f.Propagate(clone)
		}
	})

	return false, nil
}

// OnEvent is called when an event occurs
func (f *HTML) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("html", NewHTMLFilter)
}

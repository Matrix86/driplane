package filters

import (
	"github.com/Matrix86/driplane/data"
	html "html/template"
	"io/ioutil"
	text "text/template"
)

type Format struct {
	Base

	template     interface{}
	templateType string // html or text

	params map[string]string
}

func NewFormatFilter(p map[string]string) (Filter, error) {
	f := &Format{
		params:       p,
		templateType: "text",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["type"]; ok && v == "html" {
		f.templateType = "html"
	}

	if v, ok := f.params["template"]; ok {
		if f.templateType == "html" {
			t, err := html.New("formatFilterTemplate").Parse(v)
			if err != nil {
				return nil, err
			}
			f.template = t
		} else {
			t, err := text.New("formatFilterTemplate").Parse(v)
			if err != nil {
				return nil, err
			}
			f.template = t
		}
	}
	if v, ok := f.params["file"]; ok {
		content, err := ioutil.ReadFile(v)
		if err != nil {
			return nil, err
		}
		if f.templateType == "html" {
			t, err := html.New("formatFilterTemplate").Parse(string(content))
			if err != nil {
				return nil, err
			}
			f.template = t
		} else {
			t, err := text.New("formatFilterTemplate").Parse(string(content))
			if err != nil {
				return nil, err
			}
			f.template = t
		}
	}

	return f, nil
}

func (f *Format) DoFilter(msg *data.Message) (bool, error) {
	text, err := msg.ApplyPlaceholder(f.template)
	if err != nil {
		return false, err
	}
	msg.SetMessage(text)
	return true, nil
}

// Set the name of the filter
func init() {
	register("format", NewFormatFilter)
}

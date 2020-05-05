package filters

import (
	"github.com/Matrix86/driplane/data"
	"html/template"
)

type Format struct {
	Base

	template *template.Template
	params map[string]string
}

func NewFormatFilter(p map[string]string) (Filter, error) {
	f := &Format{
		params: p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["template"]; ok {
		t, err := template.New("formatFilterTemplate").Parse(v)
		if err != nil {
			return nil, err
		}
		f.template = t
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

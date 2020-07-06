package filters

import (
	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"text/template"
)

type Override struct {
	Base

	name  *template.Template
	value *template.Template

	params map[string]string
}

func NewOverrideFilter(p map[string]string) (Filter, error) {
	f := &Override{
		params: p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["name"]; ok {
		t, err := template.New("setFilterName").Parse(v)
		if err != nil {
			return nil, err
		}
		f.name = t
	}

	if v, ok := p["value"]; ok {
		t, err := template.New("setFilterValue").Parse(v)
		if err != nil {
			return nil, err
		}
		f.value = t
	}

	return f, nil
}

func (f *Override) DoFilter(msg *data.Message) (bool, error) {
	name, err := msg.ApplyPlaceholder(f.name)
	if err != nil {
		return false, err
	}
	value, err := msg.ApplyPlaceholder(f.value)
	if err != nil {
		return false, err
	}

	log.Debug("[setfilter] setting msg[%s]=%s", name, value)
	msg.SetTarget(name, value)

	return true, nil
}

// Set the name of the filter
func init() {
	register("override", NewOverrideFilter)
}

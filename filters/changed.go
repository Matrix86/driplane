package filters

import (
	"github.com/Matrix86/driplane/utils"
	"sync"

	"github.com/Matrix86/driplane/data"
)

// Changed is a Filter that call the propagation method only if
// the input Message is different from the previous one
type Changed struct {
	sync.Mutex
	Base

	target string

	params map[string]string
	cache  string
}

// NewChangedFilter is the registered method to instantiate a ChangedFilter
func NewChangedFilter(p map[string]string) (Filter, error) {
	f := &Changed{
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
func (f *Changed) DoFilter(msg *data.Message) (bool, error) {
	var text interface{}

	if f.target == "main" {
		text = msg.GetMessage()
	} else if v, ok := msg.GetExtra()[f.target]; ok {
		text = v
	} else {
		return false, nil
	}

	hash := utils.MD5Sum(text)
	if f.cache != hash {
		f.Lock()
		defer f.Unlock()
		f.cache = hash
		return true, nil
	}

	return false, nil
}

// OnEvent is called when an event occurs
func (f *Changed) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("changed", NewChangedFilter)
}

package filters

import (
	"fmt"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
)

// Echo is a filter that print the input Message on the logs.
type Echo struct {
	Base

	printExtra bool

	params map[string]string
}

// NewEchoFilter is the registered method to instantiate a EchoFilter
func NewEchoFilter(p map[string]string) (Filter, error) {
	f := &Echo{
		params:     p,
		printExtra: false,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["extra"]; ok && v == "true" {
		f.printExtra = true
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Echo) DoFilter(msg *data.Message) (bool, error) {
	text := msg.GetMessage()
	if f.printExtra {
		for k, v := range msg.GetExtra() {
			text = fmt.Sprintf("%s [%s: %s] ", text, k, v)
		}
	}
	log.Info("%s", text)
	return true, nil
}

// Set the name of the filter
func init() {
	register("echo", NewEchoFilter)
}

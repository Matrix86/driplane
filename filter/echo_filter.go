package filter

import (
	"fmt"

	"github.com/Matrix86/driplane/com"

	"github.com/evilsocket/islazy/log"
)

type EchoFilter struct {
	FilterBase

	printExtra bool

	params map[string]string
}

func NewEchoFilter(p map[string]string) (Filter, error) {
	f := &EchoFilter{
		params: p,
		printExtra: false,
	}

	if v, ok := f.params["extra"]; ok && v == "true" {
		f.printExtra = true
	}

	return f, nil
}

func (f *EchoFilter) DoFilter(msg *com.DataMessage) (bool, error) {
	text := msg.GetMessage()

	if f.printExtra {
		for n, s := range msg.GetExtra() {
			text = fmt.Sprintf("%s [%s: %s] ", text, n, s)
		}
	}
	log.Info("%s", text)

	return true, nil
}

// Set the name of the filter
func init() {
	register("echo", NewEchoFilter)
}
package filters

import (
	"os/exec"
	"text/template"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
)

// System is a Filter to exec a command on the host machine using the input Message
type System struct {
	Base

	command   *template.Template

	params map[string]string
}

// NewSystemFilter is the registered method to instantiate a SystemFilter
func NewSystemFilter(p map[string]string) (Filter, error) {
	f := &System{
		params: p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["cmd"]; ok {
		t, err := template.New("systemFilterCommand").Parse(v)
		if err != nil {
			return nil, err
		}
		f.command = t
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *System) DoFilter(msg *data.Message) (bool, error) {
	cmd, err := msg.ApplyPlaceholder(f.command)
	if err != nil {
		return false, err
	}

	log.Debug("[systemfilter] command: %s", cmd)
	c := exec.Command("sh", "-c", cmd)
	output, err := c.CombinedOutput()
	if err != nil {
		log.Debug("[systemfilter] command failed: %s %s", err, output)
		return false, err
	}

	msg.SetMessage(string(output))

	return true, nil
}

// OnEvent is called when an event occurs
func (f *System) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("system", NewSystemFilter)
}

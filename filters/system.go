package filters

import (
	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"os/exec"
	"regexp"
	"text/template"
)

type System struct {
	Base

	command   *template.Template
	rExtraCmd *regexp.Regexp

	params map[string]string
}

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

	f.rExtraCmd = regexp.MustCompile(`(%extra\.[a-z0-9]+%)`)

	return f, nil
}

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

// Set the name of the filter
func init() {
	register("system", NewSystemFilter)
}

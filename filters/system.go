package filters

import (
	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/plugin"
	"os/exec"
	"regexp"
)

type System struct {
	Base

	command   string
	rExtraCmd *regexp.Regexp

	params map[string]string
}

func NewSystemFilter(p map[string]string) (Filter, error) {
	f := &System{
		params: p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["cmd"]; ok {
		f.command = v
	}

	f.rExtraCmd = regexp.MustCompile(`(%extra\.[a-z0-9]+%)`)

	return f, nil
}

func (f *System) DoFilter(msg *data.Message) (bool, error) {
	cmd := msg.ReplacePlaceholders(f.command)

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

	plugin.Defines = map[string]interface{}{
		"log": func(s string) interface{} {
			log.Info("%s", s)
			return nil
		},
	}
}

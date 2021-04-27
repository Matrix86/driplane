package filters

import (
	"fmt"
	"math/rand"
	"strconv"
	"time"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
)

// Random is a Filter to create a random number
type Random struct {
	Base

	output string
	min   int64
	max     int64

	params map[string]string
}

// NewRandomFilter is the registered method to instantiate a RandomFilter
func NewRandomFilter(p map[string]string) (Filter, error) {
	f := &Random{
		params: p,
		output: "main",
		min: 0,
		max: 999999,
	}
	f.cbFilter = f.DoFilter


	if v, ok := p["output"]; ok {
		f.output = v
	}
	if v, ok := p["min"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		f.min = i
	}
	if v, ok := p["max"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		f.max = i
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Random) DoFilter(msg *data.Message) (bool, error) {
	s1 := rand.NewSource(time.Now().UnixNano())
	r1 := rand.New(s1)

	random := fmt.Sprintf("%d", r1.Int63n(f.max - f.min) + f.min)

	log.Debug("[randomfilter] picked %s", random)
	msg.SetTarget(f.output, random)

	return true, nil
}

// OnEvent is called when an event occurs
func (f *Random) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("random", NewRandomFilter)
}

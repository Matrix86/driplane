package filters

import (
	"context"
	"strconv"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"golang.org/x/time/rate"
)

// RateLimit is a Filter to create a RateLimit number
type RateLimit struct {
	Base

	objects       int64
	seconds       int64
	limiter       *rate.Limiter
	context       context.Context
	cancelContext context.CancelFunc

	params map[string]string
}

// NewRateLimitFilter is the registered method to instantiate a RateLimit Filter
func NewRateLimitFilter(p map[string]string) (Filter, error) {
	ctx, cancel := context.WithCancel(context.Background())
	f := &RateLimit{
		params:        p,
		seconds:       1,
		context:       ctx,
		cancelContext: cancel,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["rate"]; ok {
		i, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			return nil, err
		}
		f.objects = i
	}

	f.limiter = rate.NewLimiter(rate.Limit(f.objects), int(f.seconds))

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *RateLimit) DoFilter(msg *data.Message) (bool, error) {
	if f.objects > 0 {
		if err := f.limiter.Wait(f.context); err != nil {
			return false, nil
		}
	}

	return true, nil
}

// OnEvent is called when an event occurs
func (f *RateLimit) OnEvent(event *data.Event) {
	if event.Type == "shutdown" {
		log.Debug("shutdown event received")
		f.cancelContext()
	}
}

// Set the name of the filter
func init() {
	register("ratelimit", NewRateLimitFilter)
}

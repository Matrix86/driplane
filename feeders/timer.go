package feeders

import (
	"fmt"
	"time"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
)

// Timer is a Feeder that triggers the pipeline using a timer
type Timer struct {
	Base

	frequency time.Duration

	stopChan chan bool
	ticker   *time.Ticker
}

// NewTimerFeeder is the registered method to instantiate a TimerFeeder
func NewTimerFeeder(conf map[string]string) (Feeder, error) {
	f := &Timer{
		stopChan:  make(chan bool),
		frequency: 60 * time.Second,
	}

	if val, ok := conf["timer.freq"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("specified frequency cannot be parsed '%s': %s", val, err)
		}
		f.frequency = d
	}

	return f, nil
}

// Start propagates a message every time the ticker is fired
func (f *Timer) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case <-f.ticker.C:
				t := time.Now()
				extra := make(map[string]interface{})
				extra["timestamp"] = t.Unix()
				extra["rfc3339"]   = t.Format(time.RFC3339)
				msg := data.NewMessageWithExtra(extra["rfc3339"], extra)
				f.Propagate(msg)
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *Timer) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.stopChan <- true
	f.ticker.Stop()
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *Timer) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("timer", NewTimerFeeder)
}

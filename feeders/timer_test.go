package feeders

import (
	"fmt"
	"testing"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
)

func newTestTimer(conf map[string]string) (*Timer, chan *data.Message, error) {
	feeder, err := NewTimerFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*Timer)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *Timer")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("timerfeeder")
	f.setID(1)

	received := make(chan *data.Message, 10)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

func TestNewTimerFeeder(t *testing.T) {
	feeder, err := NewTimerFeeder(map[string]string{
		"timer.freq": "5s",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Timer); ok {
		if f.frequency != 5*time.Second {
			t.Errorf("'timer.freq' parameter ignored, expected 5s got '%s'", f.frequency)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewTimerFeederDefaults(t *testing.T) {
	feeder, err := NewTimerFeeder(map[string]string{})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Timer); ok {
		if f.frequency != 60*time.Second {
			t.Errorf("default frequency should be 60s, got '%s'", f.frequency)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewTimerFeederInvalidFreq(t *testing.T) {
	_, err := NewTimerFeeder(map[string]string{
		"timer.freq": "notaduration",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'timer.freq' is invalid")
	}
}

func TestTimerStartStop(t *testing.T) {
	f, _, err := newTestTimer(map[string]string{
		"timer.freq": "10s",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	if !f.isRunning {
		t.Errorf("feeder should be running after Start()")
	}

	time.Sleep(100 * time.Millisecond)

	f.Stop()
	if f.isRunning {
		t.Errorf("feeder should not be running after Stop()")
	}
}

func TestTimerPropagatesMessage(t *testing.T) {
	f, received, err := newTestTimer(map[string]string{
		"timer.freq": "100ms",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	select {
	case msg := <-received:
		if msg == nil {
			t.Errorf("expected a non-nil message")
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("expected a message within 500ms but got none")
	}
}

func TestTimerMessageHasTimestamp(t *testing.T) {
	f, received, err := newTestTimer(map[string]string{
		"timer.freq": "100ms",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	before := time.Now().Unix()
	f.Start()
	defer f.Stop()

	select {
	case msg := <-received:
		extra := msg.GetExtra()
		ts, ok := extra["timestamp"]
		if !ok {
			t.Errorf("expected 'timestamp' extra field to be set")
		}
		if ts.(int64) < before {
			t.Errorf("timestamp should be >= time before Start(), got %d", ts.(int64))
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("expected a message within 500ms but got none")
	}
}

func TestTimerMessageHasRFC3339(t *testing.T) {
	f, received, err := newTestTimer(map[string]string{
		"timer.freq": "100ms",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	select {
	case msg := <-received:
		extra := msg.GetExtra()
		rfc, ok := extra["rfc3339"]
		if !ok {
			t.Errorf("expected 'rfc3339' extra field to be set")
		}
		if _, err := time.Parse(time.RFC3339, rfc.(string)); err != nil {
			t.Errorf("'rfc3339' value is not a valid RFC3339 time: %s", rfc)
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("expected a message within 500ms but got none")
	}
}

func TestTimerMessageMainIsRFC3339(t *testing.T) {
	f, received, err := newTestTimer(map[string]string{
		"timer.freq": "100ms",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	select {
	case msg := <-received:
		main := msg.GetMessage()
		if main == nil || main == "" {
			t.Errorf("expected main field to be set to the RFC3339 time string")
		}
		if _, err := time.Parse(time.RFC3339, main.(string)); err != nil {
			t.Errorf("main field is not a valid RFC3339 time: %s", main)
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("expected a message within 500ms but got none")
	}
}

func TestTimerFiresMultipleTimes(t *testing.T) {
	f, received, err := newTestTimer(map[string]string{
		"timer.freq": "100ms",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	count := 0
	timeout := time.After(500 * time.Millisecond)
	for {
		select {
		case <-received:
			count++
			if count >= 3 {
				return
			}
		case <-timeout:
			t.Errorf("expected at least 3 ticks within 500ms, got %d", count)
			return
		}
	}
}

func TestTimerStopsAfterStop(t *testing.T) {
	f, received, err := newTestTimer(map[string]string{
		"timer.freq": "50ms",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()

	// wait for at least one tick
	select {
	case <-received:
	case <-time.After(500 * time.Millisecond):
		t.Fatal("expected at least one tick before Stop()")
	}

	f.Stop()

	// drain any in-flight messages
	for len(received) > 0 {
		<-received
	}

	// after Stop(), no more messages should arrive
	select {
	case <-received:
		t.Errorf("received a message after Stop()")
	case <-time.After(200 * time.Millisecond):
		// expected — no more ticks
	}
}

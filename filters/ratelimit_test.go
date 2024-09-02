package filters

import (
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewRateLimitFilter(t *testing.T) {
	filter, err := NewRateLimitFilter(map[string]string{"rate": "2", "min": "10", "max": "40"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*RateLimit); ok {
		if e.objects != 2 {
			t.Errorf("'output' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}

	_, err = NewRateLimitFilter(map[string]string{"rate": "random text", "min": "x", "max": "40"})
	if err == nil {
		t.Errorf("constructor should return error")
	}
}

func TestRateLimit_DoFilter(t *testing.T) {
	filter, err := NewRateLimitFilter(map[string]string{"rate": "5"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	if e, ok := filter.(*RateLimit); ok {
		msg := "this is a test..."
		m := data.NewMessage(msg)
		_, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if m.GetMessage() != msg {
			t.Errorf("DoFilter changed the message")
		}

		filter.OnEvent(&data.Event{
			Type:    "shutdown",
			Content: "shutdown",
		})
		v, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if v {
			t.Errorf("DoFilter has to return false after the shutdown")
		}

	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

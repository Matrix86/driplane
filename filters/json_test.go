package filters

import (
	"bytes"
	"testing"

	"github.com/Matrix86/driplane/data"

	"github.com/asaskevich/EventBus"
)

const jsonTest = `{"top": { "inside": "value"}}`

func TestNewJSONFilter(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{"none": "none", "selector": "#selector"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		if e.target != "main" {
			t.Errorf("target should be 'main' if not specified")
		}
		if e.selector != "#selector" {
			t.Errorf("selector should be '#selector'")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestNewJSONFilterParams(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "#selector",
		"target":   "othertarget",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		if e.target != "othertarget" {
			t.Errorf("target should be 'othertarget'")
		}
		if e.selector != "#selector" {
			t.Errorf("selector should be '#selector'")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestNewJSONWithoutFilterParams(t *testing.T) {
	_, err := NewJSONFilter(map[string]string{})
	if err == nil {
		t.Errorf("The test should return an error")
	}
}

func TestJSON_DoFilter(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/inside",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := jsonTest
		m := data.NewMessage(msg)
		_, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if m.GetMessage() != msg {
			t.Errorf("the message has been altered by the filter")
		}
		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "value" {
			t.Errorf("tags have not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestJSON_DoFilterOnByte(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/inside",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := []byte(jsonTest)
		m := data.NewMessage(msg)
		_, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if !bytes.Equal(m.GetMessage().([]byte), msg) {
			t.Errorf("the message has been altered by the filter")
		}
		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "value" {
			t.Errorf("tags have not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestJSON_DoFilterOnWrongType(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/inside",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := 2
		m := data.NewMessage(msg)
		x, err := e.DoFilter(m)
		if err == nil {
			t.Errorf("DoFilter should return an error")
		}
		if x {
			t.Errorf("DoFilter should return false")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestJSON_DoFilterIfNotJSON(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/inside",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := "not a JSON"
		m := data.NewMessage(msg)
		x, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter should return nil")
		}
		if x {
			t.Errorf("DoFilter should return false")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestJSON_DoFilterOnWrongSelector(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/ops",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := jsonTest
		m := data.NewMessage(msg)
		x, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter should return nil")
		}
		if x {
			t.Errorf("DoFilter should return false")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestJSON_DoFilterOnEmptyJSON(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/ops",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := ""
		m := data.NewMessage(msg)
		x, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter should return nil")
		}
		if x {
			t.Errorf("DoFilter should return false")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestJSON_DoFilterOnBadJSON(t *testing.T) {
	filter, err := NewJSONFilter(map[string]string{
		"selector": "top/ops",
		"target":   "main",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*JSON); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := jsonTest[:len(jsonTest)-1]
		m := data.NewMessage(msg)
		x, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter should return nil")
		}
		if x {
			t.Errorf("DoFilter should return false")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

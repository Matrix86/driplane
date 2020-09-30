package filters

import (
	"github.com/Matrix86/driplane/data"
	"strconv"
	"testing"
)

func TestNewRandomFilter(t *testing.T) {
	filter, err := NewRandomFilter(map[string]string{"output": "random_number", "min": "10", "max":"40" })
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Random); ok {
		if e.output != "random_number" {
			t.Errorf("'output' parameter ignored")
		}
		if e.min != 10 {
			t.Errorf("'min' parameter ignored")
		}
		if e.max != 40 {
			t.Errorf("'max' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}

	filter, err = NewRandomFilter(map[string]string{"output": "random_number", "min": "x", "max":"40" })
	if err == nil {
		t.Errorf("constructor should return error")
	}

	filter, err = NewRandomFilter(map[string]string{"output": "random_number", "min": "1", "max":"x" })
	if err == nil {
		t.Errorf("constructor should return error")
	}
}

func TestRandom_DoFilter(t *testing.T) {
	filter, err := NewRandomFilter(map[string]string{"output": "random_number", "min": "10", "max":"40" })
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	if e, ok := filter.(*Random); ok {
		msg := "this is a test..."
		m := data.NewMessage(msg)
		_, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		res := m.GetTarget("random_number")
		i, err := strconv.Atoi(res.(string))
		if err != nil {
			t.Errorf("received data is not a string")
		}
		if i < 10 && i > 40 {
			t.Errorf("random number is out of range")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

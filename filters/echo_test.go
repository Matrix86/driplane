package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
)

func TestNewEchoFilter(t *testing.T) {
	filter, err := NewEchoFilter(map[string]string{"none": "none", "extra": "true"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Echo); ok {
		if e.printExtra == false {
			t.Errorf("'extra' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestEchoDoFilter(t *testing.T) {
	filter, err := NewEchoFilter(map[string]string{"none": "none", "extra": "true"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Echo); ok {
		m := data.NewMessageWithExtra("main message", map[string]string{"extra": "1"})
		b, err := e.DoFilter(m)
		if b != true {
			t.Errorf("DoFilter cannot return false")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

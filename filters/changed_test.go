package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
)

func TestNewChangedFilter(t *testing.T) {
	filter, err := NewChangedFilter(map[string]string{"none": "none", "target": "test"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Changed); ok {
		if e.target != "test" {
			t.Errorf("'target' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestChangedDoFilterOnMain(t *testing.T) {
	filter, err := NewChangedFilter(map[string]string{"none": "none"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Changed); ok {
		m := data.NewMessageWithExtra("main message", map[string]string{"test": "1"})
		// First time it should return true
		b, err := e.DoFilter(m)
		if b == false {
			t.Errorf("it should return true the first time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// second time it should return false
		b, err = e.DoFilter(m)
		if b == true {
			t.Errorf("it should return false the second time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// third time it should return false
		b, err = e.DoFilter(m)
		if b == true {
			t.Errorf("third time it should return false")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		m = data.NewMessageWithExtra("main message changed", map[string]string{"test": "2"})
		b, err = e.DoFilter(m)
		if b == false {
			t.Errorf("it should return true, the value is changed")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestChangedDoFilterOnExtra(t *testing.T) {
	filter, err := NewChangedFilter(map[string]string{"none": "none", "target": "test"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Changed); ok {
		m := data.NewMessageWithExtra("main message", map[string]string{"test": "1"})
		// First time it should return true
		b, err := e.DoFilter(m)
		if b == false {
			t.Errorf("it should return true the first time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// second time it should return false
		b, err = e.DoFilter(m)
		if b == true {
			t.Errorf("it should return false the second time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// third time it should return false
		b, err = e.DoFilter(m)
		if b == true {
			t.Errorf("third time it should return false")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		m = data.NewMessageWithExtra("main message", map[string]string{"test": "2"})
		b, err = e.DoFilter(m)
		if b == false {
			t.Errorf("it should return true, the value is changed")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}
package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
	"time"
)

func TestNewCacheFilter(t *testing.T) {
	filter, err := NewCacheFilter(map[string]string{"none": "none", "target": "main", "ttl": "1s"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Cache); ok {
		if e.target != "main" {
			t.Errorf("'target' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestCacheDoFilterOnMain(t *testing.T) {
	filter, err := NewCacheFilter(map[string]string{"none": "none", "ttl": "1s"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Cache); ok {
		m := data.NewMessageWithExtra("main message", map[string]string{"test": "1"})
		// First time it should return true
		b, err := e.DoFilter(m)
		if b == false {
			t.Errorf("it should return true the first time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// Second time it should return false
		b, err = e.DoFilter(m)
		if b == true {
			t.Errorf("it should return false the second time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// waiting 2 secs to check if the value is expired
		time.Sleep(2 * time.Second)

		// value should be expired
		b, err = e.DoFilter(m)
		if b == false {
			t.Errorf("the value should be expired")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestCacheDoFilterOnExtra(t *testing.T) {
	filter, err := NewCacheFilter(map[string]string{"target": "extra", "ttl": "1s"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Cache); ok {
		m := data.NewMessageWithExtra("main message", map[string]string{"extra": "boooo"})
		// First time it should return true
		b, err := e.DoFilter(m)
		if b == false {
			t.Errorf("it should return true the first time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// Second time it should return false
		b, err = e.DoFilter(m)
		if b == true {
			t.Errorf("it should return false the second time")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		// waiting 2 secs to check if the value is expired
		time.Sleep(2 * time.Second)

		// value should be expired
		b, err = e.DoFilter(m)
		if b == false {
			t.Errorf("the value should be expired")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestCacheDoFilterIfExtraNotExist(t *testing.T) {
	filter, err := NewCacheFilter(map[string]string{"target": "unknown"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Cache); ok {
		m := data.NewMessageWithExtra("main message", map[string]string{"test": "1"})
		b, err := e.DoFilter(m)
		if b != false {
			t.Errorf("it should return false")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

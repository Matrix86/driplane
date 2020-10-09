package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
	"time"
)

func TestNewCacheFilter(t *testing.T) {
	filter, err := NewCacheFilter(map[string]string{"none": "none", "target": "main", "ttl": "1s", "file": "/tmp/test_file_cache.dat"})
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
		m := data.NewMessageWithExtra("main message", map[string]interface{}{"test": "1"})
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
		m := data.NewMessageWithExtra("main message", map[string]interface{}{"extra": "boooo"})
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
		m := data.NewMessageWithExtra("main message", map[string]interface{}{"test": "1"})
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

func TestCacheDoFilterWithGlobal(t *testing.T) {
	filter1, err := NewCacheFilter(map[string]string{"target": "main", "global": "true"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	filter2, err := NewCacheFilter(map[string]string{"target": "main", "global": "true"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	f1, ok1 := filter1.(*Cache)
	f2, ok2 := filter2.(*Cache)
	if ok1 && ok2 {
		m := data.NewMessageWithExtra("main message", map[string]interface{}{"test": "1"})
		b1, err := f1.DoFilter(m)
		if b1 != true {
			t.Errorf("first time the filter should return TRUE")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}

		b2, err := f2.DoFilter(m)
		if b2 != false {
			t.Errorf("second time the filter should return FALSE")
		}
		if err != nil {
			t.Errorf("DoFilter cannot return an error '%s'", err)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

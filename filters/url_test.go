package filters

import (
	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
	"testing"
)

func TestNewURLFilter(t *testing.T) {
	filter, err := NewURLFilter(map[string]string{"none": "none", "http": "false", "ftp": "false", "target": "foo"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*URL); ok {
		if e.getHTTP == true {
			t.Errorf("'http' parameter ignored")
		}
		if e.getFTP == true {
			t.Errorf("'ftp' parameter ignored")
		}
		if e.target != "foo" {
			t.Errorf("'target' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestURLDoFilterNoExtract(t *testing.T) {
	filter, err := NewURLFilter(map[string]string{"http": "true", "https": "false", "extract": "false"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*URL); ok {
		msg := "this is a test...\ncan you see the http://www.google.it/something?par=val#anchor url?"
		m := data.NewMessage(msg)
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}
		if m.GetMessage() != msg {
			t.Errorf("the message has been altered by the filter")
		}

		msg = "this is a test...\ncan you see the https://www.google.it/something?par=val#anchor url?"
		m = data.NewMessage(msg)
		b, err = e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return false")
		}
		if m.GetMessage() != msg {
			t.Errorf("the message has been altered by the filter")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestURLDoFilterExtractHTTP(t *testing.T) {
	filter, err := NewURLFilter(map[string]string{
		"http":    "true",
		"https":   "false",
		"ftp":     "false",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*URL); ok {
		m := data.NewMessage("this is a test...\ncan you see the http://www.google.it/something?par=val#anchor url? and https://lol.site.com/path/something?par=val#anchor or ftp://11.11.11.11/path.pdf ?")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "http://www.google.it/something?par=val#anchor" {
			t.Errorf("URL has not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestURLDoFilterExtractHTTPS(t *testing.T) {
	filter, err := NewURLFilter(map[string]string{
		"http":    "false",
		"https":   "true",
		"ftp":     "false",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*URL); ok {
		m := data.NewMessage("this is a test...\ncan you see the http://www.google.it/something?par=val#anchor url? and https://lol.site.com/path/something?par=val#anchor or ftp://11.11.11.11/path.pdf ?")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "https://lol.site.com/path/something?par=val#anchor" {
			t.Errorf("URL has not been extracted correctly: expected=%s had=%s", "https://lol.site.com/path/something?par=val#anchor", fb.Collected[0].GetMessage())
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestURLDoFilterExtractFTP(t *testing.T) {
	filter, err := NewURLFilter(map[string]string{
		"http":    "false",
		"https":   "false",
		"ftp":     "true",
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*URL); ok {
		m := data.NewMessage("this is a test...\ncan you see the http://www.google.it/something?par=val#anchor url? and https://lol.site.com/path/something?par=val#anchor or ftp://11.11.11.11/path.pdf ?")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "ftp://11.11.11.11/path.pdf" {
			t.Errorf("URL has not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestURLDoFilterExtractAll(t *testing.T) {
	filter, err := NewURLFilter(map[string]string{
		"extract": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	fb := NewFakeBus()

	filter.setBus(EventBus.Bus(fb))
	if e, ok := filter.(*URL); ok {
		m := data.NewMessage("this is a test...\ncan you see the http://www.google.it/something?par=val#anchor url? and https://lol.site.com/path/something?par=val#anchor or ftp://11.11.11.11/path.pdf ?")
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == true {
			t.Errorf("it should return true")
		}

		expected := []string{
			"http://www.google.it/something?par=val#anchor",
			"https://lol.site.com/path/something?par=val#anchor",
			"ftp://11.11.11.11/path.pdf",
		}

		for i, v := range fb.Collected {
			if v.GetMessage() != expected[i] {
				t.Errorf("URL has not been extracted correctly: expected %s had %s", expected[i], v.GetMessage())
			}
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

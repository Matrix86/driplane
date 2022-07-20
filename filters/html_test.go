package filters

import (
	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
	"testing"
)

const htmlTest = `
<!DOCTYPE html>
<html>
  <head>
    <title>Tests for siblings</title>
  </head>
  <BODY>
    <div id="main">
      <div id="n1" class="one even row"></div>
      <div id="n2" class="two odd row"></div>
      <div id="n3" class="three even row"></div>
      <div id="n4" class="four odd row"></div>
      <div id="n5" class="five even row"></div>
      <div id="n6" class="six odd row"></div>
    </div>
    <div id="foot">
      <div id="nf1" class="one even row"></div>
      <div id="nf2" class="two odd row"></div>
      <div id="nf3" class="three even row"></div>
      <div id="nf4" class="four odd row"></div>
      <div id="nf5" class="five even row odder"></div>
      <div id="nf6" class="six odd row"></div>
    </div>
  </BODY>
</html>`

func TestNewHTMLFilter(t *testing.T) {
	filter, err := NewHTMLFilter(map[string]string{"none": "none", "selector": "#selector"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*HTML); ok {
		if e.target != "main" {
			t.Errorf("target should be 'main' if not specified")
		}
		if e.attr != "" {
			t.Errorf("attr should be '' if not specified")
		}
		if e.getType != "html" {
			t.Errorf("get should be 'html' if not specified")
		}
		if e.selectors != "#selector" {
			t.Errorf("selector should be '#selector'")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestNewHTMLFilterParams(t *testing.T) {
	filter, err := NewHTMLFilter(map[string]string{
		"selector": "#selector",
		"target":   "othertarget",
		"get":      "text",
		"attr":     "attr",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*HTML); ok {
		if e.target != "othertarget" {
			t.Errorf("target should be 'othertarget'")
		}
		if e.attr != "attr" {
			t.Errorf("attr should be 'attr'")
		}
		if e.getType != "text" {
			t.Errorf("get should be 'text'. Received %s", e.getType)
		}
		if e.selectors != "#selector" {
			t.Errorf("selector should be '#selector'")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestNewHTMLFilterWrongGet(t *testing.T) {
	_, err := NewHTMLFilter(map[string]string{
		"selector": "#selector",
		"get":      "fake",
	})
	if err == nil {
		t.Errorf("constructor should return an error")
	}
	if err.Error() != "get param is not valid" {
		t.Errorf("constructor should return 'constructor should return an error'")
	}
}

func TestHTML_DoFilter(t *testing.T) {
	filter, err := NewHTMLFilter(map[string]string{
		"selector": ".row",
		"target":   "main",
		"get":      "attr",
		"attr":     "id",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*HTML); ok {
		fb := NewFakeBus()
		filter.setBus(EventBus.Bus(fb))

		msg := htmlTest
		m := data.NewMessage(msg)
		_, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if m.GetMessage() != msg {
			t.Errorf("the message has been altered by the filter")
		}
		if len(fb.Collected) == 0 || fb.Collected[0].GetMessage() != "n1" {
			t.Errorf("tags have not been extracted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
)

func TestNewFormatFilter(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{"none": "none", "template": "test {{.main}}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Format); ok {
		if e.template == nil {
			t.Errorf("'template' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestFormatDoFilter(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{"template": "main : {{.main}} extra : {{.test}}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Format); ok {
		m := data.NewMessageWithExtra("message", map[string]interface{}{"test": "1"})
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}

		if m.GetMessage() != "main : message extra : 1" {
			t.Errorf("message not formatted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}

	filter, err = NewFormatFilter(map[string]string{"target": "other", "template": "main : {{.main}} extra : {{.test}}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Format); ok {
		m := data.NewMessageWithExtra("message", map[string]interface{}{"test": "1"})
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}

		if m.GetTarget("other") != "main : message extra : 1" {
			t.Errorf("message not formatted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

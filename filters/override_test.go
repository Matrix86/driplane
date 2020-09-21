package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
)

func TestNewOverrideFilter(t *testing.T) {
	filter, err := NewOverrideFilter(map[string]string{"name": "name", "value": "value", "ignorethis": "ignorethis"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Override); ok {
		if e.name == nil {
			t.Errorf("'name' parameter ignored")
		}
		if e.value == nil {
			t.Errorf("'value' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestOverrideDoFilterAddNewField(t *testing.T) {
	filter, err := NewOverrideFilter(map[string]string{"name": "newname", "value": "newvalue"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Override); ok {
		msg := "main message"
		extra := make(map[string]interface{})
		m := data.NewMessageWithExtra(msg, extra)
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}
		if m.GetMessage() != msg {
			t.Errorf("the main field of the message has been altered by the filter")
		}
		if m.GetTarget("newname").(string) != "newvalue" {
			t.Errorf("the 'newname' field of the message is wrong: expected=newvalue had=%s", m.GetTarget("newname").(string))
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}

	filter2, err := NewOverrideFilter(map[string]string{"name": "{{ .field1 }}", "value": "{{ .field2 }}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter2.(*Override); ok {
		msg := "main message"
		extra := make(map[string]interface{})
		extra["field1"] = "field1"
		extra["field2"] = "field2"
		m := data.NewMessageWithExtra(msg, extra)
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}
		if m.GetMessage() != msg {
			t.Errorf("the main field of the message has been altered by the filter")
		}
		if m.GetTarget("field1").(string) != "field2" {
			t.Errorf("the 'field1' field of the message is wrong: expected=%s had=%s", "field2", m.GetTarget("field1").(string))
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

package filters

import (
	"github.com/Matrix86/driplane/data"
	"testing"
)

func TestNewSystemFilter(t *testing.T) {
	filter, err := NewSystemFilter(map[string]string{"cmd": "/bin/echo {{ .main }}" })
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*System); ok {
		m := data.NewMessage("test")
		cmd, _ := m.ApplyPlaceholder(e.command)
		if cmd != "/bin/echo test" {
			t.Errorf("'cmd' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestSystem_DoFilter(t *testing.T) {
	filter, err := NewSystemFilter(map[string]string{"cmd": "/bin/echo -n {{ .main }}" })
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}

	if e, ok := filter.(*System); ok {
		m := data.NewMessage("test")
		_, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		txt := m.GetMessage().(string)
		if txt != "test" {
			t.Errorf("TestSystem_DoFilter: wrong output: expected=%#v had=%#v", "test", txt)
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}
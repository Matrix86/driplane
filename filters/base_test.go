package filters

import (
	"fmt"
	"io/ioutil"
	"os"
	"path"
	"testing"

	"github.com/evilsocket/islazy/log"

	"github.com/Matrix86/driplane/data"
)

func TestNewFilter(t *testing.T) {
	bus := NewFakeBus()
	type Test struct {
		Name           string
		RuleName       string
		FilterName     string
		Config         map[string]string
		ID             int32
		Negative       bool
		ExpectedFilter Base
		ExpectedError  string
	}
	tests := []Test{
		{"FilterNotFound", "Rule1", "notexist", map[string]string{}, 1, false, Base{}, "filter 'notexist' doesn't exist"},
		{"FilterFound", "Rule1", "echofilter", map[string]string{}, 1, false, Base{rule: "Rule1", name: "echofilter", id: 1}, "filter 'notexist' doesn't exist"},
	}

	for _, v := range tests {
		f, err := NewFilter(v.RuleName, v.FilterName, v.Config, bus, v.ID, v.Negative)
		if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
		}
		if f != nil {
			if f.Rule() != v.ExpectedFilter.Rule() {
				t.Errorf("%s: wrong Rule: expected=%#v had=%#v", v.Name, v.ExpectedFilter.Rule(), f.Rule())
			}
			if f.Name() != v.ExpectedFilter.Name() {
				t.Errorf("%s: wrong Name: expected=%#v had=%#v", v.Name, v.ExpectedFilter.Name(), f.Name())
			}
			if f.GetIdentifier() != v.ExpectedFilter.GetIdentifier() {
				t.Errorf("%s: wrong ID: expected=%#v had=%#v", v.Name, v.ExpectedFilter.GetIdentifier(), f.GetIdentifier())
			}
		}
	}
}

func TestBase_GetIdentifier(t *testing.T) {
	b := Base{
		rule:     "rulename",
		name:     "filtername",
		id:       1,
		bus:      nil,
		negative: false,
		cbFilter: nil,
	}
	expected := fmt.Sprintf("%s:%d", b.name, b.id)
	if b.GetIdentifier() != expected {
		t.Errorf("wrong ID: expected=%#v had=%#v", expected, b.GetIdentifier())
	}
}

func TestBase_Log(t *testing.T) {
	logfile := path.Join(os.TempDir(), "log.unit_test")
	log.Output = logfile
	log.Format = "{message}"
	log.Level = log.DEBUG
	log.Open()
	defer os.Remove(logfile)

	b := Base{
		rule:     "rulename",
		name:     "filtername",
		id:       1,
		bus:      nil,
		negative: false,
		cbFilter: nil,
	}

	msg := fmt.Sprintf("this is %s", b.name)
	expected := fmt.Sprintf("[%s::%s] this is %s\n", b.rule, b.name, b.name)

	b.Log(msg)
	dat, _ := ioutil.ReadFile(logfile)
	if string(dat) != expected {
		t.Errorf("wrong error: expected=%#v had=%#v", expected, string(dat))
	}
	log.Close()
}

func TestBase_Name(t *testing.T) {
	b := Base{
		rule:     "rulename",
		name:     "filtername",
		id:       1,
		bus:      nil,
		negative: false,
		cbFilter: nil,
	}
	expected := "filtername"
	if b.Name() != expected {
		t.Errorf("wrong name: expected=%#v had=%#v", expected, b.Name())
	}
}

func TestBase_Pipe(t *testing.T) {
	type Test struct {
		Name            string
		RuleName        string
		FilterName      string
		Config          map[string]string
		ID              int32
		Negative        bool
		Callback        func(msg *data.Message) (bool, error)
		ExpectedMessage *data.Message
		ExpectedLog     string
	}

	bus := NewFakeBus()
	m := data.NewMessage("test")
	logfile := path.Join(os.TempDir(), "log.unit_test")
	log.Output = logfile
	log.Level = log.INFO
	log.Format = "{message}"

	defer os.Remove(logfile)

	callback := func(msg *data.Message) (bool, error) {
		if msg.GetMessage() == "error" {
			return false, fmt.Errorf("triggered error")
		}
		if msg.GetMessage() != "test" {
			t.Errorf("wrong msg: expected=%#v had=%#v", m, msg)
		}
		return true, nil
	}

	tests := []Test{
		{"CallbackError", "Rule1", "echo", map[string]string{}, 1, false, callback, data.NewMessage("error"), "[Rule1::echo] triggered error\n"},
		{"Negative", "Rule1", "echo", map[string]string{}, 1, false, callback, data.NewMessage("test"), ""},
	}

	for _, v := range tests {
		log.Open()

		b := Base{
			rule:     v.RuleName,
			name:     v.FilterName,
			id:       v.ID,
			bus:      bus,
			negative: v.Negative,
			cbFilter: v.Callback,
		}
		b.Pipe(v.ExpectedMessage)
		dat, _ := ioutil.ReadFile(logfile)
		if string(dat) != v.ExpectedLog {
			t.Errorf("%s : wrong error: expected=%#v had=%#v", v.Name, v.ExpectedLog, string(dat))
		}

		if len(bus.Collected) > 0 && bus.Collected[0].GetTarget("rule_name") != v.RuleName {
			t.Errorf("%s : wrong message: expected=%#v had=%#v", v.Name, v.RuleName, bus.Collected[0].GetTarget("rule_name"))
		}

		log.Close()
		os.Remove(logfile)
	}

}

func TestBase_Propagate(t *testing.T) {
	type Test struct {
		Name            string
		RuleName        string
		FilterName      string
		Config          map[string]string
		ID              int32
		Negative        bool
		Callback        func(msg *data.Message) (bool, error)
		ExpectedMessage *data.Message
		ExpectedLog     string
	}

	bus := NewFakeBus()
	tests := []Test{
		{"Single", "Rule1", "echo", map[string]string{}, 1, false, nil, data.NewMessage("error"), "[Rule1::echo] triggered error\n"},
	}

	for _, v := range tests {
		b := Base{
			rule:     v.RuleName,
			name:     v.FilterName,
			id:       v.ID,
			bus:      bus,
			negative: v.Negative,
			cbFilter: v.Callback,
		}
		b.Propagate(v.ExpectedMessage)

		if bus.Collected[0].GetTarget("rule_name") != v.RuleName {
			t.Errorf("wrong message: expected=%#v had=%#v", v.RuleName, bus.Collected[0].GetTarget("rule_name"))
		}
	}
}

func TestBase_Rule(t *testing.T) {
	b := Base{
		rule:     "rulename",
		name:     "filtername",
		id:       1,
		bus:      nil,
		negative: false,
		cbFilter: nil,
	}
	expected := "rulename"
	if b.Rule() != expected {
		t.Errorf("wrong rule: expected=%#v had=%#v", expected, b.Rule())
	}
}

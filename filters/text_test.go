package filters

import (
	"github.com/asaskevich/EventBus"
	"regexp"
	"testing"

	"github.com/Matrix86/driplane/data"

	"github.com/stretchr/testify/assert"
)

func TestNewTextFilter(t *testing.T) {
	type Test struct {
		Name           string
		Conf           map[string]string
		ExpectedFilter *Text
		ExpectedError  string
	}
	tests := []Test{
		{"PatternNotSpecified", map[string]string{"extract": "false", "regexp": "true", "target": "main"}, nil, "pattern field is required on textfilter"},
		{"WrongRegex", map[string]string{"pattern": "^\\/(?!\\/)(.*?)", "extract": "false", "regexp": "true", "target": "main"}, nil, "textfilter: cannot compile the regular expression '^\\/(?!\\/)(.*?)'"},
		{"FilterRegexpOk", map[string]string{"pattern": "test.*", "extract": "true", "regexp": "true", "target": "main"}, &Text{pattern: "test.*", target: "main", extract: true, regexp: regexp.MustCompile("test.*"), enableRegexp: true}, ""},
		{"FilterNoRegexpOk", map[string]string{"pattern": "test.*", "extract": "true", "regexp": "false", "target": "main"}, &Text{pattern: "test.*", target: "main", extract: true, regexp: nil, enableRegexp: false}, ""},
	}

	for _, v := range tests {
		filter, err := NewTextFilter(v.Conf)
		if v.ExpectedError == "" && err != nil {
			t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
		}
		if filter != nil && v.ExpectedFilter != nil {
			if e, ok := filter.(*Text); ok {
				if v.ExpectedFilter.extract != e.extract {
					t.Errorf("%s: wrong extractText: expected=%#v had=%#v", v.Name, v.ExpectedFilter.extract, e.extract)
				}
				if v.ExpectedFilter.regexp != nil && v.ExpectedFilter.regexp.String() != e.regexp.String() {
					t.Errorf("%s: wrong regexp: expected=%#v had=%#v", v.Name, v.ExpectedFilter.regexp.String(), e.regexp.String())
				}
				if v.ExpectedFilter.target != e.target {
					t.Errorf("%s: wrong target: expected=%#v had=%#v", v.Name, v.ExpectedFilter.target, e.target)
				}
				if v.ExpectedFilter.pattern != e.pattern {
					t.Errorf("%s: wrong text: expected=%#v had=%#v", v.Name, v.ExpectedFilter.pattern, e.pattern)
				}
			}
		} else if (filter != nil && v.ExpectedFilter == nil) || (filter == nil && v.ExpectedFilter != nil) {
			t.Errorf("%s: wrong pointer: expected=%#v had=%#v", v.Name, v.ExpectedFilter, filter)
		}
	}
}

func TestText_DoFilter(t *testing.T) {
	type Test struct {
		Name             string
		Conf             map[string]string
		Message          *data.Message
		MultipleMessage  bool
		ExpectedBool     bool
		ExpectedError    string
		ExpectedMessages []*data.Message
	}
	tests := []Test{
		{"TypeNotSupported", map[string]string{"pattern":""}, data.NewMessage(false), false, false, "received data is not a string", nil},
		{"TypeString", map[string]string{"pattern": "test"}, data.NewMessage("test"), false, true, "", []*data.Message{data.NewMessage("test")}},
		{"TypeByte", map[string]string{"pattern": "test"}, data.NewMessage([]byte("test")), false, true, "", []*data.Message{data.NewMessage([]byte("test"))}},
		{"TargetNotFound", map[string]string{"pattern": "test", "target": "notexist"}, data.NewMessage("test"), false,  false, "target not found", []*data.Message{}},
		{"Regexp", map[string]string{"pattern": "[a-z]est", "regexp": "true"}, data.NewMessage("test"), false,  true, "", []*data.Message{data.NewMessage("test")}},
		{"ExtractSingle", map[string]string{"pattern": "([a-z]est)", "regexp": "true", "extract": "true"}, data.NewMessage("test here"), false,  true, "", []*data.Message{data.NewMessageWithExtra("test", map[string]interface{}{"fulltext": "test here"})}},
		{"ExtractMultiple", map[string]string{"pattern": "([a-z]est)", "regexp": "true", "extract": "true"}, data.NewMessage("test here beast fast best"), true,  false, "", []*data.Message{data.NewMessageWithExtra("test", map[string]interface{}{"fulltext": "test here beast fast best", "rule_name":""}), data.NewMessageWithExtra("best", map[string]interface{}{"fulltext": "test here beast fast best", "rule_name":""})}},

	}

	fb := NewFakeBus()

	for _, v := range tests {
		filter, err := NewTextFilter(v.Conf)
		if err != nil {
			t.Errorf("constructor returned '%s'", err)
			return
		}
		filter.setBus(EventBus.Bus(fb))
		if e, ok := filter.(*Text); ok {
			orig := v.Message.Clone()
			hadBool, err := e.DoFilter(orig)

			if hadBool != v.ExpectedBool {
				t.Errorf("%s: wrong bool: expected=%#v had=%#v", v.Name, v.ExpectedBool, hadBool)
			}

			if v.ExpectedError == "" {
				if err != nil {
					t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
				}
				if v.MultipleMessage {
					if assert.Equal(t, fb.Collected, v.ExpectedMessages) == false {
						t.Errorf("%s: wrong: expected=%#v had=%#v", v.Name, v.ExpectedMessages, fb.Collected)
					}
				} else {
					if assert.Equal(t, v.ExpectedMessages[0], orig) == false {
						t.Errorf("%s: wrong: expected=%#v had=%#v", v.Name, v.ExpectedMessages, fb.Collected)
					}
				}
			} else {
				if err == nil || err.Error() != v.ExpectedError {
					t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err)
				}
			}
		}
	}
}

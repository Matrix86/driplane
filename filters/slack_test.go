package filters

import (
	"github.com/Matrix86/driplane/data"
	"html/template"
	"reflect"
	"testing"
)

func TestNewSlackFilter(t *testing.T) {
	type Test struct {
		Name           string
		Conf           map[string]string
		ExpectedFilter *Slack
		ExpectedError  string
	}
	tpl, _ := template.New("test").Parse("test")
	tests := []Test{
		{"ActionSendMessage", map[string]string{"action": "send_message", "target": "main", "token": "token", "text": "test", "to": "test", "filename": "test", "url": "test"}, &Slack{action: "send_message", target: "main", token: "token", to: tpl, filename: tpl, body: tpl, downloadURL: tpl}, ""},
		{"ToNotSpecified", map[string]string{"action": "send_message"}, nil, "destination 'to' is mandatory with this action"},
	}

	for _, v := range tests {
		filter, err := NewSlackFilter(v.Conf)
		if v.ExpectedError == "" {
			if err != nil {
				t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
			}
		} else {
			if err != nil {
				if err.Error() != v.ExpectedError {
					t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
				}
			} else {
				t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, nil)
			}
		}
		if v.ExpectedFilter != nil {
			if filter != nil {
				if e, ok := filter.(*Slack); ok {
					if v.ExpectedFilter.target != e.target {
						t.Errorf("%s: wrong target: expected=%#v had=%#v", v.Name, v.ExpectedFilter.target, e.target)
					}
					if v.ExpectedFilter.action != e.action {
						t.Errorf("%s: wrong action: expected=%#v had=%#v", v.Name, v.ExpectedFilter.action, e.action)
					}
					if v.ExpectedFilter.token != e.token {
						t.Errorf("%s: wrong token: expected=%#v had=%#v", v.Name, v.ExpectedFilter.token, e.token)
					}
					if v.ExpectedFilter.filename != nil && reflect.DeepEqual(v.ExpectedFilter.filename, e.filename) {
						t.Errorf("%s: wrong filename: expected=%#v had=%#v", v.Name, v.ExpectedFilter.filename, e.filename)
					}
					if v.ExpectedFilter.downloadURL != nil && reflect.DeepEqual(v.ExpectedFilter.downloadURL, e.downloadURL) {
						t.Errorf("%s: wrong downloadURL: expected=%#v had=%#v", v.Name, v.ExpectedFilter.downloadURL, e.downloadURL)
					}
					if v.ExpectedFilter.body != nil && reflect.DeepEqual(v.ExpectedFilter.body, e.body) {
						t.Errorf("%s: wrong body: expected=%#v had=%#v", v.Name, v.ExpectedFilter.body, e.body)
					}
					if v.ExpectedFilter.to != nil && reflect.DeepEqual(v.ExpectedFilter.to, e.to) {
						t.Errorf("%s: wrong to: expected=%#v had=%#v", v.Name, v.ExpectedFilter.to, e.to)
					}
				}
			} else {
				t.Errorf("%s: wrong filter: expected=%#v had=%#v", v.Name, v.ExpectedFilter, nil)
			}
		} else {
			if filter != nil {
				t.Errorf("%s: wrong pointer: expected=%#v had=%#v", v.Name, v.ExpectedFilter, filter)
			}
		}
	}
}

func TestSlack_DoFilter(t *testing.T) {
	type Test struct {
		Name          string
		Conf          map[string]string
		Message       *data.Message
		ExpectedBool  bool
		ExpectedError string
	}
	tests := []Test{
		{"NotTokenFound", map[string]string{"action": "send_message", "to": "test"}, data.NewMessage(""), false, "slack bot token not found"},
		{"SendMessageWithText", map[string]string{"action": "send_message", "to": "test", "text": "test", "token": "token"}, data.NewMessage(""), false, "sendMessage: slack returned: invalid_auth"},
		{"SendMessageFromMain", map[string]string{"action": "send_message", "to": "test", "target": "main", "token": "token"}, data.NewMessage(""), false, "sendMessage: slack returned: invalid_auth"},
	}

	for _, v := range tests {
		filter, _ := NewSlackFilter(v.Conf)

		b, err := filter.DoFilter(v.Message)
		if v.ExpectedError == "" {
			if err != nil {
				t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
			}
		} else {
			if err != nil {
				if err.Error() != v.ExpectedError {
					t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
				}
			} else {
				t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, nil)
			}
			if b != v.ExpectedBool {
				t.Errorf("%s: wrong bool: expected=%#v had=%#v", v.Name, v.ExpectedBool, b)
			}
		}
	}
}

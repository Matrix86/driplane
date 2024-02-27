package filters

import (
	"fmt"
	"testing"

	"github.com/asaskevich/EventBus"

	"github.com/Matrix86/driplane/data"

	"github.com/stretchr/testify/assert"
)

func TestNewNumberFilter(t *testing.T) {
	type Test struct {
		Name           string
		Conf           map[string]string
		ExpectedFilter *Number
		ExpectedError  string
	}
	tests := []Test{
		{"ValueNotNumeric", map[string]string{"value": "not a number", "target": "main"}, nil, "value parameter is not numeric: strconv.ParseFloat: parsing \"not a number\": invalid syntax"},
		{"ValueNumeric", map[string]string{"value": "10.3", "target": "main"}, &Number{value: "10.3", fValue: 10.3, target: "main", operator: ">"}, ""},
		{"ChangeTarget", map[string]string{"value": "10", "target": "another"}, &Number{value: "10", fValue: 10, target: "another", operator: ">"}, ""},
		{"ChangeOperator", map[string]string{"value": "10.3", "target": "another", "op": "<="}, &Number{value: "10.3", fValue: 10.3, target: "another", operator: "<="}, ""},
		{"WrongOperator", map[string]string{"value": "10.3", "target": "another", "op": "&"}, nil, "operator '&' not supported"},
	}

	for _, v := range tests {
		filter, err := NewNumberFilter(v.Conf)
		if v.ExpectedError == "" && err != nil {
			t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
		} else if err != nil && err.Error() != v.ExpectedError {
			t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err.Error())
		}
		if filter != nil && v.ExpectedFilter != nil {
			if e, ok := filter.(*Number); ok {
				if v.ExpectedFilter.operator != e.operator {
					t.Errorf("%s: wrong operator: expected=%#v had=%#v", v.Name, v.ExpectedFilter.operator, e.operator)
				}
				if v.ExpectedFilter.value != e.value {
					t.Errorf("%s: wrong value: expected=%#v had=%#v", v.Name, v.ExpectedFilter.value, e.value)
				}
				if v.ExpectedFilter.fValue != e.fValue {
					t.Errorf("%s: wrong fValue: expected=%#v had=%#v", v.Name, v.ExpectedFilter.fValue, e.fValue)
				}
				if v.ExpectedFilter.target != e.target {
					t.Errorf("%s: wrong target: expected=%#v had=%#v", v.Name, v.ExpectedFilter.target, e.target)
				}
			}
		} else if (filter != nil && v.ExpectedFilter == nil) || (filter == nil && v.ExpectedFilter != nil) {
			t.Errorf("%s: wrong pointer: expected=%#v had=%#v", v.Name, v.ExpectedFilter, filter)
		}
	}
}

func TestNumber_DoFilter(t *testing.T) {
	type Test struct {
		Name             string
		Conf             map[string]string
		Message          *data.Message
		MultipleMessage  bool
		ExpectedBool     bool
		ExpectedError    error
		ExpectedMessages []*data.Message
	}
	tests := []Test{
		{"TypeNotSupported", map[string]string{}, data.NewMessage([]int{1}), false, false, fmt.Errorf("received data is not a string"), nil},
		{"TypeString", map[string]string{}, data.NewMessage("10.6"), false, true, nil, []*data.Message{data.NewMessage("10.6")}},
		{"TypeByte", map[string]string{}, data.NewMessage([]byte("8")), false, true, nil, []*data.Message{data.NewMessage([]byte("8"))}},
		{"TypeNotCastable", map[string]string{}, data.NewMessage("not a number"), false, false, fmt.Errorf("received data is not a numeric value: strconv.ParseFloat: parsing \"not a number\": invalid syntax"), nil},
		{"TargetNotFound", map[string]string{"target": "none", "value": "10"}, data.NewMessage("9"), false, false, nil, nil},
		{"OpGreaterTrue", map[string]string{"op": ">", "value": "10"}, data.NewMessage("20"), false, true, nil, []*data.Message{data.NewMessage("20")}},
		{"OpGreaterFalse", map[string]string{"op": ">", "value": "10"}, data.NewMessage("9"), false, false, nil, nil},
		{"OpGreaterEqTrue", map[string]string{"op": ">=", "value": "10"}, data.NewMessage("20"), false, true, nil, []*data.Message{data.NewMessage("20")}},
		{"OpGreaterEqTrue2", map[string]string{"op": ">=", "value": "10"}, data.NewMessage("10"), false, true, nil, []*data.Message{data.NewMessage("10")}},
		{"OpGreaterEqFalse", map[string]string{"op": ">=", "value": "10"}, data.NewMessage("9"), false, false, nil, nil},
		{"OpLessTrue", map[string]string{"op": "<", "value": "10"}, data.NewMessage("9"), false, true, nil, []*data.Message{data.NewMessage("9")}},
		{"OpLessFalse", map[string]string{"op": "<", "value": "10"}, data.NewMessage("10"), false, false, nil, nil},
		{"OpLessEqTrue", map[string]string{"op": "<=", "value": "10"}, data.NewMessage("9.9999"), false, true, nil, []*data.Message{data.NewMessage("9.9999")}},
		{"OpLessEqTrue2", map[string]string{"op": "<=", "value": "10"}, data.NewMessage("10"), false, true, nil, []*data.Message{data.NewMessage("10")}},
		{"OpLessEqFalse", map[string]string{"op": "<=", "value": "10"}, data.NewMessage("90"), false, false, nil, nil},
		{"OpNotEqualTrue", map[string]string{"op": "!=", "value": "10.5"}, data.NewMessage("10"), false, true, nil, []*data.Message{data.NewMessage("10")}},
		{"OpNotEqualFalse", map[string]string{"op": "!=", "value": "10.5"}, data.NewMessage("10.5"), false, false, nil, nil},
		{"OpEqualTrue", map[string]string{"op": "==", "value": "10.5"}, data.NewMessage("10.5"), false, true, nil, []*data.Message{data.NewMessage("10.5")}},
		{"OpEqualFalse", map[string]string{"op": "==", "value": "10.5"}, data.NewMessage("10"), false, false, nil, nil},
	}

	fb := NewFakeBus()

	for _, v := range tests {
		filter, err := NewNumberFilter(v.Conf)
		if err != nil {
			t.Errorf("%s: constructor returned '%s'", v.Name, err)
			return
		}
		filter.setBus(EventBus.Bus(fb))
		if e, ok := filter.(*Number); ok {
			orig := v.Message.Clone()
			hadBool, err := e.DoFilter(orig)

			if hadBool != v.ExpectedBool {
				t.Errorf("%s: wrong bool: expected=%#v had=%#v", v.Name, v.ExpectedBool, hadBool)
			}

			if v.ExpectedError == nil {
				if err != nil {
					t.Errorf("%s: wrong error: expected=nil had=%#v", v.Name, err)
				}
				if v.MultipleMessage {
					if assert.Equal(t, fb.Collected, v.ExpectedMessages) == false {
						t.Errorf("%s: wrong: expected=%#v had=%#v", v.Name, v.ExpectedMessages, fb.Collected)
					}
				} else {
					if len(v.ExpectedMessages) != 0 && assert.Equal(t, v.ExpectedMessages[0], orig) == false {
						t.Errorf("%s: wrong: expected=%#v had=%#v", v.Name, v.ExpectedMessages, fb.Collected)
					}
				}
			} else {
				if err == nil || err.Error() != v.ExpectedError.Error() {
					t.Errorf("%s: wrong error: expected=%#v had=%#v", v.Name, v.ExpectedError, err)
				}
			}
		}
	}
}

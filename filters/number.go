package filters

import (
	"fmt"
	"strconv"

	"github.com/Matrix86/driplane/data"
)

// Number is a Filter to treat a string from the input Message as numeric value and apply some operator on it
type Number struct {
	Base

	target   string
	operator string
	value    string
	fValue   float64

	params map[string]string
}

// NewNumberFilter is the registered method to instantiate a TextFilter
func NewNumberFilter(p map[string]string) (Filter, error) {
	f := &Number{
		params:   p,
		target:   "main",
		operator: ">",
		value:    "0",
		fValue:   0.0,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["value"]; ok {
		if value, err := strconv.ParseFloat(v, 64); err != nil {
			return nil, fmt.Errorf("value parameter is not numeric: %s", err)
		} else {
			f.value = v
			f.fValue = value
		}
	}

	if v, ok := f.params["target"]; ok {
		f.target = v
	}

	if v, ok := f.params["op"]; ok {
		switch v {
		case ">", ">=", "<", "<=", "!=", "==":
			f.operator = v

		default:
			return nil, fmt.Errorf("operator '%s' not supported", v)
		}

	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Number) DoFilter(msg *data.Message) (bool, error) {
	var text string
	target := msg.GetTarget(f.target)
	if target == nil {
		return false, nil
	}

	if v, ok := target.(string); ok {
		text = v
	} else if v, ok := target.([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}

	var currentValue float64

	if value, err := strconv.ParseFloat(text, 64); err != nil {
		return false, fmt.Errorf("received data is not a numeric value: %s", err)
	} else {
		currentValue = value
	}

	switch f.operator {
	case ">":
		if currentValue > f.fValue {
			return true, nil
		}

	case ">=":
		if currentValue >= f.fValue {
			return true, nil
		}

	case "<":
		if currentValue < f.fValue {
			return true, nil
		}

	case "<=":
		if currentValue <= f.fValue {
			return true, nil
		}

	case "!=":
		if currentValue != f.fValue {
			return true, nil
		}

	case "==":
		if currentValue == f.fValue {
			return true, nil
		}
	}

	return false, nil
}

// OnEvent is called when an event occurs
func (f *Number) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("number", NewNumberFilter)
}

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
	var currentValue float64

	target := msg.GetTarget(f.target)
	if target == nil {
		return false, nil
	}

	switch v := target.(type) {
	case string:
		text = v
		if value, err := strconv.ParseFloat(text, 64); err != nil {
			return false, fmt.Errorf("received data is not a numeric value: %s", err)
		} else {
			currentValue = value
		}

	case []byte:
		text = string(v)
		if value, err := strconv.ParseFloat(text, 64); err != nil {
			return false, fmt.Errorf("received data is not a numeric value: %s", err)
		} else {
			currentValue = value
		}

	case int:
		currentValue = float64(v)
	case int8:
		currentValue = float64(v)
	case int16:
		currentValue = float64(v)
	case int32:
		currentValue = float64(v)
	case int64:
		currentValue = float64(v)
	case float32:
		currentValue = float64(v)
	case float64:
		currentValue = float64(v)

	default:
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
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

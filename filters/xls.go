package filters

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"github.com/Matrix86/driplane/data"

	"github.com/xuri/excelize/v2"
)

// XLS is a Filter that extract all the rows from an Excel file
type XLS struct {
	Base

	target   string
	filename *template.Template

	params map[string]string
}

// NewXLSFilter is the registered method to instantiate a XLSFilter
func NewXLSFilter(p map[string]string) (Filter, error) {
	f := &XLS{
		params: p,
		target: "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["filename"]; ok {
		t, err := template.New("xlsFilterFilename").Parse(v)
		if err != nil {
			return nil, err
		}
		f.filename = t
	} else if v, ok := p["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *XLS) DoFilter(msg *data.Message) (bool, error) {
	if f.filename != nil {
		text, err := msg.ApplyPlaceholder(f.filename)
		if err != nil {
			return false, err
		}

		x, err := excelize.OpenFile(text)
		if err != nil {
			return false, err
		}
		defer x.Close()

		sheets := x.GetSheetList()
		for _, sName := range sheets {
			rows, err := x.GetRows(sName)
			if err != nil {
				return false, err
			}
			for _, row := range rows {
				cloned := msg.Clone()
				cloned.SetMessage(strings.Join(row, ","))
				cloned.SetExtra("xls_sheet", sName)
				cloned.SetExtra("xls_filename", text)
				f.Propagate(cloned)
			}
		}
	} else {
		if _, ok := msg.GetTarget(f.target).([]byte); !ok {
			// ERROR this filter can't be used with different types
			return false, fmt.Errorf("received data is not supported")
		}

		buf := bytes.NewBuffer(msg.GetTarget(f.target).([]byte))

		x, err := excelize.OpenReader(bytes.NewReader(buf.Bytes()))
		if err != nil {
			return false, err
		}
		defer x.Close()

		sheets := x.GetSheetList()
		for _, sName := range sheets {
			rows, err := x.GetRows(sName)
			if err != nil {
				return false, err
			}
			for _, row := range rows {
				cloned := msg.Clone()
				cloned.SetMessage(strings.Join(row, ","))
				cloned.SetExtra("xls_sheet", sName)
				f.Propagate(cloned)
			}
		}
	}

	return true, nil
}

// OnEvent is called when an event occurs
func (f *XLS) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("xls", NewXLSFilter)
}

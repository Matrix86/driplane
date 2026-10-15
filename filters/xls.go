package filters

import (
	"bytes"
	"fmt"
	"os"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"

	"github.com/xuri/excelize/v2"
)

// oleHeader is the signature of the OLE compound file used by Excel to store encrypted workbooks
var oleHeader = []byte{0xd0, 0xcf, 0x11, 0xe0, 0xa1, 0xb1, 0x1a, 0xe1}

// XLS is a Filter that extract all the rows from an Excel file
type XLS struct {
	Base

	target   string
	filename *template.Template
	password string

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
		t, err := template.New("xlsFilterFilename").Funcs(sprig.FuncMap()).Parse(v)
		if err != nil {
			return nil, err
		}
		f.filename = t
	} else if v, ok := p["target"]; ok {
		f.target = v
	}

	if v, ok := p["password"]; ok {
		f.password = v
	}

	return f, nil
}

// open parses the workbook content, decrypting it if a password has been specified
func (f *XLS) open(content []byte) (*excelize.File, error) {
	if f.password == "" {
		return excelize.OpenReader(bytes.NewReader(content))
	}

	if !bytes.HasPrefix(content, oleHeader) {
		log.Warning("[%s::%s] password specified but the file is not protected, opening it anyway", f.Rule(), f.Name())
		return excelize.OpenReader(bytes.NewReader(content))
	}

	x, err := excelize.OpenReader(bytes.NewReader(content), excelize.Options{Password: f.password})
	if err != nil {
		return nil, fmt.Errorf("cannot decrypt the file (wrong password?): %w", err)
	}
	return x, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *XLS) DoFilter(msg *data.Message) (bool, error) {
	var content []byte
	var filename string

	if f.filename != nil {
		text, err := msg.ApplyPlaceholder(f.filename)
		if err != nil {
			return false, err
		}
		content, err = os.ReadFile(text)
		if err != nil {
			return false, err
		}
		filename = text
	} else {
		b, ok := msg.GetTarget(f.target).([]byte)
		if !ok {
			// ERROR this filter can't be used with different types
			return false, fmt.Errorf("received data is not supported")
		}
		content = b
	}

	x, err := f.open(content)
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
			if filename != "" {
				cloned.SetExtra("xls_filename", filename)
			}
			f.Propagate(cloned)
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

package filters

import (
	"bytes"
	"fmt"
	"text/template"

	"github.com/Matrix86/driplane/data"

	"github.com/ledongthuc/pdf"
)

// PDF is a Filter that extract plain text from a PDF file
type PDF struct {
	Base

	target   string
	filename *template.Template

	params map[string]string
}

// NewPDFFilter is the registered method to instantiate a TextFilter
func NewPDFFilter(p map[string]string) (Filter, error) {
	f := &PDF{
		params: p,
		target: "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["filename"]; ok {
		t, err := template.New("pdfFilterFilename").Parse(v)
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
func (f *PDF) DoFilter(msg *data.Message) (bool, error) {
	if f.filename != nil {
		text, err := msg.ApplyPlaceholder(f.filename)
		if err != nil {
			return false, err
		}

		h, r, err := pdf.Open(text)
		// remember close file
		defer h.Close()

		if err != nil {
			return false, err
		}

		var buf bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			return false, err
		}
		buf.ReadFrom(b)
		plain := buf.String()

		msg.SetMessage(plain)
		msg.SetExtra("fulltext", text)
	} else {
		if _, ok := msg.GetTarget(f.target).([]byte); !ok {
			// ERROR this filter can't be used with different types
			return false, fmt.Errorf("received data is not supported")
		}

		buf := bytes.NewBuffer(msg.GetTarget(f.target).([]byte))

		r, err := pdf.NewReader(bytes.NewReader(buf.Bytes()), int64(buf.Len()))
		if err != nil {
			return false, err
		}

		var buff bytes.Buffer
		b, err := r.GetPlainText()
		if err != nil {
			return false, err
		}
		buff.ReadFrom(b)
		plain := buff.String()

		msg.SetMessage(plain)
		msg.SetExtra("fulltext", msg.GetTarget("main"))
	}

	return true, nil
}

// OnEvent is called when an event occurs
func (f *PDF) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("pdf", NewPDFFilter)
}

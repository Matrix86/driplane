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

	filename *template.Template

	params map[string]string
}

// NewPDFFilter is the registered method to instantiate a TextFilter
func NewPDFFilter(p map[string]string) (Filter, error) {
	f := &PDF{
		params:   p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["filename"]; ok {
		t, err := template.New("pdfFilterFilename").Parse(v)
		if err != nil {
			return nil, err
		}
		f.filename = t
	} else {
		return nil, fmt.Errorf("filename is a mandatory field")
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *PDF) DoFilter(msg *data.Message) (bool, error) {
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

	return true, nil
}

// Set the name of the filter
func init() {
	register("pdf", NewPDFFilter)
}

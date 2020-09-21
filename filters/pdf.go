package filters

import (
	"bytes"

	"github.com/Matrix86/driplane/data"

	"github.com/ledongthuc/pdf"
)

// PDF is a Filter that extract plain text from a PDF file
type PDF struct {
	Base

	target string

	params map[string]string
}

// NewPDFFilter is the registered method to instantiate a TextFilter
func NewPDFFilter(p map[string]string) (Filter, error) {
	f := &PDF{
		params: p,
		target: "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *PDF) DoFilter(msg *data.Message) (bool, error) {
	var text string
	if f.target == "main" {
		text = msg.GetMessage()
	} else if v, ok := msg.GetExtra()[f.target]; ok {
		text = v
	} else {
		return false, nil
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

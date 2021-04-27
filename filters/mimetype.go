package filters

import (
	"bytes"
	"fmt"
	"github.com/Matrix86/driplane/data"
	"text/template"

	"github.com/gabriel-vasile/mimetype"
)

// Mimetype is a Filter to detect the format of an input
type Mimetype struct {
	Base

	target   string
	filename *template.Template

	params map[string]string
}

// NewMimetypeFilter is the registered method to instantiate a MimetypeFilter
func NewMimetypeFilter(p map[string]string) (Filter, error) {
	f := &Mimetype{
		params:   p,
		target: "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["filename"]; ok {
		t, err := template.New("mimeFilterFilename").Parse(v)
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
func (f *Mimetype) DoFilter(msg *data.Message) (bool, error) {
	if f.filename != nil {
		text, err := msg.ApplyPlaceholder(f.filename)
		if err != nil {
			return false, err
		}

		mime, err := mimetype.DetectFile(text)
		if err != nil {
			return false, err
		}
		msg.SetMessage(mime.String())
		msg.SetExtra("mimetype_ext", mime.Extension())
		msg.SetExtra("fulltext", text)
	} else {
		if v, ok := msg.GetTarget(f.target).([]byte); ok {
			buf := bytes.NewBuffer(v)
			mime := mimetype.Detect(buf.Bytes())
			msg.SetMessage(mime.String())
			msg.SetExtra("mimetype_ext", mime.Extension())
			msg.SetExtra("fulltext", msg.GetTarget("main"))
		} else {
			return false, fmt.Errorf("data type is not supported")
		}
	}

	return true, nil
}

// OnEvent is called when an event occurs
func (f *Mimetype) OnEvent(event *data.Event){}

// Set the name of the filter
func init() {
	register("mime", NewMimetypeFilter)
}

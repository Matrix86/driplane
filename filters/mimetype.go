package filters

import (
	"github.com/Matrix86/driplane/data"
	"text/template"

	"github.com/gabriel-vasile/mimetype"
)

// Mimetype is a Filter to detect the format of an input
type Mimetype struct {
	Base

	filename *template.Template

	params map[string]string
}

// NewMimetypeFilter is the registered method to instantiate a MimetypeFilter
func NewMimetypeFilter(p map[string]string) (Filter, error) {
	f := &Mimetype{
		params:   p,
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["filename"]; ok {
		t, err := template.New("mimeFilterFilename").Parse(v)
		if err != nil {
			return nil, err
		}
		f.filename = t
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Mimetype) DoFilter(msg *data.Message) (bool, error) {
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

	return true, nil
}

// Set the name of the filter
func init() {
	register("mime", NewMimetypeFilter)
}

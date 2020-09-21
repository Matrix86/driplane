package filters

import (
	"github.com/Matrix86/driplane/data"

	"github.com/gabriel-vasile/mimetype"
)

// Mimetype is a Filter to detect the format of an input
type Mimetype struct {
	Base

	target     string

	params map[string]string
}

// NewMimetypeFilter is the registered method to instantiate a MimetypeFilter
func NewMimetypeFilter(p map[string]string) (Filter, error) {
	f := &Mimetype{
		params:     p,
		target:     "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := p["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Mimetype) DoFilter(msg *data.Message) (bool, error) {
	var text string

	if f.target == "main" {
		text = msg.GetMessage()
	} else if v, ok := msg.GetExtra()[f.target]; ok {
		text = v
	} else {
		return false, nil
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

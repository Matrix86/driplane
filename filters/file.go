package filters

import (
	"io/ioutil"

	"github.com/Matrix86/driplane/data"
)

// File is a filter that interprets the input Message as a file path, reads it and prints it.
type File struct {
	Base
}

// NewFileFilter is the registered method to instantiate a File filter
func NewFileFilter(p map[string]string) (Filter, error) {
	f := File{}
	f.cbFilter = f.DoFilter
	return &f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *File) DoFilter(msg *data.Message) (bool, error) {
	path := msg.GetMessage().(string)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return true, err
	}
	msg.SetMessage(string(data))
	return true, nil
}

// OnEvent is called when an event occurs
func (f *File) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("file", NewFileFilter)
}

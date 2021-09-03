package filters

import (
	"io/ioutil"
	"os"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
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
	// if the message data is a string
	if path, ok := msg.GetMessage().(string); ok {
		// if the path exists and it's a file
		if stat, err := os.Stat(path); err == nil && !stat.IsDir() {
			log.Debug("path='%s' size=%d extra=%v", path, stat.Size(), msg.GetExtra())
			readData, err := ioutil.ReadFile(path)
			if err != nil {
				return true, err
			}
			msg.SetMessage(string(readData))
			return true, nil
		} else {
			log.Info("%s is not a file", path)
		}
	}
	return false, nil
}

// OnEvent is called when an event occurs
func (f *File) OnEvent(event *data.Event) {}

// Set the name of the filter
func init() {
	register("file", NewFileFilter)
}

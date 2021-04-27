package feeders

import (
	"fmt"
	"io"
	"os"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
	"github.com/hpcloud/tail"
)

// File is a Feeder that creates a stream from a file
type File struct {
	Base

	filename  string
	lastLines bool

	fp *tail.Tail
}

// NewFileFeeder is the registered method to instantiate a FileFeeder
func NewFileFeeder(conf map[string]string) (Feeder, error) {
	f := &File{
		lastLines: false,
	}

	if val, ok := conf["file.filename"]; ok {
		f.filename = val
	}
	if val, ok := conf["file.toend"]; ok && val == "true" {
		f.lastLines = true
	}

	info, err := os.Stat(f.filename)
	if os.IsNotExist(err) || info.IsDir() {
		return nil, fmt.Errorf("file '%s' does not exist", f.filename)
	}

	seek := tail.SeekInfo{
		Offset: 0,
		Whence: io.SeekStart,
	}
	if f.lastLines {
		seek.Offset = info.Size()
	}

	f.fp, err = tail.TailFile(f.filename, tail.Config{
		Logger:   tail.DiscardingLogger,
		Follow:   true,
		Location: &seek,
	})
	if err != nil {
		return nil, err
	}

	return f, nil
}

// Start propagates a message every time a new line is read
func (f *File) Start() {
	go func() {
		for line := range f.fp.Lines {
			msg := data.NewMessage(line.Text)
			msg.SetExtra("file_name", f.filename)
			f.Propagate(msg)
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *File) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.fp.Stop()
	f.fp.Cleanup()
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *File) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("file", NewFileFeeder)
}

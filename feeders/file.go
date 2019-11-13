package feeders

import (
	"fmt"
	"io"
	"os"

	"github.com/Matrix86/driplane/com"

	"github.com/evilsocket/islazy/log"
	"github.com/hpcloud/tail"
)

type File struct {
	Base

	filename  string
	lastLines bool

	fp *tail.Tail
}

func NewFileFeeder(conf map[string]string) (Feeder, error) {
	f := &File{
		lastLines:false,
	}

	if val, ok := conf["file.filename"]; ok {
		f.filename = val
	}
	if val, ok := conf["file.toend"]; ok && val == "true" {
		f.lastLines = true
	}

	info, err := os.Stat(f.filename)
	if os.IsNotExist(err) || info.IsDir() {
		return nil, fmt.Errorf("file '%s' does not exist")
	}

	seek := tail.SeekInfo{
		Offset: 0,
		Whence: io.SeekStart,
	}
	if f.lastLines {
		seek.Offset = info.Size()
	}

	f.fp, err = tail.TailFile(f.filename, tail.Config{
		Logger: tail.DiscardingLogger,
		Follow: true,
		Location: &seek,
	})
	if err != nil {
		return nil, err
	}

	return f, nil
}

func (f *File) Start() {
	go func() {
		for line := range f.fp.Lines {
			var msg com.DataMessage
			msg.SetMessage(line.Text)
			f.Propagate(msg)
		}
	}()

	f.isRunning = true
}

func (f *File) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.fp.Stop()
	f.fp.Cleanup()
	f.isRunning = false
}

// Auto factory adding
func init() {
	register("file", NewFileFeeder)
}
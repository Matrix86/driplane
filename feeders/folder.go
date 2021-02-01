package feeders

import (
	"github.com/Matrix86/driplane/data"
	"github.com/dietsche/rfsnotify"
	"github.com/evilsocket/islazy/log"
	"path"
	"path/filepath"
	"strings"
)

// Folder is a Feeder that creates a stream from a folder
type Folder struct {
	Base

	folderName string
	watcher    *rfsnotify.RWatcher
}

// NewFolderFeeder is the registered method to instantiate a FolderFeeder
func NewFolderFeeder(conf map[string]string) (Feeder, error) {
	f := &Folder{}

	if val, ok := conf["folder.name"]; ok {
		if full, err := filepath.Abs(val); err != nil {
			return nil, err
		} else {
			f.folderName = full
		}
	}

	if watcher, err := rfsnotify.NewWatcher(); err != nil {
		return nil, err
	} else if err = watcher.AddRecursive(f.folderName); err != nil {
		return nil, err
	} else {
		f.watcher = watcher
	}

	return f, nil
}

// Start propagates a message every time a new fs event happens in the folder
func (f *Folder) Start() {
	go func() {
		for event := range f.watcher.Events {
			fileName := event.Name
			if strings.Index(fileName, f.folderName) != 0 {
				fileName = path.Join(f.folderName, fileName)
			}
			msg := data.NewMessage(fileName)
			msg.SetExtra("op", event.Op.String())
			f.Propagate(msg)
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *Folder) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.watcher.Close()
	f.isRunning = false
}

// Auto factory adding
func init() {
	register("folder", NewFolderFeeder)
}

package feeders

import (
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils"

	"github.com/Matrix86/cloudwatcher"
	"github.com/evilsocket/islazy/log"
)

// Folder is a Feeder that creates a stream from a folder
type Folder struct {
	Base

	watcherConfig map[string]string
	folderName    string
	serviceName   string
	frequency     time.Duration
	stopChan      chan bool
	watcher       cloudwatcher.Watcher
}

// NewFolderFeeder is the registered method to instantiate a FolderFeeder
func NewFolderFeeder(conf map[string]string) (Feeder, error) {
	f := &Folder{
		stopChan:      make(chan bool),
		frequency:     2 * time.Second,
		watcherConfig: make(map[string]string),
	}

	for k, v := range conf {
		if k == "folder.name" {
			f.folderName = v
			// Get absolute path only if the service type is local (fsnotify)
			if t, ok := conf["folder.type"]; ok && t == "local" {
				full, err := filepath.Abs(v)
				if err != nil {
					return nil, err
				}
				f.folderName = full
			}
		} else if k == "folder.type" {
			f.serviceName = v
		} else if k == "folder.freq" {
			d, err := time.ParseDuration(v)
			if err != nil {
				return nil, fmt.Errorf("specified frequency cannot be parsed '%s': %s", v, err)
			}
			f.frequency = d
		} else if strings.HasPrefix(k, "folder.") {
			splitted := strings.Split(k, ".")
			if len(splitted) == 2 {
				f.watcherConfig[splitted[1]] = v
			}
		}
	}

	watcher, err := cloudwatcher.New(f.serviceName, f.folderName, f.frequency)
	if err != nil {
		return nil, fmt.Errorf("folder feeder: %s", err)
	}

	err = watcher.SetConfig(f.watcherConfig)
	if err != nil {
		return nil, fmt.Errorf("folder feeder: %s", err)
	}

	f.watcher = watcher

	return f, nil
}

// Start propagates a message every time a new fs event happens in the folder
func (f *Folder) Start() {
	go func() {
		err := f.watcher.Start()
		if err != nil {
			f.isRunning = false
			log.Error("%s: %s", f.Name(), err)
			return
		}

		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case event := <-f.watcher.GetEvents():
				log.Debug("received on folder feed : %#v", event)
				// check if the watcher is on a git repository and it is checking the commit/tag instead of files
				msg := data.NewMessage("")
				flat := utils.FlatStruct(event.Object)
				msg.SetExtra("folder_service", f.serviceName)
				msg.SetExtra("folder_event", event.TypeString())
				if f.serviceName == "git" {
					if v, ok := f.watcherConfig["monitor_type"]; ok && v == "repo" {
						for _, c := range event.Object.(*cloudwatcher.GitObject).Commits {
							msg := data.NewMessage(c.Message)
							msg.SetExtra("git_event_type", event.Key)
							for k, v := range flat {
								msg.SetExtra(strings.Join([]string{"git", k}, "_"), v)
							}
							f.Propagate(msg)
						}
					} else {
						msg.SetMessage(event.Key)
						msg.SetExtra("op", event.TypeString())
						msg.SetExtra("git_event_type", "file")
						for k, v := range flat {
							msg.SetExtra(strings.Join([]string{"git", k}, "_"), v)
						}
						f.Propagate(msg)
					}
				} else {
					msg.SetMessage(event.Key)
					for k, v := range flat {
						msg.SetExtra(k, v)
					}
					f.Propagate(msg)
				}

			case err := <-f.watcher.GetErrors():
				log.Error("%s: %s", f.Name(), err)
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *Folder) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.watcher.Close()
	f.stopChan <- true
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *Folder) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("folder", NewFolderFeeder)
}

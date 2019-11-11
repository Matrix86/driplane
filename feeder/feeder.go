package feeder

import (
	"fmt"
	"github.com/Matrix86/driplane/com"

	"github.com/evilsocket/islazy/log"
)

type FeederFactory func(conf map[string]string) (Feeder, error)

var feederFactories = make(map[string]FeederFactory)

// TODO: change the callback logic with publish/subscribe channels
// something like https://gist.github.com/AmirSoleimani/97298c6a94d83d3672765fb31c23194a
type FeederCallback func(msg com.DataMessage)

type Feeder interface {
	setName(name string)

	Name() string
	Start()
	Stop()
	IsRunning() bool
}

type FeederBase struct {
	Feeder
	com.Subscriber

	name        string
	isRunning   bool
	subscribers []com.DataCallback
}


func (f *FeederBase) SetEventMessageClb(clb com.DataCallback) {
	f.subscribers = append(f.subscribers, clb)
}

func (f *FeederBase) Propagate(data com.DataMessage){
	log.Debug("feeder '%s' received: []", f.Name(), data)
	log.Debug("feeder '%s' propagating to %d subscribers", f.Name(), len(f.subscribers))
	if f.subscribers != nil && len(f.subscribers) > 0 {
		for _, cb := range f.subscribers {
			cb(data)
		}
	}
}

func (f *FeederBase) setName(name string) {
	f.name = name
}

func (f *FeederBase) Name() string {
	return f.name
}

func (f *FeederBase) Start() {}
func (f *FeederBase) Stop() {}

func (f *FeederBase) IsRunning() bool {
	return f.isRunning
}

func register(name string, f FeederFactory) {
	feederName := name+"feeder"
	if f == nil {
		log.Fatal("Factory method doesn't exists")
	}
	if _, ok := feederFactories[feederName]; ok {
		log.Fatal("Factory method with the same name already exists")
	}
	feederFactories[feederName] = f
}

// Init
func init() {
}

func NewFeeder(name string, conf map[string]string) (Feeder, error) {
	if _, ok := feederFactories[name]; ok {
		f, err := feederFactories[name](conf)
		if err == nil && f != nil {
			f.setName(name)
		}

		return f, err
	}
	return nil, fmt.Errorf("feeder '%s' doesn't exist", name)
}

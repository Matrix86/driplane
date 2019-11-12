package feeder

import (
	"fmt"
	"github.com/Matrix86/driplane/com"
	"github.com/asaskevich/EventBus"

	"github.com/evilsocket/islazy/log"
)

type FeederFactory func(conf map[string]string) (Feeder, error)

var feederFactories = make(map[string]FeederFactory)

// TODO: change the callback logic with publish/subscribe channels
// something like https://gist.github.com/AmirSoleimani/97298c6a94d83d3672765fb31c23194a
type FeederCallback func(msg com.DataMessage)

type Feeder interface {
	setName(name string)
	setBus(bus EventBus.Bus)
	setId(id int32)

	Name()          string
	Start()
	Stop()
	IsRunning()     bool
	GetIdentifier() string
}

type FeederBase struct {
	Feeder
	com.Subscriber

	name        string
	id 			int32
	isRunning   bool
	bus         EventBus.Bus
}

func (f *FeederBase) Propagate(data com.DataMessage) {
	f.bus.Publish(f.GetIdentifier(), data)
}

func (f *FeederBase) setId(id int32) {
	f.id = id
}

func (f *FeederBase) setBus(bus EventBus.Bus) {
	f.bus = bus
}

func (f *FeederBase) setName(name string) {
	f.name = name
}

func (f *FeederBase) GetIdentifier() string {
	return fmt.Sprintf("%s:%d", f.name, f.id)
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

func NewFeeder(name string, conf map[string]string, bus EventBus.Bus, id int32) (Feeder, error) {
	if _, ok := feederFactories[name]; ok {
		f, err := feederFactories[name](conf)
		if err == nil && f != nil {
			f.setName(name)
			f.setBus(bus)
			f.setId(id)
		}

		return f, err
	}
	return nil, fmt.Errorf("feeder '%s' doesn't exist", name)
}

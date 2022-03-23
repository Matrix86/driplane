package feeders

import (
	"fmt"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
	"github.com/evilsocket/islazy/log"
)

// FeederFactory identifies a function to instantiate a Feeder using the Factory
type FeederFactory func(conf map[string]string) (Feeder, error)

var feederFactories = make(map[string]FeederFactory)

// Feeder defines Base methods of the object
type Feeder interface {
	setName(name string)
	setBus(bus EventBus.Bus)
	setID(id int32)
	setRuleName(name string)

	Name() string
	Rule() string
	Start()
	Stop()
	IsRunning() bool
	GetIdentifier() string
	OnEvent(e *data.Event)
}

// Base is inherited from the feeders
type Base struct {
	name      string
	rule      string
	id        int32
	isRunning bool
	bus       EventBus.Bus
}

// Propagate sends the Message to the connected Filters
func (f *Base) Propagate(data *data.Message) {
	data.SetExtra("source_feeder", f.Name())
	data.SetExtra("source_feeder_rule", f.Rule())
	data.SetExtra("rule_name", f.Rule())
	f.bus.Publish(f.GetIdentifier(), data)
}

func (f *Base) setID(id int32) {
	f.id = id
}

func (f *Base) setBus(bus EventBus.Bus) {
	f.bus = bus
}

func (f *Base) setName(name string) {
	f.name = name
}

func (f *Base) setRuleName(name string) {
	f.rule = name
}

// GetIdentifier returns the Node identifier ID used in the bus
func (f *Base) GetIdentifier() string {
	return fmt.Sprintf("%s:%d", f.name, f.id)
}

// Name returns the name of the Node
func (f *Base) Name() string {
	return f.name
}

// Rule returns the rule in which the Feeder is found
func (f *Base) Rule() string {
	return f.rule
}

// Start initializes the Node
func (f *Base) Start() {}

// Stop stops the Node
func (f *Base) Stop() {}

// IsRunning returns true if the Node is up and running
func (f *Base) IsRunning() bool {
	return f.isRunning
}

func register(name string, f FeederFactory) {
	feederName := name + "feeder"
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

// NewFeeder creates a new registered Feeder from it's name
func NewFeeder(rule string, name string, conf map[string]string, bus EventBus.Bus, id int32) (Feeder, error) {
	if _, ok := feederFactories[name]; ok {
		f, err := feederFactories[name](conf)
		if err == nil && f != nil {
			f.setName(name)
			f.setRuleName(rule)
			f.setBus(bus)
			f.setID(id)
		}

		return f, err
	}
	return nil, fmt.Errorf("feeder '%s' doesn't exist", name)
}

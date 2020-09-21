package filters

import (
	"fmt"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/plugins"

	"github.com/asaskevich/EventBus"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/plugin"
)

// FilterFactory identifies a function to instantiate a Filter using the Factory
type FilterFactory func(conf map[string]string) (Filter, error)

var filterFactories = make(map[string]FilterFactory)

// Filter defines Base methods of the object
type Filter interface {
	setRuleName(name string)
	setName(name string)
	setBus(bus EventBus.Bus)
	setID(id int32)
	setIsNegative(b bool)

	Rule() string
	Name() string
	DoFilter(msg *data.Message) (bool, error)
	Pipe(msg *data.Message)
	GetIdentifier() string
	Log(format string, args ...interface{})
}

// Base is inherited from the feeders
type Base struct {
	rule     string
	name     string
	id       int32
	bus      EventBus.Bus
	negative bool
	cbFilter func(msg *data.Message) (bool, error)
}

// Rule returns the rule in which the Filter is found
func (f *Base) Rule() string {
	return f.rule
}

// Name returns the Filter's name
func (f *Base) Name() string {
	return f.name
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

func (f *Base) setIsNegative(b bool) {
	f.negative = b
}

// GetIdentifier returns the Node identifier ID used in the bus
func (f *Base) GetIdentifier() string {
	return fmt.Sprintf("%s:%d", f.name, f.id)
}

// Log print a debug line prepending the name of the rule and of the filter
func (f *Base) Log(format string, args ...interface{}) {
	str := fmt.Sprintf("[%s::%s] %s", f.Rule(), f.Name(), format)
	log.Debug(str, args...)
}


// Pipe gets a Message from the previous Node and Propagate it to the next one if the Filter's callback will return true
func (f *Base) Pipe(msg *data.Message) {
	clone := msg.Clone()
	log.Debug("[%s::%s] received: %#v", f.rule, f.name, clone)
	b, err := f.cbFilter(clone)
	if err != nil {
		log.Error("[%s::%s] %s", f.rule, f.name, err)
	}

	// golang does not provide a logical XOR so we have to "implement" it manually
	if f.negative != b {
		log.Debug("[%s::%s] filter matched", f.rule, f.name)
		f.Propagate(clone)
	}
}

// Propagate sends the Message to the connected Filters
func (f *Base) Propagate(data *data.Message) {
	data.SetExtra("rule_name", f.Rule())
	f.bus.Publish(f.GetIdentifier(), data)
}

func register(name string, f FilterFactory) {
	filterName := name + "filter"
	if f == nil {
		log.Fatal("Filter method doesn't exists")
	}
	if _, ok := filterFactories[filterName]; ok {
		log.Fatal("Filter factory method with the same name already exists")
	}
	filterFactories[filterName] = f
}

func init() {
	// Thx @evilsocket for the hint =)
	// https://github.com/evilsocket/shellz/blob/master/plugins/plugin.go#L18
	plugin.Defines = map[string]interface{}{
		"log":     plugins.GetLog(),
		"http":    plugins.GetHTTP(),
		"file":    plugins.GetFile(),
		"util":    plugins.GetUtil(),
		"strings": plugins.GetStrings(),
	}
}

// NewFilter creates a new registered Filter from it's name
func NewFilter(rule string, name string, conf map[string]string, bus EventBus.Bus, id int32, neg bool) (Filter, error) {
	if _, ok := filterFactories[name]; ok {
		f, err := filterFactories[name](conf)
		if err == nil && f != nil {
			f.setRuleName(rule)
			f.setName(name)
			f.setBus(bus)
			f.setID(id)
			f.setIsNegative(neg)
		}
		return f, err
	}
	return nil, fmt.Errorf("filter '%s' doesn't exist", name)
}

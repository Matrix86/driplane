package filters

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/plugins"
	"github.com/asaskevich/EventBus"
	"github.com/evilsocket/islazy/log"
	"github.com/evilsocket/islazy/plugin"
)

type FilterFactory func(conf map[string]string) (Filter, error)

var filterFactories = make(map[string]FilterFactory)

type Filter interface {
	setName(name string)
	setBus(bus EventBus.Bus)
	setId(id int32)
	setIsNegative(b bool)

	Name() string
	DoFilter(msg *data.Message) (bool, error)
	Pipe(msg *data.Message)
	GetIdentifier() string
}

type Base struct {
	name     string
	id       int32
	bus      EventBus.Bus
	negative bool
	cbFilter func(msg *data.Message) (bool, error)
}

func (f *Base) Name() string {
	return f.name
}

func (f *Base) setId(id int32) {
	f.id = id
}

func (f *Base) setBus(bus EventBus.Bus) {
	f.bus = bus
}

func (f *Base) setName(name string) {
	f.name = name
}

func (f *Base) setIsNegative(b bool) {
	f.negative = b
}

func (f *Base) GetIdentifier() string {
	return fmt.Sprintf("%s:%d", f.name, f.id)
}

func (f *Base) Pipe(msg *data.Message) {
	log.Debug("[%s] received: %#v", f.name, msg)
	b, err := f.cbFilter(msg)
	if err != nil {
		log.Error("[%s] %s", f.name, err)
	}

	// golang does not provide a logical XOR so we have to "implement" it manually
	if f.negative != b {
		log.Debug("[%s] filter matched", f.name)
		f.Propagate(msg)
	}
}

func (f *Base) Propagate(data *data.Message) {
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
		"log": plugins.GetLog(),
		"http": plugins.GetHttp(),
		"file": plugins.GetFile(),
		"dio": func(format string, a ...interface{}){
			log.Info(format, a)
		},
	}
}

func NewFilter(name string, conf map[string]string, bus EventBus.Bus, id int32, neg bool) (Filter, error) {
	if _, ok := filterFactories[name]; ok {
		f, err := filterFactories[name](conf)
		if err == nil && f != nil {
			f.setName(name)
			f.setBus(bus)
			f.setId(id)
			f.setIsNegative(neg)
		}
		return f, err
	}
	return nil, fmt.Errorf("filter '%s' doesn't exist", name)
}

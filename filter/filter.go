package filter

import (
	"fmt"
	"github.com/Matrix86/driplane/com"
	"github.com/asaskevich/EventBus"

	"github.com/evilsocket/islazy/log"
)

type FilterFactory func(conf map[string]string) (Filter, error)

var filterFactories = make(map[string]FilterFactory)

type Filter interface {
	setName(name string)
	setBus(bus EventBus.Bus)
	setId(id int32)

	Name() string
	DoFilter(msg *com.DataMessage) (bool, error)
	GetIdentifier() string
}

type FilterBase struct {
	Filter
	com.Subscriber

	name string
	id   int32
	bus  EventBus.Bus
}

func (f *FilterBase) Name() string {
	return f.name
}

func (f *FilterBase) setId(id int32) {
	f.id = id
}

func (f *FilterBase) setBus(bus EventBus.Bus) {
	f.bus = bus
}

func (f *FilterBase) setName(name string) {
	f.name = name
}

func (f *FilterBase) GetIdentifier() string {
	return fmt.Sprintf("%s:%d", f.name, f.id)
}

func (f *FilterBase) Propagate(data com.DataMessage){
	f.bus.Publish(f.GetIdentifier(), data)
}

func register(name string, f FilterFactory) {
	filterName := name+"filter"
	if f == nil {
		log.Fatal("Filter method doesn't exists")
	}
	if _, ok := filterFactories[filterName]; ok {
		log.Fatal("Filter factory method with the same name already exists")
	}
	filterFactories[filterName] = f
}

func init() {
}

func NewFilter(name string, conf map[string]string, bus EventBus.Bus, id int32) (Filter, error) {
	if _, ok := filterFactories[name]; ok {
		f, err := filterFactories[name](conf)
		if err == nil && f != nil {
			f.setName(name)
			f.setBus(bus)
			f.setId(id)
		}
		return f, err
	}
	return nil, fmt.Errorf("filter '%s' doesn't exist", name)
}

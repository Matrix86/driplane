package filter

import (
	"fmt"
	"github.com/Matrix86/driplane/com"

	"github.com/evilsocket/islazy/log"
)

type FilterFactory func(conf map[string]string) (Filter, error)

var filterFactories = make(map[string]FilterFactory)

type Filter interface {
	setName(name string)

	Name() string
	DoFilter(msg *com.DataMessage) (bool, error)
}

type FilterBase struct {
	Filter
	com.Subscriber

	name string
	subscribers []com.DataCallback
}

func (f *FilterBase) Name() string {
	return f.name
}

func (f *FilterBase) setName(name string) {
	f.name = name
}

func (f *FilterBase) SetEventMessageClb(clb com.DataCallback) {
	f.subscribers = append(f.subscribers, clb)
}

func (f *FilterBase) Propagate(data com.DataMessage){
	log.Debug("filter '%s' received: [%v]", f.Name(), data)
	log.Debug("filter '%s' propagating to %d subscribers", f.Name(), len(f.subscribers))
	if f.subscribers != nil && len(f.subscribers) > 0 {
		for _, cb := range f.subscribers {
			cb(data)
		}
	}
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

func NewFilter(name string, conf map[string]string) (Filter, error) {
	if _, ok := filterFactories[name]; ok {
		f, err := filterFactories[name](conf)
		if err == nil && f != nil {
			f.setName(name)
		}
		return f, err
	}
	return nil, fmt.Errorf("filter '%s' doesn't exist", name)
}

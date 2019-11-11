package action

import (
	"fmt"
	"github.com/Matrix86/driplane/com"

	"github.com/evilsocket/islazy/log"
)

type ActionFactory func(conf []string) (Action, error)

var actionFactories = make(map[string]ActionFactory)

//type ActionCallback func(msg com.DataMessage)

type Action interface {
	Name() string
	DoAction(data com.DataMessage)
}

type ActionBase struct {
	Action
	com.Subscriber

	subscribers []com.DataCallback
}

func (a *ActionBase) SetEventMessageClb(clb com.DataCallback) {
	a.subscribers = append(a.subscribers, clb)
}

func (a *ActionBase) Propagate(data com.DataMessage){
	if a.subscribers != nil && len(a.subscribers) > 0 {
		for _, cb := range a.subscribers {
			cb(data)
		}
	}
}

func register(name string, f ActionFactory) {
	if f == nil {
		log.Fatal("Factory method doesn't exists")
	}
	if _, ok := actionFactories[name]; ok {
		log.Fatal("Factory method with the same name already exists")
	}
	actionFactories[name] = f
}

func NewAction(name string, conf []string) (Action, error) {
	if _, ok := actionFactories[name]; ok {
		return actionFactories[name](conf)
	}
	return nil, fmt.Errorf("Action '%s' doesn't exist", name)
}

// Init
func init() {
}

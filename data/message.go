package data

import "sync"

type Callback func(msg Message)

type Message struct {
	message string
	extra   sync.Map
	mutex   sync.Mutex
}

func (d *Message) SetMessage(msg string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.message = msg
}

func (d *Message) GetMessage() string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.message
}

func (d *Message) SetExtra(k string, v string) {
	d.extra.Store(k, v)
}

func (d *Message) GetExtra() *sync.Map {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return &d.extra
}

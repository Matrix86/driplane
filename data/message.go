package data

import "sync"

type Callback func(msg Message)

type Message struct {
	sync.Mutex

	message string
	extra   map[string]string
}

func NewMessage(msg string) *Message {
	return NewMessageWithExtra(msg, map[string]string{})
}

func NewMessageWithExtra(msg string, extra map[string]string) *Message {
	return &Message{
		message: msg,
		extra:   extra,
	}
}

func (d *Message) SetMessage(msg string) {
	d.Lock()
	defer d.Unlock()
	d.message = msg
}

func (d *Message) GetMessage() string {
	d.Lock()
	defer d.Unlock()
	return d.message
}

func (d *Message) SetExtra(k string, v string) {
	d.Lock()
	defer d.Unlock()
	d.extra[k] = v
}

func (d *Message) Extra(cb func(k, v string)) {
	d.Lock()
	defer d.Unlock()
	for k, v := range d.extra {
		cb(k, v)
	}
}

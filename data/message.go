package data

import (
	"bytes"
	"html/template"
	"sync"
)

type Callback func(msg Message)

type Message struct {
	sync.Mutex

	fields map[string]string
}

func NewMessage(msg string) *Message {
	return NewMessageWithExtra(msg, map[string]string{})
}

func NewMessageWithExtra(msg string, extra map[string]string) *Message {
	extra["main"] = msg
	return &Message{
		fields: extra,
	}
}

func (d *Message) SetMessage(msg string) {
	d.Lock()
	defer d.Unlock()
	d.fields["main"] = msg
}

func (d *Message) GetMessage() string {
	d.Lock()
	defer d.Unlock()
	return d.fields["main"]
}

func (d *Message) SetExtra(k string, v string) {
	d.Lock()
	defer d.Unlock()
	if k == "main" {
		return
	}
	d.fields[k] = v
}

func (d *Message) GetExtra() map[string]string {
	d.Lock()
	defer d.Unlock()

	clone := make(map[string]string)
	for key, value := range d.fields {
		if key == "main" {
			// Ignoring main content
			continue
		}
		clone[key] = value
	}
	return clone
}

//func (d *Message) Extra(cb func(k, v string)) {
//	d.Lock()
//	defer d.Unlock()
//	for k, v := range d.extra {
//		cb(k, v)
//	}
//}

func (d *Message) ApplyPlaceholder(t *template.Template) (string, error) {
	d.Lock()
	defer d.Unlock()
	var writer bytes.Buffer

	err := t.Execute(&writer, d.fields)
	if err != nil {
		return "", err
	}
	return writer.String(), nil
}

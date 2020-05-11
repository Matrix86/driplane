package data

import (
	"bytes"
	"fmt"
	html "html/template"
	"sync"
	text "text/template"
)

type Callback func(msg Message)

type Message struct {
	sync.RWMutex

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
	d.RLock()
	defer d.RUnlock()
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

func (d *Message) GetTarget(name string) string {
	d.RLock()
	defer d.RUnlock()
	if v, ok := d.fields[name]; ok {
		return v
	}
	return ""
}

func (d *Message) Clone() *Message {
	clone := &Message{
		fields: make(map[string]string, 0),
	}

	for k, v := range d.fields {
		clone.fields[k] = v
	}
	return clone
}

func (d *Message) ApplyPlaceholder(template interface{}) (string, error) {
	d.RLock()
	defer d.RUnlock()
	var writer bytes.Buffer

	switch t := template.(type) {
	case *html.Template:
		err := t.Execute(&writer, d.fields)
		if err != nil {
			return "", err
		}
		return writer.String(), nil
	case *text.Template:
		err := t.Execute(&writer, d.fields)
		if err != nil {
			return "", err
		}
		return writer.String(), nil
	}
	return "", fmt.Errorf("template type not supported")
}

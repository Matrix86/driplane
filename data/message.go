package data

import (
	"bytes"
	"fmt"
	html "html/template"
	"strings"
	"sync"
	text "text/template"
)

// Message is the data generated from a Feeder and it travels across Filters
type Message struct {
	sync.RWMutex

	fields   map[string]interface{}
	firstRun bool
}

// NewMessage creates a new Message struct with only the "main" data
func NewMessage(msg interface{}) *Message {
	return NewMessageWithExtra(msg, map[string]interface{}{})
}

// NewMessageWithExtra creates a Message struct with "main" and extra data
func NewMessageWithExtra(msg interface{}, extra map[string]interface{}) *Message {
	extra["main"] = msg
	return &Message{
		fields: extra,
	}
}

// SetMessage allows to change the "main" data in the Message struct
func (d *Message) SetMessage(msg interface{}) {
	d.Lock()
	defer d.Unlock()
	d.fields["main"] = msg
}

// GetMessage returns the "main" data in the Message struct
func (d *Message) GetMessage() interface{} {
	d.RLock()
	defer d.RUnlock()
	return d.fields["main"]
}

// SetExtra allows to change the "extra" data with key k and value v in the Message struct
func (d *Message) SetExtra(k string, v interface{}) {
	d.Lock()
	defer d.Unlock()
	if k == "main" {
		return
	}
	d.fields[k] = v
}

// GetExtra returns all the "extra" data in the Message struct
func (d *Message) GetExtra() map[string]interface{} {
	d.Lock()
	defer d.Unlock()

	clone := make(map[string]interface{})
	for key, value := range d.fields {
		if key == "main" {
			// Ignoring main content
			continue
		}
		if strings.HasPrefix(key, "_") {
			// ignore keys that starts with _
			continue
		}
		clone[key] = value
	}
	return clone
}

// SetTarget is like SetExtra but it can change also the "main" key
func (d *Message) SetTarget(name string, value interface{}) {
	d.Lock()
	defer d.Unlock()
	d.fields[name] = value
}

// SetFirstRun set the firstRun flag
func (d *Message) SetFirstRun() {
	d.Lock()
	defer d.Unlock()
	d.firstRun = true
}

// ClearFirstRun clear the firstRun flag
func (d *Message) ClearFirstRun() {
	d.Lock()
	defer d.Unlock()
	d.firstRun = false
}

// IsFirstRun return the status of the firstRun flag
func (d *Message) IsFirstRun() bool {
	d.RLock()
	defer d.RUnlock()
	return d.firstRun
}

// GetTarget returns the value of a key in the Message struct. It can return also the "main" data
func (d *Message) GetTarget(name string) interface{} {
	d.RLock()
	defer d.RUnlock()
	if v, ok := d.fields[name]; ok {
		return v
	}
	return nil
}

// Clone creates a deep copy of the Message struct
func (d *Message) Clone() *Message {
	clone := &Message{
		fields: make(map[string]interface{}, 0),
	}

	for k, v := range d.fields {
		clone.fields[k] = v
	}
	clone.firstRun = d.firstRun

	return clone
}

// ApplyPlaceholder executes the template specified using the data in the Message struct
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

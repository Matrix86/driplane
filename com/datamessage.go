package com

type DataCallback func(msg DataMessage)

type DataMessage struct {
	message string
	extra   map[string]string
}

func (d *DataMessage) SetMessage(msg string) {
	d.message = msg
}

func (d *DataMessage) GetMessage() string {
	return d.message
}

func (d *DataMessage) SetExtra(k string, v string) {
	// First Initialization
	if d.extra == nil {
		d.extra = make(map[string]string)
	}

	d.extra[k] = v
}

func (d *DataMessage) GetExtra() map[string]string {
	return d.extra
}

type Subscriber interface {
	Propagate(data DataMessage)
	//Filtering(data DataMessage)
}

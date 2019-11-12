package com

import "sync"

type DataCallback func(msg DataMessage)

type DataMessage struct {
	message string
	extra   sync.Map
	mutex   sync.Mutex
}

func (d *DataMessage) SetMessage(msg string) {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	d.message = msg
}

func (d *DataMessage) GetMessage() string {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return d.message
}

func (d *DataMessage) SetExtra(k string, v string) {
	d.extra.Store(k, v)
}

func (d *DataMessage) GetExtra() *sync.Map {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	return &d.extra
}

type Subscriber interface {
	Propagate(data DataMessage)
	//Filtering(data DataMessage)
}

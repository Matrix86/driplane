package filters

import "github.com/Matrix86/driplane/data"

type FakeBus struct {
	Collected []*data.Message
}

func NewFakeBus() *FakeBus {
	return &FakeBus{
		Collected: make([]*data.Message, 0),
	}
}

func (b *FakeBus) Reset() {
	b.Collected = make([]*data.Message, 0)
}

func (b *FakeBus) Publish(topic string, args ...interface{}) {
	for _, k := range args {
		if v, ok := k.(*data.Message); ok {
			b.Collected = append(b.Collected, v)
		}
	}
}

func (b *FakeBus) HasCallback(topic string) bool                                         { return true }
func (b *FakeBus) WaitAsync()                                                            {}
func (b *FakeBus) Subscribe(topic string, fn interface{}) error                          { return nil }
func (b *FakeBus) SubscribeAsync(topic string, fn interface{}, transactional bool) error { return nil }
func (b *FakeBus) SubscribeOnce(topic string, fn interface{}) error                      { return nil }
func (b *FakeBus) SubscribeOnceAsync(topic string, fn interface{}) error                 { return nil }
func (b *FakeBus) Unsubscribe(topic string, handler interface{}) error                   { return nil }

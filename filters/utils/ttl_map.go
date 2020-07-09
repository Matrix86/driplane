package utils

import (
	"sync"
	"time"
)

type item struct {
	value      interface{}
	expiration int64
}

func (i *item) expired() bool {
	return i.expiration <= time.Now().Unix()
}

type TTLMap struct {
	sync.RWMutex

	dict    map[interface{}]*item
	gcdelay time.Duration
}

func NewTTLMap(gcdelay time.Duration) (m *TTLMap) {
	m = &TTLMap{
		dict:    make(map[interface{}]*item, 0),
		gcdelay: gcdelay,
	}

	// Cleaning method is called every X seconds because we
	// check if an item is expired in the Get method
	go func() {
		for _ = range time.Tick(gcdelay) {
			m.Lock()
			for k, v := range m.dict {
				if v.expired() {
					delete(m.dict, k)
				}
			}
			m.Unlock()
		}
	}()
	return
}

func (m *TTLMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.dict)
}

func (m *TTLMap) Put(k, v interface{}, ttl int64) {
	m.Lock()
	defer m.Unlock()

	if i, ok := m.dict[k]; ok {
		// refresh ttl
		i.expiration = time.Now().Unix() + ttl
	} else {
		i := &item{
			value:      v,
			expiration: time.Now().Unix() + ttl,
		}
		m.dict[k] = i
	}
}

func (m *TTLMap) Get(k string) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()
	if i, ok := m.dict[k]; ok {
		if i.expired() {
			return nil, false
		}
		return i, true
	}
	return nil, false
}

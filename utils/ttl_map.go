package utils

import (
	"encoding/gob"
	"os"
	"sync"
	"time"

	"github.com/evilsocket/islazy/log"
	"github.com/juju/fslock"
)

type item struct {
	Value      interface{}
	Expiration int64
}

func (i *item) expired() bool {
	return i.Expiration <= time.Now().Unix()
}

// TTLMap is a cache with ttl
type TTLMap struct {
	sync.RWMutex

	filename string
	dict     map[interface{}]*item
	gcdelay  time.Duration
}

// NewTTLMap creates a TTLMap instance
func NewTTLMap(gcdelay time.Duration) (m *TTLMap) {
	m = &TTLMap{
		dict:    make(map[interface{}]*item, 0),
		gcdelay: gcdelay,
	}

	// Cleaning method is called every X seconds because we
	// check if an item is expired in the Get method
	go func() {
		for range time.Tick(gcdelay) {
			m.Lock()
			for k, v := range m.dict {
				if v.expired() {
					delete(m.dict, k)
				}
			}
			err := m.syncFile()
			if err != nil {
				log.Error("TTLMap syncFile: %s", err)
			}
			m.Unlock()
		}
	}()
	return
}

// Len returns the number of cached keys
func (m *TTLMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	return len(m.dict)
}

// Put inserts a key => value pair in the cache
func (m *TTLMap) Put(k, v interface{}, ttl int64) {
	m.Lock()
	defer m.Unlock()

	if i, ok := m.dict[k]; ok {
		i.Value = v
		// refresh ttl
		i.Expiration = time.Now().Unix() + ttl
	} else {
		i := &item{
			Value:      v,
			Expiration: time.Now().Unix() + ttl,
		}
		m.dict[k] = i
	}
}

// Get returns the value of the associated key from the cache
func (m *TTLMap) Get(k interface{}) (interface{}, bool) {
	m.RLock()
	defer m.RUnlock()
	if i, ok := m.dict[k]; ok {
		if i.expired() {
			return nil, false
		}
		return i.Value, true
	}
	return nil, false
}

func (m *TTLMap) syncFile() error {
	log.Debug("called!")
	if m.filename == "" {
		return nil
	}

	// Try to lock the file during the sync
	lock := fslock.New(m.filename)
	err := lock.LockWithTimeout(m.gcdelay)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	file, err := os.Create(m.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(m.dict); err != nil {
		return err
	}

	log.Debug("cache file synced : %s", m.filename)
	return nil
}

// SetPersistence load the map from a file (it can be used only the first time)
func (m *TTLMap) SetPersistence(filename string) error {
	m.Lock()
	defer m.Unlock()

	if m.filename != "" {
		return nil
	}

	info, err := os.Stat(filename)
	// file doesn't not exist or empty
	if os.IsNotExist(err) || (!info.IsDir() && info.Size() == 0) {
		file, err := os.Create(filename)
		if err != nil {
			return err
		}
		defer file.Close()
		m.filename = filename
		return nil
	}

	// Try to lock the file during the sync
	lock := fslock.New(filename)
	err = lock.LockWithTimeout(m.gcdelay)
	if err != nil {
		return err
	}
	defer lock.Unlock()

	file, err := os.Open(filename)
	if err != nil {
		return err
	}
	defer file.Close()
	decoder := gob.NewDecoder(file)
	if err := decoder.Decode(&m.dict); err != nil {
		return err
	}
	m.filename = filename
	return nil
}

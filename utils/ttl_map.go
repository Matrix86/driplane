package utils

import (
	"context"
	"encoding/gob"
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/evilsocket/islazy/log"
	"github.com/gofrs/flock"
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
	ticker   *time.Ticker
	close    chan bool
	wg       sync.WaitGroup
}

// NewTTLMap creates a TTLMap instance
func NewTTLMap(gcdelay time.Duration) (m *TTLMap) {
	m = &TTLMap{
		dict:    make(map[interface{}]*item, 0),
		gcdelay: gcdelay,
		ticker:  time.NewTicker(gcdelay),
		close:   make(chan bool, 1),
		wg:      sync.WaitGroup{},
	}

	m.wg.Add(1)

	// Cleaning method is called every X seconds because we
	// check if an item is expired in the Get method
	go func() {
		defer m.wg.Done()
		for {
			select {
			case <-m.close:
				m.Lock()
				_ = m.syncFile()
				m.Unlock()
				return

			case <-m.ticker.C:
				if m.dict != nil {
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
			}
		}
	}()
	return
}

// Len returns the number of cached keys
func (m *TTLMap) Len() int {
	m.RLock()
	defer m.RUnlock()
	if m.dict == nil {
		return 0
	}

	return len(m.dict)
}

// Put inserts a key => value pair in the cache
func (m *TTLMap) Put(k, v interface{}, ttl int64) {
	if m.dict == nil {
		return
	}

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
	if m.dict == nil {
		return nil, false
	}

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
	if m.filename == "" {
		return nil
	}

	if m.dict == nil {
		return fmt.Errorf("map has been closed")
	}

	// Try to lock the file during the sync
	lock := flock.New(m.filename + ".lock")
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := lock.TryLockContext(lockCtx, 500*time.Millisecond)
	if err != nil {
		return fmt.Errorf("file locking: %s", err)
	}
	if locked {
		defer lock.Unlock()
	}

	file, err := os.Create(m.filename)
	if err != nil {
		return err
	}
	defer file.Close()
	encoder := gob.NewEncoder(file)
	if err := encoder.Encode(m.dict); err != nil {
		return fmt.Errorf("encoding: %s", err)
	}

	log.Debug("cache file synced : %s", m.filename)
	return nil
}

// SetPersistence load the map from a file (it can be used only the first time)
func (m *TTLMap) SetPersistence(filename string) error {
	if m.dict == nil {
		return fmt.Errorf("map has been closed")
	}

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
	lock := flock.New(filename + ".lock")
	lockCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	locked, err := lock.TryLockContext(lockCtx, 500*time.Millisecond)
	if err != nil {
		return fmt.Errorf("file locking: %s", err)
	}
	if locked {
		defer lock.Unlock()
	}

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

func (m *TTLMap) Close() {
	if m.dict == nil {
		return
	}
	m.close <- true
	// waiting the cleaning close
	m.wg.Wait()
	close(m.close)
	m.dict = nil
}

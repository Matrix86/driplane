package filters

import (
	"fmt"
	"sync"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils"
)

// Cache handles a cache usable in the rule
type Cache struct {
	sync.Mutex
	Base

	target       string
	ttl          time.Duration
	refreshOnGet bool
	global       bool
	syncTime     time.Duration

	persistentFile string

	params map[string]string
	cache  *utils.TTLMap
}

// NewCacheFilter is the registered method to instantiate a CacheFilter
func NewCacheFilter(p map[string]string) (Filter, error) {
	f := &Cache{
		params:       p,
		target:       "main",
		refreshOnGet: true,
		global:       false,
		ttl:          24 * time.Hour,
		syncTime:     5 * time.Minute,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["target"]; ok {
		f.target = v
	}
	if v, ok := f.params["refresh_on_get"]; ok && v == "false" {
		f.refreshOnGet = false
	}
	if v, ok := f.params["sync_time"]; ok {
		// https://golang.org/pkg/time/#ParseDuration
		i, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		f.syncTime = i
	}
	if v, ok := f.params["ttl"]; ok {
		// https://golang.org/pkg/time/#ParseDuration
		i, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		f.ttl = i
	}
	if v, ok := f.params["global"]; ok && v == "true" {
		f.global = true
		f.cache = utils.GetGlobalTTLMapInstance(f.syncTime).Cache
	} else {
		f.cache = utils.NewTTLMap(f.syncTime)
	}

	if v, ok := f.params["file"]; ok {
		err := f.cache.SetPersistence(v)
		if err != nil {
			return nil, fmt.Errorf("cache tried to load %s: %s", v, err)
		}
		f.persistentFile = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Cache) DoFilter(msg *data.Message) (bool, error) {
	var text interface{}

	if f.target == "main" {
		text = msg.GetMessage()
	} else if v, ok := msg.GetExtra()[f.target]; ok {
		text = v
	} else {
		return false, nil
	}

	hash := utils.MD5Sum(text)
	if _, ok := f.cache.Get(hash); !ok {
		f.cache.Put(hash, true, int64(f.ttl.Seconds()))
		return true, nil
	} else if f.refreshOnGet {
		f.cache.Put(hash, true, int64(f.ttl.Seconds()))
	}
	return false, nil
}

// Set the name of the filter
func init() {
	register("cache", NewCacheFilter)
}

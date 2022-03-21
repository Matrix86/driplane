package filters

import (
	"fmt"
	"sync"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils"

	"github.com/evilsocket/islazy/log"
)

// Cache handles a cache usable in the rule
type Cache struct {
	sync.Mutex
	Base

	target         string
	ttl            time.Duration
	refreshOnGet   bool
	global         bool
	syncTime       time.Duration
	ignoreFirstRun bool
	cacheName      string

	persistentFile string

	params map[string]string
	cache  *utils.TTLMap
}

// NewCacheFilter is the registered method to instantiate a CacheFilter
func NewCacheFilter(p map[string]string) (Filter, error) {
	f := &Cache{
		params:         p,
		target:         "main",
		refreshOnGet:   true,
		global:         false,
		ttl:            24 * time.Hour,
		syncTime:       5 * time.Minute,
		ignoreFirstRun: false,
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
		f.cacheName = "global"
		f.cache = utils.GetNamedTTLMap("global", f.syncTime)
	} else if v, ok := f.params["name"]; ok {
		f.cacheName = v
		f.cache = utils.GetNamedTTLMap(v, f.syncTime)
	} else {
		f.cache = utils.NewTTLMap(f.syncTime)
	}

	if v, ok := f.params["ignore_first_run"]; ok && v == "true" {
		f.ignoreFirstRun = true
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
	}

	if text == nil {
		return true, nil
	}

	hash := utils.MD5Sum(text)
	if _, ok := f.cache.Get(hash); !ok {
		f.cache.Put(hash, true, int64(f.ttl.Seconds()))
		if f.ignoreFirstRun {
			log.Debug("ignoring first run")
			return true, nil
		}
		log.Debug("caching '%s' firstRun:%v", text, msg.IsFirstRun())
		// don't propagate the message is it is the first run
		return !msg.IsFirstRun(), nil
	} else if f.refreshOnGet {
		f.cache.Put(hash, true, int64(f.ttl.Seconds()))
	}
	return false, nil
}

// OnEvent is called when an event occurs
func (f *Cache) OnEvent(event *data.Event) {
	if event.Type == "shutdown" {
		f.cache.Close()
	}
}

// Set the name of the filter
func init() {
	register("cache", NewCacheFilter)
}

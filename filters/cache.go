package filters

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/filters/utils"
	"sync"
	"time"
)

type Cache struct {
	sync.Mutex
	Base

	target string
	ttl    time.Duration

	params map[string]string
	cache  *utils.TTLMap
}

func NewCacheFilter(p map[string]string) (Filter, error) {
	f := &Cache{
		params: p,
		target: "main",
		ttl: 24 * time.Hour,
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["target"]; ok {
		f.target = v
	}
	if v, ok := f.params["ttl"]; ok {
		// https://golang.org/pkg/time/#ParseDuration
		i, err := time.ParseDuration(v)
		if err != nil {
			return nil, err
		}
		f.ttl = i
	}

	f.cache = utils.NewTTLMap(f.ttl.Seconds(), 5*time.Minute)

	return f, nil
}

func (f *Cache) getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

func (f *Cache) DoFilter(msg *data.Message) (bool, error) {
	var text string

	if f.target == "main" {
		text = msg.GetMessage()
	} else if v, ok := msg.GetExtra()[f.target]; ok {
		text = v
	} else {
		return false, nil
	}

	hash := f.getMD5Hash(text)
	if _, ok := f.cache.Get(hash); !ok {
		f.cache.Put(hash, true)
		return true, nil
	}
	return false, nil
}

// Set the name of the filter
func init() {
	register("cache", NewCacheFilter)
}

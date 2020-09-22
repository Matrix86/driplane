package plugins

import (
	"testing"
	"time"
)

func TestCachePluginGetPutMethods(t *testing.T) {
	cache := GetCache()

	res := cache.Get("key")
	if res.Status != false {
		t.Errorf("cache.Get should return false" )
	}

	cache.Put("key", "newvalue", 1)
	res = cache.Get("key")
	if res.Status == false {
		t.Errorf("cache.Get should return true" )
	}
	if res.Value != "newvalue" {
		t.Errorf("cache.Get should return 'newvalue'" )
	}

	time.Sleep(1 * time.Second)

	res = cache.Get("key")
	if res.Status != false {
		t.Errorf("cache.Get should return false" )
	}
}
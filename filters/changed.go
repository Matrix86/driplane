package filters

import (
	"crypto/md5"
	"encoding/hex"
	"github.com/Matrix86/driplane/data"
	"sync"
)

// Changed is a Filter that call the propagation method only if
// the input Message is different from the previous one
type Changed struct {
	sync.Mutex
	Base

	target string

	params map[string]string
	cache  string
}

// NewChangedFilter is the registered method to instantiate a ChangedFilter
func NewChangedFilter(p map[string]string) (Filter, error) {
	f := &Changed{
		params: p,
		target: "main",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["target"]; ok {
		f.target = v
	}

	return f, nil
}

func (f *Changed) getMD5Hash(text string) string {
	hash := md5.Sum([]byte(text))
	return hex.EncodeToString(hash[:])
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Changed) DoFilter(msg *data.Message) (bool, error) {
	var text string

	if f.target == "main" {
		text = msg.GetMessage()
	} else if v, ok := msg.GetExtra()[f.target]; ok {
		text = v
	} else {
		return false, nil
	}

	hash := f.getMD5Hash(text)
	if f.cache != hash {
		f.Lock()
		defer f.Unlock()
		f.cache = hash
		return true, nil
	}

	return false, nil
}

// Set the name of the filter
func init() {
	register("changed", NewChangedFilter)
}

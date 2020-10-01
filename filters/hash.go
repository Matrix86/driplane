package filters

import (
	"fmt"
	"regexp"

	"github.com/Matrix86/driplane/data"
)

// Hash is a Filter that searches for hashes in the Message
type Hash struct {
	Base

	regex *regexp.Regexp

	useMd5      bool
	useSha1     bool
	useSha256   bool
	useSha512   bool
	extractHash bool
	target      string

	//filter_extra string

	params map[string]string
}

// NewHashFilter is the registered method to instantiate a HashFilter
func NewHashFilter(p map[string]string) (Filter, error) {
	f := &Hash{
		params:      p,
		useMd5:      true,
		useSha1:     true,
		useSha256:   true,
		useSha512:   true,
		extractHash: false,
		target:      "main",
	}
	f.cbFilter = f.DoFilter

	f.regex = regexp.MustCompile(`(?i)[a-f0-9]{32,128}`)

	if v, ok := f.params["md5"]; ok && v == "false" {
		f.useMd5 = false
	}
	if v, ok := f.params["sha1"]; ok && v == "false" {
		f.useSha1 = false
	}
	if v, ok := f.params["sha256"]; ok && v == "false" {
		f.useSha256 = false
	}
	if v, ok := f.params["sha512"]; ok && v == "false" {
		f.useSha512 = false
	}
	if v, ok := f.params["extract"]; ok && v == "true" {
		f.extractHash = true
	}
	if v, ok := f.params["target"]; ok {
		f.target = v
	}

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *Hash) DoFilter(msg *data.Message) (bool, error) {
	var text string

	if v, ok := msg.GetTarget(f.target).(string); ok {
		text = v
	} else if v, ok := msg.GetTarget(f.target).([]byte); ok {
		text = string(v)
	} else {
		// ERROR this filter can't be used with different types
		return false, fmt.Errorf("received data is not a string")
	}
	match := f.regex.FindAllStringSubmatch(text, -1)
	if match != nil {
		for _, m := range match {
			length := len(m[0])
			found := false
			if f.useMd5 && length == 32 {
				found = true
			} else if f.useSha1 && length == 40 {
				found = true
			} else if f.useSha256 && length == 64 {
				found = true
			} else if f.useSha512 && length == 128 {
				found = true
			}

			if found {
				if f.extractHash {
					clone := msg.Clone()
					clone.SetMessage(m[0])
					clone.SetExtra("fulltext", text)
					f.Propagate(clone)
				} else {
					return true, nil
				}
			}
		}
	}
	return false, nil
}

// Set the name of the filter
func init() {
	register("hash", NewHashFilter)
}

package filters

import (
	"github.com/Matrix86/driplane/data"
	"regexp"
)

type Hash struct {
	Base

	regex *regexp.Regexp

	useMd5      bool
	useSha1     bool
	useSha256   bool
	useSha512   bool
	extractHash bool

	//filter_extra string

	params map[string]string
}

func NewHashFilter(p map[string]string) (Filter, error) {
	f := &Hash{
		params:      p,
		useMd5:      true,
		useSha1:     true,
		useSha256:   true,
		useSha512:   true,
		extractHash: false,
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
		f.useSha256 = false
	}
	if v, ok := f.params["extract"]; ok && v == "true" {
		f.extractHash = true
	}
	//if v, ok := f.params["filter_extra"]; ok {
	//	f.filter_extra = v
	//}

	return f, nil
}

func (f *Hash) DoFilter(msg *data.Message) (bool, error) {
	text := msg.GetMessage()
	msg.SetMessage(text)

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
					clone := *msg
					clone.SetMessage(m[0])
					clone.SetExtra("fulltext", text)
					f.Propagate(&clone)
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

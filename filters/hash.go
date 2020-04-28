package filters

import (
	"github.com/Matrix86/driplane/data"
	"regexp"
)

type Hash struct {
	Base

	rMd5    *regexp.Regexp
	rSha1   *regexp.Regexp
	rSha256 *regexp.Regexp
	rSha512 *regexp.Regexp

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
		extractHash: true,
	}
	f.cbFilter = f.DoFilter

	f.rMd5 = regexp.MustCompile(`(?i)[a-f0-9]{32}`)
	f.rSha1 = regexp.MustCompile(`(?i)[a-f0-9]{40}`)
	f.rSha256 = regexp.MustCompile(`(?i)[a-f0-9]{64}`)
	f.rSha512 = regexp.MustCompile(`(?i)[a-f0-9]{128}`)

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
	if v, ok := f.params["extract"]; ok && v == "false" {
		f.extractHash = false
	}
	//if v, ok := f.params["filter_extra"]; ok {
	//	f.filter_extra = v
	//}

	return f, nil
}

func (f *Hash) DoFilter(msg *data.Message) (bool, error) {
	text := msg.GetMessage()
	msg.SetMessage(text)

	if f.extractHash {
		matched := make([]string, 0)
		if f.useSha512 {
			match := f.rSha512.FindAllStringSubmatch(text, -1)
			if match != nil {
				for _, m := range match {
					matched = append(matched, m[0])
				}
			}
		}
		if f.useSha256 {
			match := f.rSha256.FindAllStringSubmatch(text, -1)
			if match != nil {
				for _, m := range match {
					matched = append(matched, m[0])
				}
			}
		}
		if f.useSha1 {
			match := f.rSha1.FindAllStringSubmatch(text, -1)
			if match != nil {
				for _, m := range match {
					matched = append(matched, m[0])
				}
			}
		}
		if f.useMd5 {
			match := f.rMd5.FindAllStringSubmatch(text, -1)
			if match != nil {
				for _, m := range match {
					matched = append(matched, m[0])
				}
			}
		}

		if len(matched) == 1 {
			msg.SetMessage(matched[0])
			msg.SetExtra("fulltext", text)
			return true, nil
		} else if len(matched) > 1 {
			for _, m := range matched {
				clone := *msg
				clone.SetMessage(m)
				clone.SetExtra("fulltext", text)
				f.Propagate(&clone)
			}
			return false, nil
		}
	} else {
		if f.useMd5 && f.rMd5.MatchString(text) {
			return true, nil
		} else if f.useSha1 && f.rSha1.MatchString(text) {
			return true, nil
		} else if f.useSha256 && f.rSha256.MatchString(text) {
			return true, nil
		} else if f.useSha512 && f.rSha512.MatchString(text) {
			return true, nil
		}
	}
	return false, nil
}

// Set the name of the filter
func init() {
	register("hash", NewHashFilter)
}

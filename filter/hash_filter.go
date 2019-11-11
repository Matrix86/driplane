package filter

import (
	"github.com/Matrix86/driplane/com"
	"regexp"
)

type HashFilter struct {
	FilterBase

	rMd5    *regexp.Regexp
	rSha1   *regexp.Regexp
	rSha256 *regexp.Regexp
	rSha512 *regexp.Regexp

	useMd5      bool
	useSha1     bool
	useSha256   bool
	useSha512   bool
	extractHash bool

	params map[string]string
}

func NewHashFilter(p map[string]string) (Filter, error) {
	f := &HashFilter{
		params: p,
		useMd5: true,
		useSha1: true,
		useSha256: true,
		extractHash: true,
	}

	f.rMd5    = regexp.MustCompile(`(?i)[a-f0-9]{32}`)
	f.rSha1   = regexp.MustCompile(`(?i)[a-f0-9]{40}`)
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
	if v, ok := f.params["extract"]; ok && v == "false" {
		f.extractHash = false
	}

	return f, nil
}

func (f *HashFilter) DoFilter(msg *com.DataMessage) (bool, error) {
	text := msg.GetMessage()

	if f.extractHash {
		if f.useMd5 {
			match := f.rMd5.FindStringSubmatch(text)
			if match != nil {
				msg.SetMessage(match[0])
			}
			return match != nil, nil
		} else if f.useSha1 {
			match := f.rSha1.FindStringSubmatch(text)
			if match != nil {
				msg.SetMessage(match[0])
			}
			return match != nil, nil
		} else if f.useSha256 {
			match := f.rSha256.FindStringSubmatch(text)
			if match != nil {
				msg.SetMessage(match[0])
			}
			return match != nil, nil
		} else if f.useSha512 {
			match := f.rSha512.FindStringSubmatch(text)
			if match != nil {
				msg.SetMessage(match[0])
			}
			return match != nil, nil
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
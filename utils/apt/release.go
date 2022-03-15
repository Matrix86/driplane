package apt

import (
	"fmt"
	"io"
	"pault.ag/go/debian/control"
	"strconv"
	"strings"
)

// Release represents the Release file
type Release struct {
	// Optional
	Description string `control:"Description"`
	Origin      string `control:"Origin"`
	Label       string `control:"Label"`
	Version     string `control:"Version"`
	Suite       string `control:"Suite"`
	Codename    string `control:"Codename"`

	Components    []string `control:"Components"`
	Architectures []string `control:"Architectures"`

	Date         string      `control:"Date"`
	ValidUntil   string      `control:"Valid-Until"`
	PackagePaths []string    `control:"-"`
	MD5Sum       []IndexHash `control:"MD5Sum" delim:"\n" strip:"\n\r\t "`
	SHA1         []IndexHash `control:"SHA1" delim:"\n" strip:"\n\r\t "`
	SHA256       []IndexHash `control:"SHA256" delim:"\n" strip:"\n\r\t "`
}

// IndexHash contains hash and size of a Packages file
type IndexHash struct {
	Hash string
	Size int64
	Path string
}

// UnmarshalControl tells how to unmarshal the control files
func (i *IndexHash) UnmarshalControl(data string) error {
	splitter := func(r rune) bool {
		return r == '\t' || r == ' '
	}
	parts := strings.FieldsFunc(data, splitter)
	if len(parts) != 3 {
		return nil
	}
	i.Hash = parts[0]
	s, err := strconv.ParseInt(parts[1], 10, 64)
	if err != nil {
		return fmt.Errorf("can't unmarshal size: %s", err)
	}
	i.Size = s
	i.Path = parts[2]
	return nil
}

// ParseRelease parses the Release file of a repository
func ParseRelease(r io.Reader) (*Release, error) {
	release := &Release{}
	if err := control.Unmarshal(release, r); err != nil {
		return nil, err
	}

	if len(release.MD5Sum) != 0 {
		release.PackagePaths = make([]string, len(release.MD5Sum))
		for i, h := range release.MD5Sum {
			release.PackagePaths[i] = h.Path
		}
	} else if len(release.SHA1) != 0 {
		release.PackagePaths = make([]string, len(release.SHA1))
		for i, h := range release.SHA1 {
			release.PackagePaths[i] = h.Path
		}
	} else if len(release.SHA256) != 0 {
		release.PackagePaths = make([]string, len(release.SHA256))
		for i, h := range release.SHA256 {
			release.PackagePaths[i] = h.Path
		}
	}
	return release, nil
}

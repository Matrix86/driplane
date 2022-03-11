package apt

import (
	"compress/bzip2"
	"compress/gzip"
	"fmt"
	"io"
	"pault.ag/go/debian/control"
)

type Index struct {
	Type     string
	Binaries []BinaryPackage
	// TODO: implement also the SourcePackages
}

type BinaryPackage struct {
	// mandatory
	Filename string
	Size     string

	// optional
	BinaryPackage  string
	MD5sum         string
	SHA1           string
	SHA256         string
	DescriptionMD5 string
	Depends        []string `delim:", " strip:"\n\r\t "`
	InstalledSize  string   `control:"Installed-Size"`
	Package        string
	Architecture   string
	Version        string
	Section        string
	Maintainer     string
	Homepage       string
	Description    string
	Tag            string
	Author         string
	Name           string
}

func ParsePackageIndex(r io.Reader, mtype string) (*Index, error) {
	reader := r
	var err error
	if mtype == "bz2" {
		reader = bzip2.NewReader(r)
	} else if mtype == "gz" {
		reader, err = gzip.NewReader(r)
		if err != nil {
			return nil, fmt.Errorf("gzip error: %s", err)
		}
	}
	p := &Index{Type: mtype}
	if err := control.Unmarshal(&p.Binaries, reader); err != nil {
		return nil, err
	}
	return p, nil
}

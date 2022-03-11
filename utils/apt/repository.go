package apt

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

type Repository struct {
	rootArchiveURL string
	distribution   string
	releaseURL     string
	indexURL       string
	isFlat         *bool
	architecture   string

	releaseFile  *Release
	packagesFile *Index
}

func NewRepository(root string, suite string) (*Repository, error) {
	_, err := url.Parse(root)
	if err != nil {
		return nil, err
	}

	r := &Repository{
		rootArchiveURL: root,
		distribution:   suite,
	}
	if !r.IsFlat() {
		if err := r.findRelease(); err != nil {
			return nil, err
		}
	}
	return r, nil
}

func (r *Repository) GetDistribution() string {
	return r.distribution
}

func (r *Repository) GetReleaseURL() string {
	return r.releaseURL
}

func (r *Repository) GetIndexURL() string {
	return r.indexURL
}

func (r *Repository) ForceIndexURL(u string) {
	b := true
	r.isFlat = &b
	r.indexURL = u
}

func (r *Repository) GetArchitectures() []string {
	if r.releaseFile != nil {
		return r.releaseFile.Architectures
	}
	return nil
}

func (r *Repository) SetArchitectures(arch string) error {
	if r.IsFlat() {
		return nil
	}
	if r.releaseFile == nil {
		return fmt.Errorf("release file not parsed")
	}
	for _, a := range r.releaseFile.Architectures {
		if a == arch {
			r.architecture = arch
			// reset the packages file if it has been already parsed for a different arch
			if r.packagesFile != nil {
				r.packagesFile = nil
			}
			return nil
		}
	}
	return fmt.Errorf("architecture '%s' not found in the Release list: %v", arch, r.releaseFile.Architectures)
}

func getFileTo(url string, w io.Writer) error {
	req, err := http.Get(url)
	if err != nil {
		return fmt.Errorf("cannot get %s: %s", url, err)
	}
	defer req.Body.Close()
	if req.StatusCode != 200 {
		return fmt.Errorf("cannot download %s: %v", url, req.Status)
	}
	_, err = io.Copy(w, req.Body)
	return err
}

func (r *Repository) IsFlat() bool {
	isFlat := false
	if r.isFlat == nil {
		paths := []string{
			fmt.Sprintf("%s/Packages.bz2", r.rootArchiveURL),
			fmt.Sprintf("%s/Packages.gz", r.rootArchiveURL),
			fmt.Sprintf("%s/Packages", r.rootArchiveURL),
		}
		for _, packageURL := range paths {
			res, err := http.Head(packageURL)
			if err != nil {
				continue
			}
			if res.StatusCode == 200 {
				isFlat = true
				r.indexURL = packageURL
			} else {
				continue
			}
			break
		}
		r.isFlat = &isFlat
	}
	return *r.isFlat
}

func (r *Repository) GetIndex() *Index {
	return r.packagesFile
}

func (r *Repository) GetRelease() *Release {
	return r.releaseFile
}

func (r *Repository) ReloadPackages() error {
	err := r.findIndex()
	if err != nil {
		return fmt.Errorf("retrieving index: %s", err)
	}
	return nil
}

func (r *Repository) GetPackages() ([]BinaryPackage, error) {
	if r.packagesFile == nil {
		err := r.ReloadPackages()
		if err != nil {
			return nil, err
		}
	}
	return r.packagesFile.Binaries, nil
}

func (r *Repository) findIndex() error {
	if r.IsFlat() {
		buff := bytes.NewBuffer(nil)
		err := getFileTo(r.indexURL, buff)
		if err != nil {
			return fmt.Errorf("can't download index file '%s': %s", r.indexURL, err)
		}

		mtype := ""
		if strings.HasSuffix(r.indexURL, ".bz2") {
			mtype = "bz2"
		} else if strings.HasSuffix(r.indexURL, ".gz") {
			mtype = "bz2"
		}

		p, err := ParsePackageIndex(buff, mtype)
		if err != nil {
			return fmt.Errorf("can't parse the index '%s': %s", r.indexURL, err)
		}
		r.packagesFile = p
		return nil
	}

	if r.releaseFile == nil {
		return fmt.Errorf("release file not parsed")
	}

	if r.architecture == "" {
		return fmt.Errorf("need to set the architecture")
	}

	archPaths := []string{
		fmt.Sprintf("/binary-%s/Packages.bz2", r.architecture),
		fmt.Sprintf("/binary-%s/Packages.gz", r.architecture),
		fmt.Sprintf("/binary-%s/Packages", r.architecture),
	}
	for _, archPath := range archPaths {
		for _, path := range r.releaseFile.PackagePaths {
			if strings.HasSuffix(path, archPath) {
				packagesURL := fmt.Sprintf("%s/dists/%s/%s", r.rootArchiveURL, r.distribution, path)
				buff := bytes.NewBuffer(nil)
				err := getFileTo(packagesURL, buff)
				if err != nil {
					continue
				}

				r.indexURL = packagesURL
				mtype := ""
				if strings.HasSuffix(r.indexURL, ".bz2") {
					mtype = "bz2"
				} else if strings.HasSuffix(r.indexURL, ".gz") {
					mtype = "bz2"
				}

				p, err := ParsePackageIndex(buff, mtype)
				if err != nil {
					return fmt.Errorf("can't parse the index '%s': %s", r.indexURL, err)
				}
				r.packagesFile = p
				return nil
			}
		}
	}
	return fmt.Errorf("can't find package indexes")
}

func (r *Repository) findRelease() error {
	if r.IsFlat() {
		return fmt.Errorf("flat repositories don't have the Release file")
	}

	releaseURL := fmt.Sprintf("%s/dists/%s/Release", r.rootArchiveURL, r.distribution)
	res, err := http.Head(releaseURL)
	if err != nil {
		return err
	}
	if res.StatusCode == 200 {
		r.releaseURL = releaseURL
		buf := bytes.NewBuffer(nil)
		if err = getFileTo(releaseURL, buf); err != nil {
			return fmt.Errorf("cannot download %s: %s", releaseURL, err)
		}
		release, err := ParseRelease(buf)
		if err != nil {
			return fmt.Errorf("cannot parse %s: %s", releaseURL, err)
		}
		r.releaseFile = release
	} else {
		return fmt.Errorf("release file not found on '%s'", releaseURL)
	}
	return nil
}

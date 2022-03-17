package apt

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Repository is the object that represents an APT repository
type Repository struct {
	rootArchiveURL string
	distribution   string
	releaseURL     string
	indexURL       string
	isFlat         *bool
	architecture   string
	userAgent      string

	releaseFile  *Release
	packagesFile *Index
	context      context.Context
}

// NewRepository creates a new Repository object
func NewRepository(ctx context.Context, root string, suite string, userAgent string) (*Repository, error) {
	_, err := url.Parse(root)
	if err != nil {
		return nil, err
	}

	r := &Repository{
		rootArchiveURL: root,
		distribution:   suite,
		context:        ctx,
		userAgent:      userAgent,
	}
	if root != "" && !r.IsFlat() {
		if err := r.findRelease(); err != nil {
			return nil, err
		}
	}
	return r, nil
}

// GetDistribution returns the distribution
func (r *Repository) GetDistribution() string {
	return r.distribution
}

// GetReleaseURL returns the URL of the Release file
func (r *Repository) GetReleaseURL() string {
	return r.releaseURL
}

// GetIndexURL returns the URL of the Packages file
func (r *Repository) GetIndexURL() string {
	return r.indexURL
}

// ForceIndexURL force the URL of the Packages file to parse bypassing the parsing of the Release
func (r *Repository) ForceIndexURL(u string) {
	b := true
	r.isFlat = &b
	r.indexURL = u
}

// GetArchitectures returns the architectures found in the Release file
func (r *Repository) GetArchitectures() []string {
	if r.releaseFile != nil {
		return r.releaseFile.Architectures
	}
	return nil
}

// SetArchitecture sets the architecture to parse the corrispondent Packages file
func (r *Repository) SetArchitecture(arch string) error {
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

func (r *Repository) getFileTo(url string, w io.Writer) error {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(r.context, "GET", url, nil)
	if err != nil {
		return fmt.Errorf("cannot get %s: %s", url, err)
	}
	if r.userAgent != "" {
		req.Header.Set("User-Agent", r.userAgent)
	}
	resp, err := client.Do(req)
	if err != nil {
		return fmt.Errorf("cannot get %s: %s", url, err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return fmt.Errorf("cannot download %s: %v", url, resp.Status)
	}
	_, err = io.Copy(w, resp.Body)
	return err
}

func (r *Repository) headRequest(url string) (int, error) {
	client := &http.Client{}
	req, err := http.NewRequestWithContext(r.context, "HEAD", url, nil)
	if err != nil {
		return -1, fmt.Errorf("cannot get %s: %s", url, err)
	}
	if r.userAgent != "" {
		req.Header.Set("User-Agent", r.userAgent)
	}
	resp, err := client.Do(req)
	if err != nil {
		return -1, fmt.Errorf("cannot get %s: %s", url, err)
	}
	defer resp.Body.Close()

	return resp.StatusCode, nil
}

// IsFlat returns true is the repository is a flat repo
func (r *Repository) IsFlat() bool {
	isFlat := false
	if r.isFlat == nil {
		paths := []string{
			fmt.Sprintf("%s/Packages.bz2", r.rootArchiveURL),
			fmt.Sprintf("%s/Packages.gz", r.rootArchiveURL),
			fmt.Sprintf("%s/Packages", r.rootArchiveURL),
		}
		for _, packageURL := range paths {
			statusCode, err := r.headRequest(packageURL)
			if err != nil {
				continue
			}
			if statusCode == 200 {
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

// GetIndex returns the parsed Index object
func (r *Repository) GetIndex() *Index {
	return r.packagesFile
}

// GetRelease returns the parsed Release object
func (r *Repository) GetRelease() *Release {
	return r.releaseFile
}

// ReloadPackages parses again the Packages file
func (r *Repository) ReloadPackages() error {
	err := r.findIndex()
	if err != nil {
		return fmt.Errorf("retrieving index: %s", err)
	}
	return nil
}

// GetPackages returns the binary packages from the Packages file
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
		err := r.getFileTo(r.indexURL, buff)
		if err != nil {
			return fmt.Errorf("can't download index file '%s': %s", r.indexURL, err)
		}

		mtype := ""
		if strings.HasSuffix(r.indexURL, ".bz2") {
			mtype = "bz2"
		} else if strings.HasSuffix(r.indexURL, ".gz") {
			mtype = "gz"
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
				err := r.getFileTo(packagesURL, buff)
				if err != nil {
					continue
				}

				r.indexURL = packagesURL
				mtype := ""
				if strings.HasSuffix(r.indexURL, ".bz2") {
					mtype = "bz2"
				} else if strings.HasSuffix(r.indexURL, ".gz") {
					mtype = "gz"
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
	statusCode, err := r.headRequest(releaseURL)
	if err != nil {
		return err
	}
	if statusCode == 200 {
		r.releaseURL = releaseURL
		buf := bytes.NewBuffer(nil)
		if err = r.getFileTo(releaseURL, buf); err != nil {
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

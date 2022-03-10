package apt

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

var (
	releaseFile = `Origin: ModMyi (Archive)
Label: ModMyi
Suite: stable
Codename: stable
Version: 1.0
Support: http://cydia.saurik.com/support/*
Architectures: iphoneos-arm
Components: main
Description: ModMyi.com - they hosted your apps!
MD5Sum:
 4f03f476fbbfcfc04b7a471bfb6b32e3 15040211 main/binary-iphoneos-arm/Packages
 52e2b38b0ca40710930966579ea81639 1940768 main/binary-iphoneos-arm/Packages.bz2`

	packageFile = `BinaryPackage: me.absidue.ignoexplore
architecture: iphoneos-arm
Version: 0.0.6
Section: Tweaks
Maintainer: absidue
Installed-Size: 88
Depends: firmware (>= 13.0), org.coolstar.libhooker | mobilesubstrate
Filename: ./pool/me.absidue.ignoexplore_0.0.6_iphoneos-arm.deb
Size: 3750
MD5sum: 99b33f7a6b0ff67be3603d1a5cdf57de
SHA1: d0656d068121bd82c81d425ef4af9de9222ebd21
SHA256: d84f3843586229f3f5682a301e23e25a34d9b86e0296ebf99d6ffdddc0c3101f
SHA512: 97cfee4cdd3c7fec1f31a23110bc5c84032f0b6126f353b82ad7da9d97c85c55a588955b38b54b7f8c83f27cd2bf9ede1276161a1c1554a11cfbace7adba9dc5
Homepage: https://github.com/absidue/IGNoExplore/
Description: Hides the explore feed in Instagram and enters search field straight away
 Hides the explore feed on the Instagram search page and focuses the search field immediately.
 .
 Works with the latest Instagram version at the time of release, please file an issue on the GitHub page if it's broken for you.
 .
 No options to configure.
Tag: compatible_min::ios13.0
Name: IGNoExplore
Author: absidue
Icon: https://apt.absidue.me/assets/icons/me.absidue.ignoexplore.png

BinaryPackage: me.absidue.musicrecognitionmoreapps
architecture: iphoneos-arm
Version: 1.0.1
Section: Tweaks
Maintainer: absidue
Installed-Size: 168
Depends: firmware (>= 14.0), mobilesubstrate (>= 0.9.5000)
Filename: ./pool/me.absidue.musicrecognitionmoreapps_1.0.1_iphoneos-arm.deb
Size: 8006
MD5sum: c0b64cea641a9da67ba7de1513d844eb
SHA1: 661811ebc49314e88b414971be3aa9c47d71603d
SHA256: c4a6d899f8a4cb2877bd5a8fa4a8130790dc65c805ff4cbbf9a34a24c1ca628f
SHA512: 6c31b0456cf2e923cd408837a1db61daac927ca5e3387201893de71a7fa5e1ede29a667d9a16fc4b29d28418a4a8748b150511746820d598c28b16c6a74a0134
Homepage: https://github.com/absidue/MusicRecognitionMoreApps/
Description: Open Music Recognition results in more apps
 Open Music Recognition results in more apps.
 Automatically detects which apps you have installed and shows the relevant apps.
 .
 Currently supported apps:
 - YouTube
 - YouTube Music
 - Spotify
 - Deezer
 .
 If your music app is not supported yet, please open an issue or pull request on the GitHub page.
 .
 No options to configure.
Tag: compatible_min::ios14.0
Name: MusicRecognitionMoreApps
Author: absidue

BinaryPackage: me.absidue.musicrecognitionmoreapps
architecture: iphoneos-arm
Version: 1.0
Section: Tweaks
Maintainer: absidue
Installed-Size: 168
Depends: firmware (>= 14.0), mobilesubstrate (>= 0.9.5000)
Filename: ./pool/me.absidue.musicrecognitionmoreapps_1.0_iphoneos-arm.deb
Size: 7088
MD5sum: d7e8d18c16fb947d923969af9f82ba4d
SHA1: 5762101b97655fc7a2313cc75d3f5ebff15c03c3
SHA256: fefc36170b48e391c96a1a8c3aac78d1ab3f95b9eb461408c22be53e039eb570
SHA512: 8d8784919fadf14032889abb6c5e3c933fd561d4da5d6f92c6c24f9ba774f15c0e539df32d0d01a6644c37984b54218a6af05b12361825587c0047ba0df9e12f
Homepage: https://github.com/absidue/MusicRecognitionMoreApps/
Description: Open Music Recognition results in more apps
 Open Music Recognition results in more apps.
 Automatically detects which apps you have installed and shows the relevant apps.
 .
 Currently supported apps:
 - YouTube
 - YouTube Music
 - Spotify
 - Deezer
 .
 If your music app is not supported yet, please open an issue or pull request on the GitHub page.
 .
 No options to configure.
Tag: compatible_min::ios14.0
Name: MusicRecognitionMoreApps
Author: absidue`
)

func TestNewRepository(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else if strings.Contains(r.URL.String(), "/error") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else {
			_, _ = fmt.Fprintf(w, releaseFile)
		}
	}))
	defer svr.Close()

	_, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}

	_, err = NewRepository("http://192.168.0.%31:8080/", "stable")
	if err == nil {
		t.Errorf("error expected but it is nil")
	}

	_, err = NewRepository(svr.URL, "error")
	if err == nil {
		t.Errorf("error expected but it is nil")
	}
}

func TestRepository_GetArchitectures(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else if strings.Contains(r.URL.String(), "/error") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else {
			_, _ = fmt.Fprintf(w, releaseFile)
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if len(repo.GetArchitectures()) != 1 {
		t.Errorf("it should contains only 1 architecture")
	}
}

func TestRepository_GetDistribution(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else {
			_, _ = fmt.Fprintf(w, releaseFile)
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if repo.GetDistribution() != "stable" {
		t.Errorf("wrong distribution returned")
	}
}

func TestRepository_GetReleaseURL(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else {
			_, _ = fmt.Fprintf(w, releaseFile)
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if repo.GetReleaseURL() != fmt.Sprintf("%s/dists/stable/Release", svr.URL) {
		t.Errorf("wrong distribution returned: %s", repo.GetReleaseURL())
	}
}

func TestRepository_GetRelease(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else {
			_, _ = fmt.Fprintf(w, releaseFile)
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if repo.GetRelease() == nil {
		t.Errorf("the release hasn't been parsed")
	}
}

func TestRepository_SetArchitectures(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else {
			_, _ = fmt.Fprintf(w, releaseFile)
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if repo.SetArchitectures("iphoneos-arm") != nil {
		t.Errorf("the architecture hasn't been found")
	}
	if repo.SetArchitectures("fake-one") == nil {
		t.Errorf("the function should return nil")
	}
}

func TestRepository_GetPackages(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.String(), "binary-iphoneos-arm/Packages") {
			_, _ = fmt.Fprintf(w, packageFile)
		} else if strings.HasSuffix(r.URL.String(), "/Packages") {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		} else if strings.HasSuffix(r.URL.String(), "stable/Release") {
			_, _ = fmt.Fprintf(w, releaseFile)
		} else {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	err = repo.SetArchitectures("iphoneos-arm")
	if err != nil {
		t.Errorf("set architecture: %s", err)
	}
	_, err = repo.GetPackages()
	if err != nil {
		t.Errorf("problem on GetPackages: %s", err)
	}
	packages, err := repo.GetPackages()
	if err != nil {
		t.Errorf("problem on GetPackages: %s", err)
	}
	if len(packages) != 3 {
		t.Errorf("wrong number of packages: %d", len(packages))
	}
}

func TestRepository_GetPackages2(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.String(), "/Packages") {
			_, _ = fmt.Fprintf(w, packageFile)
		} else if strings.HasSuffix(r.URL.String(), "stable/Release") {
			_, _ = fmt.Fprintf(w, releaseFile)
		} else {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if !repo.IsFlat() {
		t.Errorf("it should be a flat repo")
	}
	_, err = repo.GetPackages()
	if err != nil {
		t.Errorf("problem on GetPackages: %s", err)
	}
	packages, err := repo.GetPackages()
	if err != nil {
		t.Errorf("problem on GetPackages: %s", err)
	}
	if len(packages) != 3 {
		t.Errorf("wrong number of packages: %d", len(packages))
	}
}

func TestRepository_ForceIndexURL(t *testing.T) {
	svr := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.String(), "/Packages") {
			_, _ = fmt.Fprintf(w, packageFile)
		} else if strings.HasSuffix(r.URL.String(), "stable/Release") {
			_, _ = fmt.Fprintf(w, releaseFile)
		} else {
			w.WriteHeader(404)
			_, _ = w.Write([]byte("not found"))
		}
	}))
	defer svr.Close()

	repo, err := NewRepository(svr.URL, "stable")
	if err != nil {
		t.Errorf("can't create repository: %s", err)
	}
	if !repo.IsFlat() {
		t.Errorf("it should be a flat repo")
	}
	_, err = repo.GetPackages()
	if err != nil {
		t.Errorf("problem on GetPackages: %s", err)
	}
	packages, err := repo.GetPackages()
	if err != nil {
		t.Errorf("problem on GetPackages: %s", err)
	}
	if len(packages) != 3 {
		t.Errorf("wrong number of packages: %d", len(packages))
	}
}

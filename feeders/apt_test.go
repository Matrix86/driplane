package feeders

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/Matrix86/driplane/utils/apt"
	"github.com/asaskevich/EventBus"
)

var testReleaseFile = `Origin: TestRepo
Label: TestRepo
Suite: stable
Codename: stable
Version: 1.0
Architectures: amd64 arm64
Components: main
Description: A test APT repository
MD5Sum:
 d41d8cd98f00b204e9800998ecf8427e 12345 main/binary-amd64/Packages
 d41d8cd98f00b204e9800998ecf8427e 12345 main/binary-arm64/Packages`

var testPackageFile = `Package: testpkg-one
Architecture: amd64
Version: 1.2.3
Section: utils
Maintainer: tester@example.com
Installed-Size: 256
Depends: libc6 (>= 2.17), libssl1.1
Filename: pool/main/t/testpkg-one/testpkg-one_1.2.3_amd64.deb
Size: 50000
MD5sum: abcdef1234567890abcdef1234567890
SHA1: 1234567890abcdef1234567890abcdef12345678
SHA256: abcdef1234567890abcdef1234567890abcdef1234567890abcdef1234567890
Homepage: https://example.com/testpkg-one
Description: First test package
Name: TestPkgOne
Author: Tester

Package: testpkg-two
Architecture: amd64
Version: 4.5.6
Section: net
Maintainer: dev@example.com
Installed-Size: 512
Filename: pool/main/t/testpkg-two/testpkg-two_4.5.6_amd64.deb
Size: 70000
MD5sum: fedcba0987654321fedcba0987654321
SHA1: abcdef1234567890abcdef1234567890abcdef12
SHA256: fedcba0987654321fedcba0987654321fedcba0987654321fedcba0987654321
Homepage: https://example.com/testpkg-two
Description: Second test package
Name: TestPkgTwo
Author: Developer`

func newAptTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "binary-amd64/Packages"):
			fmt.Fprint(w, testPackageFile)
		case strings.HasSuffix(r.URL.Path, "stable/Release"):
			fmt.Fprint(w, testReleaseFile)
		case strings.HasSuffix(r.URL.Path, "/Packages"):
			// Not the flat-repo index; 404 so IsFlat() returns false
			w.WriteHeader(http.StatusNotFound)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

func newAptFlatTestServer() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/Packages") {
			fmt.Fprint(w, testPackageFile)
		} else {
			w.WriteHeader(http.StatusNotFound)
		}
	}))
}

// newTestApt creates an Apt feeder wired to an EventBus so Propagate works.
func newTestApt(conf map[string]string) (*Apt, chan *data.Message, error) {
	feeder, err := NewAptFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*Apt)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *Apt")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("aptfeeder")
	f.setID(1)

	received := make(chan *data.Message, 50)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

func TestNewAptFeeder(t *testing.T) {
	ts := newAptTestServer()
	defer ts.Close()

	feeder, err := NewAptFeeder(map[string]string{
		"apt.url":   ts.URL,
		"apt.suite": "stable",
		"apt.arch":  "amd64",
		"apt.freq":  "5m",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := feeder.(*Apt)
	if !ok {
		t.Fatal("cannot cast to *Apt")
	}
	if f.url != ts.URL {
		t.Errorf("'apt.url' not set correctly, got '%s'", f.url)
	}
	if f.distribution != "stable" {
		t.Errorf("'apt.suite' not set correctly, got '%s'", f.distribution)
	}
	if f.architecture != "amd64" {
		t.Errorf("'apt.arch' not set correctly, got '%s'", f.architecture)
	}
	if f.frequency != 5*time.Minute {
		t.Errorf("'apt.freq' not set correctly, got '%s'", f.frequency)
	}
}

func TestNewAptFeederDefaults(t *testing.T) {
	feeder, err := NewAptFeeder(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := feeder.(*Apt)
	if !ok {
		t.Fatal("cannot cast to *Apt")
	}
	if f.distribution != "stable" {
		t.Errorf("default 'apt.suite' should be 'stable', got '%s'", f.distribution)
	}
	if f.frequency != 60*time.Second {
		t.Errorf("default 'apt.freq' should be 60s, got '%s'", f.frequency)
	}
	if f.insecure {
		t.Errorf("default 'apt.insecure' should be false")
	}
}

func TestNewAptFeederInvalidFreq(t *testing.T) {
	_, err := NewAptFeeder(map[string]string{
		"apt.freq": "notaduration",
	})
	if err == nil {
		t.Errorf("constructor should return an error for invalid frequency")
	}
}

func TestNewAptFeederUserAgent(t *testing.T) {
	feeder, err := NewAptFeeder(map[string]string{
		"apt.useragent": "TestBot/1.0",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Apt)
	if f.userAgent != "TestBot/1.0" {
		t.Errorf("'apt.useragent' not set correctly, got '%s'", f.userAgent)
	}
}

func TestNewAptFeederInsecure(t *testing.T) {
	feeder, err := NewAptFeeder(map[string]string{
		"apt.insecure": "true",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Apt)
	if !f.insecure {
		t.Errorf("'apt.insecure' should be true")
	}
}

func TestNewAptFeederIndexURL(t *testing.T) {
	feeder, err := NewAptFeeder(map[string]string{
		"apt.index": "http://example.com/repo/Packages",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Apt)
	if f.indexURL != "http://example.com/repo/Packages" {
		t.Errorf("'apt.index' not set correctly, got '%s'", f.indexURL)
	}
	if f.url != "http://example.com/repo" {
		t.Errorf("url should be derived from index path, got '%s'", f.url)
	}
}

// ---------- parseFeed tests ----------

func TestAptParseFeedWithRelease(t *testing.T) {
	ts := newAptTestServer()
	defer ts.Close()

	f, received, err := newTestApt(map[string]string{
		"apt.url":   ts.URL,
		"apt.suite": "stable",
		"apt.arch":  "amd64",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	var msgs []*data.Message
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages propagated, got %d", len(msgs))
	}

	// all messages from first run should have firstRun set
	for i, msg := range msgs {
		if !msg.IsFirstRun() {
			t.Errorf("message %d should have firstRun set", i)
		}
	}

	// check that extras are populated
	first := msgs[0]
	extra := first.GetExtra()
	if _, ok := extra["package"]; !ok {
		t.Errorf("expected 'package' extra field")
	}
	if _, ok := extra["version"]; !ok {
		t.Errorf("expected 'version' extra field")
	}
	if _, ok := extra["filename"]; !ok {
		t.Errorf("expected 'filename' extra field")
	}
	if _, ok := extra["link"]; !ok {
		t.Errorf("expected 'link' extra field")
	}

	// link should be the full URL
	link, _ := extra["link"].(string)
	if !strings.HasPrefix(link, ts.URL) {
		t.Errorf("expected link to start with server URL, got '%s'", link)
	}

	// main should be the filename
	main, _ := first.GetMessage().(string)
	if main == "" {
		t.Errorf("expected non-empty main message (filename)")
	}
}

func TestAptParseFeedWithIndex(t *testing.T) {
	ts := newAptFlatTestServer()
	defer ts.Close()

	indexURL := ts.URL + "/Packages"

	f, received, err := newTestApt(map[string]string{
		"apt.index": indexURL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(false)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	var msgs []*data.Message
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages propagated, got %d", len(msgs))
	}

	// messages should NOT have firstRun set
	for i, msg := range msgs {
		if msg.IsFirstRun() {
			t.Errorf("message %d should not have firstRun set", i)
		}
	}
}

func TestAptParseFeedAutoArch(t *testing.T) {
	ts := newAptTestServer()
	defer ts.Close()

	// Do not set apt.arch — should auto-pick the first architecture
	f, _, err := newTestApt(map[string]string{
		"apt.url":   ts.URL,
		"apt.suite": "stable",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	if f.architecture != "amd64" {
		t.Errorf("expected auto-selected architecture 'amd64', got '%s'", f.architecture)
	}
}

// ---------- getExtraFromPackage tests ----------

func TestGetExtraFromPackage(t *testing.T) {
	ts := newAptTestServer()
	defer ts.Close()

	f, _, err := newTestApt(map[string]string{
		"apt.url":   ts.URL,
		"apt.suite": "stable",
		"apt.arch":  "amd64",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// Manually build a BinaryPackage to test extra extraction
	pkg := newTestBinaryPackage()
	extra := f.getExtraFromPackage(&pkg)

	if v, ok := extra["package"]; !ok || v != "testpkg" {
		t.Errorf("expected extra['package'] = 'testpkg', got '%v'", v)
	}
	if v, ok := extra["version"]; !ok || v != "1.0.0" {
		t.Errorf("expected extra['version'] = '1.0.0', got '%v'", v)
	}
	if v, ok := extra["filename"]; !ok || v != "pool/testpkg_1.0.0.deb" {
		t.Errorf("expected extra['filename'] = 'pool/testpkg_1.0.0.deb', got '%v'", v)
	}
	if v, ok := extra["link"]; !ok {
		t.Errorf("expected 'link' extra field")
	} else {
		link := v.(string)
		if !strings.Contains(link, "pool/testpkg_1.0.0.deb") {
			t.Errorf("expected link to contain filename, got '%s'", link)
		}
	}
	// depends is a []string field → should be comma-joined quoted values
	if v, ok := extra["depends"]; !ok {
		t.Errorf("expected 'depends' extra field")
	} else {
		dep := v.(string)
		if !strings.Contains(dep, "'libc6'") {
			t.Errorf("expected depends to contain quoted 'libc6', got '%s'", dep)
		}
	}
}

func TestGetExtraFromPackageNoFilename(t *testing.T) {
	ts := newAptTestServer()
	defer ts.Close()

	f, _, err := newTestApt(map[string]string{
		"apt.url":   ts.URL,
		"apt.suite": "stable",
		"apt.arch":  "amd64",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	pkg := newTestBinaryPackage()
	pkg.Filename = ""
	extra := f.getExtraFromPackage(&pkg)

	// The code always sets "link" when the "filename" key exists in the extra map,
	// even if the value is an empty string. So link will be "<url>/".
	link, ok := extra["link"]
	if !ok {
		t.Fatalf("expected 'link' extra field to exist")
	}
	linkStr := link.(string)
	expectedSuffix := ts.URL + "/"
	if linkStr != expectedSuffix {
		t.Errorf("expected link '%s', got '%s'", expectedSuffix, linkStr)
	}
}

// ---------- Start / Stop tests ----------

func TestAptStartStop(t *testing.T) {
	ts := newAptTestServer()
	defer ts.Close()

	f, _, err := newTestApt(map[string]string{
		"apt.url":   ts.URL,
		"apt.suite": "stable",
		"apt.arch":  "amd64",
		"apt.freq":  "10s",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	if !f.isRunning {
		t.Errorf("feeder should be running after Start()")
	}

	// give the goroutine time to do the initial parse
	time.Sleep(300 * time.Millisecond)

	f.Stop()
	if f.isRunning {
		t.Errorf("feeder should not be running after Stop()")
	}
}

// ---------- OnEvent test ----------

func TestAptOnEvent(t *testing.T) {
	f, _, err := newTestApt(map[string]string{})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// OnEvent is a no-op; just ensure it doesn't panic
	event := &data.Event{}
	f.OnEvent(event)
}

// ---------- init registration test ----------

func TestAptFeederRegistered(t *testing.T) {
	if _, ok := feederFactories["aptfeeder"]; !ok {
		t.Errorf("apt feeder should be registered as 'aptfeeder'")
	}
}

// ---------- test helpers ----------

func newTestBinaryPackage() apt.BinaryPackage {
	return apt.BinaryPackage{
		Package:      "testpkg",
		Architecture: "amd64",
		Version:      "1.0.0",
		Filename:     "pool/testpkg_1.0.0.deb",
		Size:         "12345",
		MD5sum:       "abcdef1234567890",
		SHA1:         "1234567890abcdef",
		SHA256:       "abcdef1234567890abcdef",
		Maintainer:   "tester@example.com",
		Homepage:     "https://example.com",
		Description:  "A test package",
		Section:      "utils",
		Name:         "TestPkg",
		Author:       "Tester",
		Depends:      []string{"libc6", "libssl1.1"},
	}
}

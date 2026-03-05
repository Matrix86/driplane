package feeders

import (
	"fmt"
	"os"
	"testing"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
)

// ---------- helpers ----------

// newTestFolder creates a Folder feeder wired to an EventBus so Propagate works.
func newTestFolder(conf map[string]string) (*Folder, chan *data.Message, error) {
	feeder, err := NewFolderFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*Folder)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *Folder")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("folderfeeder")
	f.setID(1)

	received := make(chan *data.Message, 50)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

// ---------- constructor tests ----------

func TestNewFolderFeederLocal(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_test_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	feeder, err := NewFolderFeeder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := feeder.(*Folder)
	if !ok {
		t.Fatal("cannot cast to *Folder")
	}
	if f.serviceName != "local" {
		t.Errorf("expected serviceName 'local', got '%s'", f.serviceName)
	}
	if f.folderName == "" {
		t.Errorf("expected folderName to be set")
	}
	if f.watcher == nil {
		t.Errorf("expected watcher to be initialized")
	}
}

func TestNewFolderFeederDefaults(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_defaults_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	feeder, err := NewFolderFeeder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Folder)
	if f.frequency != 2*time.Second {
		t.Errorf("default frequency should be 2s, got '%s'", f.frequency)
	}
}

func TestNewFolderFeederCustomFreq(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_freq_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	feeder, err := NewFolderFeeder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
		"folder.freq": "10s",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Folder)
	if f.frequency != 10*time.Second {
		t.Errorf("expected frequency 10s, got '%s'", f.frequency)
	}
}

func TestNewFolderFeederInvalidFreq(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_badfreq_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = NewFolderFeeder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
		"folder.freq": "notaduration",
	})
	if err == nil {
		t.Errorf("expected error for invalid frequency")
	}
}

func TestNewFolderFeederInvalidServiceType(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_badtype_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = NewFolderFeeder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "nonexistent_service_type",
	})
	if err == nil {
		t.Errorf("expected error for unknown service type")
	}
}

func TestNewFolderFeederLocalAbsPath(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_abs_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	feeder, err := NewFolderFeeder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Folder)
	// For local type, the path should be absolute
	if len(f.folderName) == 0 || f.folderName[0] != '/' {
		t.Errorf("expected absolute path for local watcher, got '%s'", f.folderName)
	}
}

func TestNewFolderFeederExtraConfig(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_extra_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	feeder, err := NewFolderFeeder(map[string]string{
		"folder.name":       tmpDir,
		"folder.type":       "local",
		"folder.customkey1": "val1",
		"folder.customkey2": "val2",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*Folder)
	if v, ok := f.watcherConfig["customkey1"]; !ok || v != "val1" {
		t.Errorf("expected watcherConfig['customkey1'] = 'val1', got '%v'", v)
	}
	if v, ok := f.watcherConfig["customkey2"]; !ok || v != "val2" {
		t.Errorf("expected watcherConfig['customkey2'] = 'val2', got '%v'", v)
	}
}

// ---------- Start / Stop ----------

func TestFolderStartStop(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_startstop_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	f, _, err := newTestFolder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	if !f.isRunning {
		t.Errorf("feeder should be running after Start()")
	}

	time.Sleep(200 * time.Millisecond)

	f.Stop()
	if f.isRunning {
		t.Errorf("feeder should not be running after Stop()")
	}
}

// ---------- file event propagation ----------

func TestFolderPropagatesOnFileCreate(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_event_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	f, received, err := newTestFolder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	// Give the watcher time to initialize
	time.Sleep(500 * time.Millisecond)

	// Create a file in the watched folder
	tmpFile, err := os.CreateTemp(tmpDir, "event_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpFile.WriteString("hello\n")
	tmpFile.Close()

	select {
	case msg := <-received:
		main := msg.GetMessage()
		if main == nil || main == "" {
			t.Errorf("expected non-empty main message on file create event")
		}
	case <-time.After(10 * time.Second):
		t.Fatalf("timed out waiting for folder event")
	}
}

// ---------- OnEvent ----------

func TestFolderOnEvent(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "folderfeeder_onevent_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	f, _, err := newTestFolder(map[string]string{
		"folder.name": tmpDir,
		"folder.type": "local",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// OnEvent is a no-op; just verify it doesn't panic
	f.OnEvent(&data.Event{})
}

// ---------- init registration ----------

func TestFolderFeederRegistered(t *testing.T) {
	if _, ok := feederFactories["folderfeeder"]; !ok {
		t.Errorf("folder feeder should be registered as 'folderfeeder'")
	}
}

package feeders

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
)

// newTestFile creates a File feeder wired to an EventBus so Propagate works.
func newTestFile(conf map[string]string) (*File, chan *data.Message, error) {
	feeder, err := NewFileFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*File)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *File")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("filefeeder")
	f.setID(1)

	received := make(chan *data.Message, 50)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

func TestNewFileFeederValid(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_test_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	feeder, err := NewFileFeeder(map[string]string{
		"file.filename": tmpFile.Name(),
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := feeder.(*File)
	if !ok {
		t.Fatal("cannot cast to *File")
	}
	if f.filename != tmpFile.Name() {
		t.Errorf("expected filename '%s', got '%s'", tmpFile.Name(), f.filename)
	}
	if f.lastLines {
		t.Errorf("expected lastLines to be false by default")
	}
	f.fp.Stop()
	f.fp.Cleanup()
}

func TestNewFileFeederToEnd(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_toend_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	// Write some content so size > 0
	tmpFile.WriteString("existing content\n")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	feeder, err := NewFileFeeder(map[string]string{
		"file.filename": tmpFile.Name(),
		"file.toend":    "true",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := feeder.(*File)
	if !f.lastLines {
		t.Errorf("expected lastLines to be true when file.toend is 'true'")
	}
	f.fp.Stop()
	f.fp.Cleanup()
}

func TestNewFileFeederNonexistentFile(t *testing.T) {
	_, err := NewFileFeeder(map[string]string{
		"file.filename": "/tmp/this_file_should_not_exist_ever_12345.txt",
	})
	if err == nil {
		t.Errorf("expected error for non-existent file")
	}
}

func TestNewFileFeederDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filefeeder_dir_test_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	_, err = NewFileFeeder(map[string]string{
		"file.filename": tmpDir,
	})
	if err == nil {
		t.Errorf("expected error when filename is a directory")
	}
}

func TestNewFileFeederEmptyFilename(t *testing.T) {
	_, err := NewFileFeeder(map[string]string{})
	if err == nil {
		t.Errorf("expected error when filename is empty")
	}
}

func TestFileStartStopPropagation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_startstop_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpName)

	f, received, err := newTestFile(map[string]string{
		"file.filename": tmpName,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	if !f.isRunning {
		t.Errorf("feeder should be running after Start()")
	}

	// Append a line to the file
	fp, err := os.OpenFile(tmpName, os.O_APPEND|os.O_WRONLY, 0644)
	if err != nil {
		t.Fatalf("cannot open temp file for writing: %s", err)
	}
	_, err = fp.WriteString("test line\n")
	fp.Close()
	if err != nil {
		t.Fatalf("cannot write to temp file: %s", err)
	}

	// Wait for the message
	select {
	case msg := <-received:
		if msg.GetMessage() != "test line" {
			t.Errorf("expected main message 'test line', got '%v'", msg.GetMessage())
		}
		extra := msg.GetExtra()
		if v, ok := extra["file_name"]; !ok || v != tmpName {
			t.Errorf("expected file_name '%s', got '%v'", tmpName, v)
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for propagated message")
	}

	f.Stop()
	if f.isRunning {
		t.Errorf("feeder should not be running after Stop()")
	}
}

func TestFileMultipleLines(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_multi_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpName := tmpFile.Name()
	tmpFile.Close()
	defer os.Remove(tmpName)

	f, received, err := newTestFile(map[string]string{
		"file.filename": tmpName,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	// Write 3 lines
	fp, _ := os.OpenFile(tmpName, os.O_APPEND|os.O_WRONLY, 0644)
	fp.WriteString("line1\nline2\nline3\n")
	fp.Close()

	lines := make([]string, 0, 3)
	for i := 0; i < 3; i++ {
		select {
		case msg := <-received:
			lines = append(lines, msg.GetMessage().(string))
		case <-time.After(5 * time.Second):
			t.Fatalf("timed out waiting for message %d", i+1)
		}
	}

	expected := []string{"line1", "line2", "line3"}
	for i, exp := range expected {
		if lines[i] != exp {
			t.Errorf("message %d: expected '%s', got '%s'", i, exp, lines[i])
		}
	}
}

func TestFileToEndSkipsExisting(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_toend2_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpName := tmpFile.Name()
	tmpFile.WriteString("old line 1\nold line 2\n")
	tmpFile.Close()
	defer os.Remove(tmpName)

	f, received, err := newTestFile(map[string]string{
		"file.filename": tmpName,
		"file.toend":    "true",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	// Append a new line after start
	fp, _ := os.OpenFile(tmpName, os.O_APPEND|os.O_WRONLY, 0644)
	fp.WriteString("new line\n")
	fp.Close()

	select {
	case msg := <-received:
		if msg.GetMessage() != "new line" {
			t.Errorf("expected 'new line', got '%v'", msg.GetMessage())
		}
	case <-time.After(5 * time.Second):
		t.Fatalf("timed out waiting for new line")
	}
}

func TestFileOnEvent(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_event_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	defer os.Remove(tmpFile.Name())
	tmpFile.Close()

	f, _, err := newTestFile(map[string]string{
		"file.filename": tmpFile.Name(),
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}
	defer func() { f.fp.Stop(); f.fp.Cleanup() }()

	// OnEvent is a no-op; just verify it doesn't panic
	f.OnEvent(&data.Event{})
}

func TestFileFeederRegistered(t *testing.T) {
	if _, ok := feederFactories["filefeeder"]; !ok {
		t.Errorf("file feeder should be registered as 'filefeeder'")
	}
}

func TestFileFeederWithSymlink(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefeeder_sym_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	symlink := filepath.Join(os.TempDir(), "filefeeder_symlink_test.txt")
	os.Remove(symlink) // clean up any leftover
	if err := os.Symlink(tmpFile.Name(), symlink); err != nil {
		t.Skipf("cannot create symlink: %s", err)
	}
	defer os.Remove(symlink)

	feeder, err := NewFileFeeder(map[string]string{
		"file.filename": symlink,
	})
	if err != nil {
		t.Fatalf("constructor returned error for symlink: %s", err)
	}

	f := feeder.(*File)
	f.fp.Stop()
	f.fp.Cleanup()
}

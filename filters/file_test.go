package filters

import (
	"os"
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewFileFilter(t *testing.T) {
	filter, err := NewFileFilter(map[string]string{
		"target": "custom",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*File)
	if !ok {
		t.Fatal("cannot cast to *File")
	}
	if f.target != "custom" {
		t.Errorf("expected target 'custom', got '%s'", f.target)
	}
}

func TestNewFileFilterDefaults(t *testing.T) {
	filter, err := NewFileFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*File)
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
}

func TestFileDoFilterValidFile(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefilter_test_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	content := "file content here"
	tmpFile.WriteString(content)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	filter, _ := NewFileFilter(map[string]string{})
	f := filter.(*File)

	msg := data.NewMessage(tmpFile.Name())
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true for valid file")
	}
	if msg.GetMessage() != content {
		t.Errorf("expected message '%s', got '%v'", content, msg.GetMessage())
	}
}

func TestFileDoFilterCustomTarget(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "filefilter_target_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	content := "custom target content"
	tmpFile.WriteString(content)
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	filter, _ := NewFileFilter(map[string]string{"target": "output"})
	f := filter.(*File)

	msg := data.NewMessage(tmpFile.Name())
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true for valid file")
	}
	if msg.GetTarget("output") != content {
		t.Errorf("expected target 'output' = '%s', got '%v'", content, msg.GetTarget("output"))
	}
}

func TestFileDoFilterNonexistentFile(t *testing.T) {
	filter, _ := NewFileFilter(map[string]string{})
	f := filter.(*File)

	msg := data.NewMessage("/tmp/this_file_should_not_exist_ever_99999.txt")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Errorf("DoFilter should not return error for nonexistent file, got: %s", err)
	}
	if ok {
		t.Errorf("DoFilter should return false for nonexistent file")
	}
}

func TestFileDoFilterDirectory(t *testing.T) {
	tmpDir, err := os.MkdirTemp("", "filefilter_dir_*")
	if err != nil {
		t.Fatalf("cannot create temp dir: %s", err)
	}
	defer os.RemoveAll(tmpDir)

	filter, _ := NewFileFilter(map[string]string{})
	f := filter.(*File)

	msg := data.NewMessage(tmpDir)
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Errorf("DoFilter should not return error for directory, got: %s", err)
	}
	if ok {
		t.Errorf("DoFilter should return false for directory")
	}
}

func TestFileDoFilterNonStringMessage(t *testing.T) {
	filter, _ := NewFileFilter(map[string]string{})
	f := filter.(*File)

	msg := data.NewMessage(12345)
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Errorf("DoFilter should not return error for non-string, got: %s", err)
	}
	if ok {
		t.Errorf("DoFilter should return false for non-string message")
	}
}

func TestFileOnEvent(t *testing.T) {
	filter, _ := NewFileFilter(map[string]string{})
	f := filter.(*File)
	f.OnEvent(&data.Event{})
}

func TestFileFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["filefilter"]; !ok {
		t.Errorf("file filter should be registered as 'filefilter'")
	}
}

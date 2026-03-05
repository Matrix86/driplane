package filters

import (
	"os"
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewMimetypeFilter(t *testing.T) {
	filter, err := NewMimetypeFilter(map[string]string{
		"filename": "/tmp/{{.main}}",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*Mimetype)
	if !ok {
		t.Fatal("cannot cast to *Mimetype")
	}
	if f.filename == nil {
		t.Errorf("expected filename template to be set")
	}
}

func TestNewMimetypeFilterWithTarget(t *testing.T) {
	filter, err := NewMimetypeFilter(map[string]string{
		"target": "body",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Mimetype)
	if f.target != "body" {
		t.Errorf("expected target 'body', got '%s'", f.target)
	}
	if f.filename != nil {
		t.Errorf("expected filename to be nil when target is set")
	}
}

func TestNewMimetypeFilterDefaults(t *testing.T) {
	filter, err := NewMimetypeFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Mimetype)
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
}

func TestNewMimetypeFilterInvalidTemplate(t *testing.T) {
	_, err := NewMimetypeFilter(map[string]string{
		"filename": "{{.invalid",
	})
	if err == nil {
		t.Errorf("expected error for invalid filename template")
	}
}

func TestMimetypeDoFilterWithFilename(t *testing.T) {
	// Create a temp text file
	tmpFile, err := os.CreateTemp("", "mimetype_test_*.txt")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpFile.WriteString("hello world text content")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	filter, _ := NewMimetypeFilter(map[string]string{
		"filename": tmpFile.Name(),
	})
	f := filter.(*Mimetype)

	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
	// Should detect a text mime type
	result := msg.GetMessage().(string)
	if len(result) == 0 {
		t.Errorf("expected non-empty mime type")
	}
	extra := msg.GetExtra()
	if _, ok := extra["mimetype_ext"]; !ok {
		t.Errorf("expected 'mimetype_ext' extra field")
	}
	if _, ok := extra["fulltext"]; !ok {
		t.Errorf("expected 'fulltext' extra field")
	}
}

func TestMimetypeDoFilterWithFilenameNonexistent(t *testing.T) {
	filter, _ := NewMimetypeFilter(map[string]string{
		"filename": "/tmp/nonexistent_file_mimetype_99999.bin",
	})
	f := filter.(*Mimetype)

	msg := data.NewMessage("test")
	_, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for nonexistent file")
	}
}

func TestMimetypeDoFilterWithBytes(t *testing.T) {
	filter, _ := NewMimetypeFilter(map[string]string{})
	f := filter.(*Mimetype)

	// PNG header bytes
	pngHeader := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A}
	msg := data.NewMessage(pngHeader)
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true for valid bytes")
	}
	result := msg.GetMessage().(string)
	if len(result) == 0 {
		t.Errorf("expected non-empty mime type result")
	}
}

func TestMimetypeDoFilterUnsupportedType(t *testing.T) {
	filter, _ := NewMimetypeFilter(map[string]string{})
	f := filter.(*Mimetype)

	msg := data.NewMessage("not bytes, just a string")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for unsupported data type (string without filename)")
	}
	if ok {
		t.Errorf("DoFilter should return false for unsupported type")
	}
}

func TestMimetypeOnEvent(t *testing.T) {
	filter, _ := NewMimetypeFilter(map[string]string{})
	f := filter.(*Mimetype)
	f.OnEvent(&data.Event{})
}

func TestMimetypeFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["mimefilter"]; !ok {
		t.Errorf("mime filter should be registered as 'mimefilter'")
	}
}

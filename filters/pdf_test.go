package filters

import (
	"os"
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewPDFFilter(t *testing.T) {
	filter, err := NewPDFFilter(map[string]string{
		"filename": "/tmp/{{.main}}",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*PDF)
	if !ok {
		t.Fatal("cannot cast to *PDF")
	}
	if f.filename == nil {
		t.Errorf("expected filename template to be set")
	}
}

func TestNewPDFFilterWithTarget(t *testing.T) {
	filter, err := NewPDFFilter(map[string]string{
		"target": "body",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*PDF)
	if f.target != "body" {
		t.Errorf("expected target 'body', got '%s'", f.target)
	}
	if f.filename != nil {
		t.Errorf("expected filename to be nil when target is set")
	}
}

func TestNewPDFFilterDefaults(t *testing.T) {
	filter, err := NewPDFFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*PDF)
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
}

func TestNewPDFFilterInvalidTemplate(t *testing.T) {
	_, err := NewPDFFilter(map[string]string{
		"filename": "{{.invalid",
	})
	if err == nil {
		t.Errorf("expected error for invalid filename template")
	}
}

func TestPDFDoFilterNonexistentFile(t *testing.T) {
	filter, _ := NewPDFFilter(map[string]string{
		"filename": "/tmp/nonexistent_pdf_99999.pdf",
	})
	f := filter.(*PDF)

	msg := data.NewMessage("test")
	_, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for nonexistent PDF file")
	}
}

func TestPDFDoFilterWithBytesUnsupported(t *testing.T) {
	filter, _ := NewPDFFilter(map[string]string{})
	f := filter.(*PDF)

	// non-[]byte message should fail
	msg := data.NewMessage("not bytes")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for non-[]byte data")
	}
	if ok {
		t.Errorf("DoFilter should return false for unsupported type")
	}
}

func TestPDFDoFilterWithInvalidPDFBytes(t *testing.T) {
	filter, _ := NewPDFFilter(map[string]string{})
	f := filter.(*PDF)

	// Invalid data that is []byte but not a valid PDF
	msg := data.NewMessage([]byte("not a valid pdf content"))
	_, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for invalid PDF bytes")
	}
}

func TestPDFDoFilterWithFilenameInvalidPDF(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "pdf_test_*.pdf")
	if err != nil {
		t.Fatalf("cannot create temp file: %s", err)
	}
	tmpFile.WriteString("not a valid pdf")
	tmpFile.Close()
	defer os.Remove(tmpFile.Name())

	filter, _ := NewPDFFilter(map[string]string{
		"filename": tmpFile.Name(),
	})
	f := filter.(*PDF)

	msg := data.NewMessage("test")
	_, err = f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for invalid PDF file")
	}
}

func TestPDFOnEvent(t *testing.T) {
	filter, _ := NewPDFFilter(map[string]string{})
	f := filter.(*PDF)
	f.OnEvent(&data.Event{})
}

func TestPDFFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["pdffilter"]; !ok {
		t.Errorf("pdf filter should be registered as 'pdffilter'")
	}
}

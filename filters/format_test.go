package filters

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewFormatFilter(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{"none": "none", "template": "test {{.main}}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Format); ok {
		if e.template == nil {
			t.Errorf("'template' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestNewFormatFilterDefaults(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
	if f.templateType != "text" {
		t.Errorf("default templateType should be 'text', got '%s'", f.templateType)
	}
}

func TestNewFormatFilterHTMLType(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{
		"type":     "html",
		"template": "<b>{{.main}}</b>",
	})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)
	if f.templateType != "html" {
		t.Errorf("expected templateType 'html', got '%s'", f.templateType)
	}
	if f.template == nil {
		t.Error("template should be set")
	}
}

func TestNewFormatFilterInvalidTextTemplate(t *testing.T) {
	_, err := NewFormatFilter(map[string]string{
		"template": "{{.invalid",
	})
	if err == nil {
		t.Error("expected error for invalid text template")
	}
}

func TestNewFormatFilterInvalidHTMLTemplate(t *testing.T) {
	_, err := NewFormatFilter(map[string]string{
		"type":     "html",
		"template": "{{.invalid",
	})
	if err == nil {
		t.Error("expected error for invalid html template")
	}
}

func TestNewFormatFilterFileWithTemplatesPath(t *testing.T) {
	dir := t.TempDir()
	content := "Hello {{.main}}!"
	err := os.WriteFile(filepath.Join(dir, "test.fmt"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create template file: %s", err)
	}

	filter, err := NewFormatFilter(map[string]string{
		"file":                   "test.fmt",
		"general.templates_path": dir,
	})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)
	if f.template == nil {
		t.Error("template should be set from file")
	}
}

func TestNewFormatFilterFileWithRulesPath(t *testing.T) {
	dir := t.TempDir()
	content := "Rule {{.main}}"
	err := os.WriteFile(filepath.Join(dir, "test.fmt"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create template file: %s", err)
	}

	filter, err := NewFormatFilter(map[string]string{
		"file":               "test.fmt",
		"general.rules_path": dir,
	})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)
	if f.template == nil {
		t.Error("template should be set from file")
	}
}

func TestNewFormatFilterFileNoPathConfig(t *testing.T) {
	_, err := NewFormatFilter(map[string]string{
		"file": "test.fmt",
	})
	if err == nil {
		t.Error("expected error when no templates_path or rules_path is configured")
	}
}

func TestNewFormatFilterFileNotFound(t *testing.T) {
	dir := t.TempDir()
	_, err := NewFormatFilter(map[string]string{
		"file":                   "nonexistent.fmt",
		"general.templates_path": dir,
	})
	if err == nil {
		t.Error("expected error for nonexistent template file")
	}
}

func TestNewFormatFilterFileHTMLType(t *testing.T) {
	dir := t.TempDir()
	content := "<b>{{.main}}</b>"
	err := os.WriteFile(filepath.Join(dir, "test.html"), []byte(content), 0644)
	if err != nil {
		t.Fatalf("failed to create template file: %s", err)
	}

	filter, err := NewFormatFilter(map[string]string{
		"type":                   "html",
		"file":                   "test.html",
		"general.templates_path": dir,
	})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)
	if f.templateType != "html" {
		t.Errorf("expected templateType 'html', got '%s'", f.templateType)
	}
	if f.template == nil {
		t.Error("template should be set from file")
	}
}

func TestNewFormatFilterFileInvalidTextTemplate(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad.fmt"), []byte("{{.invalid"), 0644)
	if err != nil {
		t.Fatalf("failed to create template file: %s", err)
	}

	_, err = NewFormatFilter(map[string]string{
		"file":                   "bad.fmt",
		"general.templates_path": dir,
	})
	if err == nil {
		t.Error("expected error for invalid text template from file")
	}
}

func TestNewFormatFilterFileInvalidHTMLTemplate(t *testing.T) {
	dir := t.TempDir()
	err := os.WriteFile(filepath.Join(dir, "bad.html"), []byte("{{.invalid"), 0644)
	if err != nil {
		t.Fatalf("failed to create template file: %s", err)
	}

	_, err = NewFormatFilter(map[string]string{
		"type":                   "html",
		"file":                   "bad.html",
		"general.templates_path": dir,
	})
	if err == nil {
		t.Error("expected error for invalid html template from file")
	}
}

func TestFormatDoFilter(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{"template": "main : {{.main}} extra : {{.test}}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Format); ok {
		m := data.NewMessageWithExtra("message", map[string]interface{}{"test": "1"})
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}

		if m.GetMessage() != "main : message extra : 1" {
			t.Errorf("message not formatted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}

	filter, err = NewFormatFilter(map[string]string{"target": "other", "template": "main : {{.main}} extra : {{.test}}"})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if e, ok := filter.(*Format); ok {
		m := data.NewMessageWithExtra("message", map[string]interface{}{"test": "1"})
		b, err := e.DoFilter(m)
		if err != nil {
			t.Errorf("DoFilter returned an error '%s'", err)
		}
		if b == false {
			t.Errorf("it should return true")
		}

		if m.GetTarget("other") != "main : message extra : 1" {
			t.Errorf("message not formatted correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Filter...")
	}
}

func TestFormatDoFilterHTML(t *testing.T) {
	filter, err := NewFormatFilter(map[string]string{
		"type":     "html",
		"template": "<p>{{.main}}</p>",
	})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)

	m := data.NewMessage("<script>alert(1)</script>")
	ok, err := f.DoFilter(m)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Error("DoFilter should return true")
	}
	// html/template escapes dangerous content
	got := m.GetMessage()
	if got == "<p><script>alert(1)</script></p>" {
		t.Error("html template should escape script tags")
	}
}

func TestFormatDoFilterTemplateError(t *testing.T) {
	// Create a filter without setting a template so f.template is nil.
	// ApplyPlaceholder will return "template type not supported".
	filter, err := NewFormatFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned '%s'", err)
	}
	f := filter.(*Format)

	m := data.NewMessage("test")
	ok, err := f.DoFilter(m)
	if err == nil {
		t.Error("expected error when template is nil")
	}
	if ok {
		t.Error("DoFilter should return false on error")
	}
}

func TestFormatOnEvent(t *testing.T) {
	filter, _ := NewFormatFilter(map[string]string{})
	f := filter.(*Format)
	f.OnEvent(&data.Event{})
}

func TestFormatFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["formatfilter"]; !ok {
		t.Error("format filter should be registered as 'formatfilter'")
	}
}

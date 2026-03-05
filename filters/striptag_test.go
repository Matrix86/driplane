package filters

import (
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewStripTagFilter(t *testing.T) {
	filter, err := NewStripTagFilter(map[string]string{
		"target": "body",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*StripTag)
	if !ok {
		t.Fatal("cannot cast to *StripTag")
	}
	if f.target != "body" {
		t.Errorf("expected target 'body', got '%s'", f.target)
	}
}

func TestNewStripTagFilterDefaults(t *testing.T) {
	filter, err := NewStripTagFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*StripTag)
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
}

func TestStripTagDoFilterSimpleHTML(t *testing.T) {
	filter, _ := NewStripTagFilter(map[string]string{})
	f := filter.(*StripTag)

	msg := data.NewMessage("<html><body><p>Hello <b>World</b></p></body></html>")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}

	result := msg.GetMessage().(string)
	if len(result) == 0 {
		t.Errorf("expected non-empty stripped text")
	}
	// Should not contain HTML tags
	if containsHTMLTags(result) {
		t.Errorf("result should not contain HTML tags, got '%s'", result)
	}

	extra := msg.GetExtra()
	if _, ok := extra["fulltext"]; !ok {
		t.Errorf("expected 'fulltext' extra field with original HTML")
	}
}

func TestStripTagDoFilterPlainText(t *testing.T) {
	filter, _ := NewStripTagFilter(map[string]string{})
	f := filter.(*StripTag)

	msg := data.NewMessage("plain text without tags")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
}

func TestStripTagDoFilterBytesMessage(t *testing.T) {
	filter, _ := NewStripTagFilter(map[string]string{})
	f := filter.(*StripTag)

	msg := data.NewMessage([]byte("<p>bytes content</p>"))
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true for []byte input")
	}
}

func TestStripTagDoFilterNonStringMessage(t *testing.T) {
	filter, _ := NewStripTagFilter(map[string]string{})
	f := filter.(*StripTag)

	msg := data.NewMessage(12345)
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for non-string message")
	}
	if ok {
		t.Errorf("DoFilter should return false for non-string message")
	}
}

func TestStripTagDoFilterCustomTarget(t *testing.T) {
	filter, _ := NewStripTagFilter(map[string]string{"target": "html_body"})
	f := filter.(*StripTag)

	msg := data.NewMessage("ignored")
	msg.SetExtra("html_body", "<div>custom target</div>")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
}

func TestStripTagOnEvent(t *testing.T) {
	filter, _ := NewStripTagFilter(map[string]string{})
	f := filter.(*StripTag)
	f.OnEvent(&data.Event{})
}

func TestStripTagFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["striptagfilter"]; !ok {
		t.Errorf("striptag filter should be registered as 'striptagfilter'")
	}
}

// helper
func containsHTMLTags(s string) bool {
	inTag := false
	for _, c := range s {
		if c == '<' {
			inTag = true
		} else if c == '>' && inTag {
			return true
		}
	}
	return false
}

package data

import (
	html "html/template"
	"sync"
	"testing"
	text "text/template"
)

func TestNewMessage(t *testing.T) {
	msg := NewMessage("hello")
	if msg == nil {
		t.Fatal("NewMessage returned nil")
	}
	if msg.GetMessage() != "hello" {
		t.Errorf("expected 'hello', got '%v'", msg.GetMessage())
	}
}

func TestNewMessageWithExtra(t *testing.T) {
	msg := NewMessageWithExtra("hello", map[string]interface{}{"key": "value"})
	if msg == nil {
		t.Fatal("NewMessageWithExtra returned nil")
	}
	if msg.GetMessage() != "hello" {
		t.Errorf("expected main 'hello', got '%v'", msg.GetMessage())
	}
	extra := msg.GetExtra()
	if extra["key"] != "value" {
		t.Errorf("expected extra 'key' to be 'value', got '%v'", extra["key"])
	}
}

func TestSetMessage(t *testing.T) {
	msg := NewMessage("initial")
	msg.SetMessage("updated")
	if msg.GetMessage() != "updated" {
		t.Errorf("expected 'updated', got '%v'", msg.GetMessage())
	}
}

func TestSetExtra(t *testing.T) {
	msg := NewMessage("hello")
	msg.SetExtra("foo", "bar")
	extra := msg.GetExtra()
	if extra["foo"] != "bar" {
		t.Errorf("expected 'bar', got '%v'", extra["foo"])
	}
}

func TestSetExtraIgnoresMainKey(t *testing.T) {
	msg := NewMessage("hello")
	msg.SetExtra("main", "should be ignored")
	if msg.GetMessage() != "hello" {
		t.Errorf("SetExtra with 'main' key should be ignored, got '%v'", msg.GetMessage())
	}
}

func TestGetExtraExcludesMain(t *testing.T) {
	msg := NewMessageWithExtra("hello", map[string]interface{}{"key": "value"})
	extra := msg.GetExtra()
	if _, ok := extra["main"]; ok {
		t.Errorf("GetExtra should not return 'main' key")
	}
}

func TestGetExtraExcludesUnderscoreKeys(t *testing.T) {
	msg := NewMessage("hello")
	msg.SetTarget("_hidden", "secret")
	extra := msg.GetExtra()
	if _, ok := extra["_hidden"]; ok {
		t.Errorf("GetExtra should not return keys starting with '_'")
	}
}

func TestSetTarget(t *testing.T) {
	msg := NewMessage("hello")
	msg.SetTarget("foo", "bar")
	if msg.GetTarget("foo") != "bar" {
		t.Errorf("expected 'bar', got '%v'", msg.GetTarget("foo"))
	}
}

func TestSetTargetCanOverwriteMain(t *testing.T) {
	msg := NewMessage("hello")
	msg.SetTarget("main", "overwritten")
	if msg.GetMessage() != "overwritten" {
		t.Errorf("SetTarget should be able to overwrite 'main', got '%v'", msg.GetMessage())
	}
}

func TestGetTargetReturnsNilIfMissing(t *testing.T) {
	msg := NewMessage("hello")
	if msg.GetTarget("nonexistent") != nil {
		t.Errorf("expected nil for missing key, got '%v'", msg.GetTarget("nonexistent"))
	}
}

func TestSetFirstRun(t *testing.T) {
	msg := NewMessage("hello")
	if msg.IsFirstRun() {
		t.Errorf("firstRun should be false by default")
	}
	msg.SetFirstRun()
	if !msg.IsFirstRun() {
		t.Errorf("firstRun should be true after SetFirstRun()")
	}
}

func TestClearFirstRun(t *testing.T) {
	msg := NewMessage("hello")
	msg.SetFirstRun()
	msg.ClearFirstRun()
	if msg.IsFirstRun() {
		t.Errorf("firstRun should be false after ClearFirstRun()")
	}
}

func TestClone(t *testing.T) {
	original := NewMessageWithExtra("hello", map[string]interface{}{"key": "value"})
	original.SetFirstRun()

	clone := original.Clone()

	if clone.GetMessage() != original.GetMessage() {
		t.Errorf("cloned main mismatch: expected '%v', got '%v'", original.GetMessage(), clone.GetMessage())
	}
	if clone.IsFirstRun() != original.IsFirstRun() {
		t.Errorf("cloned firstRun mismatch")
	}
	extra := clone.GetExtra()
	if extra["key"] != "value" {
		t.Errorf("cloned extra mismatch: expected 'value', got '%v'", extra["key"])
	}
}

func TestCloneIsDeepCopy(t *testing.T) {
	original := NewMessageWithExtra("hello", map[string]interface{}{"key": "value"})
	clone := original.Clone()

	// mutating the clone should not affect the original
	clone.SetMessage("mutated")
	clone.SetExtra("key", "changed")

	if original.GetMessage() != "hello" {
		t.Errorf("original main should not be affected by clone mutation, got '%v'", original.GetMessage())
	}
	if original.GetTarget("key") != "value" {
		t.Errorf("original extra should not be affected by clone mutation, got '%v'", original.GetTarget("key"))
	}
}

func TestApplyPlaceholderTextTemplate(t *testing.T) {
	msg := NewMessageWithExtra("hello", map[string]interface{}{"name": "world"})
	tmpl, err := text.New("test").Parse("Hello {{.name}}!")
	if err != nil {
		t.Fatalf("failed to parse template: %s", err)
	}
	result, err := msg.ApplyPlaceholder(tmpl)
	if err != nil {
		t.Errorf("ApplyPlaceholder returned error: %s", err)
	}
	if result != "Hello world!" {
		t.Errorf("expected 'Hello world!', got '%s'", result)
	}
}

func TestApplyPlaceholderHTMLTemplate(t *testing.T) {
	msg := NewMessageWithExtra("hello", map[string]interface{}{"name": "world"})
	tmpl, err := html.New("test").Parse("<b>Hello {{.name}}!</b>")
	if err != nil {
		t.Fatalf("failed to parse template: %s", err)
	}
	result, err := msg.ApplyPlaceholder(tmpl)
	if err != nil {
		t.Errorf("ApplyPlaceholder returned error: %s", err)
	}
	if result != "<b>Hello world!</b>" {
		t.Errorf("expected '<b>Hello world!</b>', got '%s'", result)
	}
}

func TestApplyPlaceholderMainField(t *testing.T) {
	msg := NewMessage("hello world")
	tmpl, err := text.New("test").Parse("main is: {{.main}}")
	if err != nil {
		t.Fatalf("failed to parse template: %s", err)
	}
	result, err := msg.ApplyPlaceholder(tmpl)
	if err != nil {
		t.Errorf("ApplyPlaceholder returned error: %s", err)
	}
	if result != "main is: hello world" {
		t.Errorf("expected 'main is: hello world', got '%s'", result)
	}
}

func TestApplyPlaceholderUnsupportedType(t *testing.T) {
	msg := NewMessage("hello")
	_, err := msg.ApplyPlaceholder("not a template")
	if err == nil {
		t.Errorf("expected error for unsupported template type")
	}
}

func TestApplyPlaceholderTextTemplateError(t *testing.T) {
	msg := NewMessage("hello")
	// template references a missing field with a strict option (option "missingkey=error")
	tmpl, err := text.New("test").Option("missingkey=error").Parse("{{.nonexistent}}")
	if err != nil {
		t.Fatalf("failed to parse template: %s", err)
	}
	_, err = msg.ApplyPlaceholder(tmpl)
	if err == nil {
		t.Errorf("expected error for missing key with missingkey=error option")
	}
}

func TestMessageConcurrency(t *testing.T) {
	msg := NewMessage("initial")
	var wg sync.WaitGroup

	// concurrent writers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			msg.SetMessage(i)
			msg.SetExtra("key", i)
		}(i)
	}

	// concurrent readers
	for i := 0; i < 50; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			_ = msg.GetMessage()
			_ = msg.GetExtra()
			_ = msg.IsFirstRun()
		}()
	}

	wg.Wait()
}

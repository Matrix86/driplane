package filters

import (
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewMailFilter(t *testing.T) {
	filter, err := NewMailFilter(map[string]string{
		"body":     "<b>{{.main}}</b>",
		"username": "user@example.com",
		"password": "secret",
		"host":     "smtp.example.com",
		"port":     "587",
		"fromAddr": "sender@example.com",
		"fromName": "Sender",
		"to":       "a@b.com,c@d.com",
		"subject":  "Test Subject",
		"use_auth": "true",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*Mail)
	if !ok {
		t.Fatal("cannot cast to *Mail")
	}
	if f.username != "user@example.com" {
		t.Errorf("expected username 'user@example.com', got '%s'", f.username)
	}
	if f.password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", f.password)
	}
	if f.server != "smtp.example.com" {
		t.Errorf("expected server 'smtp.example.com', got '%s'", f.server)
	}
	if f.port != 587 {
		t.Errorf("expected port 587, got %d", f.port)
	}
	if f.fromAddr != "sender@example.com" {
		t.Errorf("expected fromAddr 'sender@example.com', got '%s'", f.fromAddr)
	}
	if f.fromName != "Sender" {
		t.Errorf("expected fromName 'Sender', got '%s'", f.fromName)
	}
	if len(f.to) != 2 {
		t.Errorf("expected 2 recipients, got %d", len(f.to))
	}
	if f.subject != "Test Subject" {
		t.Errorf("expected subject 'Test Subject', got '%s'", f.subject)
	}
	if !f.useAuth {
		t.Errorf("expected useAuth to be true")
	}
	if f.template == nil {
		t.Errorf("expected template to be set")
	}
}

func TestNewMailFilterDefaults(t *testing.T) {
	filter, err := NewMailFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Mail)
	if f.useAuth {
		t.Errorf("default useAuth should be false")
	}
	if f.template != nil {
		t.Errorf("default template should be nil")
	}
}

func TestNewMailFilterInvalidPort(t *testing.T) {
	_, err := NewMailFilter(map[string]string{
		"port": "notanumber",
	})
	if err == nil {
		t.Errorf("expected error for invalid port")
	}
}

func TestNewMailFilterInvalidBodyTemplate(t *testing.T) {
	_, err := NewMailFilter(map[string]string{
		"body": "{{.invalid",
	})
	if err == nil {
		t.Errorf("expected error for invalid body template")
	}
}

func TestMailDoFilterNonStringMessage(t *testing.T) {
	filter, _ := NewMailFilter(map[string]string{
		"host": "localhost",
		"port": "25",
	})
	f := filter.(*Mail)

	msg := data.NewMessage(12345)
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for non-string message")
	}
	if ok {
		t.Errorf("DoFilter should return false for non-string message")
	}
}

func TestMailDoFilterConnectionFailure(t *testing.T) {
	filter, _ := NewMailFilter(map[string]string{
		"host": "127.0.0.1",
		"port": "1",
		"to":   "test@example.com",
	})
	f := filter.(*Mail)

	msg := data.NewMessage("test message")
	_, err := f.DoFilter(msg)
	// Should fail to connect to SMTP on port 1
	if err == nil {
		t.Errorf("expected error when SMTP is not reachable")
	}
}

func TestMailOnEvent(t *testing.T) {
	filter, _ := NewMailFilter(map[string]string{})
	f := filter.(*Mail)
	f.OnEvent(&data.Event{})
}

func TestMailFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["mailfilter"]; !ok {
		t.Errorf("mail filter should be registered as 'mailfilter'")
	}
}

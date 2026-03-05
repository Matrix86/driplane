package filters

import (
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewTelegramFilter(t *testing.T) {
	filter, err := NewTelegramFilter(map[string]string{
		"action":   "send_message",
		"to":       "user123",
		"text":     "Hello {{.main}}",
		"filename": "/tmp/{{.main}}",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*Telegram)
	if !ok {
		t.Fatal("cannot cast to *Telegram")
	}
	if f.action != "send_message" {
		t.Errorf("expected action 'send_message', got '%s'", f.action)
	}
	if f.to == nil {
		t.Errorf("expected 'to' template to be set")
	}
	if f.message == nil {
		t.Errorf("expected 'message' template to be set")
	}
	if f.downloadPath == nil {
		t.Errorf("expected 'downloadPath' template to be set")
	}
}

func TestNewTelegramFilterDefaultAction(t *testing.T) {
	filter, err := NewTelegramFilter(map[string]string{
		"to":   "user123",
		"text": "hello",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Telegram)
	if f.action != "send_message" {
		t.Errorf("default action should be 'send_message', got '%s'", f.action)
	}
}

func TestNewTelegramFilterWithChatID(t *testing.T) {
	filter, err := NewTelegramFilter(map[string]string{
		"action":    "send_message",
		"to_chatid": "12345",
		"text":      "hello",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Telegram)
	if f.toChat == nil {
		t.Errorf("expected 'toChat' template to be set")
	}
}

func TestNewTelegramFilterDownloadFile(t *testing.T) {
	filter, err := NewTelegramFilter(map[string]string{
		"action":   "download_file",
		"filename": "/tmp/downloads/{{.main}}",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*Telegram)
	if f.action != "download_file" {
		t.Errorf("expected action 'download_file', got '%s'", f.action)
	}
}

func TestNewTelegramFilterDownloadFileMissingFilename(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action": "download_file",
	})
	if err == nil {
		t.Errorf("expected error when download_file action is missing filename")
	}
}

func TestNewTelegramFilterSendMessageMissingText(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action": "send_message",
		"to":     "user",
	})
	if err == nil {
		t.Errorf("expected error when send_message action is missing text")
	}
}

func TestNewTelegramFilterSendMessageMissingRecipient(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action": "send_message",
		"text":   "hello",
	})
	if err == nil {
		t.Errorf("expected error when send_message has no to or to_chatid")
	}
}

func TestNewTelegramFilterInvalidAction(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action": "invalid_action",
	})
	if err == nil {
		t.Errorf("expected error for invalid action")
	}
}

func TestNewTelegramFilterInvalidToTemplate(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action": "send_message",
		"to":     "{{.invalid",
		"text":   "hello",
	})
	if err == nil {
		t.Errorf("expected error for invalid 'to' template")
	}
}

func TestNewTelegramFilterInvalidTextTemplate(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action": "send_message",
		"to":     "user",
		"text":   "{{.invalid",
	})
	if err == nil {
		t.Errorf("expected error for invalid 'text' template")
	}
}

func TestNewTelegramFilterInvalidChatIDTemplate(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action":    "send_message",
		"to_chatid": "{{.invalid",
		"text":      "hello",
	})
	if err == nil {
		t.Errorf("expected error for invalid 'to_chatid' template")
	}
}

func TestNewTelegramFilterInvalidFilenameTemplate(t *testing.T) {
	_, err := NewTelegramFilter(map[string]string{
		"action":   "download_file",
		"filename": "{{.invalid",
	})
	if err == nil {
		t.Errorf("expected error for invalid 'filename' template")
	}
}

func TestTelegramDoFilterNoTelegramAPI(t *testing.T) {
	filter, _ := NewTelegramFilter(map[string]string{
		"action": "send_message",
		"to":     "user",
		"text":   "hello",
	})
	f := filter.(*Telegram)

	// Message without _telegram_api target → should return false
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Errorf("expected no error when telegram API is missing, got: %s", err)
	}
	if ok {
		t.Errorf("DoFilter should return false without telegram API")
	}
}

func TestTelegramOnEvent(t *testing.T) {
	filter, _ := NewTelegramFilter(map[string]string{
		"action": "send_message",
		"to":     "user",
		"text":   "hello",
	})
	f := filter.(*Telegram)
	f.OnEvent(&data.Event{})
}

func TestTelegramFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["telegramfilter"]; !ok {
		t.Errorf("telegram filter should be registered as 'telegramfilter'")
	}
}

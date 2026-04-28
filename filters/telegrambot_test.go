package filters

import (
	"bytes"
	"encoding/json"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"github.com/Matrix86/driplane/data"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type mockHandlerFn func(t *testing.T, w http.ResponseWriter, r *http.Request)

type recordedRequest struct {
	method  string
	body    []byte
	headers http.Header
}

type mockBotAPI struct {
	t           *testing.T
	handlers    map[string]mockHandlerFn
	rawHandlers map[string]mockHandlerFn
	files       map[string][]byte
	requests    []recordedRequest
}

func newMockBotAPI(t *testing.T) *mockBotAPI {
	return &mockBotAPI{t: t, handlers: map[string]mockHandlerFn{}, rawHandlers: map[string]mockHandlerFn{}, files: map[string][]byte{}}
}

func (m *mockBotAPI) handle(method string, fn mockHandlerFn) { m.handlers[method] = fn }

// handleRaw registers a handler that receives the ORIGINAL request body (no
// multipart-to-JSON conversion). Use it when a test needs to inspect
// multipart/form-data parts directly (e.g. local file uploads).
func (m *mockBotAPI) handleRaw(method string, fn mockHandlerFn) { m.rawHandlers[method] = fn }

func (m *mockBotAPI) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/file/") {
		rest := strings.TrimPrefix(r.URL.Path, "/file/")
		idx := strings.Index(rest, "/")
		if idx < 0 {
			http.NotFound(w, r)
			return
		}
		relPath := rest[idx+1:]
		if data, ok := m.files[relPath]; ok {
			w.Write(data)
			return
		}
		http.NotFound(w, r)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/")
	idx := strings.Index(rest, "/")
	if idx < 0 {
		http.NotFound(w, r)
		return
	}
	methodName := rest[idx+1:]
	bodyBytes, _ := io.ReadAll(r.Body)
	r.Body.Close()
	// Raw handlers see the original body untouched (multipart preserved).
	if h, ok := m.rawHandlers[methodName]; ok {
		r.Body = io.NopCloser(bytes.NewReader(bodyBytes))
		m.requests = append(m.requests, recordedRequest{method: methodName, body: bodyBytes, headers: r.Header.Clone()})
		h(m.t, w, r)
		return
	}
	// Convert multipart body to a JSON object so test handlers can decode it
	// uniformly. The go-telegram/bot client posts as multipart/form-data with
	// each field already JSON-encoded (strings have outer quotes stripped).
	jsonBody := bodyBytes
	ct := r.Header.Get("Content-Type")
	if mt, params, err := mime.ParseMediaType(ct); err == nil && strings.HasPrefix(mt, "multipart/") {
		if jb, err := multipartToJSON(bodyBytes, params["boundary"]); err == nil {
			jsonBody = jb
		}
	}
	r.Body = io.NopCloser(bytes.NewReader(jsonBody))
	m.requests = append(m.requests, recordedRequest{method: methodName, body: jsonBody, headers: r.Header.Clone()})
	if h, ok := m.handlers[methodName]; ok {
		h(m.t, w, r)
		return
	}
	http.Error(w, "no handler for "+methodName, http.StatusNotFound)
}

// multipartToJSON parses multipart/form-data into a JSON object.
// Each field is decoded heuristically: numbers as numbers, valid JSON
// values (objects/arrays/booleans) as their parsed forms, everything else as
// a string. This matches how the go-telegram/bot library serialises params.
func multipartToJSON(body []byte, boundary string) ([]byte, error) {
	mr := multipart.NewReader(bytes.NewReader(body), boundary)
	out := map[string]interface{}{}
	for {
		part, err := mr.NextPart()
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, err
		}
		name := part.FormName()
		raw, _ := io.ReadAll(part)
		s := string(raw)
		// Try integer.
		if n, err := strconv.ParseInt(s, 10, 64); err == nil {
			out[name] = float64(n)
			continue
		}
		// Try float.
		if f, err := strconv.ParseFloat(s, 64); err == nil {
			out[name] = f
			continue
		}
		// Try parsing as JSON value (object/array/bool/null).
		trimmed := strings.TrimSpace(s)
		if strings.HasPrefix(trimmed, "{") || strings.HasPrefix(trimmed, "[") || trimmed == "true" || trimmed == "false" || trimmed == "null" {
			var v interface{}
			if err := json.Unmarshal([]byte(trimmed), &v); err == nil {
				out[name] = v
				continue
			}
		}
		out[name] = s
	}
	return json.Marshal(out)
}

func newMockedBot(t *testing.T, mock *mockBotAPI) (*bot.Bot, *httptest.Server) {
	srv := httptest.NewServer(mock)
	t.Cleanup(srv.Close)
	b, err := bot.New("test:token", bot.WithServerURL(srv.URL), bot.WithSkipGetMe())
	if err != nil {
		t.Fatalf("bot.New: %v", err)
	}
	return b, srv
}

func newTestMessage(b *bot.Bot, main string, extras map[string]interface{}) *data.Message {
	if extras == nil {
		extras = map[string]interface{}{}
	}
	msg := data.NewMessageWithExtra(main, extras)
	msg.SetTarget("_telegrambot_api", b)
	return msg
}

func newTestFilter(t *testing.T, params map[string]string) *TelegramBot {
	f, err := NewTelegramBotFilter(params)
	if err != nil {
		t.Fatalf("NewTelegramBotFilter: %v", err)
	}
	tg, ok := f.(*TelegramBot)
	if !ok {
		t.Fatalf("filter is not *TelegramBot: %T", f)
	}
	return tg
}

func writeJSON(t *testing.T, w http.ResponseWriter, payload string) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := io.WriteString(w, payload); err != nil {
		t.Fatalf("write response: %v", err)
	}
}

func TestTelegramBotFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["telegrambotfilter"]; !ok {
		t.Fatalf("telegrambot filter not registered (expected key 'telegrambotfilter' in filterFactories)")
	}
}

func TestTelegramBotFilterDefaultActionIsSendMessage(t *testing.T) {
	f, err := NewTelegramBotFilter(map[string]string{"text": "hi", "to_chat": "1"})
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	tg, ok := f.(*TelegramBot)
	if !ok {
		t.Fatalf("expected *TelegramBot, got %T", f)
	}
	if tg.action != "send_message" {
		t.Fatalf("default action: want send_message, got %q", tg.action)
	}
}

func TestTelegramBotFilterRejectsUnknownAction(t *testing.T) {
	_, err := NewTelegramBotFilter(map[string]string{"action": "no_such_action"})
	if err == nil {
		t.Fatalf("expected error for unknown action, got nil")
	}
}

func TestTelegramBotFilterRequiredParamsPerAction(t *testing.T) {
	cases := []struct {
		name    string
		params  map[string]string
		wantErr bool
	}{
		{"send_message_missing_text", map[string]string{"action": "send_message"}, true},
		{"send_message_ok", map[string]string{"action": "send_message", "text": "hi"}, false},

		{"edit_message_missing_text", map[string]string{"action": "edit_message", "message_id": "1"}, true},
		{"edit_message_missing_id", map[string]string{"action": "edit_message", "text": "hi"}, true},
		{"edit_message_ok", map[string]string{"action": "edit_message", "text": "hi", "message_id": "1"}, false},

		{"delete_message_missing_id", map[string]string{"action": "delete_message"}, true},
		{"delete_message_ok", map[string]string{"action": "delete_message", "message_id": "1"}, false},

		{"answer_callback_missing_text", map[string]string{"action": "answer_callback"}, true},
		{"answer_callback_ok", map[string]string{"action": "answer_callback", "text": "hi"}, false},

		{"send_media_missing_media", map[string]string{"action": "send_media"}, true},
		{"send_media_ok", map[string]string{"action": "send_media", "media": "/tmp/x.png"}, false},

		{"download_file_missing_filename", map[string]string{"action": "download_file"}, true},
		{"download_file_ok", map[string]string{"action": "download_file", "filename": "/tmp/x.bin"}, false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			_, err := NewTelegramBotFilter(c.params)
			if c.wantErr && err == nil {
				t.Fatalf("expected error, got nil")
			}
			if !c.wantErr && err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
		})
	}
}

func TestTelegramBotFilterInvalidTemplate(t *testing.T) {
	_, err := NewTelegramBotFilter(map[string]string{
		"action": "send_message",
		"text":   "{{ .broken",
	})
	if err == nil {
		t.Fatalf("expected error for malformed template, got nil")
	}
}

func TestTelegramBotResolveBotMissing(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "1"})
	msg := data.NewMessageWithExtra("x", map[string]interface{}{})
	b, ok := f.resolveBot(msg)
	if ok {
		t.Fatalf("expected ok=false when _telegrambot_api missing, got bot=%v", b)
	}
}

func TestTelegramBotResolveBotWrongType(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "1"})
	msg := data.NewMessageWithExtra("x", map[string]interface{}{})
	msg.SetTarget("_telegrambot_api", "not a bot")
	_, ok := f.resolveBot(msg)
	if ok {
		t.Fatalf("expected ok=false for wrong-type bot extra")
	}
}

func TestTelegramBotResolveBotOK(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "1"})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", nil)
	got, ok := f.resolveBot(msg)
	if !ok || got != b {
		t.Fatalf("expected resolved bot, got ok=%v bot=%v", ok, got)
	}
}

func TestTelegramBotResolveChatID_Explicit(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "12345"})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", nil)
	chatID, ok := f.resolveChatID(msg)
	if !ok || chatID != int64(12345) {
		t.Fatalf("want 12345 ok=true, got %d ok=%v", chatID, ok)
	}
}

func TestTelegramBotResolveChatID_FromExtra(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi"})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", map[string]interface{}{"chat_id": "77"})
	chatID, ok := f.resolveChatID(msg)
	if !ok || chatID != 77 {
		t.Fatalf("want 77, got %d ok=%v", chatID, ok)
	}
}

func TestTelegramBotResolveChatID_Missing(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi"})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", nil)
	_, ok := f.resolveChatID(msg)
	if ok {
		t.Fatalf("expected ok=false when no chat_id available")
	}
}

func TestTelegramBotResolveParseMode(t *testing.T) {
	cases := map[string]models.ParseMode{
		"":           models.ParseMode(""),
		"html":       models.ParseModeHTML,
		"HTML":       models.ParseModeHTML,
		"markdown":   models.ParseModeMarkdownV1,
		"MarkdownV2": models.ParseModeMarkdown,
		"garbage":    models.ParseMode(""), // unknown → plain
	}
	for in, want := range cases {
		got := mapParseMode(in)
		if got != want {
			t.Fatalf("mapParseMode(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestTelegramBotParseReplyMarkup_Empty(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "1"})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", nil)
	mk, ok := f.resolveReplyMarkup(msg)
	if !ok {
		t.Fatalf("expected ok=true for unset reply_markup")
	}
	if mk != nil {
		t.Fatalf("expected nil markup when unset, got %T", mk)
	}
}

func TestTelegramBotParseReplyMarkup_JSON(t *testing.T) {
	f := newTestFilter(t, map[string]string{
		"action":       "send_message",
		"text":         "hi",
		"to_chat":      "1",
		"reply_markup": `{"inline_keyboard":[[{"text":"OK","callback_data":"ok"}]]}`,
	})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", nil)
	mk, ok := f.resolveReplyMarkup(msg)
	if !ok {
		t.Fatalf("expected ok=true for valid JSON")
	}
	kb, isKB := mk.(models.InlineKeyboardMarkup)
	if !isKB {
		t.Fatalf("expected InlineKeyboardMarkup, got %T", mk)
	}
	if len(kb.InlineKeyboard) != 1 || len(kb.InlineKeyboard[0]) != 1 {
		t.Fatalf("unexpected keyboard shape: %#v", kb)
	}
	if kb.InlineKeyboard[0][0].Text != "OK" {
		t.Fatalf("button text mismatch: %s", kb.InlineKeyboard[0][0].Text)
	}
}

func TestTelegramBotParseReplyMarkup_Invalid(t *testing.T) {
	f := newTestFilter(t, map[string]string{
		"action":       "send_message",
		"text":         "hi",
		"to_chat":      "1",
		"reply_markup": `{not json`,
	})
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	msg := newTestMessage(b, "x", nil)
	_, ok := f.resolveReplyMarkup(msg)
	if ok {
		t.Fatalf("expected ok=false for invalid JSON")
	}
}

func TestTelegramBotSendMessage_Success(t *testing.T) {
	mock := newMockBotAPI(t)
	mock.handle("sendMessage", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		// Verify body
		body, _ := io.ReadAll(r.Body)
		var got map[string]interface{}
		if err := json.Unmarshal(body, &got); err != nil {
			t.Fatalf("body json: %v", err)
		}
		if got["chat_id"].(float64) != 99 {
			t.Fatalf("chat_id: %v", got["chat_id"])
		}
		if got["text"] != "hello alice" {
			t.Fatalf("text: %v", got["text"])
		}
		if got["parse_mode"] != "HTML" {
			t.Fatalf("parse_mode: %v", got["parse_mode"])
		}
		writeJSON(t, w, `{"ok":true,"result":{"message_id":42,"chat":{"id":99,"type":"private"},"date":1234567890,"text":"hello alice"}}`)
	})
	b, _ := newMockedBot(t, mock)

	f := newTestFilter(t, map[string]string{
		"action":  "send_message",
		"text":    "hello {{ .user_username }}",
		"to_chat": "99",
	})
	msg := newTestMessage(b, "incoming", map[string]interface{}{"user_username": "alice"})
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if !ok {
		t.Fatalf("expected ok=true on success")
	}

	if msg.GetMessage() != "hello alice" {
		t.Fatalf("main: %s", msg.GetMessage())
	}
	if msg.GetTarget("sent_message_id") != "42" {
		t.Fatalf("sent_message_id: %v", msg.GetTarget("sent_message_id"))
	}
	if msg.GetTarget("sent_chat_id") != "99" {
		t.Fatalf("sent_chat_id: %v", msg.GetTarget("sent_chat_id"))
	}
	if msg.GetTarget("sent_text") != "hello alice" {
		t.Fatalf("sent_text: %v", msg.GetTarget("sent_text"))
	}
	if msg.GetTarget("sent_date") != "1234567890" {
		t.Fatalf("sent_date: %v", msg.GetTarget("sent_date"))
	}
	// incoming extras preserved
	if msg.GetTarget("user_username") != "alice" {
		t.Fatalf("incoming extras lost")
	}
}

func TestTelegramBotSendMessage_FallbackChatID(t *testing.T) {
	mock := newMockBotAPI(t)
	mock.handle("sendMessage", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]interface{}
		json.Unmarshal(body, &got)
		if got["chat_id"].(float64) != 7 {
			t.Fatalf("expected chat_id=7, got %v", got["chat_id"])
		}
		writeJSON(t, w, `{"ok":true,"result":{"message_id":1,"chat":{"id":7,"type":"private"},"date":1,"text":"hi"}}`)
	})
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi"})
	msg := newTestMessage(b, "x", map[string]interface{}{"chat_id": "7"})
	ok, _ := f.DoFilter(msg)
	if !ok {
		t.Fatalf("want ok=true")
	}
}

func TestTelegramBotSendMessage_NoChatID(t *testing.T) {
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi"})
	msg := newTestMessage(b, "x", nil)
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ok {
		t.Fatalf("want ok=false when no chat_id available")
	}
}

func TestTelegramBotSendMessage_APIError(t *testing.T) {
	mock := newMockBotAPI(t)
	mock.handle("sendMessage", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, `{"ok":false,"description":"chat not found","error_code":400}`)
	})
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "1"})
	msg := newTestMessage(b, "x", nil)
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ok {
		t.Fatalf("want ok=false on API error")
	}
}

func TestTelegramBotSendMessage_MissingBot(t *testing.T) {
	f := newTestFilter(t, map[string]string{"action": "send_message", "text": "hi", "to_chat": "1"})
	msg := data.NewMessageWithExtra("x", map[string]interface{}{}) // no _telegrambot_api
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("err: %v", err)
	}
	if ok {
		t.Fatalf("want ok=false")
	}
}

func TestTelegramBotSendMessage_WithReplyMarkup(t *testing.T) {
	mock := newMockBotAPI(t)
	mock.handle("sendMessage", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]interface{}
		json.Unmarshal(body, &got)
		rm, ok := got["reply_markup"].(map[string]interface{})
		if !ok {
			t.Fatalf("reply_markup missing or wrong shape: %v", got["reply_markup"])
		}
		kb, ok := rm["inline_keyboard"].([]interface{})
		if !ok || len(kb) != 1 {
			t.Fatalf("inline_keyboard malformed: %v", rm)
		}
		writeJSON(t, w, `{"ok":true,"result":{"message_id":1,"chat":{"id":1,"type":"private"},"date":1,"text":"hi"}}`)
	})
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{
		"action":       "send_message",
		"text":         "hi",
		"to_chat":      "1",
		"reply_markup": `{"inline_keyboard":[[{"text":"yes","callback_data":"y"}]]}`,
	})
	msg := newTestMessage(b, "x", nil)
	ok, _ := f.DoFilter(msg)
	if !ok {
		t.Fatalf("want ok=true")
	}
}

func TestTelegramBotEditMessage_Success(t *testing.T) {
	mock := newMockBotAPI(t)
	mock.handle("editMessageText", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var got map[string]interface{}
		json.Unmarshal(body, &got)
		if got["chat_id"].(float64) != 5 { t.Fatalf("chat_id: %v", got["chat_id"]) }
		if got["message_id"].(float64) != 42 { t.Fatalf("message_id: %v", got["message_id"]) }
		if got["text"] != "edited" { t.Fatalf("text: %v", got["text"]) }
		writeJSON(t, w, `{"ok":true,"result":{"message_id":42,"chat":{"id":5,"type":"private"},"date":1,"text":"edited"}}`)
	})
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{
		"action":     "edit_message",
		"text":       "edited",
		"to_chat":    "5",
		"message_id": "42",
	})
	msg := newTestMessage(b, "x", nil)
	ok, _ := f.DoFilter(msg)
	if !ok { t.Fatalf("want ok=true") }
	if msg.GetTarget("edit_success") != "true" { t.Fatalf("edit_success: %v", msg.GetTarget("edit_success")) }
	if msg.GetTarget("edited_message_id") != "42" { t.Fatalf("edited_message_id: %v", msg.GetTarget("edited_message_id")) }
	if msg.GetTarget("edited_chat_id") != "5" { t.Fatalf("edited_chat_id: %v", msg.GetTarget("edited_chat_id")) }
}

func TestTelegramBotEditMessage_BadMessageID(t *testing.T) {
	mock := newMockBotAPI(t)
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{
		"action":     "edit_message",
		"text":       "x",
		"to_chat":    "1",
		"message_id": "not-a-number",
	})
	msg := newTestMessage(b, "x", nil)
	ok, _ := f.DoFilter(msg)
	if ok { t.Fatalf("want ok=false on unparseable message_id") }
}

func TestTelegramBotEditMessage_APIError(t *testing.T) {
	mock := newMockBotAPI(t)
	mock.handle("editMessageText", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		writeJSON(t, w, `{"ok":false,"description":"message not found","error_code":400}`)
	})
	b, _ := newMockedBot(t, mock)
	f := newTestFilter(t, map[string]string{"action": "edit_message", "text": "x", "to_chat": "1", "message_id": "1"})
	msg := newTestMessage(b, "x", nil)
	ok, _ := f.DoFilter(msg)
	if ok { t.Fatalf("want ok=false on API error") }
}

func TestTelegramBotDeleteMessage_Success(t *testing.T) {
    mock := newMockBotAPI(t)
    mock.handle("deleteMessage", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["chat_id"].(float64) != 5 || got["message_id"].(float64) != 9 {
            t.Fatalf("unexpected body: %v", got)
        }
        writeJSON(t, w, `{"ok":true,"result":true}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{
        "action":     "delete_message",
        "to_chat":    "5",
        "message_id": "9",
    })
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
    if msg.GetTarget("delete_success") != "true" { t.Fatalf("delete_success: %v", msg.GetTarget("delete_success")) }
}

func TestTelegramBotDeleteMessage_APIError(t *testing.T) {
    mock := newMockBotAPI(t)
    mock.handle("deleteMessage", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        writeJSON(t, w, `{"ok":false,"description":"message not found","error_code":400}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "delete_message", "to_chat": "1", "message_id": "1"})
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if ok { t.Fatalf("want ok=false") }
}

func TestTelegramBotAnswerCallback_FromExtra(t *testing.T) {
    mock := newMockBotAPI(t)
    mock.handle("answerCallbackQuery", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["callback_query_id"] != "cbq-1" || got["text"] != "ack" {
            t.Fatalf("unexpected body: %v", got)
        }
        writeJSON(t, w, `{"ok":true,"result":true}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "answer_callback", "text": "ack"})
    msg := newTestMessage(b, "x", map[string]interface{}{"callback_id": "cbq-1"})
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
    if msg.GetTarget("callback_answered") != "true" { t.Fatalf("callback_answered missing") }
}

func TestTelegramBotAnswerCallback_ExplicitID(t *testing.T) {
    mock := newMockBotAPI(t)
    mock.handle("answerCallbackQuery", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["callback_query_id"] != "explicit-id" { t.Fatalf("got: %v", got) }
        writeJSON(t, w, `{"ok":true,"result":true}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "answer_callback", "text": "ok", "callback_id": "explicit-id"})
    msg := newTestMessage(b, "x", map[string]interface{}{"callback_id": "from-extra"})
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
}

func TestTelegramBotAnswerCallback_NoID(t *testing.T) {
    mock := newMockBotAPI(t)
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "answer_callback", "text": "ok"})
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if ok { t.Fatalf("want ok=false when no callback_id available") }
}

func TestTelegramBotSendMedia_URL(t *testing.T) {
    mock := newMockBotAPI(t)
    mock.handle("sendDocument", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["document"] != "https://example.com/x.pdf" { t.Fatalf("document: %v", got["document"]) }
        if got["chat_id"].(float64) != 1 { t.Fatalf("chat_id: %v", got["chat_id"]) }
        writeJSON(t, w, `{"ok":true,"result":{"message_id":11,"chat":{"id":1,"type":"private"},"date":1}}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "send_media", "media": "https://example.com/x.pdf", "to_chat": "1"})
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
}

func TestTelegramBotSendMedia_FileID(t *testing.T) {
    mock := newMockBotAPI(t)
    mock.handle("sendPhoto", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["photo"] != "abcde-fileid" { t.Fatalf("photo: %v", got["photo"]) }
        writeJSON(t, w, `{"ok":true,"result":{"message_id":1,"chat":{"id":1,"type":"private"},"date":1}}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{
        "action": "send_media", "media": "abcde-fileid", "media_type": "photo", "to_chat": "1",
    })
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
}

func TestTelegramBotSendMedia_LocalUpload(t *testing.T) {
    tmpDir := t.TempDir()
    path := filepath.Join(tmpDir, "hello.txt")
    if err := os.WriteFile(path, []byte("hello-bytes"), 0600); err != nil { t.Fatalf("%v", err) }

    mock := newMockBotAPI(t)
    mock.handleRaw("sendDocument", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        ct := r.Header.Get("Content-Type")
        mediaType, params, err := mime.ParseMediaType(ct)
        if err != nil { t.Fatalf("content-type: %v", err) }
        if mediaType != "multipart/form-data" { t.Fatalf("expected multipart/form-data, got %s", mediaType) }
        mr := multipart.NewReader(r.Body, params["boundary"])
        sawFile := false
        for {
            part, err := mr.NextPart()
            if err != nil { break }
            if part.FormName() == "document" {
                buf, _ := io.ReadAll(part)
                if string(buf) != "hello-bytes" { t.Fatalf("unexpected file content: %q", buf) }
                if part.FileName() != "hello.txt" { t.Fatalf("filename: %s", part.FileName()) }
                sawFile = true
            }
        }
        if !sawFile { t.Fatalf("multipart did not contain document part") }
        writeJSON(t, w, `{"ok":true,"result":{"message_id":1,"chat":{"id":1,"type":"private"},"date":1}}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "send_media", "media": path, "to_chat": "1"})
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
}

func TestTelegramBotSendMedia_UnknownMediaType(t *testing.T) {
    mock := newMockBotAPI(t)
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{
        "action": "send_media", "media": "x", "media_type": "hologram", "to_chat": "1",
    })
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if ok { t.Fatalf("want ok=false for unknown media_type") }
}

func TestTelegramBotDownloadFile_FromExtra(t *testing.T) {
    tmpDir := t.TempDir()
    out := filepath.Join(tmpDir, "downloads", "saved.bin")

    mock := newMockBotAPI(t)
    mock.handle("getFile", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["file_id"] != "FID-1" { t.Fatalf("file_id: %v", got["file_id"]) }
        writeJSON(t, w, `{"ok":true,"result":{"file_id":"FID-1","file_unique_id":"u","file_size":11,"file_path":"docs/hello.bin"}}`)
    })
    mock.files["docs/hello.bin"] = []byte("hello-bytes")
    b, _ := newMockedBot(t, mock)

    f := newTestFilter(t, map[string]string{"action": "download_file", "filename": out})
    msg := newTestMessage(b, "x", map[string]interface{}{"msg_file_id": "FID-1"})
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }

    got, err := os.ReadFile(out)
    if err != nil { t.Fatalf("read: %v", err) }
    if string(got) != "hello-bytes" { t.Fatalf("file content: %q", got) }
    if msg.GetTarget("downloaded_path") != out { t.Fatalf("downloaded_path: %v", msg.GetTarget("downloaded_path")) }
}

func TestTelegramBotDownloadFile_ExplicitID(t *testing.T) {
    tmpDir := t.TempDir()
    out := filepath.Join(tmpDir, "out.bin")

    mock := newMockBotAPI(t)
    mock.handle("getFile", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        body, _ := io.ReadAll(r.Body)
        var got map[string]interface{}
        json.Unmarshal(body, &got)
        if got["file_id"] != "explicit" { t.Fatalf("file_id: %v", got["file_id"]) }
        writeJSON(t, w, `{"ok":true,"result":{"file_id":"explicit","file_unique_id":"u","file_size":4,"file_path":"a/b.bin"}}`)
    })
    mock.files["a/b.bin"] = []byte("DATA")
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "download_file", "filename": out, "file_id": "explicit"})
    msg := newTestMessage(b, "x", map[string]interface{}{"msg_file_id": "from-extra"})
    ok, _ := f.DoFilter(msg)
    if !ok { t.Fatalf("want ok=true") }
}

func TestTelegramBotDownloadFile_GetFileError(t *testing.T) {
    tmpDir := t.TempDir()
    out := filepath.Join(tmpDir, "out.bin")
    mock := newMockBotAPI(t)
    mock.handle("getFile", func(t *testing.T, w http.ResponseWriter, r *http.Request) {
        w.WriteHeader(http.StatusBadRequest)
        writeJSON(t, w, `{"ok":false,"description":"file not found","error_code":400}`)
    })
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "download_file", "filename": out})
    msg := newTestMessage(b, "x", map[string]interface{}{"msg_file_id": "x"})
    ok, _ := f.DoFilter(msg)
    if ok { t.Fatalf("want ok=false") }
}

func TestTelegramBotDownloadFile_NoFileID(t *testing.T) {
    tmpDir := t.TempDir()
    out := filepath.Join(tmpDir, "out.bin")
    mock := newMockBotAPI(t)
    b, _ := newMockedBot(t, mock)
    f := newTestFilter(t, map[string]string{"action": "download_file", "filename": out})
    msg := newTestMessage(b, "x", nil)
    ok, _ := f.DoFilter(msg)
    if ok { t.Fatalf("want ok=false when no file_id available") }
}

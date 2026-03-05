package feeders

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
)

const testWebPage = `<!DOCTYPE html>
<html>
<head>
	<title>Test Page</title>
	<meta name="description" content="Test description">
	<meta property="og:image" content="https://example.com/image.png">
	<meta property="og:site_name" content="Test Site">
</head>
<body>
	<p>Hello World</p>
</body>
</html>`

func newWebTestServer(handler http.HandlerFunc) *httptest.Server {
	return httptest.NewServer(handler)
}

func newDefaultWebTestServer() *httptest.Server {
	return newWebTestServer(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, testWebPage)
	})
}

func newTestWeb(conf map[string]string) (*Web, chan *data.Message, error) {
	feeder, err := NewWebFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*Web)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *Web")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("webfeeder")
	f.setID(1)

	received := make(chan *data.Message, 10)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

// --- Constructor tests ---

func TestNewWebFeeder(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	feeder, err := NewWebFeeder(map[string]string{
		"web.url":  ts.URL,
		"web.freq": "30s",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.url != ts.URL {
			t.Errorf("'web.url' parameter ignored")
		}
		if f.frequency != 30*time.Second {
			t.Errorf("'web.freq' parameter ignored, expected 30s got '%s'", f.frequency)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederDefaults(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.frequency != 60*time.Second {
			t.Errorf("default frequency should be 60s, got '%s'", f.frequency)
		}
		if f.method != "GET" {
			t.Errorf("default method should be GET, got '%s'", f.method)
		}
		if f.textOnly != false {
			t.Errorf("default textOnly should be false")
		}
		if f.checkStatus != 0 {
			t.Errorf("default checkStatus should be 0, got %d", f.checkStatus)
		}
		if !f.lastParsing.IsZero() {
			t.Errorf("default lastParsing should be zero time")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederInvalidFreq(t *testing.T) {
	_, err := NewWebFeeder(map[string]string{
		"web.freq": "notaduration",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'web.freq' is invalid")
	}
}

func TestNewWebFeederInvalidStatus(t *testing.T) {
	_, err := NewWebFeeder(map[string]string{
		"web.status": "notanumber",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'web.status' is not a number")
	}
}

func TestNewWebFeederInvalidHeaders(t *testing.T) {
	_, err := NewWebFeeder(map[string]string{
		"web.headers": "not-valid-json",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'web.headers' is invalid JSON")
	}
}

func TestNewWebFeederInvalidData(t *testing.T) {
	_, err := NewWebFeeder(map[string]string{
		"web.data": "not-valid-json",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'web.data' is invalid JSON")
	}
}

func TestNewWebFeederMethod(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{
		"web.method": "POST",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.method != "POST" {
			t.Errorf("'web.method' parameter ignored, expected POST got '%s'", f.method)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederTextOnly(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{
		"web.text_only": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if !f.textOnly {
			t.Errorf("'web.text_only' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederHeaders(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{
		"web.headers": `{"Authorization": "Bearer token123", "X-Custom": "value"}`,
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.headers["Authorization"] != "Bearer token123" {
			t.Errorf("'web.headers' Authorization not parsed correctly")
		}
		if f.headers["X-Custom"] != "value" {
			t.Errorf("'web.headers' X-Custom not parsed correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederPostData(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{
		"web.data": `{"key": "value", "foo": "bar"}`,
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.dataPost["key"] != "value" {
			t.Errorf("'web.data' key not parsed correctly")
		}
		if f.dataPost["foo"] != "bar" {
			t.Errorf("'web.data' foo not parsed correctly")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederRawData(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{
		"web.rawData": "raw body content",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.rawData != "raw body content" {
			t.Errorf("'web.rawData' parameter ignored, got '%s'", f.rawData)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebFeederStatus(t *testing.T) {
	feeder, err := NewWebFeeder(map[string]string{
		"web.status": "200",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*Web); ok {
		if f.checkStatus != 200 {
			t.Errorf("'web.status' parameter ignored, expected 200 got %d", f.checkStatus)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

// --- parseURL tests ---

func TestWebParseURLPropagatesMessage(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	var msgs []*data.Message
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
}

func TestWebParseURLFirstRun(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	for msg := range received {
		if !msg.IsFirstRun() {
			t.Errorf("expected IsFirstRun() to be true on first run")
		}
	}
}

func TestWebParseURLNotFirstRun(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(false)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	for msg := range received {
		if msg.IsFirstRun() {
			t.Errorf("expected IsFirstRun() to be false on subsequent runs")
		}
	}
}

func TestWebParseURLExtraFields(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	var msgs []*data.Message
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) == 0 {
		t.Fatal("no messages received")
	}

	extra := msgs[0].GetExtra()
	expectedKeys := []string{"url", "title", "description", "image", "sitename"}
	for _, key := range expectedKeys {
		if _, ok := extra[key]; !ok {
			t.Errorf("expected extra field '%s' to be set", key)
		}
	}
	if extra["url"] != ts.URL {
		t.Errorf("expected extra 'url' to be '%s', got '%s'", ts.URL, extra["url"])
	}
}

func TestWebParseURLUpdatesLastParsing(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	before := time.Now()
	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	for range received {
	}

	if f.lastParsing.Before(before) {
		t.Errorf("lastParsing should be updated after parseURL")
	}
}

func TestWebParseURLInvalidURL(t *testing.T) {
	f, _, err := newTestWeb(map[string]string{
		"web.url": "http://127.0.0.1:0/invalid",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err == nil {
		t.Errorf("expected error for invalid URL")
	}
}

func TestWebParseURLCheckStatusMatch(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url":    ts.URL,
		"web.status": "200",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Errorf("parseURL should not return error when status matches, got: %s", err)
	}

	close(received)
	msgCount := 0
	for range received {
		msgCount++
	}
	if msgCount != 1 {
		t.Errorf("expected 1 message when status matches, got %d", msgCount)
	}
}

func TestWebParseURLCheckStatusMismatch(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, _, err := newTestWeb(map[string]string{
		"web.url":    ts.URL,
		"web.status": "404",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err == nil {
		t.Errorf("expected error when status does not match")
	}
}

func TestWebParseURLSendsHeaders(t *testing.T) {
	receivedHeader := ""
	ts := newWebTestServer(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Custom-Header")
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, testWebPage)
	})
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url":     ts.URL,
		"web.headers": `{"X-Custom-Header": "test-value"}`,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	for range received {
	}

	if receivedHeader != "test-value" {
		t.Errorf("expected server to receive header 'test-value', got '%s'", receivedHeader)
	}
}

func TestWebParseURLPostData(t *testing.T) {
	receivedMethod := ""
	receivedBody := ""
	ts := newWebTestServer(func(w http.ResponseWriter, r *http.Request) {
		receivedMethod = r.Method
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, testWebPage)
	})
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url":    ts.URL,
		"web.method": "POST",
		"web.data":   `{"field": "value"}`,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	for range received {
	}

	if receivedMethod != "POST" {
		t.Errorf("expected POST request, got '%s'", receivedMethod)
	}
	if receivedBody == "" {
		t.Errorf("expected non-empty body for POST with data")
	}
}

func TestWebParseURLRawData(t *testing.T) {
	receivedBody := ""
	ts := newWebTestServer(func(w http.ResponseWriter, r *http.Request) {
		buf := make([]byte, 1024)
		n, _ := r.Body.Read(buf)
		receivedBody = string(buf[:n])
		w.WriteHeader(http.StatusOK)
		fmt.Fprint(w, testWebPage)
	})
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url":     ts.URL,
		"web.method":  "POST",
		"web.rawData": "raw content here",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseURL(true)
	if err != nil {
		t.Fatalf("parseURL returned error: %s", err)
	}

	close(received)
	for range received {
	}

	if receivedBody != "raw content here" {
		t.Errorf("expected raw body 'raw content here', got '%s'", receivedBody)
	}
}

func TestWebStartStop(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, _, err := newTestWeb(map[string]string{
		"web.url":  ts.URL,
		"web.freq": "10s",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	if !f.isRunning {
		t.Errorf("feeder should be running after Start()")
	}

	time.Sleep(200 * time.Millisecond)

	f.Stop()
	if f.isRunning {
		t.Errorf("feeder should not be running after Stop()")
	}
}

func TestWebStartPropagatesOnFirstRun(t *testing.T) {
	ts := newDefaultWebTestServer()
	defer ts.Close()

	f, received, err := newTestWeb(map[string]string{
		"web.url":  ts.URL,
		"web.freq": "10s",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.Start()
	defer f.Stop()

	select {
	case msg := <-received:
		if !msg.IsFirstRun() {
			t.Errorf("first message on Start() should have IsFirstRun() = true")
		}
	case <-time.After(500 * time.Millisecond):
		t.Errorf("expected a message on Start() but got none")
	}
}

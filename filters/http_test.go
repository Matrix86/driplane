package filters

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewHTTPFilter(t *testing.T) {
	filter, err := NewHTTPFilter(map[string]string{
		"url":     "http://example.com",
		"method":  "POST",
		"status":  "200",
		"headers": `{"X-Custom":"value"}`,
		"data":    `{"key":"val"}`,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*HTTP)
	if !ok {
		t.Fatal("cannot cast to *HTTP")
	}
	if f.method != "POST" {
		t.Errorf("expected method 'POST', got '%s'", f.method)
	}
	if f.checkStatus != 200 {
		t.Errorf("expected checkStatus 200, got %d", f.checkStatus)
	}
	if v, ok := f.headers["X-Custom"]; !ok || v != "value" {
		t.Errorf("expected header X-Custom=value, got '%v'", f.headers)
	}
	if _, ok := f.dataPost["key"]; !ok {
		t.Errorf("expected dataPost key 'key'")
	}
}

func TestNewHTTPFilterDefaults(t *testing.T) {
	filter, err := NewHTTPFilter(map[string]string{
		"url": "http://example.com",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	if f.method != "GET" {
		t.Errorf("default method should be 'GET', got '%s'", f.method)
	}
	if f.checkStatus != 0 {
		t.Errorf("default checkStatus should be 0, got %d", f.checkStatus)
	}
	if f.textOnly {
		t.Errorf("default textOnly should be false")
	}
}

func TestNewHTTPFilterTextOnly(t *testing.T) {
	filter, err := NewHTTPFilter(map[string]string{
		"url":       "http://example.com",
		"text_only": "true",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	if !f.textOnly {
		t.Errorf("expected textOnly to be true")
	}
}

func TestNewHTTPFilterInvalidStatus(t *testing.T) {
	_, err := NewHTTPFilter(map[string]string{
		"url":    "http://example.com",
		"status": "notanumber",
	})
	if err == nil {
		t.Errorf("expected error for invalid status")
	}
}

func TestNewHTTPFilterInvalidHeaders(t *testing.T) {
	_, err := NewHTTPFilter(map[string]string{
		"url":     "http://example.com",
		"headers": "invalid json",
	})
	if err == nil {
		t.Errorf("expected error for invalid headers JSON")
	}
}

func TestNewHTTPFilterInvalidData(t *testing.T) {
	_, err := NewHTTPFilter(map[string]string{
		"url":  "http://example.com",
		"data": "invalid json",
	})
	if err == nil {
		t.Errorf("expected error for invalid data JSON")
	}
}

func TestNewHTTPFilterRawData(t *testing.T) {
	filter, err := NewHTTPFilter(map[string]string{
		"url":     "http://example.com",
		"rawData": `{"raw":"body"}`,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	if f.rawData == nil {
		t.Errorf("expected rawData template to be set")
	}
}

func TestNewHTTPFilterDownloadTo(t *testing.T) {
	filter, err := NewHTTPFilter(map[string]string{
		"url":         "http://example.com",
		"download_to": "/tmp/{{.main}}",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	if f.downloadTo == nil {
		t.Errorf("expected downloadTo template to be set")
	}
}

func TestHTTPDoFilterGET(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("expected GET, got %s", r.Method)
		}
		fmt.Fprint(w, "response body")
	}))
	defer ts.Close()

	filter, err := NewHTTPFilter(map[string]string{
		"url": ts.URL,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
	// The body is returned as []byte
	if body, ok := msg.GetMessage().([]byte); ok {
		if string(body) != "response body" {
			t.Errorf("expected body 'response body', got '%s'", string(body))
		}
	} else {
		t.Errorf("expected []byte message, got %T", msg.GetMessage())
	}
}

func TestHTTPDoFilterPOSTWithData(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST, got %s", r.Method)
		}
		body, err := io.ReadAll(r.Body)
		if err != nil {
			t.Errorf("failed to read body: %s", err)
		}
		values, err := url.ParseQuery(string(body))
		if err != nil {
			t.Errorf("failed to parse body as form values: %s", err)
		}
		if values.Get("key") != "value" {
			t.Errorf("expected form key=value, got '%s'", values.Get("key"))
		}
		fmt.Fprint(w, "post response")
	}))
	defer ts.Close()

	filter, err := NewHTTPFilter(map[string]string{
		"url":    ts.URL,
		"method": "POST",
		"data":   `{"key":"value"}`,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
}

func TestHTTPDoFilterWithHeaders(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Header.Get("X-Test") != "hello" {
			t.Errorf("expected X-Test header 'hello', got '%s'", r.Header.Get("X-Test"))
		}
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	filter, err := NewHTTPFilter(map[string]string{
		"url":     ts.URL,
		"headers": `{"X-Test":"hello"}`,
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	_, err = f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
}

func TestHTTPDoFilterCheckStatusMatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		fmt.Fprint(w, "ok")
	}))
	defer ts.Close()

	filter, _ := NewHTTPFilter(map[string]string{
		"url":    ts.URL,
		"status": "200",
	})
	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true when status matches")
	}
}

func TestHTTPDoFilterCheckStatusMismatch(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
	}))
	defer ts.Close()

	filter, _ := NewHTTPFilter(map[string]string{
		"url":    ts.URL,
		"status": "200",
	})
	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error when status doesn't match")
	}
	if ok {
		t.Errorf("DoFilter should return false when status doesn't match")
	}
}

func TestHTTPDoFilterTextOnly(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "<html><body><p>hello world</p></body></html>")
	}))
	defer ts.Close()

	filter, _ := NewHTTPFilter(map[string]string{
		"url":       ts.URL,
		"text_only": "true",
	})
	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
	// Result should be plain text (string), not contain HTML tags
	result := fmt.Sprintf("%v", msg.GetMessage())
	if len(result) == 0 {
		t.Errorf("expected non-empty text result")
	}
}

func TestHTTPDoFilterURLTemplate(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, "templated")
	}))
	defer ts.Close()

	filter, err := NewHTTPFilter(map[string]string{
		"url": ts.URL + "/{{.main}}",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*HTTP)
	msg := data.NewMessage("page")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Errorf("DoFilter should return true")
	}
}

func TestHTTPDoFilterInvalidURL(t *testing.T) {
	filter, _ := NewHTTPFilter(map[string]string{
		"url": "http://127.0.0.1:1/invalid",
	})
	f := filter.(*HTTP)
	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error for unreachable URL")
	}
	if ok {
		t.Errorf("DoFilter should return false for unreachable URL")
	}
}

func TestHTTPOnEvent(t *testing.T) {
	filter, _ := NewHTTPFilter(map[string]string{"url": "http://example.com"})
	f := filter.(*HTTP)
	f.OnEvent(&data.Event{})
}

func TestHTTPFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["httpfilter"]; !ok {
		t.Errorf("http filter should be registered as 'httpfilter'")
	}
}

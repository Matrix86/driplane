package filters

import (
	"testing"

	"github.com/Matrix86/driplane/data"
)

func TestNewElasticSearchFilter(t *testing.T) {
	filter, err := NewElasticSearchFilter(map[string]string{
		"address":  "http://localhost:9200",
		"username": "elastic",
		"password": "secret",
		"index":    "test-index",
		"target":   "main",
		"retries":  "3",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*ElasticSearch)
	if !ok {
		t.Fatal("cannot cast to *ElasticSearch")
	}
	if f.address != "http://localhost:9200" {
		t.Errorf("expected address 'http://localhost:9200', got '%s'", f.address)
	}
	if f.username != "elastic" {
		t.Errorf("expected username 'elastic', got '%s'", f.username)
	}
	if f.password != "secret" {
		t.Errorf("expected password 'secret', got '%s'", f.password)
	}
	if f.index != "test-index" {
		t.Errorf("expected index 'test-index', got '%s'", f.index)
	}
	if f.target != "main" {
		t.Errorf("expected target 'main', got '%s'", f.target)
	}
	if f.retries != 3 {
		t.Errorf("expected retries 3, got %d", f.retries)
	}
}

func TestNewElasticSearchFilterDefaults(t *testing.T) {
	filter, err := NewElasticSearchFilter(map[string]string{})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*ElasticSearch)
	if f.address != "localhost:9200" {
		t.Errorf("default address should be 'localhost:9200', got '%s'", f.address)
	}
	if f.retries != 1 {
		t.Errorf("default retries should be 1, got %d", f.retries)
	}
	if f.target != "main" {
		t.Errorf("default target should be 'main', got '%s'", f.target)
	}
}

func TestNewElasticSearchFilterInvalidRetries(t *testing.T) {
	_, err := NewElasticSearchFilter(map[string]string{
		"retries": "notanumber",
	})
	if err == nil {
		t.Errorf("expected error for invalid retries value")
	}
}

func TestElasticSearchDoFilterNoClient(t *testing.T) {
	filter, _ := NewElasticSearchFilter(map[string]string{
		"address": "http://127.0.0.1:1",
		"retries": "1",
		"index":   "test",
	})

	f := filter.(*ElasticSearch)
	msg := data.NewMessage(`{"key":"value"}`)
	// DoFilter should fail because it can't connect
	_, err := f.DoFilter(msg)
	if err == nil {
		t.Errorf("expected error when ES is not reachable")
	}
}

func TestElasticSearchOnEvent(t *testing.T) {
	filter, _ := NewElasticSearchFilter(map[string]string{})
	f := filter.(*ElasticSearch)
	// OnEvent is a no-op; just ensure no panic
	f.OnEvent(&data.Event{})
}

func TestElasticSearchFilterRegistered(t *testing.T) {
	if _, ok := filterFactories["elasticsearchfilter"]; !ok {
		t.Errorf("elasticsearch filter should be registered as 'elasticsearchfilter'")
	}
}

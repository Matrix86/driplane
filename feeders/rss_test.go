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

const testRSSFeed = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed</title>
    <link>https://example.com</link>
    <description>Test RSS Feed</description>
    <language>en</language>
    <copyright>Test Copyright</copyright>
    <generator>Test Generator</generator>
    <item>
      <title>First Article</title>
      <link>https://example.com/first</link>
      <description>First description</description>
      <pubDate>Mon, 01 Jan 2024 00:00:00 +0000</pubDate>
      <author>John Doe</author>
    </item>
    <item>
      <title>Second Article</title>
      <link>https://example.com/second</link>
      <description>Second description</description>
      <pubDate>Tue, 02 Jan 2024 00:00:00 +0000</pubDate>
      <author>Jane Doe</author>
    </item>
  </channel>
</rss>`

const testRSSFeedNoDate = `<?xml version="1.0" encoding="UTF-8"?>
<rss version="2.0">
  <channel>
    <title>Test Feed No Date</title>
    <link>https://example.com</link>
    <description>Test RSS Feed without dates</description>
    <item>
      <title>No Date Article</title>
      <link>https://example.com/nodate</link>
      <description>No date description</description>
    </item>
  </channel>
</rss>`

func newRSSTestServer(feed string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/rss+xml; charset=utf-8")
		fmt.Fprint(w, feed)
	}))
}

// newTestRSS creates an RSS feeder wired to a real EventBus so Propagate works
func newTestRSS(conf map[string]string) (*RSS, chan *data.Message, error) {
	feeder, err := NewRSSFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*RSS)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *RSS")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("rssfeeder")
	f.setID(1)

	received := make(chan *data.Message, 10)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

func TestNewRSSFeeder(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	feeder, err := NewRSSFeeder(map[string]string{
		"rss.url":  ts.URL,
		"rss.freq": "30s",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*RSS); ok {
		if f.url != ts.URL {
			t.Errorf("'rss.url' parameter ignored")
		}
		if f.frequency != 30*time.Second {
			t.Errorf("'rss.freq' parameter ignored, expected 30s got '%s'", f.frequency)
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewRSSFeederDefaults(t *testing.T) {
	feeder, err := NewRSSFeeder(map[string]string{})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*RSS); ok {
		if f.frequency != 60*time.Second {
			t.Errorf("default frequency should be 60s, got '%s'", f.frequency)
		}
		if f.ignorePubDate != false {
			t.Errorf("default ignorePubDate should be false")
		}
		if !f.lastParsing.IsZero() {
			t.Errorf("default lastParsing should be zero time")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewRSSFeederInvalidFreq(t *testing.T) {
	_, err := NewRSSFeeder(map[string]string{
		"rss.url":  "https://example.com",
		"rss.freq": "notaduration",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'rss.freq' is invalid")
	}
}

func TestNewRSSFeederStartFromBeginning(t *testing.T) {
	before := time.Now()
	feeder, err := NewRSSFeeder(map[string]string{
		"rss.start_from_beginning": "false",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*RSS); ok {
		if f.lastParsing.Before(before) {
			t.Errorf("lastParsing should be set to ~now when start_from_beginning=false")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewRSSFeederIgnorePubDate(t *testing.T) {
	feeder, err := NewRSSFeeder(map[string]string{
		"rss.ignore_pubdate": "true",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*RSS); ok {
		if !f.ignorePubDate {
			t.Errorf("'rss.ignore_pubdate' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestRSSParseFeedArticles(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	var msgs []*data.Message
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages, got %d", len(msgs))
	}
}

func TestRSSParseFeedFirstRun(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	for msg := range received {
		if !msg.IsFirstRun() {
			t.Errorf("expected IsFirstRun() to be true on first run")
		}
	}
}

func TestRSSParseFeedNotFirstRun(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// reset lastParsing to zero so all items pass the date filter
	f.lastParsing = time.Time{}

	err = f.parseFeed(false)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	for msg := range received {
		if msg.IsFirstRun() {
			t.Errorf("expected IsFirstRun() to be false on subsequent runs")
		}
	}
}

func TestRSSParseFeedExtraFields(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	msgs := make([]*data.Message, 0)
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) == 0 {
		t.Fatal("no messages received")
	}

	first := msgs[0]
	extra := first.GetExtra()

	expectedKeys := []string{"feed_title", "feed_link", "feed_language", "feed_copyright", "feed_generator"}
	for _, key := range expectedKeys {
		if _, ok := extra[key]; !ok {
			t.Errorf("expected extra field '%s' to be set", key)
		}
	}
}

func TestRSSParseFeedMainIsTitle(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	for msg := range received {
		if msg.GetMessage() == "" {
			t.Errorf("expected main field to be the article title, got empty string")
		}
	}
}

func TestRSSParseFeedSkipsOldItems(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// set lastParsing to future so all items are considered old
	f.lastParsing = time.Now().Add(24 * time.Hour)

	err = f.parseFeed(false)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	msgCount := 0
	for range received {
		msgCount++
	}

	if msgCount != 0 {
		t.Errorf("expected 0 messages for old items, got %d", msgCount)
	}
}

func TestRSSParseFeedIgnorePubDate(t *testing.T) {
	ts := newRSSTestServer(testRSSFeedNoDate)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url":            ts.URL,
		"rss.ignore_pubdate": "true",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	msgCount := 0
	for range received {
		msgCount++
	}

	if msgCount != 1 {
		t.Errorf("expected 1 message with ignore_pubdate=true even without date, got %d", msgCount)
	}
}

func TestRSSParseFeedUpdatesLastParsing(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	zeroBefore := f.lastParsing.IsZero()

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	for range received {
	}

	if zeroBefore && f.lastParsing.IsZero() {
		t.Errorf("lastParsing should be updated after parseFeed")
	}
}

func TestRSSParseFeedInvalidURL(t *testing.T) {
	f, _, err := newTestRSS(map[string]string{
		"rss.url": "http://127.0.0.1:0/invalid",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err == nil {
		t.Errorf("expected error for invalid URL")
	}
}

func TestRSSStartStop(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, _, err := newTestRSS(map[string]string{
		"rss.url":  ts.URL,
		"rss.freq": "10s",
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

func TestGetPublishDate(t *testing.T) {
	ts := newRSSTestServer(testRSSFeed)
	defer ts.Close()

	f, received, err := newTestRSS(map[string]string{
		"rss.url": ts.URL,
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	err = f.parseFeed(true)
	if err != nil {
		t.Fatalf("parseFeed returned error: %s", err)
	}

	close(received)
	msgs := make([]*data.Message, 0)
	for msg := range received {
		msgs = append(msgs, msg)
	}

	// items should be ordered by pubdate, lastParsing should reflect the latest
	if f.lastParsing.IsZero() {
		t.Errorf("lastParsing should be set after parsing feed with dated items")
	}
}

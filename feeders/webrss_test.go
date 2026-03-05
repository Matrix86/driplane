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

func newWebRSSTestServer(html string) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, html)
	}))
}

const testWebRSSPage = `
<html><body>
	<div class="post">
		<h2 class="title">First Article</h2>
		<a class="link" href="/blog/first-article">Read more</a>
		<p class="desc">First description</p>
		<span class="date">2024-01-01</span>
	</div>
	<div class="post">
		<h2 class="title">Second Article</h2>
		<a class="link" href="/blog/second-article">Read more</a>
		<p class="desc">Second description</p>
		<span class="date">2024-01-02</span>
	</div>
</body></html>
`

// newTestWebRSS creates a WebRSS feeder wired to a real EventBus so Propagate works
func newTestWebRSS(conf map[string]string) (*WebRSS, chan *data.Message, error) {
	feeder, err := NewWebRSSFeeder(conf)
	if err != nil {
		return nil, nil, err
	}

	f, ok := feeder.(*WebRSS)
	if !ok {
		return nil, nil, fmt.Errorf("cannot cast to *WebRSS")
	}

	bus := EventBus.New()
	f.setBus(bus)
	f.setName("webrssfeeder")
	f.setID(1)

	received := make(chan *data.Message, 10)
	bus.Subscribe(f.GetIdentifier(), func(msg *data.Message) {
		received <- msg
	})

	return f, received, nil
}

func TestNewWebRSSFeeder(t *testing.T) {
	ts := newWebRSSTestServer(testWebRSSPage)
	defer ts.Close()

	feeder, err := NewWebRSSFeeder(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
		"webrss.desc_selector":  ".desc",
		"webrss.date_selector":  ".date",
		"webrss.freq":           "1h",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*WebRSS); ok {
		if f.url != ts.URL {
			t.Errorf("'webrss.url' parameter ignored")
		}
		if f.itemSelector != ".post" {
			t.Errorf("'webrss.item_selector' parameter ignored")
		}
		if f.titleSelector != ".title" {
			t.Errorf("'webrss.title_selector' parameter ignored")
		}
		if f.linkSelector != ".link" {
			t.Errorf("'webrss.link_selector' parameter ignored")
		}
		if f.descSelector != ".desc" {
			t.Errorf("'webrss.desc_selector' parameter ignored")
		}
		if f.dateSelector != ".date" {
			t.Errorf("'webrss.date_selector' parameter ignored")
		}
		if f.frequency != 1*time.Hour {
			t.Errorf("'webrss.freq' parameter ignored")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebRSSFeederDefaults(t *testing.T) {
	ts := newWebRSSTestServer(testWebRSSPage)
	defer ts.Close()

	feeder, err := NewWebRSSFeeder(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
	})
	if err != nil {
		t.Errorf("constructor returned '%s'", err)
	}
	if f, ok := feeder.(*WebRSS); ok {
		if f.linkAttr != "href" {
			t.Errorf("default 'link_attr' should be 'href', got '%s'", f.linkAttr)
		}
		if f.frequency != 60*time.Minute {
			t.Errorf("default 'freq' should be 60m, got '%s'", f.frequency)
		}
		if f.descSelector != "" {
			t.Errorf("default 'desc_selector' should be empty")
		}
		if f.dateSelector != "" {
			t.Errorf("default 'date_selector' should be empty")
		}
	} else {
		t.Errorf("cannot cast to proper Feeder...")
	}
}

func TestNewWebRSSFeederMissingURL(t *testing.T) {
	_, err := NewWebRSSFeeder(map[string]string{
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'webrss.url' is missing")
	}
}

func TestNewWebRSSFeederMissingItemSelector(t *testing.T) {
	_, err := NewWebRSSFeeder(map[string]string{
		"webrss.url":            "https://example.com",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'webrss.item_selector' is missing")
	}
}

func TestNewWebRSSFeederMissingTitleSelector(t *testing.T) {
	_, err := NewWebRSSFeeder(map[string]string{
		"webrss.url":           "https://example.com",
		"webrss.item_selector": ".post",
		"webrss.link_selector": ".link",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'webrss.title_selector' is missing")
	}
}

func TestNewWebRSSFeederMissingLinkSelector(t *testing.T) {
	_, err := NewWebRSSFeeder(map[string]string{
		"webrss.url":            "https://example.com",
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'webrss.link_selector' is missing")
	}
}

func TestNewWebRSSFeederInvalidFreq(t *testing.T) {
	_, err := NewWebRSSFeeder(map[string]string{
		"webrss.url":            "https://example.com",
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
		"webrss.freq":           "notaduration",
	})
	if err == nil {
		t.Errorf("constructor should return an error if 'webrss.freq' is invalid")
	}
}

func TestWebRSSScrapeArticles(t *testing.T) {
	ts := newWebRSSTestServer(testWebRSSPage)
	defer ts.Close()

	f, received, err := newTestWebRSS(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
		"webrss.desc_selector":  ".desc",
		"webrss.date_selector":  ".date",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.scrape(true)

	if len(f.seenLinks) != 2 {
		t.Errorf("expected 2 articles to be scraped, got %d", len(f.seenLinks))
	}

	// collect propagated messages
	close(received)
	var msgs []*data.Message
	for msg := range received {
		msgs = append(msgs, msg)
	}

	if len(msgs) != 2 {
		t.Fatalf("expected 2 messages propagated, got %d", len(msgs))
	}

	// check first message fields
	first := msgs[0]
	if first.GetMessage() == "" {
		t.Errorf("expected non-empty title in main field")
	}
	if v, ok := first.GetExtra()["link"]; !ok || v == "" {
		t.Errorf("expected 'link' extra field to be set")
	}
	if v, ok := first.GetExtra()["description"]; !ok || v == "" {
		t.Errorf("expected 'description' extra field to be set")
	}
	if v, ok := first.GetExtra()["published_at"]; !ok || v == "" {
		t.Errorf("expected 'published_at' extra field to be set")
	}
}

func TestWebRSSScrapeDeduplication(t *testing.T) {
	ts := newWebRSSTestServer(testWebRSSPage)
	defer ts.Close()

	f, received, err := newTestWebRSS(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	// first scrape — should see 2 articles
	f.scrape(true)
	firstCount := len(f.seenLinks)

	// second scrape — same page, nothing new should be propagated
	f.scrape(false)
	secondCount := len(f.seenLinks)

	close(received)
	msgCount := 0
	for range received {
		msgCount++
	}

	if firstCount != 2 {
		t.Errorf("expected 2 seen links after first scrape, got %d", firstCount)
	}
	if secondCount != firstCount {
		t.Errorf("expected no new links after second scrape, got %d (was %d)", secondCount, firstCount)
	}
	if msgCount != 2 {
		t.Errorf("expected exactly 2 messages propagated total (no duplicates), got %d", msgCount)
	}
}

func TestWebRSSScrapeEmptyPage(t *testing.T) {
	ts := newWebRSSTestServer(`<html><body></body></html>`)
	defer ts.Close()

	f, received, err := newTestWebRSS(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.scrape(true)
	close(received)

	if len(f.seenLinks) != 0 {
		t.Errorf("expected 0 seen links on empty page, got %d", len(f.seenLinks))
	}
}

func TestWebRSSScrapeRelativeURLResolution(t *testing.T) {
	ts := newWebRSSTestServer(testWebRSSPage)
	defer ts.Close()

	f, _, err := newTestWebRSS(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
	})
	if err != nil {
		t.Fatalf("setup failed: %s", err)
	}

	f.scrape(true)

	for link := range f.seenLinks {
		if len(link) < 4 || link[:4] != "http" {
			t.Errorf("expected absolute URL, got '%s'", link)
		}
	}
}

func TestWebRSSStartStop(t *testing.T) {
	ts := newWebRSSTestServer(testWebRSSPage)
	defer ts.Close()

	f, _, err := newTestWebRSS(map[string]string{
		"webrss.url":            ts.URL,
		"webrss.item_selector":  ".post",
		"webrss.title_selector": ".title",
		"webrss.link_selector":  ".link",
		"webrss.freq":           "10s",
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

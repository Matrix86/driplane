package feeders

import (
	"fmt"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"github.com/gocolly/colly/v2"
)

// WebRSS is a Feeder that scrapes a website and emits one Message per article found
type WebRSS struct {
	Base

	url           string
	frequency     time.Duration
	itemSelector  string
	titleSelector string
	linkSelector  string
	descSelector  string
	dateSelector  string
	linkAttr      string

	stopChan  chan bool
	ticker    *time.Ticker
	seenLinks map[string]bool
}

// NewWebRSSFeeder is the registered method to instantiate a WebRSSFeeder
func NewWebRSSFeeder(conf map[string]string) (Feeder, error) {
	f := &WebRSS{
		stopChan:  make(chan bool),
		frequency: 60 * time.Minute,
		linkAttr:  "href",
		seenLinks: make(map[string]bool),
	}

	if val, ok := conf["webrss.url"]; ok {
		f.url = val
	} else {
		return nil, fmt.Errorf("WebRSSFeeder: 'webrss.url' is required")
	}
	if val, ok := conf["webrss.item_selector"]; ok {
		f.itemSelector = val
	} else {
		return nil, fmt.Errorf("WebRSSFeeder: 'webrss.item_selector' is required")
	}
	if val, ok := conf["webrss.title_selector"]; ok {
		f.titleSelector = val
	} else {
		return nil, fmt.Errorf("WebRSSFeeder: 'webrss.title_selector' is required")
	}
	if val, ok := conf["webrss.link_selector"]; ok {
		f.linkSelector = val
	} else {
		return nil, fmt.Errorf("WebRSSFeeder: 'webrss.link_selector' is required")
	}
	if val, ok := conf["webrss.desc_selector"]; ok {
		f.descSelector = val
	}
	if val, ok := conf["webrss.date_selector"]; ok {
		f.dateSelector = val
	}
	if val, ok := conf["webrss.link_attr"]; ok {
		f.linkAttr = val
	}
	if val, ok := conf["webrss.freq"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("WebRSSFeeder: invalid frequency '%s': %s", val, err)
		}
		f.frequency = d
	}

	return f, nil
}

func (f *WebRSS) scrape(firstRun bool) {
	c := colly.NewCollector()

	c.OnResponse(func(r *colly.Response) {
		log.Debug("got response %d: %s", r.StatusCode, r.Headers.Get("Content-Type"))
	})

	c.OnHTML("body", func(e *colly.HTMLElement) {
		e.ForEach(f.itemSelector, func(_ int, el *colly.HTMLElement) {
			fmt.Printf("element: %#v\n", el)
			var link string
			if f.linkSelector == "" || f.linkSelector == "self" {
				link = el.Request.AbsoluteURL(el.Attr(f.linkAttr))
			} else {
				link = el.Request.AbsoluteURL(el.ChildAttr(f.linkSelector, f.linkAttr))
			}

			if link == "" {
				return
			}
			if f.seenLinks[link] {
				return
			}

			title := el.ChildText(f.titleSelector)
			desc := ""
			date := ""

			if f.descSelector != "" {
				desc = el.ChildText(f.descSelector)
			}
			if f.dateSelector != "" {
				date = el.ChildText(f.dateSelector)
			}

			if title == "" {
				log.Debug("Title not found")
				return
			}

			extra := map[string]interface{}{
				"title":        title,
				"description":  desc,
				"published_at": date,
				"link":         link,
			}

			msg := data.NewMessageWithExtra(title, extra)
			if firstRun {
				msg.SetFirstRun()
			}

			f.seenLinks[link] = true
			f.Propagate(msg)
		})
	})

	c.OnError(func(r *colly.Response, err error) {
		log.Error("[WebRSSFeeder] error scraping %s: %v", f.url, err)
	})

	if err := c.Visit(f.url); err != nil {
		log.Error("[WebRSSFeeder] visit error: %v", err)
	}
}

// Start propagates a message every time a new article is found
func (f *WebRSS) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		f.scrape(true)

		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case <-f.ticker.C:
				f.scrape(false)
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *WebRSS) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.stopChan <- true
	f.ticker.Stop()
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *WebRSS) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("webrss", NewWebRSSFeeder)
}

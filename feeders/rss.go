package feeders

import (
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
	"github.com/mmcdole/gofeed"
)

// RSS is a Feeder that creates a stream from a RSS feed
type RSS struct {
	Base

	url           string
	frequency     time.Duration
	ignorePubDate bool

	parser      *gofeed.Parser
	stopChan    chan bool
	ticker      *time.Ticker
	lastParsing time.Time
}

// NewRSSFeeder is the registered method to instantiate a RSSFeeder
func NewRSSFeeder(conf map[string]string) (Feeder, error) {
	f := &RSS{
		parser:      gofeed.NewParser(),
		stopChan:    make(chan bool),
		frequency:   60 * time.Second,
		ignorePubDate: false,
		lastParsing: time.Time{},
	}

	if val, ok := conf["rss.url"]; ok {
		f.url = val
	}
	if val, ok := conf["rss.freq"]; ok {
		d, err := time.ParseDuration(val)
		if err != nil {
			return nil, fmt.Errorf("specified frequency cannot be parsed '%s': %s", val, err)
		}
		f.frequency = d
	}
	if val, ok := conf["rss.start_from_beginning"]; ok && val == "false" {
		f.lastParsing = time.Now()
	}
	if val, ok := conf["rss.ignore_pubdate"]; ok && val == "true" {
		f.ignorePubDate = true
	}

	return f, nil
}

func (f *RSS) parseFeed() error {
	var lastPubDate time.Time
	log.Debug("Start RSS parsing: %s", f.url)
	feed, err := f.parser.ParseURL(f.url)
	if err != nil {
		return err
	}

	log.Debug("Found %d items", len(feed.Items))
	for _, item := range feed.Items {
		extra := make(map[string]interface{})

		//log.Debug("time : %s", item.PublishedParsed.Format("2006-01-02 15:04:05"))

		// send messages only if pubDate is setted
		if f.ignorePubDate || (item.PublishedParsed != nil && f.lastParsing.Before(*item.PublishedParsed)) {
			extra["feed_title"] = feed.Title
			extra["feed_link"] = feed.Link
			extra["feed_feedlink"] = feed.FeedLink
			extra["feed_updated"] = feed.Updated
			extra["feed_published"] = feed.Published
			if feed.Author != nil {
				extra["feed_author"] = fmt.Sprintf("%s <%s>", feed.Author.Name, feed.Author.Email)
			}
			extra["feed_language"] = feed.Language
			extra["feed_copyright"] = feed.Copyright
			extra["feed_generator"] = feed.Generator

			elems := reflect.ValueOf(item).Elem()
			typeOfT := elems.Type()
			// Get all the RSS string fields as extra fields
			for i := 0; i < elems.NumField(); i++ {
				f := elems.Field(i)
				if f.Type().String() == "string" {
					extra[strings.ToLower(typeOfT.Field(i).Name)] = f.Interface().(string)
				} else if f.Type().String() == "[]string" {
					slice := f.Interface().([]string)
					quoted := make([]string, len(slice))
					for x, v := range slice {
						quoted[x] = fmt.Sprintf("'%s'", v)
					}
					extra[strings.ToLower(typeOfT.Field(i).Name)] = strings.Join(quoted, ",")
				}
			}

			for k, v := range item.Custom {
				extra[strings.ToLower(k)] = v
			}

			main := ""
			if item.Title != "" {
				main = item.Title
			}

			msg := data.NewMessageWithExtra(main, extra)
			f.Propagate(msg)
		}

		// pubDate of the last rss item
		if item.PublishedParsed != nil && lastPubDate.Before(*item.PublishedParsed) {
			lastPubDate = *item.PublishedParsed
		}
	}

	f.lastParsing = lastPubDate
	log.Debug("Latest item has been published on %s...updating date", lastPubDate.Format("2006-01-02 15:04:05"))

	return nil
}

// Start propagates a message every time a new row is published
func (f *RSS) Start() {
	f.ticker = time.NewTicker(f.frequency)
	go func() {
		// first start!
		_ = f.parseFeed()

		for {
			select {
			case <-f.stopChan:
				log.Debug("%s: stop arrived on the channel", f.Name())
				return
			case <-f.ticker.C:
				_ = f.parseFeed()
			}
		}
	}()

	f.isRunning = true
}

// Stop handles the Feeder shutdown
func (f *RSS) Stop() {
	log.Debug("feeder '%s' stream stop", f.Name())
	f.stopChan <- true
	f.ticker.Stop()
	f.isRunning = false
}

// OnEvent is called when an event occurs
func (f *RSS) OnEvent(event *data.Event) {}

// Auto factory adding
func init() {
	register("rss", NewRSSFeeder)
}

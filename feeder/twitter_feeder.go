package feeder

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/Matrix86/driplane/com"

	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/evilsocket/islazy/log"
)

type TwitterFeeder struct {
	FeederBase

	consumerKey    string
	consumerSecret string
	accessToken    string
	accessSecret   string

	keywords      string
	languages     string
	stallWarnings bool

	stream      *twitter.Stream
}

func NewTwitterFeeder(conf map[string]string) (Feeder, error) {
	t := &TwitterFeeder{
		stallWarnings: false,
	}

	if val, ok := conf["twitter.consumerKey"]; ok {
		t.consumerKey = val
	}
	if val, ok := conf["twitter.consumerSecret"]; ok {
		t.consumerSecret = val
	}
	if val, ok := conf["twitter.accessToken"]; ok {
		t.accessToken = val
	}
	if val, ok := conf["twitter.accessSecret"]; ok {
		t.accessSecret = val
	}
	if val, ok := conf["twitter.keywords"]; ok {
		t.keywords = val
	}
	if val, ok := conf["twitter.languages"]; ok {
		t.languages = val
	}
	if val, ok := conf["twitter.stallWarnings"]; ok {
		b, err := strconv.ParseBool(val)
		if err != nil {
			log.Error("Value for 'twitter.stallWarnings' is invalid '%s'", val)
		} else {
			t.stallWarnings = b
		}
	}

	return t, nil
}

func (t *TwitterFeeder) Start() {
	var err error

	log.Debug("Initialization of TwitterFeeder")
	config := oauth1.NewConfig(t.consumerKey, t.consumerSecret)
	token := oauth1.NewToken(t.accessToken, t.accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	client := twitter.NewClient(httpClient)

	log.Debug("Setting demuxer")
	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		text := tweet.Text
		if tweet.ExtendedTweet != nil {
			text = tweet.ExtendedTweet.FullText
		}
		if strings.HasPrefix(text, "RT ") == false {

			var msg com.DataMessage
			msg.SetMessage(text)
			msg.SetExtra("link", fmt.Sprintf("https://twitter.com/statuses/%d", tweet.ID))
			msg.SetExtra("language", tweet.Lang)
			msg.SetExtra("username", tweet.User.ScreenName)
			t.Propagate(msg)
		}
	}

	// we don't track Direct Messages or Events
	demux.DM = func(dm *twitter.DirectMessage) {
	}
	demux.Event = func(event *twitter.Event) {
	}

	// FILTER
	log.Debug("Setting keywords: '%s'", t.keywords)
	keywords := strings.Split(t.keywords, ",")
	filterParams := &twitter.StreamFilterParams{
		Track:         keywords,
		StallWarnings: twitter.Bool(true),
	}
	if t.languages != "" {
		languages := strings.Split(t.languages, ",")
		filterParams.Language = languages
	}
	t.stream, err = client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal("twitter stream filter error: %s", err)
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(t.stream.Messages)
	t.isRunning = true
}

func (t *TwitterFeeder) Stop() {
	log.Debug("feeder '%s' stream stop", t.Name())
	t.stream.Stop()
	t.isRunning = false
}

// Auto factory adding
func init() {
	register("twitter", NewTwitterFeeder)
}
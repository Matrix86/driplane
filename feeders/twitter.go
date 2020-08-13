package feeders

import (
	"fmt"
	"github.com/Matrix86/driplane/data"
	"github.com/dghubble/go-twitter/twitter"
	"github.com/dghubble/oauth1"
	"github.com/evilsocket/islazy/log"
	"strconv"
	"strings"
	"time"
)

type Twitter struct {
	Base

	consumerKey    string
	consumerSecret string
	accessToken    string
	accessSecret   string

	keywords      string
	users         string
	languages     string
	retweet       bool
	quoted        bool
	stallWarnings bool

	stream *twitter.Stream
	client *twitter.Client
}

// Doc
// https://developer.twitter.com/en/docs/tweets/filter-realtime/guides/basic-stream-parameters
func NewTwitterFeeder(conf map[string]string) (Feeder, error) {
	t := &Twitter{
		stallWarnings: false,
		retweet:       true,
		quoted:        true,
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
	if val, ok := conf["twitter.users"]; ok {
		t.users = val
	}
	if val, ok := conf["twitter.languages"]; ok {
		t.languages = val
	}
	if val, ok := conf["twitter.disable_retweet"]; ok && val == "true" {
		t.retweet = false
	}
	if val, ok := conf["twitter.disable_quoted"]; ok && val == "true" {
		t.quoted = false
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

func (t *Twitter) getIdsFromUsernames(usernames []string) (map[string]int64, error) {
	ids := make(map[string]int64)
	// This API supports max 100 ids per request and 300 requests/15min window
	for start, end := 0, 100; start < len(usernames); start, end = start+100, end+100 {
		if end > len(usernames) {
			end = len(usernames)
		}
		params := &twitter.UserLookupParams{ScreenName: usernames[start:end]}
		idsList, _, err := t.client.Users.Lookup(params)
		if err != nil {
			if err.Error() == "twitter: 88 Rate limit exceeded" {
				log.Info("Twitter rate limit exceeded...waiting 15 minutes")
				time.Sleep(15 * time.Minute)
				continue
			} else {
				return nil, err
			}
		}

		for _, u := range idsList {
			ids[u.ScreenName] = u.ID
		}
	}
	return ids, nil
}

func (t *Twitter) getTweetExtendedText(tweet *twitter.Tweet) string {
	text := tweet.Text
	if tweet.ExtendedTweet != nil {
		text = tweet.ExtendedTweet.FullText
	}
	return text
}

func (t *Twitter) Start() {
	var err error

	log.Debug("Initialization of Twitter")
	config := oauth1.NewConfig(t.consumerKey, t.consumerSecret)
	token := oauth1.NewToken(t.accessToken, t.accessSecret)
	httpClient := config.Client(oauth1.NoContext, token)

	t.client = twitter.NewClient(httpClient)

	log.Debug("Setting demuxer")
	// Convenience Demux demultiplexed stream messages
	demux := twitter.NewSwitchDemux()
	demux.Tweet = func(tweet *twitter.Tweet) {
		if t.retweet && tweet.RetweetedStatus != nil {
			if tweet.RetweetedStatus != nil {
				retweet := tweet.RetweetedStatus
				txt := t.getTweetExtendedText(retweet)
				t.Propagate(data.NewMessageWithExtra(txt, map[string]string{
					"link":              fmt.Sprintf("https://twitter.com/%s/statuses/%d", tweet.User.ScreenName, tweet.ID),
					"language":          retweet.Lang,
					"username":          retweet.User.ScreenName,
					"quoted":            "false",
					"retweet":           "true",
					"original_username": tweet.User.ScreenName,
					"original_language": tweet.Lang,
					"original_status":   txt,
					"original_link":     fmt.Sprintf("https://twitter.com/%s/statuses/%d", retweet.User.ScreenName, retweet.ID),
				}))
			}
		} else if t.quoted && tweet.QuotedStatus != nil {
			if tweet.QuotedStatus != nil {
				quoted := tweet.QuotedStatus
				txt := t.getTweetExtendedText(tweet)
				t.Propagate(data.NewMessageWithExtra(txt, map[string]string{
					"link":            fmt.Sprintf("https://twitter.com/%s/statuses/%d", tweet.User.ScreenName, tweet.ID),
					"language":        tweet.Lang,
					"username":        tweet.User.ScreenName,
					"quoted":          "true",
					"retweet":         "false",
					"original_username": quoted.User.ScreenName,
					"original_status":   t.getTweetExtendedText(quoted),
					"original_language": quoted.Lang,
					"original_link": fmt.Sprintf("https://twitter.com/%s/statuses/%d", quoted.User.ScreenName, quoted.ID),
				}))
			}
		} else {
			text := t.getTweetExtendedText(tweet)
			t.Propagate(data.NewMessageWithExtra(text, map[string]string{
				"link":     fmt.Sprintf("https://twitter.com/%s/statuses/%d", tweet.User.ScreenName, tweet.ID),
				"language": tweet.Lang,
				"username": tweet.User.ScreenName,
				"quoted":   "false",
				"retweet":  "false",
			}))
		}
	}

	// we don't track Direct Messages or Events
	demux.DM = func(dm *twitter.DirectMessage) {
		log.Debug("[DIRECTMESSAGE] %s", dm.Text)
	}
	demux.Event = func(event *twitter.Event) {
		log.Debug("[EVENT] %s", event.Event)
	}

	// FILTER
	filterParams := &twitter.StreamFilterParams{
		StallWarnings: twitter.Bool(true),
	}
	if t.users != "" {
		users := strings.Split(t.users, ",")
		for i, k := range users {
			users[i] = strings.TrimSpace(k)
		}
		ids, err := t.getIdsFromUsernames(users)
		if err != nil {
			log.Error("getIdsFromUsernames: %s", err)
		} else {
			idss := []string{}
			for _, v := range ids {
				idss = append(idss, fmt.Sprintf("%d", v))
			}
			log.Debug("Setting users to follow: '%s' [ids: %s]", strings.Join(users, ","), strings.Join(idss, ","))
			filterParams.Follow = idss
		}
	}
	if t.keywords != "" {
		keywords := strings.Split(t.keywords, ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		log.Debug("Setting keywords: '%s'", strings.Join(keywords, ","))
		filterParams.Track = keywords
	}
	if t.languages != "" {
		languages := strings.Split(t.languages, ",")
		filterParams.Language = languages
	}
	t.stream, err = t.client.Streams.Filter(filterParams)
	if err != nil {
		log.Fatal("twitter stream filter error: %s", err)
	}

	// Receive messages until stopped or stream quits
	go demux.HandleChan(t.stream.Messages)
	t.isRunning = true
}

func (t *Twitter) Stop() {
	log.Debug("feeder '%s' stream stop", t.Name())
	t.stream.Stop()
	t.isRunning = false
}

// Auto factory adding
func init() {
	register("twitter", NewTwitterFeeder)
}

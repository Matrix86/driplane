package feeders

import (
	"context"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/Matrix86/driplane/data"
	"github.com/k0kubun/pp"

	//"github.com/dghubble/go-twitter/twitter"
	//"github.com/dghubble/oauth1"

	twitter "github.com/g8rswimmer/go-twitter/v2"

	"github.com/evilsocket/islazy/log"
)

// Twitter is a Feeder that feeds a pipeline with tweets
type Twitter struct {
	Base

	consumerKey    string
	consumerSecret string
	accessToken    string
	accessSecret   string
	bearerToken    string

	keywords      string
	twitterRule   string
	users         string
	languages     string
	retweet       bool
	quoted        bool
	stallWarnings bool
	closeChan     chan int

	client *twitter.Client
}

// NewTwitterFeeder is the registered method to instantiate a TwitterFeeder
// https://developer.twitter.com/en/docs/tweets/filter-realtime/guides/basic-stream-parameters
func NewTwitterFeeder(conf map[string]string) (Feeder, error) {
	t := &Twitter{
		stallWarnings: false,
		retweet:       true,
		quoted:        true,
		closeChan:     make(chan int),
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
	if val, ok := conf["twitter.bearerToken"]; ok {
		t.bearerToken = val
	}
	if val, ok := conf["twitter.keywords"]; ok {
		t.keywords = val
	}
	if val, ok := conf["twitter.rule"]; ok {
		t.twitterRule = val
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

//func (t *Twitter) getIdsFromUsernames(usernames []string) (map[string]int64, error) {
//	ids := make(map[string]int64)
//	// This API supports max 100 ids per request and 300 requests/15min window
//	for start, end := 0, 100; start < len(usernames); start, end = start+100, end+100 {
//		if end > len(usernames) {
//			end = len(usernames)
//		}
//		params := &twitter.UserLookupParams{ScreenName: usernames[start:end]}
//		idsList, _, err := t.client.Users.Lookup(params)
//		if err != nil {
//			if err.Error() == "twitter: 88 Rate limit exceeded" {
//				log.Info("Twitter rate limit exceeded...waiting 15 minutes")
//				time.Sleep(15 * time.Minute)
//				continue
//			} else {
//				return nil, err
//			}
//		}
//
//		for _, u := range idsList {
//			ids[u.ScreenName] = u.ID
//		}
//	}
//	return ids, nil
//}

//func (t *Twitter) getTweetExtendedText(tweet *twitter.Tweet) string {
//	text := tweet.Text
//	if tweet.ExtendedTweet != nil {
//		text = tweet.ExtendedTweet.FullText
//	}
//	return text
//}

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

func (t *Twitter) getUsernameFromIds(IDs []string) ([]string, error) {
	res, err := t.client.UserLookup(context.Background(), IDs, twitter.UserLookupOpts{UserFields: []twitter.UserField{twitter.UserFieldName}})
	if err != nil {
		return nil, fmt.Errorf("can't retrieve info for userID %s : %s", strings.Join(IDs, ","), err)
	}
	if len(res.Raw.Users) != 0 {
		names := make([]string, len(res.Raw.Users))
		for idx, n := range res.Raw.Users {
			names[idx] = n.Name
		}
		return names, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (t *Twitter) handleTweet(tm *twitter.TweetMessage) {
	for _, tweet := range tm.Raw.Tweets {
		ids := []string{tweet.AuthorID}
		if tweet.InReplyToUserID != "" {
			ids = append(ids, tweet.InReplyToUserID)
		}
		usernames, err := t.getUsernameFromIds(ids)
		if err != nil {
			log.Error("handleTweet couldn't handle the tweet: %s", err)
			return
		}

		//if tweet.ReferencedTweets != nil && len(tweet.ReferencedTweets) != 0 {
		//	if tweet.ReferencedTweets[0].Type == "quoted" {
		//
		//	}
		//}

		t.Propagate(data.NewMessageWithExtra(tweet.Text, map[string]interface{}{
			"link":          fmt.Sprintf("https://twitter.com/%s/statuses/%d", usernames[0], tweet.ID),
			"language":      tweet.Language,
			"username":      usernames[0],
			"source_client": tweet.Source,
			//"quoted":            "false",
			//"retweet":           "true",
			//"original_username": retweet.User.ScreenName,
			//"original_language": retweet.Lang,
			//"original_status":   txt,
			//"original_link":     fmt.Sprintf("https://twitter.com/%s/statuses/%d", retweet.User.ScreenName, retweet.ID),
		}))
	}
}

// Start propagates a message every time a new tweet is published
func (t *Twitter) Start() {
	var err error

	log.Debug("Initialization of Twitter client")
	t.client = &twitter.Client{
		Authorizer: authorize{
			Token: t.bearerToken,
		},
		Client: http.DefaultClient,
		Host:   "https://api.twitter.com",
	}

	// Setup rules
	log.Debug("Adding Twitter Rules")
	if t.keywords != "" {
		keywords := strings.Split(t.keywords, ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		log.Debug("Setting keywords: '%s'", strings.Join(keywords, ","))

		streamRule := twitter.TweetSearchStreamRule{
			Value: strings.Join(keywords, " OR "),
			Tag:   "keywords",
		}

		searchStreamRules, err := t.client.TweetSearchStreamAddRule(context.Background(), []twitter.TweetSearchStreamRule{streamRule}, false)
		if err != nil {
			log.Fatal("can't create a rule for Twitter stream: %s", err)
		}
		log.Debug("Rule created: %#v", searchStreamRules)
	}

	opts := twitter.TweetSearchStreamOpts{
		TweetFields: []twitter.TweetField{
			twitter.TweetFieldID,
			twitter.TweetFieldText,
			twitter.TweetFieldAuthorID,
			twitter.TweetFieldLanguage,
			twitter.TweetFieldSource,
			twitter.TweetFieldInReplyToUserID,
			twitter.TweetFieldReferencedTweets,
		},
	}
	tweetStream, err := t.client.TweetSearchStream(context.Background(), opts)
	if err != nil {
		log.Fatal("can't start Twitter stream: %s", err)
	}

	go func() {
		defer tweetStream.Close()
		for {
			select {
			case <-t.closeChan:
				return

			case tm := <-tweetStream.Tweets():
				//log.Info("Tweet received: %s", pp.Sprint(tm))
				t.handleTweet(tm)

			case sm := <-tweetStream.SystemMessages():
				log.Error("System message received: %#v", sm)

			case de := <-tweetStream.DisconnectionError():
				log.Error("Disconnection error: %#v", pp.Sprint(de))

			case strErr := <-tweetStream.Err():
				log.Error("Error: %#v", strErr)

			default:
			}
			if !tweetStream.Connection() {
				log.Error("Connection retry")
				t.Start()
				return
			}
		}
	}()

	//log.Debug("Setting demuxer")
	//// Convenience Demux demultiplexed stream messages
	//demux := twitter.NewSwitchDemux()
	//demux.Tweet = func(tweet *twitter.Tweet) {
	//	if t.retweet && tweet.RetweetedStatus != nil {
	//		if tweet.RetweetedStatus != nil {
	//			retweet := tweet.RetweetedStatus
	//			txt := t.getTweetExtendedText(retweet)
	//			t.Propagate(data.NewMessageWithExtra(txt, map[string]interface{}{
	//				"link":              fmt.Sprintf("https://twitter.com/%s/statuses/%d", tweet.User.ScreenName, tweet.ID),
	//				"language":          tweet.Lang,
	//				"username":          tweet.User.ScreenName,
	//				"quoted":            "false",
	//				"retweet":           "true",
	//				"original_username": retweet.User.ScreenName,
	//				"original_language": retweet.Lang,
	//				"original_status":   txt,
	//				"original_link":     fmt.Sprintf("https://twitter.com/%s/statuses/%d", retweet.User.ScreenName, retweet.ID),
	//			}))
	//		}
	//	} else if t.quoted && tweet.QuotedStatus != nil {
	//		if tweet.QuotedStatus != nil {
	//			quoted := tweet.QuotedStatus
	//			txt := t.getTweetExtendedText(tweet)
	//			t.Propagate(data.NewMessageWithExtra(txt, map[string]interface{}{
	//				"link":              fmt.Sprintf("https://twitter.com/%s/statuses/%d", tweet.User.ScreenName, tweet.ID),
	//				"language":          tweet.Lang,
	//				"username":          tweet.User.ScreenName,
	//				"status":            txt,
	//				"quoted":            "true",
	//				"retweet":           "false",
	//				"original_username": quoted.User.ScreenName,
	//				"original_status":   t.getTweetExtendedText(quoted),
	//				"original_language": quoted.Lang,
	//				"original_link":     fmt.Sprintf("https://twitter.com/%s/statuses/%d", quoted.User.ScreenName, quoted.ID),
	//			}))
	//		}
	//	} else {
	//		text := t.getTweetExtendedText(tweet)
	//		t.Propagate(data.NewMessageWithExtra(text, map[string]interface{}{
	//			"link":     fmt.Sprintf("https://twitter.com/%s/statuses/%d", tweet.User.ScreenName, tweet.ID),
	//			"language": tweet.Lang,
	//			"username": tweet.User.ScreenName,
	//			"status":   text,
	//			"quoted":   "false",
	//			"retweet":  "false",
	//		}))
	//	}
	//}
	//
	//// we don't track Direct Messages or Events
	//demux.DM = func(dm *twitter.DirectMessage) {
	//	log.Debug("[DIRECTMESSAGE] %s", dm.Text)
	//}
	//demux.Event = func(event *twitter.Event) {
	//	log.Debug("[EVENT] %s", event.Event)
	//}
	//
	//// FILTER
	//filterParams := &twitter.StreamFilterParams{
	//	StallWarnings: twitter.Bool(true),
	//}
	//if t.users != "" {
	//	users := strings.Split(t.users, ",")
	//	for i, k := range users {
	//		users[i] = strings.TrimSpace(k)
	//	}
	//	ids, err := t.getIdsFromUsernames(users)
	//	if err != nil {
	//		log.Error("getIdsFromUsernames: %s", err)
	//	} else {
	//		idss := []string{}
	//		for _, v := range ids {
	//			idss = append(idss, fmt.Sprintf("%d", v))
	//		}
	//		log.Debug("Setting users to follow: '%s' [ids: %s]", strings.Join(users, ","), strings.Join(idss, ","))
	//		filterParams.Follow = idss
	//	}
	//}
	//if t.keywords != "" {
	//	keywords := strings.Split(t.keywords, ",")
	//	for i, k := range keywords {
	//		keywords[i] = strings.TrimSpace(k)
	//	}
	//	log.Debug("Setting keywords: '%s'", strings.Join(keywords, ","))
	//	filterParams.Track = keywords
	//}
	//if t.languages != "" {
	//	languages := strings.Split(t.languages, ",")
	//	filterParams.Language = languages
	//}
	//t.stream, err = t.client.Streams.Filter(filterParams)
	//if err != nil {
	//	log.Fatal("twitter stream filter error: %s", err)
	//}
	//
	//// Receive messages until stopped or stream quits
	//go demux.HandleChan(t.stream.Messages)
	t.isRunning = true
	log.Warning("OLE")
}

// Stop handles the Feeder shutdown
func (t *Twitter) Stop() {
	// TODO: delete rules

	log.Debug("feeder '%s' stream stop", t.Name())
	close(t.closeChan)
	t.isRunning = false
}

// OnEvent is called when an event occurs
func (t *Twitter) OnEvent(event *data.Event) {
	if event.Type == "shutdown" && t.isRunning {
		t.Stop()
	}
}

// Auto factory adding
func init() {
	register("twitter", NewTwitterFeeder)
}

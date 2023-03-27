package feeders

import (
	"context"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/Matrix86/driplane/data"

	"github.com/evilsocket/islazy/log"
	twitter "github.com/g8rswimmer/go-twitter/v2"
	"github.com/k0kubun/pp"
)

// Twitter is a Feeder that feeds a pipeline with tweets
type Twitter struct {
	Base

	bearerToken string

	keywords      string
	twitterRule   string
	users         string
	languages     map[string]bool
	retweet       bool
	quoted        bool
	stallWarnings bool
	closeChan     chan int

	retry int

	rules  []string
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
		rules:         make([]string, 0),
		languages:     make(map[string]bool),
		retry:         10,
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
		langs := strings.Split(val, ",")
		for _, k := range langs {
			t.languages[strings.TrimSpace(k)] = true
		}
	}
	if val, ok := conf["twitter.disable_retweet"]; ok && val == "true" {
		t.retweet = false
	}
	if val, ok := conf["twitter.disable_quoted"]; ok && val == "true" {
		t.quoted = false
	}

	return t, nil
}

type authorize struct {
	Token string
}

func (a authorize) Add(req *http.Request) {
	req.Header.Add("Authorization", fmt.Sprintf("Bearer %s", a.Token))
}

func (t *Twitter) getUsernameFromIds(IDs []string) ([]string, error) {
	res, err := t.client.UserLookup(context.Background(), IDs, twitter.UserLookupOpts{UserFields: []twitter.UserField{twitter.UserFieldUserName}})
	if err != nil {
		return nil, fmt.Errorf("can't retrieve info for userID %s : %s", strings.Join(IDs, ","), err)
	}
	if len(res.Raw.Users) != 0 {
		names := make([]string, len(res.Raw.Users))
		for idx, n := range res.Raw.Users {
			names[idx] = n.UserName
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

		// skip if the filter by language is active
		if len(t.languages) > 0 {
			if _, ok := t.languages[tweet.Language]; !ok {
				continue
			}
		}

		msg := data.NewMessageWithExtra(tweet.Text, map[string]interface{}{
			"link":          fmt.Sprintf("https://twitter.com/%s/statuses/%s", usernames[0], tweet.ID),
			"language":      tweet.Language,
			"username":      usernames[0],
			"source_client": tweet.Source,
			"quoted":        "false",
			"retweet":       "false",
			"response":      "false",
		})

		isQuote, isRetweet := false, false
		extra := make(map[string]string)

		if tweet.ReferencedTweets != nil && len(tweet.ReferencedTweets) != 0 {
			for idx := range tweet.ReferencedTweets {
				if tweet.ReferencedTweets[idx].Type == "quoted" {
					extra["quoted"] = "true"
					extra["original_link"] = fmt.Sprintf("https://twitter.com/dummy/statuses/%s", tweet.ReferencedTweets[idx].ID)
					isQuote = true
				} else if tweet.ReferencedTweets[idx].Type == "replied_to" {
					extra["response"] = "true"
					extra["original_link"] = fmt.Sprintf("https://twitter.com/dummy/statuses/%s", tweet.ReferencedTweets[idx].ID)
				} else if tweet.ReferencedTweets[idx].Type == "retweeted" {
					extra["retweet"] = "true"
					extra["original_link"] = fmt.Sprintf("https://twitter.com/dummy/statuses/%s", tweet.ReferencedTweets[idx].ID)
					isRetweet = true
				}
			}
		}
		if tweet.InReplyToUserID != "" {
			extra["reply_for_user"] = tweet.InReplyToUserID
		}
		for k, v := range extra {
			msg.SetExtra(k, v)
		}

		// do not share if retweets and quotes are disabled
		if !t.retweet && isRetweet {
			continue
		}
		if !t.quoted && isQuote {
			continue
		}

		t.Propagate(msg)
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
	rules := []twitter.TweetSearchStreamRule{}
	if t.keywords != "" {
		keywords := strings.Split(t.keywords, ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
		}
		log.Debug("Setting keywords rule: '%s'", strings.Join(keywords, ","))

		keywordRule := twitter.TweetSearchStreamRule{
			Value: strings.Join(keywords, " OR "),
			Tag:   "keywords_rule",
		}

		t.rules = append(t.rules, strings.Join(keywords, " OR "))
		rules = append(rules, keywordRule)
	}

	if t.users != "" {
		users := strings.Split(t.users, ",")
		for i, k := range users {
			users[i] = fmt.Sprintf("@%s", strings.TrimSpace(k))
		}
		log.Debug("Setting users rule: '%s'", strings.Join(users, ","))

		userRule := twitter.TweetSearchStreamRule{
			Value: strings.Join(users, " OR "),
			Tag:   "users_rule",
		}

		t.rules = append(t.rules, strings.Join(users, " OR "))
		rules = append(rules, userRule)
	}

	if t.twitterRule != "" {
		log.Debug("Setting custom rule: '%s'", t.twitterRule)

		customRule := twitter.TweetSearchStreamRule{
			Value: t.twitterRule,
			Tag:   "custom_rule",
		}

		t.rules = append(t.rules, t.twitterRule)
		rules = append(rules, customRule)
	}

	if len(rules) > 0 {
		log.Debug("TwitterFeeder: adding %d rules", len(t.rules))
		searchStreamRules, err := t.client.TweetSearchStreamAddRule(context.Background(), rules, false)
		if err != nil {
			log.Error("TwitterFeeder: can't create a rule for Twitter stream: %s", err)
		} else {
			if searchStreamRules.Errors != nil && len(searchStreamRules.Errors) > 0 {
				for _, e := range searchStreamRules.Errors {
					log.Error("TwitterFeeder: error adding rule '%s': %s", e.Value, e.Title)
				}
			}

			if searchStreamRules.Rules != nil {
				log.Debug("TwitterFeeder: %d rules have been created", len(searchStreamRules.Rules))
			}
		}
	} else {
		log.Info("TwitterFeeder: no rule found, waiting for rules specified externally")
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
				t.handleTweet(tm)

			case sm := <-tweetStream.SystemMessages():
				log.Info("TwitterFeeder: System message received: %#v", sm)

			case de := <-tweetStream.DisconnectionError():
				log.Error("TwitterFeeder: disconnection error: %#v", pp.Sprint(de))

			case strErr := <-tweetStream.Err():
				log.Error("TwitterFeeder: error on the stream: %#v", strErr)

			default:
			}
			if !tweetStream.Connection() {
				log.Error("Connection retry")
				if t.retry > 0 {
					t.retry--
					time.Sleep(10 * time.Second)
					t.Start()
				} else {
					log.Error("Connection retries finished...need to restart the app")
				}
				return
			} else {
				// resetting the retries
				t.retry = 10
			}
		}
	}()
	t.isRunning = true
}

// Stop handles the Feeder shutdown
func (t *Twitter) Stop() {
	if len(t.rules) > 0 {
		log.Debug("TwitterFeeder: removing %d rules", len(t.rules))
		_, err := t.client.TweetSearchStreamDeleteRuleByValue(context.Background(), t.rules, false)
		if err != nil {
			log.Error("twitterFeeder: couldn't delete the rules: %s", err)
		}
	}

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

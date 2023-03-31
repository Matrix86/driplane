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
)

// Twitter is a Feeder that feeds a pipeline with tweets
type Twitter struct {
	Base

	bearerToken string

	keywords      string
	twitterRules  map[string]string
	users         string
	languages     map[string]bool
	retweet       bool
	quoted        bool
	stallWarnings bool
	closeChan     chan int

	retry int

	ruleIDs []twitter.TweetSearchStreamRuleID
	client  *twitter.Client
}

// NewTwitterFeeder is the registered method to instantiate a TwitterFeeder
// https://developer.twitter.com/en/docs/tweets/filter-realtime/guides/basic-stream-parameters
func NewTwitterFeeder(conf map[string]string) (Feeder, error) {
	t := &Twitter{
		stallWarnings: false,
		retweet:       true,
		quoted:        true,
		closeChan:     make(chan int),
		ruleIDs:       make([]twitter.TweetSearchStreamRuleID, 0),
		languages:     make(map[string]bool),
		twitterRules:  make(map[string]string),
		retry:         10,
	}

	if val, ok := conf["twitter.bearerToken"]; ok {
		t.bearerToken = val
	}
	if val, ok := conf["twitter.keywords"]; ok {
		t.keywords = val
	}
	if val, ok := conf["twitter.rules"]; ok {
		// rules are separated by | char and in the form tag:rule
		rules := strings.Split(val, "|")
		for _, rule := range rules {
			if tag, value, found := strings.Cut(rule, ":"); found {
				t.twitterRules[tag] = value
			}
		}
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

func (t *Twitter) handleTweet(tm *twitter.TweetMessage) {
	var author *twitter.UserObj
	var ok bool
	for _, tweet := range tm.Raw.Tweets {
		// skip if the filter by language is active
		if len(t.languages) > 0 {
			if _, ok := t.languages[tweet.Language]; !ok {
				continue
			}
		}

		// convert user and tweet array to map
		users := tm.Raw.Includes.UsersByID()
		tweets := tm.Raw.Includes.TweetsByID()

		if author, ok = users[tweet.AuthorID]; !ok {
			log.Error("couldn't find user by ID=%s in the includes", tweet.AuthorID)
			continue
		}

		matching := []string{}
		for _, m := range tm.Raw.MatchingRules {
			matching = append(matching, m.Tag)
		}

		msg := data.NewMessageWithExtra(tweet.Text, map[string]interface{}{
			"link":          fmt.Sprintf("https://twitter.com/%s/statuses/%s", author.UserName, tweet.ID),
			"language":      tweet.Language,
			"username":      author.UserName,
			"name":          author.Name,
			"author_id":     author.ID,
			"source_client": tweet.Source,
			"quoted":        "false",
			"retweet":       "false",
			"response":      "false",
			"matched_rules": strings.Join(matching, ","),
		})

		isQuote, isRetweet := false, false
		extra := make(map[string]string)

		if tweet.ReferencedTweets != nil && len(tweet.ReferencedTweets) != 0 {
			var originalAuthor *twitter.UserObj
			var originalTweet *twitter.TweetObj
			for idx := range tweet.ReferencedTweets {
				// getting the original tweet
				if originalTweet, ok = tweets[tweet.ReferencedTweets[idx].ID]; !ok {
					log.Error("couldn't find tweet by ID=%s in the includes", tweet.ReferencedTweets[idx].ID)
					continue
				}

				// getting the original author
				if originalAuthor, ok = users[originalTweet.AuthorID]; !ok {
					log.Error("couldn't find original user by ID=%s in the includes", originalTweet.AuthorID)
					continue
				}

				extra["original_link"] = fmt.Sprintf("https://twitter.com/%s/statuses/%s", originalAuthor.UserName, tweet.ReferencedTweets[idx].ID)
				extra["original_username"] = originalAuthor.UserName
				extra["original_name"] = originalAuthor.Name
				extra["original_text"] = originalTweet.Text
				extra["original_userid"] = originalAuthor.ID

				if tweet.ReferencedTweets[idx].Type == "quoted" {
					extra["quoted"] = "true"
					isQuote = true
				} else if tweet.ReferencedTweets[idx].Type == "replied_to" {
					extra["response"] = "true"
				} else if tweet.ReferencedTweets[idx].Type == "retweeted" {
					extra["retweet"] = "true"
					// if it is a retweet the text could be truncated and a "RT" word is added as prefix
					msg.SetMessage(originalTweet.Text)
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

		rules = append(rules, userRule)
	}

	if len(t.twitterRules) > 0 {
		log.Debug("Setting custom rules:")
		for k, v := range t.twitterRules {
			log.Debug("rule %s = %s", k, v)
			rules = append(rules, twitter.TweetSearchStreamRule{
				Value: v,
				Tag:   k,
			})
		}
	}

	if len(rules) > 0 {
		log.Debug("TwitterFeeder: adding rules")
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

			for _, r := range searchStreamRules.Rules {
				t.ruleIDs = append(t.ruleIDs, r.ID)
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
		Expansions: []twitter.Expansion{
			twitter.ExpansionAuthorID,
			twitter.ExpansionInReplyToUserID,
			twitter.ExpansionReferencedTweetsID,
			twitter.ExpansionEntitiesMentionsUserName,
		},
		UserFields: []twitter.UserField{
			twitter.UserFieldUserName,
			twitter.UserFieldName,
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
				if len(de.Connections) > 0 {
					for _, e := range de.Connections {
						log.Debug("TwitterFeeder: connection error: Title='%s' ConnectionIssue='%s' Detail='%s' Type='%s'", e.Title, e.ConnectionIssue, e.Detail, e.Type)
					}
				} else if len(de.Disconnections) > 0 {
					for _, e := range de.Disconnections {
						log.Error("TwitterFeeder: disconnection error: Title='%s' DisconnectType='%s' Detail='%s' Type='%s'", e.Title, e.DisconnectType, e.Detail, e.Type)
					}
				}

			case strErr := <-tweetStream.Err():
				log.Error("TwitterFeeder: error on the stream: %s", strErr)

			default:
			}
			if !tweetStream.Connection() {
				log.Error("TwitterFeeder: disconnection detected")
				for ; t.retry > 0 && !tweetStream.Connection(); t.retry-- {
					waitTime := time.Duration(2*(10-t.retry)) * time.Second
					log.Error("TwitterFeeder: connection retry...waiting %f", waitTime.Seconds())
					time.Sleep(waitTime)
					tweetStream, err = t.client.TweetSearchStream(context.Background(), opts)
					if err != nil {
						log.Error("TwitterFeeder: stream connection error: %s", err)
					} else {
						t.retry = 10
					}
				}
				if t.retry == 0 {
					log.Fatal("TwitterFeeder: too many connection retries. Closing")
				}
				return
			}
		}
	}()
	t.isRunning = true
}

// Stop handles the Feeder shutdown
func (t *Twitter) Stop() {
	if len(t.ruleIDs) > 0 {
		log.Debug("TwitterFeeder: removing %d rules", len(t.ruleIDs))
		_, err := t.client.TweetSearchStreamDeleteRuleByID(context.Background(), t.ruleIDs, false)
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

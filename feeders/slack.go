package feeders

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/evilsocket/islazy/tui"
	ll "log"
	"net/http"
	"os"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"github.com/localtunnel/go-localtunnel"
	"github.com/slack-go/slack"
	"github.com/slack-go/slack/slackevents"
)

// Slack is a Feeder that get events from Slack
type Slack struct {
	Base

	token             string
	verificationToken string
	addr              string
	enableLocalTunnel bool
	ltSubdomain       string
	port              int
	events            map[string]bool
	ignoreBot         bool

	Team   string
	User   string
	TeamID string
	UserID string
	BotID  string

	client *slack.Client
	server *http.Server
	wg     *sync.WaitGroup
	tunnel *localtunnel.Listener
}

// NewSlackFeeder is the registered method to instantiate a SlackFeeder
func NewSlackFeeder(conf map[string]string) (Feeder, error) {
	s := &Slack{
		addr:              ":3000",
		port:              3000,
		enableLocalTunnel: false,
		ignoreBot:         true,
		events: map[string]bool{
			"app_mention":             true,
			"app_home_opened":         true,
			"app_uninstalled":         true,
			"grid_migration_finished": true,
			"grid_migration_started":  true,
			"link_shared":             true,
			"message":                 true,
			"member_joined_channel":   true,
			"member_left_channel":     true,
			"pin_added":               true,
			"pin_removed":             true,
			"reaction_added":          true,
			"reaction_removed":        true,
			"tokens_revoked":          true,
			"file_shared":             true,
		},
		wg: &sync.WaitGroup{},
	}

	if val, ok := conf["slack.token"]; ok {
		s.token = val
	}

	if val, ok := conf["slack.verification_token"]; ok {
		s.verificationToken = val
	}
	if val, ok := conf["slack.addr"]; ok {
		s.addr = val
	}
	if val, ok := conf["slack.lt_enable"]; ok && val == "true" {
		s.enableLocalTunnel = true
	}
	if val, ok := conf["slack.lt_subdomain"]; ok {
		s.ltSubdomain = val
	}
	if val, ok := conf["slack.events"]; ok {
		keywords := strings.Split(val, ",")
		for i, k := range keywords {
			keywords[i] = strings.TrimSpace(k)
			s.events[strings.TrimSpace(k)] = true
		}
		log.Debug("Slack: listen on events: '%s'", strings.Join(keywords, ","))
	}
	if val, ok := conf["ignore_bot"]; ok && val == "false" {
		s.ignoreBot = false
	}

	// Get port from addr and convert it to int
	parts := strings.Split(s.addr, ":")
	i, err := strconv.Atoi(parts[1])
	if err != nil {
		return nil, err
	}
	s.port = i

	s.client = slack.New(
		s.token,
		slack.OptionDebug(false),
		slack.OptionLog(ll.New(os.Stdout, "slackfeeder: ", ll.Lshortfile|ll.LstdFlags)),
	)

	return s, nil
}

// GetClient return the slack.client pointer
func (s *Slack) GetClient() *slack.Client {
	return s.client
}

func (s *Slack) propagateFiles(msg string, files []slackevents.File) {
	for _, file := range files {
		extraf := make(map[string]interface{})
		extraf["type"] = "file_shared"
		extraf["slackfeeder.token"] = s.token
		fr := reflect.ValueOf(file)
		for x := 0; x < fr.NumField(); x++ {
			if fr.Field(x).CanInterface() {
				switch sv := fr.Field(x).Interface().(type) {
				case string:
					extraf[strings.ToLower(fr.Type().Field(x).Name)] = sv
				case int:
					extraf[strings.ToLower(fr.Type().Field(x).Name)] = fmt.Sprintf("%d", sv)
				case bool:
					extraf[strings.ToLower(fr.Type().Field(x).Name)] = strconv.FormatBool(sv)
				}
			}
		}
		s.Propagate(data.NewMessageWithExtra(msg, extraf))
	}
}

func (s *Slack) startEventsEndpoint() {
	s.wg.Add(1)
	s.server = &http.Server{Addr: s.addr}

	http.HandleFunc("/events", func(w http.ResponseWriter, r *http.Request) {
		buf := new(bytes.Buffer)
		buf.ReadFrom(r.Body)
		body := buf.String()
		eventsAPIEvent, e := slackevents.ParseEvent(json.RawMessage(body), slackevents.OptionVerifyToken(&slackevents.TokenComparator{VerificationToken: s.verificationToken}))
		if e != nil {
			log.Error("VerificationToken error: %s", e)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}

		if eventsAPIEvent.Type == slackevents.URLVerification {
			var r *slackevents.ChallengeResponse
			err := json.Unmarshal([]byte(body), &r)
			if err != nil {
				w.WriteHeader(http.StatusInternalServerError)
				return
			}
			w.Header().Set("Content-Type", "text")
			w.Write([]byte(r.Challenge))
		}

		if eventsAPIEvent.Type == slackevents.CallbackEvent {
			innerEvent := eventsAPIEvent.InnerEvent

			//pp.Println(innerEvent.Data)
			log.Debug("Slack event: %s :\n%#v", innerEvent.Type, innerEvent.Data)

			if _, ok := s.events[innerEvent.Type]; ok && innerEvent.Type != "file_shared" {
				propagate := true
				txt := innerEvent.Type
				extra := make(map[string]interface{})
				extra["slackfeeder.token"] = s.token

				switch ev := innerEvent.Data.(type) {
				case *slackevents.AppMentionEvent:
					if (s.ignoreBot && ev.BotID != "") || ev.BotID == s.BotID {
						return
					}
					v := reflect.ValueOf(*ev)
					txt = ev.Text
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							if str, ok := v.Field(i).Interface().(string); ok {
								extra[strings.ToLower(v.Type().Field(i).Name)] = str
							}
						}
					}
				case *slackevents.MessageEvent:
					if (s.ignoreBot && ev.BotID != "") || ev.BotID == s.BotID {
						return
					}
					v := reflect.ValueOf(*ev)
					txt = ev.Text
					// propagate multiple messages, one for each shared file
					if _, ok := s.events["file_shared"]; ok {
						s.propagateFiles(txt, ev.Files)
					}
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							if str, ok := v.Field(i).Interface().(string); ok {
								extra[strings.ToLower(v.Type().Field(i).Name)] = str
							}
						}
					}
				case *slackevents.MemberJoinedChannelEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							if str, ok := v.Field(i).Interface().(string); ok {
								extra[strings.ToLower(v.Type().Field(i).Name)] = str
							}
						}
					}
				case *slackevents.MemberLeftChannelEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							if str, ok := v.Field(i).Interface().(string); ok {
								extra[strings.ToLower(v.Type().Field(i).Name)] = str
							}
						}
					}
				case *slackevents.AppHomeOpenedEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							if str, ok := v.Field(i).Interface().(string); ok {
								extra[strings.ToLower(v.Type().Field(i).Name)] = str
							}
						}
					}
				case *slackevents.AppUninstalledEvent:
					extra["type"] = ev.Type
				case *slackevents.GridMigrationFinishedEvent:
					extra["type"] = ev.Type
					extra["enterprise_id"] = ev.EnterpriseID
				case *slackevents.GridMigrationStartedEvent:
					extra["type"] = ev.Type
					extra["enterprise_id"] = ev.EnterpriseID
				case *slackevents.LinkSharedEvent:
					propagate = false
					for _, link := range ev.Links {
						ex := map[string]interface{}{}
						ex["user"] = ev.User
						ex["timestamp"] = ev.TimeStamp
						ex["threadtimestamp"] = ev.ThreadTimeStamp
						ex["domain"] = link.Domain
						ex["link"] = link.URL
						ex["type"] = "link_shared"
						ex["slackfeeder.token"] = s.token
						s.Propagate(data.NewMessageWithExtra(link.URL, ex))
					}
				case *slackevents.ReactionAddedEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							extra[strings.ToLower(v.Type().Field(i).Name)] = v.Field(i).Interface()
						}
					}
				case *slackevents.ReactionRemovedEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							extra[strings.ToLower(v.Type().Field(i).Name)] = v.Field(i).Interface()
						}
					}
				case *slackevents.PinAddedEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							extra[strings.ToLower(v.Type().Field(i).Name)] = v.Field(i).Interface()
						}
					}
				case *slackevents.PinRemovedEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							extra[strings.ToLower(v.Type().Field(i).Name)] = v.Field(i).Interface()
						}
					}
				case *slackevents.TokensRevokedEvent:
					v := reflect.ValueOf(*ev)
					for i := 0; i < v.NumField(); i++ {
						if v.Field(i).CanInterface() {
							if str, ok := v.Field(i).Interface().(string); ok {
								extra[strings.ToLower(v.Type().Field(i).Name)] = str
							}
						}
					}
				}

				if propagate {
					s.Propagate(data.NewMessageWithExtra(txt, extra))
				}
			}
		}
	})

	go func() {
		defer s.wg.Done() // let main know we are done cleaning up
		if s.enableLocalTunnel {
			opts := localtunnel.Options{
				Subdomain: s.ltSubdomain,
			}
			tunnel, err := localtunnel.Listen(opts)
			if err != nil {
				log.Error("LocalTunnel: %s", err)
				return
			}
			s.tunnel = tunnel
			log.Info("Slack feeder tunnel started. Your URL is: %s", s.tunnel.URL())
		}

		log.Info("Slack endpoint server listening on: %s", s.addr)
		if err := s.server.Serve(s.tunnel); err != http.ErrServerClosed {
			log.Fatal("Slack::ListenAndServe(): %s", err)
		}
	}()
}

func (s *Slack) getBotInfo() {
	// Get Information
	r, err := s.client.AuthTest()
	if err != nil {
		log.Fatal("Slack: AuthTest failed: %s", err)
	}
	s.Team = r.Team
	s.User = r.User
	s.TeamID = r.TeamID
	s.UserID = r.UserID
	s.BotID = r.BotID

	log.Info("Slack: bot connected as %s in the %s team", tui.Bold(s.User), tui.Bold(s.Team))
}

func (s *Slack) stopEventsEndpoint() {
	if err := s.server.Shutdown(context.Background()); err != nil {
		log.Fatal("Slack::Shutdown(): %s", err)
	}
	s.wg.Wait()
}

// Start propagates a message every time a new tweet is published
func (s *Slack) Start() {
	log.Debug("Initialization of Slack")
	s.getBotInfo()
	s.startEventsEndpoint()
	s.isRunning = true
}

// Stop handles the Feeder shutdown
func (s *Slack) Stop() {
	log.Debug("feeder '%s' stream stop", s.Name())
	s.stopEventsEndpoint()
	s.isRunning = false
}

// Auto factory adding
func init() {
	register("slack", NewSlackFeeder)
}

package feeders

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"github.com/gotd/td/telegram"
	"github.com/gotd/td/telegram/auth"
	"github.com/gotd/td/telegram/updates"
	"github.com/gotd/td/tg"
)

const dateLayout = "2006-01-02_15-04-05"

type Telegram struct {
	Base

	phoneNumber   string
	appID         int
	appHash       string
	sessionFolder string
	context       context.Context
	cancelContext context.CancelFunc

	userMap    map[int64]*tg.User
	channelMap map[int64]*tg.Channel
	chatMap    map[int64]*tg.Chat

	api *tg.Client
}

// NewTelegramFeeder is the registered method to instantiate a TelegramFeeder
func NewTelegramFeeder(conf map[string]string) (Feeder, error) {
	context, cancel := context.WithCancel(context.Background())
	t := &Telegram{
		context:       context,
		cancelContext: cancel,
		userMap:       make(map[int64]*tg.User),
		channelMap:    make(map[int64]*tg.Channel),
		chatMap:       make(map[int64]*tg.Chat),
	}

	if val, ok := conf["telegram.phone_number"]; ok {
		t.phoneNumber = val
	}
	if val, ok := conf["telegram.app_id"]; ok {
		i, err := strconv.Atoi(val)
		if err != nil {
			return nil, err
		}
		t.appID = i
	}
	if val, ok := conf["telegram.app_hash"]; ok {
		t.appHash = val
	}
	if val, ok := conf["telegram.session_folder"]; ok {
		t.sessionFolder = val
	}

	if t.phoneNumber == "" {
		return nil, fmt.Errorf("the param 'phone_number' is requested by telegram feeder")
	}
	if t.appHash == "" {
		return nil, fmt.Errorf("the param 'app_hash' is requested by telegram feeder")
	}
	if t.appID == 0 {
		return nil, fmt.Errorf("the param 'app_id' is requested by telegram feeder")
	}
	return t, nil
}

func (t *Telegram) updateMaps(e tg.Entities) {
	for userID, user := range e.Users {
		t.userMap[userID] = user
	}

	for channelID, channel := range e.Channels {
		t.channelMap[channelID] = channel
	}

	for chatID, chat := range e.Chats {
		t.chatMap[chatID] = chat
	}
}

func (t *Telegram) getUserByID(id int64) (*tg.User, error) {
	if user, ok := t.userMap[id]; ok {
		return user, nil
	}
	return nil, fmt.Errorf("user not found")
}

func (t *Telegram) getChannelByID(id int64) (*tg.Channel, error) {
	if channel, ok := t.channelMap[id]; ok {
		return channel, nil
	}
	return nil, fmt.Errorf("channel not found")
}

func (t *Telegram) getChatByID(id int64) (*tg.Chat, error) {
	if chat, ok := t.chatMap[id]; ok {
		return chat, nil
	}
	return nil, fmt.Errorf("chat not found")
}

func (t *Telegram) fillExtraFromUser(extra map[string]interface{}, sender *tg.User) {
	extra["user_bot"] = strconv.FormatBool(sender.GetBot())
	extra["user_isclosefriend"] = strconv.FormatBool(sender.GetCloseFriend())
	extra["user_iscontact"] = strconv.FormatBool(sender.GetContact())
	extra["user_isdeleted"] = strconv.FormatBool(sender.GetDeleted())
	extra["user_isfake"] = strconv.FormatBool(sender.GetFake())
	extra["user_id"] = fmt.Sprintf("%d", sender.GetID())
	extra["user_mutualcontact"] = strconv.FormatBool(sender.GetMutualContact())
	extra["user_premium"] = strconv.FormatBool(sender.GetPremium())
	extra["user_verified"] = strconv.FormatBool(sender.GetVerified())
	if h, ok := sender.GetAccessHash(); ok {
		extra["user_accesshash"] = h
	}
	if h, ok := sender.GetFirstName(); ok {
		extra["user_firstname"] = h
	}
	if h, ok := sender.GetUsername(); ok {
		extra["user_username"] = h
	}
	if h, ok := sender.GetLangCode(); ok {
		extra["user_language"] = h
	}
	if h, ok := sender.GetLastName(); ok {
		extra["user_lastname"] = h
	}
	if h, ok := sender.GetPhone(); ok {
		extra["user_phone"] = h
	}
}

func (t *Telegram) fillExtraFromChat(extra map[string]interface{}, chat *tg.Chat) {
	extra["chat_callactive"] = strconv.FormatBool(chat.GetCallActive())
	extra["chat_creator"] = strconv.FormatBool(chat.GetCreator())
	extra["chat_deactivated"] = strconv.FormatBool(chat.GetDeactivated())
	extra["chat_id"] = fmt.Sprintf("%d", chat.GetID())
	extra["chat_partecipantscount"] = fmt.Sprintf("%d", chat.GetParticipantsCount())
	extra["chat_title"] = chat.GetTitle()
	extra["chat_version"] = fmt.Sprintf("%d", chat.GetVersion())
}

func (t *Telegram) fillExtraFromChannel(extra map[string]interface{}, channel *tg.Channel) {
	extra["chan_broadcast"] = strconv.FormatBool(channel.GetBroadcast())
	extra["chan_callactive"] = strconv.FormatBool(channel.GetCallActive())
	extra["chan_creator"] = strconv.FormatBool(channel.GetCreator())
	extra["chan_fake"] = strconv.FormatBool(channel.GetFake())
	extra["chan_forum"] = strconv.FormatBool(channel.GetForum())
	extra["chan_gigagroup"] = strconv.FormatBool(channel.GetGigagroup())
	extra["chan_hasgeo"] = strconv.FormatBool(channel.GetHasGeo())
	extra["chan_haslink"] = strconv.FormatBool(channel.GetHasLink())
	extra["chan_id"] = fmt.Sprintf("%d", channel.GetID())
	extra["chan_hasJoinRequest"] = strconv.FormatBool(channel.GetJoinRequest())
	extra["chan_ismegagroup"] = strconv.FormatBool(channel.GetMegagroup())
	extra["chan_isrestricted"] = strconv.FormatBool(channel.GetRestricted())
	extra["chan_title"] = channel.GetTitle()
	extra["chan_verified"] = strconv.FormatBool(channel.GetVerified())
	if n, ok := channel.GetParticipantsCount(); ok {
		extra["chan_partecipantscount"] = fmt.Sprintf("%d", n)
	}
	if u, ok := channel.GetUsername(); ok {
		extra["chan_username"] = u
	}
}

func (t *Telegram) retrieveEntities(pts int, e *tg.Entities) {
	diff, err := t.api.UpdatesGetDifference(context.Background(), &tg.UpdatesGetDifferenceRequest{
		Pts:  pts - 1,
		Date: int(time.Now().Unix()),
	})
	// Silently add catched entities to *tg.Entities
	if err == nil {
		if value, ok := diff.(*tg.UpdatesDifference); ok {
			for _, vu := range value.Chats {
				switch chat := vu.(type) {
				case *tg.Chat:
					e.Chats[chat.ID] = chat
				case *tg.Channel:
					e.Channels[chat.ID] = chat
				}
			}
			for _, vu := range value.Users {
				user, ok := vu.AsNotEmpty()
				if !ok {
					continue
				}
				e.Users[user.ID] = user
			}
		}
	}

	t.updateMaps(*e)
}

func (f *Telegram) getDocFilename(doc *tg.Document) (string, string) {
	var filename, ext string
	for _, attr := range doc.Attributes {
		switch v := attr.(type) {
		case *tg.DocumentAttributeImageSize:
			switch doc.MimeType {
			case "image/png":
				ext = ".png"
			case "image/webp":
				ext = ".webp"
			case "image/tiff":
				ext = ".tif"
			default:
				ext = ".jpg"
			}
		case *tg.DocumentAttributeAnimated:
			ext = ".gif"
		case *tg.DocumentAttributeSticker:
			ext = ".webp"
		case *tg.DocumentAttributeVideo:
			switch doc.MimeType {
			case "video/mpeg":
				ext = ".mpeg"
			case "video/webm":
				ext = ".webm"
			case "video/ogg":
				ext = ".ogg"
			default:
				ext = ".mp4"
			}
		case *tg.DocumentAttributeAudio:
			switch doc.MimeType {
			case "audio/webm":
				ext = ".webm"
			case "audio/aac":
				ext = ".aac"
			case "audio/ogg":
				ext = ".ogg"
			default:
				ext = ".mp3"
			}
		case *tg.DocumentAttributeFilename:
			filename = v.FileName
		}
	}

	if filename == "" {
		filename = fmt.Sprintf(
			"doc%d_%s%s", doc.GetID(),
			time.Unix(int64(doc.Date), 0).Format(dateLayout),
			ext,
		)
	}

	if ext == "" {
		ext = filepath.Ext(filename)
	}

	return filename, ext
}

func (t *Telegram) getMediaInfo(mediaClass tg.MessageMediaClass) (string, string, int64) {
	switch media := mediaClass.(type) {
	case *tg.MessageMediaDocument:
		if docClass, ok := media.GetDocument(); ok {
			if d, ok := docClass.AsNotEmpty(); ok {
				name, ext := t.getDocFilename(d)
				size := d.GetSize()
				return name, ext, size
			}
		}

	case *tg.MessageMediaPhoto:
		if photoClass, ok := media.GetPhoto(); ok {
			if photo, ok := photoClass.AsNotEmpty(); ok {
				maxH := 0
				maxW := 0
				size := 0

				// Get the biggest picture
				for _, g := range photo.Sizes {
					if sz, ok := g.(*tg.PhotoSize); ok {
						if sz.GetH() > maxH || sz.GetW() > maxW {
							maxH = sz.GetH()
							maxW = sz.GetW()
							size = sz.GetSize()
						}
					}
				}

				return fmt.Sprintf(
					"photo%d_%s.jpg", photo.GetID(),
					time.Unix(int64(photo.Date), 0).Format(dateLayout),
				), "jpg", int64(size)
			}
		}
	}

	return "", "", -1
}

func (t *Telegram) onMessage(ctx context.Context, e tg.Entities, pts int, msg *tg.Message, edit bool, media tg.MessageMediaClass) error {
	var err error
	var chat *tg.Chat
	var sender *tg.User
	extra := make(map[string]interface{})

	extra["type"] = "chat_message"
	extra["msg_edited"] = strconv.FormatBool(edit)
	extra["msg_hasmedia"] = false
	if media != nil {
		extra["msg_hasmedia"] = true
		extra["_msg_media"] = media
		fname, ext, size := t.getMediaInfo(media)
		extra["msg_medianame"] = fname
		extra["msg_mediaext"] = ext
		extra["msg_mediasize"] = size
	}
	extra["_telegram_api"] = t.api

	tm := time.Unix(int64(msg.Date), 0)
	extra["msg_timestamp"] = msg.Date
	extra["msg_date"] = tm.Format(time.DateOnly)
	extra["msg_time"] = tm.Format(time.TimeOnly)

	t.updateMaps(e)

	if chatPeer, ok := msg.GetPeerID().(*tg.PeerChat); ok {
		chat, err = t.getChatByID(chatPeer.GetChatID())
		if err != nil {
			t.retrieveEntities(pts, &e)
			chat, _ = t.getChatByID(chatPeer.GetChatID())
		}
		t.fillExtraFromChat(extra, chat)
	}

	if peerClass, ok := msg.GetFromID(); ok {
		if userPeer, ok := peerClass.(*tg.PeerUser); ok {
			sender, err = t.getUserByID(userPeer.GetUserID())
			if err != nil {
				t.retrieveEntities(pts, &e)
				sender, _ = t.getUserByID(userPeer.GetUserID())
			}
			t.fillExtraFromUser(extra, sender)
		}
	}

	t.Propagate(data.NewMessageWithExtra(msg.Message, extra))
	return nil
}

func (t *Telegram) onChannelMessage(ctx context.Context, e tg.Entities, pts int, msg *tg.Message, edit bool, media tg.MessageMediaClass) error {
	var err error
	var chat *tg.Chat
	var channel *tg.Channel
	var sender *tg.User

	extra := make(map[string]interface{})
	extra["type"] = "channel_message"
	extra["msg_edited"] = strconv.FormatBool(edit)
	extra["msg_hasmedia"] = false
	if media != nil {
		extra["msg_hasmedia"] = "true"
		extra["_msg_media"] = media
		fname, ext, size := t.getMediaInfo(media)
		extra["msg_medianame"] = fname
		extra["msg_mediaext"] = ext
		extra["msg_mediasize"] = size
	}
	extra["_telegram_api"] = t.api

	tm := time.Unix(int64(msg.Date), 0)
	extra["msg_timestamp"] = msg.Date
	extra["msg_date"] = tm.Format(time.DateOnly)
	extra["msg_time"] = tm.Format(time.TimeOnly)

	t.updateMaps(e)

	if chatPeer, ok := msg.GetPeerID().(*tg.PeerChat); ok {
		chat, err = t.getChatByID(chatPeer.GetChatID())
		if err != nil {
			t.retrieveEntities(pts, &e)
			chat, _ = t.getChatByID(chatPeer.GetChatID())
		}
		t.fillExtraFromChat(extra, chat)
	} else if chanPeer, ok := msg.GetPeerID().(*tg.PeerChannel); ok {
		channel, err = t.getChannelByID(chanPeer.GetChannelID())
		if err != nil {
			t.retrieveEntities(pts, &e)
			channel, _ = t.getChannelByID(chanPeer.GetChannelID())
		}
		t.fillExtraFromChannel(extra, channel)
	}

	if peerClass, ok := msg.GetFromID(); ok {
		if userPeer, ok := peerClass.(*tg.PeerUser); ok {
			sender, err = t.getUserByID(userPeer.GetUserID())
			if err != nil {
				t.retrieveEntities(pts, &e)
				sender, _ = t.getUserByID(userPeer.GetUserID())
			}
			t.fillExtraFromUser(extra, sender)
		}
	}

	t.Propagate(data.NewMessageWithExtra(msg.Message, extra))
	return nil
}

// Start propagates a message every time a new tweet is published
func (t *Telegram) Start() {
	log.Debug("Initialization of Telegram")
	dispatcher := tg.NewUpdateDispatcher()
	gaps := updates.New(updates.Config{
		Handler: dispatcher,
	})

	options := telegram.Options{
		UpdateHandler: gaps,
	}

	if t.sessionFolder != "" {
		if _, err := os.Stat(t.sessionFolder); os.IsNotExist(err) {
			if err := os.MkdirAll(t.sessionFolder, 0700); err != nil {
				log.Error("can't create the folder '%s': %s", t.sessionFolder, err)
				return
			}
		}
		sessionPath := filepath.Join(t.sessionFolder, fmt.Sprintf("phone%s-session.json", t.phoneNumber))
		log.Info("setting Session storage for Telegram: '%s'", sessionPath)
		options.SessionStorage = &telegram.FileSessionStorage{
			Path: sessionPath,
		}
	}

	client := telegram.NewClient(t.appID, t.appHash, options)
	t.api = client.API()

	dispatcher.OnEditMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditMessage) error {
		t.updateMaps(e)
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		if msg.Out {
			// Outgoing message.
			return nil
		}

		var media tg.MessageMediaClass
		if m, ok := msg.GetMedia(); ok {
			media = m
		}

		return t.onMessage(ctx, e, u.Pts, msg, true, media)
	})

	dispatcher.OnNewMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewMessage) error {
		t.updateMaps(e)
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		if msg.Out {
			// Outgoing message.
			return nil
		}

		var media tg.MessageMediaClass
		if m, ok := msg.GetMedia(); ok {
			media = m
		}

		return t.onMessage(ctx, e, u.Pts, msg, false, media)
	})

	dispatcher.OnEditChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateEditChannelMessage) error {
		t.updateMaps(e)
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		if msg.Out {
			// Outgoing message.
			return nil
		}

		var media tg.MessageMediaClass
		if m, ok := msg.GetMedia(); ok {
			media = m
		}

		return t.onChannelMessage(ctx, e, u.Pts, msg, true, media)
	})

	dispatcher.OnNewChannelMessage(func(ctx context.Context, e tg.Entities, u *tg.UpdateNewChannelMessage) error {
		t.updateMaps(e)
		msg, ok := u.Message.(*tg.Message)
		if !ok {
			return nil
		}
		if msg.Out {
			// Outgoing message.
			return nil
		}

		var media tg.MessageMediaClass
		if m, ok := msg.GetMedia(); ok {
			media = m
		}

		return t.onChannelMessage(ctx, e, u.Pts, msg, false, media)
	})

	t.isRunning = true

	go func() {
		// Authentication flow handles authentication process, like prompting for code and 2FA password.
		flow := auth.NewFlow(Terminal{PhoneNumber: t.phoneNumber}, auth.SendCodeOptions{})

		if err := client.Run(t.context, func(ctx context.Context) error {
			// Perform auth if no session is available.
			if err := client.Auth().IfNecessary(ctx, flow); err != nil {
				return fmt.Errorf("auth error: %s", err)
			}

			// Getting info about current user.
			self, err := client.Self(ctx)
			if err != nil {
				return fmt.Errorf("getting self info: %s", err)
			}

			name := self.FirstName
			if self.Username != "" {
				// Username is optional.
				name = fmt.Sprintf("%s (@%s) ID=%d", name, self.Username, self.ID)
			}
			log.Info("Connected as user: %s", name)

			return gaps.Run(ctx, t.api, self.ID, updates.AuthOptions{
				IsBot: self.Bot,
				OnStart: func(ctx context.Context) {
					log.Info("Telegram: update recovery initialized and started, listening for events")
				},
			})
		}); err != nil {
			log.Error("run: %s", err)
			t.isRunning = false
		}
	}()
}

// Stop handles the Feeder shutdown
func (t *Telegram) Stop() {
	log.Debug("feeder '%s' stream stop", t.Name())
	t.isRunning = false
	t.cancelContext()
}

// OnEvent is called when an event occurs
func (t *Telegram) OnEvent(event *data.Event) {
	if event.Type == "shutdown" && t.isRunning {
		log.Debug("shutdown event received")
		t.Stop()
	}
}

// Auto factory adding
func init() {
	register("telegram", NewTelegramFeeder)
}

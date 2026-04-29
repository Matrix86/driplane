package feeders

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/evilsocket/islazy/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

// TelegramBot is a Feeder that consumes Telegram Bot API updates.
type TelegramBot struct {
	Base

	token            string
	mode             string
	addr             string
	webhookURL       string
	webhookSecret    string
	deleteWebhook    bool
	commands         []string
	allowedUsers     map[int64]bool
	allowedUsernames map[string]bool
	allowedChats     map[int64]bool
	events           map[string]bool
	debug            bool

	bot    *bot.Bot
	server *http.Server
	ctx    context.Context
	cancel context.CancelFunc
}

// NewTelegramBotFeeder is the factory method registered for this feeder.
func NewTelegramBotFeeder(conf map[string]string) (Feeder, error) {
	t := &TelegramBot{
		mode:             "polling",
		addr:             ":3000",
		allowedUsers:     map[int64]bool{},
		allowedUsernames: map[string]bool{},
		allowedChats:     map[int64]bool{},
		events:           defaultTelegramBotEvents(),
	}

	if v, ok := conf["telegrambot.token"]; ok {
		t.token = v
	}
	if v, ok := conf["telegrambot.mode"]; ok {
		t.mode = strings.ToLower(strings.TrimSpace(v))
	}
	if v, ok := conf["telegrambot.addr"]; ok && v != "" {
		t.addr = v
	}
	if v, ok := conf["telegrambot.webhook_url"]; ok {
		t.webhookURL = v
	}
	if v, ok := conf["telegrambot.webhook_secret"]; ok {
		t.webhookSecret = v
	}
	if v, ok := conf["telegrambot.delete_webhook_on_stop"]; ok && v == "true" {
		t.deleteWebhook = true
	}
	if v, ok := conf["telegrambot.debug"]; ok && v == "true" {
		t.debug = true
	}

	if v, ok := conf["telegrambot.allowed_users"]; ok {
		t.allowedUsers, t.allowedUsernames = parseAllowedUsers(v)
	}
	if v, ok := conf["telegrambot.allowed_chats"]; ok {
		t.allowedChats = parseAllowedChats(v)
	}
	if v, ok := conf["telegrambot.events"]; ok {
		t.events = parseEvents(v)
	}
	if v, ok := conf["telegrambot.commands"]; ok {
		t.commands = parseCommands(v)
	}

	if t.token == "" {
		return nil, fmt.Errorf("the param 'telegrambot.token' is required by telegrambot feeder")
	}
	if t.mode != "polling" && t.mode != "webhook" {
		return nil, fmt.Errorf("telegrambot.mode must be 'polling' or 'webhook', got '%s'", t.mode)
	}
	if t.mode == "webhook" && t.webhookURL == "" {
		return nil, fmt.Errorf("telegrambot.webhook_url is required when mode='webhook'")
	}

	return t, nil
}

func defaultTelegramBotEvents() map[string]bool {
	return map[string]bool{
		"message":             true,
		"command":             true,
		"callback_query":      true,
		"edited_message":      true,
		"channel_post":        true,
		"edited_channel_post": true,
		"chat_member":         true,
		"my_chat_member":      true,
	}
}

func parseAllowedUsers(raw string) (map[int64]bool, map[string]bool) {
	ids := map[int64]bool{}
	names := map[string]bool{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if strings.HasPrefix(item, "@") {
			names[strings.ToLower(item[1:])] = true
			continue
		}
		n, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			log.Warning("telegrambot: ignoring malformed allowed_users entry '%s'", item)
			continue
		}
		ids[n] = true
	}
	return ids, names
}

func parseAllowedChats(raw string) map[int64]bool {
	out := map[int64]bool{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		n, err := strconv.ParseInt(item, 10, 64)
		if err != nil {
			log.Warning("telegrambot: ignoring malformed allowed_chats entry '%s'", item)
			continue
		}
		out[n] = true
	}
	return out
}

func parseEvents(raw string) map[string]bool {
	all := defaultTelegramBotEvents()
	out := map[string]bool{}
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		if _, ok := all[item]; !ok {
			log.Warning("telegrambot: ignoring unknown event '%s'", item)
			continue
		}
		out[item] = true
	}
	return out
}

func parseCommands(raw string) []string {
	var out []string
	for _, item := range strings.Split(raw, ",") {
		item = strings.TrimSpace(item)
		if item == "" {
			continue
		}
		out = append(out, item)
	}
	return out
}

// Start initializes the bot client and begins consuming updates.
func (t *TelegramBot) Start() {
	log.Debug("Initialization of TelegramBot")

	opts := []bot.Option{
		bot.WithDefaultHandler(t.onUpdate),
	}
	if t.debug {
		opts = append(opts, bot.WithDebug())
	}
	if t.mode == "webhook" && t.webhookSecret != "" {
		opts = append(opts, bot.WithWebhookSecretToken(t.webhookSecret))
	}

	b, err := bot.New(t.token, opts...)
	if err != nil {
		log.Error("telegrambot: bot.New failed: %s", err)
		return
	}
	t.bot = b

	t.ctx, t.cancel = context.WithCancel(context.Background())

	switch t.mode {
	case "polling":
		go t.bot.Start(t.ctx)
	case "webhook":
		if err := t.startWebhook(); err != nil {
			log.Error("telegrambot: webhook start failed: %s", err)
			return
		}
	}
	t.isRunning = true
}

// Stop tears down the bot client and any HTTP server.
func (t *TelegramBot) Stop() {
	log.Debug("feeder '%s' stream stop", t.Name())
	if !t.isRunning {
		return
	}
	t.isRunning = false

	if t.cancel != nil {
		t.cancel()
	}

	if t.mode == "webhook" {
		t.stopWebhook()
	}
}

func (t *TelegramBot) startWebhook() error {
	_, err := t.bot.SetWebhook(t.ctx, &bot.SetWebhookParams{
		URL:         t.webhookURL,
		SecretToken: t.webhookSecret,
	})
	if err != nil {
		return fmt.Errorf("SetWebhook: %w", err)
	}

	path := "/"
	if u, err := url.Parse(t.webhookURL); err == nil && u.Path != "" {
		path = u.Path
	}

	mux := http.NewServeMux()
	mux.Handle(path, t.bot.WebhookHandler())
	t.server = &http.Server{Addr: t.addr, Handler: mux}

	go func() {
		log.Info("telegrambot: webhook server listening on %s", t.addr)
		if err := t.server.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Error("telegrambot: webhook server error: %s", err)
		}
	}()
	go t.bot.StartWebhook(t.ctx)
	return nil
}

func (t *TelegramBot) stopWebhook() {
	if t.deleteWebhook && t.bot != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := t.bot.DeleteWebhook(ctx, &bot.DeleteWebhookParams{}); err != nil {
			log.Error("telegrambot: DeleteWebhook: %s", err)
		}
	}
	if t.server != nil {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := t.server.Shutdown(ctx); err != nil {
			log.Error("telegrambot: server.Shutdown: %s", err)
		}
	}
}

// OnEvent handles bus events.
func (t *TelegramBot) OnEvent(e *data.Event) {
	if e.Type == "shutdown" && t.isRunning {
		t.Stop()
	}
}

// detectType returns the canonical event type for an update, or "" if the
// update has no handleable payload.
func (t *TelegramBot) detectType(u *models.Update) string {
	switch {
	case u.Message != nil:
		if t.looksLikeCommand(u.Message.Text) {
			return "command"
		}
		return "message"
	case u.EditedMessage != nil:
		return "edited_message"
	case u.ChannelPost != nil:
		return "channel_post"
	case u.EditedChannelPost != nil:
		return "edited_channel_post"
	case u.CallbackQuery != nil:
		return "callback_query"
	case u.MyChatMember != nil:
		return "my_chat_member"
	case u.ChatMember != nil:
		return "chat_member"
	default:
		return ""
	}
}

// looksLikeCommand decides whether a message text should be classified as a
// command for this feeder (respects the configured explicit list).
func (t *TelegramBot) looksLikeCommand(text string) bool {
	if !strings.HasPrefix(text, "/") {
		return false
	}
	cmd, _ := splitCommand(text)
	if len(t.commands) == 0 {
		return true
	}
	for _, c := range t.commands {
		if c == cmd {
			return true
		}
	}
	return false
}

// allowed reports whether an event passes the configured allowlists.
//
// Semantics:
//   - When allowed_users is configured, the sender MUST be in the list. Updates
//     without sender identity (e.g. channel posts) skip the user check.
//   - When allowed_chats is configured, the chat must be in the list, EXCEPT
//     for private chats from a verified allowed user — those bypass the chat
//     list so users in allowed_users can DM the bot regardless of allowed_chats.
func (t *TelegramBot) allowed(userID int64, username string, chatID int64, chatType models.ChatType) bool {
	userListConfigured := len(t.allowedUsers) > 0 || len(t.allowedUsernames) > 0
	chatListConfigured := len(t.allowedChats) > 0

	if userListConfigured && (userID != 0 || username != "") {
		userInList := (userID != 0 && t.allowedUsers[userID]) ||
			(username != "" && t.allowedUsernames[strings.ToLower(username)])
		if !userInList {
			return false
		}
	}

	if chatListConfigured {
		if t.allowedChats[chatID] {
			return true
		}
		if userListConfigured && chatType == models.ChatTypePrivate {
			return true
		}
		return false
	}

	return true
}

// extractIdentity pulls (userID, username, chatID, chatType) from an update.
// Any field may be zero when the update has no corresponding party.
func extractIdentity(u *models.Update) (int64, string, int64, models.ChatType) {
	switch {
	case u.Message != nil:
		uid, name := userInfo(u.Message.From)
		return uid, name, u.Message.Chat.ID, u.Message.Chat.Type
	case u.EditedMessage != nil:
		uid, name := userInfo(u.EditedMessage.From)
		return uid, name, u.EditedMessage.Chat.ID, u.EditedMessage.Chat.Type
	case u.ChannelPost != nil:
		uid, name := userInfo(u.ChannelPost.From)
		return uid, name, u.ChannelPost.Chat.ID, u.ChannelPost.Chat.Type
	case u.EditedChannelPost != nil:
		uid, name := userInfo(u.EditedChannelPost.From)
		return uid, name, u.EditedChannelPost.Chat.ID, u.EditedChannelPost.Chat.Type
	case u.CallbackQuery != nil:
		chat, _ := callbackChat(u.CallbackQuery)
		var chatID int64
		var ct models.ChatType
		if chat != nil {
			chatID = chat.ID
			ct = chat.Type
		}
		return u.CallbackQuery.From.ID, u.CallbackQuery.From.Username, chatID, ct
	case u.MyChatMember != nil:
		return u.MyChatMember.From.ID, u.MyChatMember.From.Username, u.MyChatMember.Chat.ID, u.MyChatMember.Chat.Type
	case u.ChatMember != nil:
		return u.ChatMember.From.ID, u.ChatMember.From.Username, u.ChatMember.Chat.ID, u.ChatMember.Chat.Type
	default:
		return 0, "", 0, ""
	}
}

func userInfo(u *models.User) (int64, string) {
	if u == nil {
		return 0, ""
	}
	return u.ID, u.Username
}

// callbackChat returns the chat the inline-keyboard message lives in, plus the
// message ID, regardless of whether the bot still has access to that message.
// Returns (nil, 0) for inline_message_id-only callbacks (no chat context).
func callbackChat(cq *models.CallbackQuery) (*models.Chat, int) {
	if cq == nil {
		return nil, 0
	}
	switch cq.Message.Type {
	case models.MaybeInaccessibleMessageTypeMessage:
		if cq.Message.Message != nil {
			return &cq.Message.Message.Chat, cq.Message.Message.ID
		}
	case models.MaybeInaccessibleMessageTypeInaccessibleMessage:
		if cq.Message.InaccessibleMessage != nil {
			return &cq.Message.InaccessibleMessage.Chat, cq.Message.InaccessibleMessage.MessageID
		}
	}
	return nil, 0
}

// splitCommand returns (command, args-remainder). Strips any @botname suffix.
func splitCommand(text string) (string, string) {
	parts := strings.SplitN(text, " ", 2)
	head := parts[0]
	rest := ""
	if len(parts) == 2 {
		rest = parts[1]
	}
	if idx := strings.Index(head, "@"); idx > 0 {
		head = head[:idx]
	}
	return head, rest
}

func fillExtraFromUser(extra map[string]interface{}, u *models.User) {
	if u == nil {
		return
	}
	extra["user_id"] = strconv.FormatInt(u.ID, 10)
	extra["user_username"] = u.Username
	extra["user_firstname"] = u.FirstName
	extra["user_lastname"] = u.LastName
	extra["user_language"] = u.LanguageCode
	extra["user_isbot"] = strconv.FormatBool(u.IsBot)
	extra["user_ispremium"] = strconv.FormatBool(u.IsPremium)
}

func fillExtraFromChat(extra map[string]interface{}, c models.Chat) {
	extra["chat_id"] = strconv.FormatInt(c.ID, 10)
	extra["chat_type"] = string(c.Type)
	extra["chat_title"] = c.Title
	extra["chat_username"] = c.Username
}

func fillExtraFromMessage(extra map[string]interface{}, m *models.Message, edited bool) {
	extra["msg_id"] = strconv.Itoa(m.ID)
	extra["msg_timestamp"] = strconv.Itoa(m.Date)
	tm := time.Unix(int64(m.Date), 0)
	extra["msg_date"] = tm.Format(time.DateOnly)
	extra["msg_time"] = tm.Format(time.TimeOnly)
	extra["msg_edited"] = strconv.FormatBool(edited)
	extra["text"] = m.Text

	if m.Caption != "" {
		extra["msg_caption"] = m.Caption
	}
	if m.ReplyToMessage != nil {
		extra["msg_reply_to_id"] = strconv.Itoa(m.ReplyToMessage.ID)
	}
	if m.ForwardOrigin != nil {
		switch {
		case m.ForwardOrigin.MessageOriginUser != nil:
			extra["msg_forward_from"] = strconv.FormatInt(m.ForwardOrigin.MessageOriginUser.SenderUser.ID, 10)
		case m.ForwardOrigin.MessageOriginChat != nil:
			extra["msg_forward_from_chat"] = strconv.FormatInt(m.ForwardOrigin.MessageOriginChat.SenderChat.ID, 10)
		case m.ForwardOrigin.MessageOriginChannel != nil:
			extra["msg_forward_from_chat"] = strconv.FormatInt(m.ForwardOrigin.MessageOriginChannel.Chat.ID, 10)
		}
	}

	hasMedia, kind, fileID, fileUID, name, ext, size := mediaInfo(m)
	extra["msg_hasmedia"] = strconv.FormatBool(hasMedia)
	if hasMedia {
		extra["msg_mediatype"] = kind
		extra["msg_file_id"] = fileID
		extra["msg_file_unique_id"] = fileUID
		extra["msg_medianame"] = name
		extra["msg_mediaext"] = ext
		extra["msg_mediasize"] = strconv.FormatInt(size, 10)
	}
}

// mediaInfo inspects a Message and returns (hasMedia, kind, fileID, fileUniqueID, name, ext, size).
func mediaInfo(m *models.Message) (bool, string, string, string, string, string, int64) {
	switch {
	case len(m.Photo) > 0:
		biggest := m.Photo[0]
		for _, p := range m.Photo[1:] {
			if p.FileSize > biggest.FileSize {
				biggest = p
			}
		}
		name := fmt.Sprintf("photo_%s.jpg", biggest.FileUniqueID)
		return true, "photo", biggest.FileID, biggest.FileUniqueID, name, ".jpg", int64(biggest.FileSize)
	case m.Document != nil:
		d := m.Document
		name := d.FileName
		ext := filepath.Ext(name)
		if name == "" {
			name = "doc_" + d.FileUniqueID
		}
		return true, "document", d.FileID, d.FileUniqueID, name, ext, int64(d.FileSize)
	case m.Video != nil:
		v := m.Video
		name := fmt.Sprintf("video_%s.mp4", v.FileUniqueID)
		return true, "video", v.FileID, v.FileUniqueID, name, ".mp4", int64(v.FileSize)
	case m.Audio != nil:
		a := m.Audio
		name := a.FileName
		ext := filepath.Ext(name)
		if name == "" {
			name = fmt.Sprintf("audio_%s.mp3", a.FileUniqueID)
			ext = ".mp3"
		}
		return true, "audio", a.FileID, a.FileUniqueID, name, ext, int64(a.FileSize)
	case m.Voice != nil:
		v := m.Voice
		return true, "voice", v.FileID, v.FileUniqueID, "voice_" + v.FileUniqueID + ".ogg", ".ogg", int64(v.FileSize)
	case m.Animation != nil:
		a := m.Animation
		name := a.FileName
		ext := filepath.Ext(name)
		if name == "" {
			name = "animation_" + a.FileUniqueID + ".mp4"
			ext = ".mp4"
		}
		return true, "animation", a.FileID, a.FileUniqueID, name, ext, int64(a.FileSize)
	case m.Sticker != nil:
		s := m.Sticker
		return true, "sticker", s.FileID, s.FileUniqueID, "sticker_" + s.FileUniqueID + ".webp", ".webp", int64(s.FileSize)
	}
	return false, "", "", "", "", "", 0
}

// onUpdate is the single entry point registered as the library's default handler.
func (t *TelegramBot) onUpdate(_ context.Context, _ *bot.Bot, u *models.Update) {
	defer func() {
		if r := recover(); r != nil {
			log.Error("telegrambot: panic in onUpdate: %v", r)
		}
	}()

	evType := t.detectType(u)
	if evType == "" {
		return
	}
	if !t.events[evType] {
		return
	}

	userID, username, chatID, chatType := extractIdentity(u)
	if !t.allowed(userID, username, chatID, chatType) {
		log.Debug("telegrambot: update rejected by allowlist (user=%d chat=%d)", userID, chatID)
		return
	}

	extra := map[string]interface{}{
		"type":             evType,
		"update_id":        strconv.FormatInt(u.ID, 10),
		"_telegrambot_api": t.bot,
	}

	main := t.fillByType(extra, evType, u)

	t.Propagate(data.NewMessageWithExtra(main, extra))
}

// fillByType routes to the per-type filler and returns the "main" text.
func (t *TelegramBot) fillByType(extra map[string]interface{}, evType string, u *models.Update) string {
	switch evType {
	case "message", "command":
		m := u.Message
		fillExtraFromUser(extra, m.From)
		fillExtraFromChat(extra, m.Chat)
		fillExtraFromMessage(extra, m, false)
		if evType == "command" {
			cmd, args := splitCommand(m.Text)
			extra["command"] = cmd
			extra["command_args"] = args
		}
		return mainOf(m)
	case "edited_message":
		m := u.EditedMessage
		fillExtraFromUser(extra, m.From)
		fillExtraFromChat(extra, m.Chat)
		fillExtraFromMessage(extra, m, true)
		if m.EditDate != 0 {
			extra["msg_edit_date"] = strconv.Itoa(m.EditDate)
		}
		return mainOf(m)
	case "channel_post":
		m := u.ChannelPost
		fillExtraFromUser(extra, m.From)
		fillExtraFromChat(extra, m.Chat)
		fillExtraFromMessage(extra, m, false)
		return mainOf(m)
	case "edited_channel_post":
		m := u.EditedChannelPost
		fillExtraFromUser(extra, m.From)
		fillExtraFromChat(extra, m.Chat)
		fillExtraFromMessage(extra, m, true)
		if m.EditDate != 0 {
			extra["msg_edit_date"] = strconv.Itoa(m.EditDate)
		}
		return mainOf(m)
	case "callback_query":
		cq := u.CallbackQuery
		fillExtraFromUser(extra, &cq.From)
		extra["callback_id"] = cq.ID
		extra["callback_data"] = cq.Data
		extra["callback_chatinstance"] = cq.ChatInstance
		if chat, msgID := callbackChat(cq); chat != nil {
			fillExtraFromChat(extra, *chat)
			extra["msg_id"] = strconv.Itoa(msgID)
		}
		return cq.Data
	case "chat_member", "my_chat_member":
		var cmu *models.ChatMemberUpdated
		if evType == "chat_member" {
			cmu = u.ChatMember
		} else {
			cmu = u.MyChatMember
		}
		fillExtraFromUser(extra, &cmu.From)
		fillExtraFromChat(extra, cmu.Chat)
		extra["member_old_status"] = chatMemberStatus(cmu.OldChatMember)
		extra["member_new_status"] = chatMemberStatus(cmu.NewChatMember)
		mid, mname := chatMemberIdentity(cmu.NewChatMember)
		if mid == 0 {
			mid, mname = chatMemberIdentity(cmu.OldChatMember)
		}
		extra["member_user_id"] = strconv.FormatInt(mid, 10)
		extra["member_user_username"] = mname
		return ""
	}
	return ""
}

func mainOf(m *models.Message) string {
	if m.Text != "" {
		return m.Text
	}
	return m.Caption
}

// chatMemberStatus returns the Telegram status string for a ChatMember union.
func chatMemberStatus(cm models.ChatMember) string {
	return string(cm.Type)
}

// chatMemberIdentity extracts (user_id, username) from a ChatMember union by
// walking its non-nil variant.
func chatMemberIdentity(cm models.ChatMember) (int64, string) {
	switch {
	case cm.Owner != nil && cm.Owner.User != nil:
		return cm.Owner.User.ID, cm.Owner.User.Username
	case cm.Administrator != nil:
		return cm.Administrator.User.ID, cm.Administrator.User.Username
	case cm.Member != nil && cm.Member.User != nil:
		return cm.Member.User.ID, cm.Member.User.Username
	case cm.Restricted != nil && cm.Restricted.User != nil:
		return cm.Restricted.User.ID, cm.Restricted.User.Username
	case cm.Left != nil && cm.Left.User != nil:
		return cm.Left.User.ID, cm.Left.User.Username
	case cm.Banned != nil && cm.Banned.User != nil:
		return cm.Banned.User.ID, cm.Banned.User.Username
	}
	return 0, ""
}

func init() {
	register("telegrambot", NewTelegramBotFeeder)
}

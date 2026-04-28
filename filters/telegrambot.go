package filters

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"path"
	"path/filepath"
	"strconv"
	"strings"
	"text/template"

	"github.com/Masterminds/sprig/v3"
	"github.com/evilsocket/islazy/log"
	"github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"

	"github.com/Matrix86/driplane/data"
)

// TelegramBot is a Filter to perform Bot API actions on top of the telegrambot feeder.
type TelegramBot struct {
	Base

	action string

	toChat      *template.Template
	text        *template.Template
	caption     *template.Template
	parseMode   *template.Template
	messageID   *template.Template
	callbackID  *template.Template
	fileID      *template.Template
	filename    *template.Template
	media       *template.Template
	mediaType   *template.Template
	replyMarkup *template.Template

	params map[string]string
}

var validTelegramBotActions = map[string]bool{
	"send_message":    true,
	"edit_message":    true,
	"delete_message":  true,
	"answer_callback": true,
	"send_media":      true,
	"download_file":   true,
}

// NewTelegramBotFilter is the registered method to instantiate a TelegramBot filter.
func NewTelegramBotFilter(p map[string]string) (Filter, error) {
	f := &TelegramBot{
		params: p,
		action: "send_message",
	}
	f.cbFilter = f.DoFilter

	if v, ok := f.params["action"]; ok {
		f.action = v
	}
	if !validTelegramBotActions[f.action] {
		return nil, fmt.Errorf("telegrambot: action '%s' is not valid", f.action)
	}

	if err := f.compileTemplates(); err != nil {
		return nil, err
	}
	if err := f.validateRequired(); err != nil {
		return nil, err
	}
	return f, nil
}

// compileTemplates parses every recognised string param as a text/template with sprig funcs.
func (f *TelegramBot) compileTemplates() error {
	type entry struct {
		key string
		dst **template.Template
	}
	entries := []entry{
		{"to_chat", &f.toChat},
		{"text", &f.text},
		{"caption", &f.caption},
		{"parse_mode", &f.parseMode},
		{"message_id", &f.messageID},
		{"callback_id", &f.callbackID},
		{"file_id", &f.fileID},
		{"filename", &f.filename},
		{"media", &f.media},
		{"media_type", &f.mediaType},
		{"reply_markup", &f.replyMarkup},
	}
	for _, e := range entries {
		v, ok := f.params[e.key]
		if !ok || v == "" {
			continue
		}
		tpl, err := template.New("telegrambot:" + e.key).Funcs(sprig.FuncMap()).Parse(v)
		if err != nil {
			return fmt.Errorf("telegrambot: invalid template for %q: %w", e.key, err)
		}
		*e.dst = tpl
	}
	return nil
}

// validateRequired asserts mandatory params are present per action.
func (f *TelegramBot) validateRequired() error {
	switch f.action {
	case "send_message":
		if f.text == nil {
			return fmt.Errorf("telegrambot: 'text' is mandatory for action 'send_message'")
		}
	case "edit_message":
		if f.text == nil {
			return fmt.Errorf("telegrambot: 'text' is mandatory for action 'edit_message'")
		}
		if f.messageID == nil {
			return fmt.Errorf("telegrambot: 'message_id' is mandatory for action 'edit_message'")
		}
	case "delete_message":
		if f.messageID == nil {
			return fmt.Errorf("telegrambot: 'message_id' is mandatory for action 'delete_message'")
		}
	case "answer_callback":
		if f.text == nil {
			return fmt.Errorf("telegrambot: 'text' is mandatory for action 'answer_callback'")
		}
	case "send_media":
		if f.media == nil {
			return fmt.Errorf("telegrambot: 'media' is mandatory for action 'send_media'")
		}
	case "download_file":
		if f.filename == nil {
			return fmt.Errorf("telegrambot: 'filename' is mandatory for action 'download_file'")
		}
	}
	return nil
}

// resolveBot fetches the *bot.Bot exported by the telegrambot feeder.
func (f *TelegramBot) resolveBot(msg *data.Message) (*bot.Bot, bool) {
	target := msg.GetTarget("_telegrambot_api")
	if target == nil {
		log.Error("[%s::%s] telegrambot filter: missing _telegrambot_api extra", f.Rule(), f.Name())
		return nil, false
	}
	b, ok := target.(*bot.Bot)
	if !ok || b == nil {
		log.Error("[%s::%s] telegrambot filter: _telegrambot_api has wrong type %T", f.Rule(), f.Name(), target)
		return nil, false
	}
	return b, true
}

// resolveChatID renders the to_chat template if set, otherwise falls back to the
// incoming chat_id extra. Returns false on any rendering / parsing error.
func (f *TelegramBot) resolveChatID(msg *data.Message) (int64, bool) {
	if f.toChat != nil {
		s, err := msg.ApplyPlaceholder(f.toChat)
		if err != nil {
			log.Error("[%s::%s] telegrambot: render to_chat: %s", f.Rule(), f.Name(), err)
			return 0, false
		}
		s = strings.TrimSpace(s)
		if s == "" {
			log.Error("[%s::%s] telegrambot: to_chat rendered empty", f.Rule(), f.Name())
			return 0, false
		}
		n, err := strconv.ParseInt(s, 10, 64)
		if err != nil {
			log.Error("[%s::%s] telegrambot: parse to_chat: %s", f.Rule(), f.Name(), err)
			return 0, false
		}
		return n, true
	}
	raw := msg.GetTarget("chat_id")
	if raw == nil {
		log.Error("[%s::%s] telegrambot: no chat_id available (set to_chat or rely on incoming chat_id)", f.Rule(), f.Name())
		return 0, false
	}
	s, ok := raw.(string)
	if !ok {
		log.Error("[%s::%s] telegrambot: incoming chat_id has wrong type %T", f.Rule(), f.Name(), raw)
		return 0, false
	}
	n, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		log.Error("[%s::%s] telegrambot: parse incoming chat_id: %s", f.Rule(), f.Name(), err)
		return 0, false
	}
	return n, true
}

// mapParseMode maps user-facing strings to library constants. Defaults to plain
// (empty parse_mode) for empty or unknown values.
func mapParseMode(s string) models.ParseMode {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "":
		return models.ParseMode("")
	case "html":
		return models.ParseModeHTML
	case "markdown":
		return models.ParseModeMarkdownV1
	case "markdownv2":
		return models.ParseModeMarkdown
	default:
		return models.ParseMode("")
	}
}

// resolveParseMode returns the configured parse_mode (default HTML).
func (f *TelegramBot) resolveParseMode(msg *data.Message) (models.ParseMode, bool) {
	if f.parseMode == nil {
		return models.ParseModeHTML, true
	}
	s, err := msg.ApplyPlaceholder(f.parseMode)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render parse_mode: %s", f.Rule(), f.Name(), err)
		return models.ParseMode(""), false
	}
	return mapParseMode(s), true
}

// resolveReplyMarkup parses the reply_markup template (if set) into models.ReplyMarkup.
// Returns (nil, true) when reply_markup is unset or renders to whitespace.
func (f *TelegramBot) resolveReplyMarkup(msg *data.Message) (models.ReplyMarkup, bool) {
	if f.replyMarkup == nil {
		return nil, true
	}
	raw, err := msg.ApplyPlaceholder(f.replyMarkup)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render reply_markup: %s", f.Rule(), f.Name(), err)
		return nil, false
	}
	if strings.TrimSpace(raw) == "" {
		return nil, true
	}
	var kb models.InlineKeyboardMarkup
	if err := json.Unmarshal([]byte(raw), &kb); err != nil {
		log.Error("[%s::%s] telegrambot: invalid reply_markup JSON: %s", f.Rule(), f.Name(), err)
		return nil, false
	}
	return kb, true
}

// DoFilter dispatches by action.
func (f *TelegramBot) DoFilter(msg *data.Message) (bool, error) {
	b, ok := f.resolveBot(msg)
	if !ok {
		return false, nil
	}

	ctx := context.Background()
	switch f.action {
	case "send_message":
		return f.doSendMessage(ctx, b, msg), nil
	case "edit_message":
		return f.doEditMessage(ctx, b, msg), nil
	case "delete_message":
		return f.doDeleteMessage(ctx, b, msg), nil
	case "answer_callback":
		return f.doAnswerCallback(ctx, b, msg), nil
	case "send_media":
		return f.doSendMedia(ctx, b, msg), nil
	case "download_file":
		return f.doDownloadFile(ctx, b, msg), nil
	}
	return false, nil
}

// doSendMessage performs sendMessage and enriches the message with sent_* extras.
func (f *TelegramBot) doSendMessage(ctx context.Context, b *bot.Bot, msg *data.Message) bool {
	chatID, ok := f.resolveChatID(msg)
	if !ok {
		return false
	}

	text, err := msg.ApplyPlaceholder(f.text)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render text: %s", f.Rule(), f.Name(), err)
		return false
	}
	pm, ok := f.resolveParseMode(msg)
	if !ok {
		return false
	}
	rm, ok := f.resolveReplyMarkup(msg)
	if !ok {
		return false
	}

	sent, err := b.SendMessage(ctx, &bot.SendMessageParams{
		ChatID:      chatID,
		Text:        text,
		ParseMode:   pm,
		ReplyMarkup: rm,
	})
	if err != nil {
		log.Error("[%s::%s] telegrambot: sendMessage failed: %s", f.Rule(), f.Name(), err)
		return false
	}

	msg.SetTarget("main", text)
	msg.SetTarget("sent_message_id", strconv.Itoa(sent.ID))
	msg.SetTarget("sent_chat_id", strconv.FormatInt(sent.Chat.ID, 10))
	msg.SetTarget("sent_text", text)
	msg.SetTarget("sent_date", strconv.Itoa(sent.Date))
	return true
}

// doEditMessage performs editMessageText and enriches the message with edited_* extras.
func (f *TelegramBot) doEditMessage(ctx context.Context, b *bot.Bot, msg *data.Message) bool {
	chatID, ok := f.resolveChatID(msg)
	if !ok {
		return false
	}

	msgIDStr, err := msg.ApplyPlaceholder(f.messageID)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render message_id: %s", f.Rule(), f.Name(), err)
		return false
	}
	msgID, err := strconv.Atoi(strings.TrimSpace(msgIDStr))
	if err != nil {
		log.Error("[%s::%s] telegrambot: parse message_id: %s", f.Rule(), f.Name(), err)
		return false
	}

	text, err := msg.ApplyPlaceholder(f.text)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render text: %s", f.Rule(), f.Name(), err)
		return false
	}
	pm, ok := f.resolveParseMode(msg)
	if !ok {
		return false
	}
	rm, ok := f.resolveReplyMarkup(msg)
	if !ok {
		return false
	}

	if _, err := b.EditMessageText(ctx, &bot.EditMessageTextParams{
		ChatID:      chatID,
		MessageID:   msgID,
		Text:        text,
		ParseMode:   pm,
		ReplyMarkup: rm,
	}); err != nil {
		log.Error("[%s::%s] telegrambot: editMessageText failed: %s", f.Rule(), f.Name(), err)
		return false
	}

	msg.SetTarget("edit_success", "true")
	msg.SetTarget("edited_message_id", strconv.Itoa(msgID))
	msg.SetTarget("edited_chat_id", strconv.FormatInt(chatID, 10))
	return true
}

func (f *TelegramBot) doDeleteMessage(ctx context.Context, b *bot.Bot, msg *data.Message) bool {
    chatID, ok := f.resolveChatID(msg)
    if !ok { return false }

    msgIDStr, err := msg.ApplyPlaceholder(f.messageID)
    if err != nil {
        log.Error("[%s::%s] telegrambot: render message_id: %s", f.Rule(), f.Name(), err)
        return false
    }
    msgID, err := strconv.Atoi(strings.TrimSpace(msgIDStr))
    if err != nil {
        log.Error("[%s::%s] telegrambot: parse message_id: %s", f.Rule(), f.Name(), err)
        return false
    }

    if _, err := b.DeleteMessage(ctx, &bot.DeleteMessageParams{ChatID: chatID, MessageID: msgID}); err != nil {
        log.Error("[%s::%s] telegrambot: deleteMessage failed: %s", f.Rule(), f.Name(), err)
        return false
    }
    msg.SetTarget("delete_success", "true")
    return true
}

func (f *TelegramBot) doAnswerCallback(ctx context.Context, b *bot.Bot, msg *data.Message) bool {
	var cbID string
	if f.callbackID != nil {
		s, err := msg.ApplyPlaceholder(f.callbackID)
		if err != nil {
			log.Error("[%s::%s] telegrambot: render callback_id: %s", f.Rule(), f.Name(), err)
			return false
		}
		cbID = strings.TrimSpace(s)
	} else if raw := msg.GetTarget("callback_id"); raw != nil {
		if s, ok := raw.(string); ok {
			cbID = s
		}
	}
	if cbID == "" {
		log.Error("[%s::%s] telegrambot: no callback_id available (set callback_id or rely on incoming extra)", f.Rule(), f.Name())
		return false
	}

	text, err := msg.ApplyPlaceholder(f.text)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render text: %s", f.Rule(), f.Name(), err)
		return false
	}

	if _, err := b.AnswerCallbackQuery(ctx, &bot.AnswerCallbackQueryParams{
		CallbackQueryID: cbID,
		Text:            text,
	}); err != nil {
		log.Error("[%s::%s] telegrambot: answerCallbackQuery failed: %s", f.Rule(), f.Name(), err)
		return false
	}
	msg.SetTarget("callback_answered", "true")
	return true
}

// resolveMediaSource decides whether the rendered media value is a URL, a local path,
// or a file_id, and returns the matching models.InputFile + an optional file handle to close.
func (f *TelegramBot) resolveMediaSource(rendered string) (models.InputFile, func(), bool) {
	s := strings.TrimSpace(rendered)
	if strings.HasPrefix(s, "http://") || strings.HasPrefix(s, "https://") {
		return &models.InputFileString{Data: s}, func() {}, true
	}
	if info, err := os.Stat(s); err == nil && !info.IsDir() {
		fh, err := os.Open(s)
		if err != nil {
			log.Error("[%s::%s] telegrambot: open media file: %s", f.Rule(), f.Name(), err)
			return nil, nil, false
		}
		return &models.InputFileUpload{Filename: filepath.Base(s), Data: fh}, func() { fh.Close() }, true
	}
	return &models.InputFileString{Data: s}, func() {}, true
}

func (f *TelegramBot) doSendMedia(ctx context.Context, b *bot.Bot, msg *data.Message) bool {
	chatID, ok := f.resolveChatID(msg)
	if !ok {
		return false
	}

	mediaStr, err := msg.ApplyPlaceholder(f.media)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render media: %s", f.Rule(), f.Name(), err)
		return false
	}
	mediaStr = strings.TrimSpace(mediaStr)
	if mediaStr == "" {
		log.Error("[%s::%s] telegrambot: media rendered empty", f.Rule(), f.Name())
		return false
	}

	mediaType := "document"
	if f.mediaType != nil {
		mt, err := msg.ApplyPlaceholder(f.mediaType)
		if err != nil {
			log.Error("[%s::%s] telegrambot: render media_type: %s", f.Rule(), f.Name(), err)
			return false
		}
		mediaType = strings.ToLower(strings.TrimSpace(mt))
	}

	caption := ""
	if f.caption != nil {
		c, err := msg.ApplyPlaceholder(f.caption)
		if err != nil {
			log.Error("[%s::%s] telegrambot: render caption: %s", f.Rule(), f.Name(), err)
			return false
		}
		caption = c
	}
	pm, ok := f.resolveParseMode(msg)
	if !ok {
		return false
	}
	rm, ok := f.resolveReplyMarkup(msg)
	if !ok {
		return false
	}

	// Validate media_type up-front so unknown types don't open file handles.
	switch mediaType {
	case "photo", "document", "video", "audio", "voice", "animation", "sticker":
	default:
		log.Error("[%s::%s] telegrambot: unknown media_type %q", f.Rule(), f.Name(), mediaType)
		return false
	}

	src, closeFn, ok := f.resolveMediaSource(mediaStr)
	if !ok {
		return false
	}
	defer closeFn()

	var sent *models.Message
	var apiErr error
	switch mediaType {
	case "photo":
		sent, apiErr = b.SendPhoto(ctx, &bot.SendPhotoParams{ChatID: chatID, Photo: src, Caption: caption, ParseMode: pm, ReplyMarkup: rm})
	case "document":
		sent, apiErr = b.SendDocument(ctx, &bot.SendDocumentParams{ChatID: chatID, Document: src, Caption: caption, ParseMode: pm, ReplyMarkup: rm})
	case "video":
		sent, apiErr = b.SendVideo(ctx, &bot.SendVideoParams{ChatID: chatID, Video: src, Caption: caption, ParseMode: pm, ReplyMarkup: rm})
	case "audio":
		sent, apiErr = b.SendAudio(ctx, &bot.SendAudioParams{ChatID: chatID, Audio: src, Caption: caption, ParseMode: pm, ReplyMarkup: rm})
	case "voice":
		sent, apiErr = b.SendVoice(ctx, &bot.SendVoiceParams{ChatID: chatID, Voice: src, Caption: caption, ParseMode: pm, ReplyMarkup: rm})
	case "animation":
		sent, apiErr = b.SendAnimation(ctx, &bot.SendAnimationParams{ChatID: chatID, Animation: src, Caption: caption, ParseMode: pm, ReplyMarkup: rm})
	case "sticker":
		sent, apiErr = b.SendSticker(ctx, &bot.SendStickerParams{ChatID: chatID, Sticker: src, ReplyMarkup: rm})
	}
	if apiErr != nil {
		log.Error("[%s::%s] telegrambot: send %s failed: %s", f.Rule(), f.Name(), mediaType, apiErr)
		return false
	}

	msg.SetTarget("main", caption)
	msg.SetTarget("sent_message_id", strconv.Itoa(sent.ID))
	msg.SetTarget("sent_chat_id", strconv.FormatInt(sent.Chat.ID, 10))
	msg.SetTarget("sent_text", caption)
	msg.SetTarget("sent_date", strconv.Itoa(sent.Date))
	return true
}

func (f *TelegramBot) doDownloadFile(ctx context.Context, b *bot.Bot, msg *data.Message) bool {
	var fileID string
	if f.fileID != nil {
		s, err := msg.ApplyPlaceholder(f.fileID)
		if err != nil {
			log.Error("[%s::%s] telegrambot: render file_id: %s", f.Rule(), f.Name(), err)
			return false
		}
		fileID = strings.TrimSpace(s)
	} else if raw := msg.GetTarget("msg_file_id"); raw != nil {
		if s, ok := raw.(string); ok {
			fileID = s
		}
	}
	if fileID == "" {
		log.Error("[%s::%s] telegrambot: no file_id available (set file_id or rely on msg_file_id extra)", f.Rule(), f.Name())
		return false
	}

	file, err := b.GetFile(ctx, &bot.GetFileParams{FileID: fileID})
	if err != nil {
		log.Error("[%s::%s] telegrambot: getFile failed: %s", f.Rule(), f.Name(), err)
		return false
	}

	// Make file_path basename available for filename templates that want to reuse it.
	msg.SetTarget("msg_filename", path.Base(file.FilePath))

	out, err := msg.ApplyPlaceholder(f.filename)
	if err != nil {
		log.Error("[%s::%s] telegrambot: render filename: %s", f.Rule(), f.Name(), err)
		return false
	}
	if dir := filepath.Dir(out); dir != "" {
		if err := os.MkdirAll(dir, 0700); err != nil {
			log.Error("[%s::%s] telegrambot: mkdir %s: %s", f.Rule(), f.Name(), dir, err)
			return false
		}
	}

	link := b.FileDownloadLink(file)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, link, nil)
	if err != nil {
		log.Error("[%s::%s] telegrambot: build download request: %s", f.Rule(), f.Name(), err)
		return false
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		log.Error("[%s::%s] telegrambot: download GET: %s", f.Rule(), f.Name(), err)
		return false
	}
	defer resp.Body.Close()
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		log.Error("[%s::%s] telegrambot: download non-2xx: %s", f.Rule(), f.Name(), resp.Status)
		return false
	}

	fh, err := os.Create(out)
	if err != nil {
		log.Error("[%s::%s] telegrambot: create %s: %s", f.Rule(), f.Name(), out, err)
		return false
	}
	if _, err := io.Copy(fh, resp.Body); err != nil {
		fh.Close()
		log.Error("[%s::%s] telegrambot: write %s: %s", f.Rule(), f.Name(), out, err)
		return false
	}
	fh.Close()

	msg.SetTarget("downloaded_path", out)
	return true
}

// OnEvent is called when an event occurs.
func (f *TelegramBot) OnEvent(event *data.Event) {}

func init() {
	register("telegrambot", NewTelegramBotFilter)
}

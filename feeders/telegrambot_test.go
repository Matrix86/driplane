package feeders

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/Matrix86/driplane/data"
	"github.com/asaskevich/EventBus"
	"github.com/go-telegram/bot/models"
)

func TestTelegramBotRegistered(t *testing.T) {
	if _, ok := feederFactories["telegrambotfeeder"]; !ok {
		t.Errorf("telegrambot feeder should be registered as 'telegrambotfeeder'")
	}
}

func TestTelegramBotTokenRequired(t *testing.T) {
	_, err := NewTelegramBotFeeder(map[string]string{})
	if err == nil {
		t.Errorf("expected error when token is missing")
	}
}

func TestTelegramBotMinimalConfig(t *testing.T) {
	feeder, err := NewTelegramBotFeeder(map[string]string{
		"telegrambot.token": "fake-token",
	})
	if err != nil {
		t.Fatalf("unexpected error: %s", err)
	}
	tb, ok := feeder.(*TelegramBot)
	if !ok {
		t.Fatal("cannot cast to *TelegramBot")
	}
	if tb.token != "fake-token" {
		t.Errorf("expected token 'fake-token', got '%s'", tb.token)
	}
	if tb.mode != "polling" {
		t.Errorf("expected default mode 'polling', got '%s'", tb.mode)
	}
	if tb.addr != ":3000" {
		t.Errorf("expected default addr ':3000', got '%s'", tb.addr)
	}
}

// newTestTelegramBot builds a feeder wired to a capture channel.
func newTestTelegramBot(t *testing.T, conf map[string]string) (*TelegramBot, chan *data.Message) {
	t.Helper()
	if _, ok := conf["telegrambot.token"]; !ok {
		conf["telegrambot.token"] = "fake-token"
	}
	feeder, err := NewTelegramBotFeeder(conf)
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}
	tb, ok := feeder.(*TelegramBot)
	if !ok {
		t.Fatal("cannot cast to *TelegramBot")
	}
	bus := EventBus.New()
	tb.setBus(bus)
	tb.setName("telegrambotfeeder")
	tb.setID(1)

	ch := make(chan *data.Message, 16)
	bus.Subscribe(tb.GetIdentifier(), func(m *data.Message) { ch <- m })
	return tb, ch
}

var _ = fmt.Sprintf // keep import around as helpers evolve

func TestTelegramBotAllowedUsersParsing(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "123, @alice , 456,@Bob",
	})
	if !tb.allowedUsers[123] || !tb.allowedUsers[456] {
		t.Errorf("expected numeric IDs 123, 456 in allowedUsers: %v", tb.allowedUsers)
	}
	if !tb.allowedUsernames["alice"] || !tb.allowedUsernames["bob"] {
		t.Errorf("expected usernames 'alice', 'bob' (lowercase): %v", tb.allowedUsernames)
	}
}

func TestTelegramBotAllowedChatsParsing(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_chats": "-100123, 456",
	})
	if !tb.allowedChats[-100123] || !tb.allowedChats[456] {
		t.Errorf("expected chat IDs -100123, 456: %v", tb.allowedChats)
	}
}

func TestTelegramBotEventsFilter(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.events": "message, command ,  callback_query",
	})
	if len(tb.events) != 3 {
		t.Errorf("expected 3 events enabled, got %d: %v", len(tb.events), tb.events)
	}
	if !tb.events["message"] || !tb.events["command"] || !tb.events["callback_query"] {
		t.Errorf("expected message+command+callback_query, got: %v", tb.events)
	}
	if tb.events["edited_message"] {
		t.Errorf("expected edited_message disabled")
	}
}

func TestTelegramBotCommandsParsing(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.commands": "/start, /help,/status",
	})
	expected := []string{"/start", "/help", "/status"}
	if len(tb.commands) != len(expected) {
		t.Fatalf("expected %d commands, got %d: %v", len(expected), len(tb.commands), tb.commands)
	}
	for i, c := range expected {
		if tb.commands[i] != c {
			t.Errorf("commands[%d]: expected '%s', got '%s'", i, c, tb.commands[i])
		}
	}
}

func TestTelegramBotMalformedAllowedUserSkipped(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "123,not_a_number,@alice",
	})
	if !tb.allowedUsers[123] {
		t.Errorf("expected 123 parsed: %v", tb.allowedUsers)
	}
	if len(tb.allowedUsers) != 1 {
		t.Errorf("expected 1 numeric ID, got: %v", tb.allowedUsers)
	}
	if !tb.allowedUsernames["alice"] {
		t.Errorf("expected @alice parsed: %v", tb.allowedUsernames)
	}
}

func TestTelegramBotWebhookModeRequiresURL(t *testing.T) {
	_, err := NewTelegramBotFeeder(map[string]string{
		"telegrambot.token": "fake",
		"telegrambot.mode":  "webhook",
	})
	if err == nil {
		t.Errorf("expected error when mode=webhook without webhook_url")
	}
}

func TestTelegramBotInvalidModeRejected(t *testing.T) {
	_, err := NewTelegramBotFeeder(map[string]string{
		"telegrambot.token": "fake",
		"telegrambot.mode":  "gibberish",
	})
	if err == nil {
		t.Errorf("expected error for invalid mode")
	}
}

func TestTelegramBotDetectType(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{})

	cases := []struct {
		name string
		u    *models.Update
		want string
	}{
		{"message", &models.Update{Message: &models.Message{Text: "hi"}}, "message"},
		{"command_slash", &models.Update{Message: &models.Message{Text: "/start"}}, "command"},
		{"command_slash_with_args", &models.Update{Message: &models.Message{Text: "/help me"}}, "command"},
		{"edited_message", &models.Update{EditedMessage: &models.Message{Text: "hi"}}, "edited_message"},
		{"channel_post", &models.Update{ChannelPost: &models.Message{Text: "news"}}, "channel_post"},
		{"edited_channel_post", &models.Update{EditedChannelPost: &models.Message{Text: "news"}}, "edited_channel_post"},
		{"callback_query", &models.Update{CallbackQuery: &models.CallbackQuery{Data: "x"}}, "callback_query"},
		{"chat_member", &models.Update{ChatMember: &models.ChatMemberUpdated{}}, "chat_member"},
		{"my_chat_member", &models.Update{MyChatMember: &models.ChatMemberUpdated{}}, "my_chat_member"},
		{"unknown", &models.Update{}, ""},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := tb.detectType(tc.u)
			if got != tc.want {
				t.Errorf("detectType(%s): want '%s', got '%s'", tc.name, tc.want, got)
			}
		})
	}
}

func TestTelegramBotDetectTypeCommandListMismatch(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.commands": "/start",
	})
	got := tb.detectType(&models.Update{Message: &models.Message{Text: "/help"}})
	if got != "message" {
		t.Errorf("when commands list is configured and command not in list, expected 'message', got '%s'", got)
	}
}

func TestSplitCommand(t *testing.T) {
	cases := []struct {
		in       string
		wantCmd  string
		wantArgs string
	}{
		{"/start", "/start", ""},
		{"/start arg1 arg2", "/start", "arg1 arg2"},
		{"/start@mybot", "/start", ""},
		{"/start@mybot arg1 arg2", "/start", "arg1 arg2"},
		{"hello", "hello", ""},
		{"", "", ""},
	}
	for _, tc := range cases {
		t.Run(tc.in, func(t *testing.T) {
			cmd, args := splitCommand(tc.in)
			if cmd != tc.wantCmd || args != tc.wantArgs {
				t.Errorf("splitCommand(%q) = (%q,%q), want (%q,%q)", tc.in, cmd, args, tc.wantCmd, tc.wantArgs)
			}
		})
	}
}

func TestTelegramBotAccessAllowAll(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{})
	if !tb.allowed(42, "alice", 100) {
		t.Errorf("empty allowlists should allow everyone")
	}
}

func TestTelegramBotAccessUserID(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "42",
	})
	if !tb.allowed(42, "alice", 100) {
		t.Errorf("user 42 should pass")
	}
	if tb.allowed(43, "alice", 100) {
		t.Errorf("user 43 should be blocked")
	}
}

func TestTelegramBotAccessUsername(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "@alice",
	})
	if !tb.allowed(0, "Alice", 100) {
		t.Errorf("username Alice should match @alice case-insensitively")
	}
	if tb.allowed(0, "bob", 100) {
		t.Errorf("username bob should be blocked")
	}
}

func TestTelegramBotAccessChats(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_chats": "-100",
	})
	if !tb.allowed(1, "x", -100) {
		t.Errorf("chat -100 should pass")
	}
	if tb.allowed(1, "x", 200) {
		t.Errorf("chat 200 should be blocked")
	}
}

func TestTelegramBotAccessUserAndChatCombined(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "42",
		"telegrambot.allowed_chats": "-100",
	})
	if !tb.allowed(42, "x", -100) {
		t.Errorf("user 42 in chat -100 should pass")
	}
	if tb.allowed(42, "x", 200) {
		t.Errorf("user 42 in chat 200 should be blocked (chat mismatch)")
	}
	if tb.allowed(43, "x", -100) {
		t.Errorf("user 43 in chat -100 should be blocked (user mismatch)")
	}
}

func TestTelegramBotAccessNoUserPresent(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "42",
		"telegrambot.allowed_chats": "-100",
	})
	if !tb.allowed(0, "", -100) {
		t.Errorf("no-sender event in allowed chat should pass")
	}
	if tb.allowed(0, "", 200) {
		t.Errorf("no-sender event in disallowed chat should be blocked")
	}
}

func TestTelegramBotExtractIdentityMessage(t *testing.T) {
	u := &models.Update{
		Message: &models.Message{
			From: &models.User{ID: 42, Username: "alice"},
			Chat: models.Chat{ID: -100, Type: models.ChatTypeGroup},
		},
	}
	uid, name, cid := extractIdentity(u)
	if uid != 42 || name != "alice" || cid != -100 {
		t.Errorf("message: got (%d,%s,%d), want (42,alice,-100)", uid, name, cid)
	}
}

func TestTelegramBotExtractIdentityCallback(t *testing.T) {
	u := &models.Update{
		CallbackQuery: &models.CallbackQuery{
			From: models.User{ID: 7, Username: "bob"},
		},
	}
	uid, name, cid := extractIdentity(u)
	if uid != 7 || name != "bob" || cid != 0 {
		t.Errorf("callback: got (%d,%s,%d), want (7,bob,0)", uid, name, cid)
	}
}

func TestTelegramBotExtractIdentityChannelPost(t *testing.T) {
	u := &models.Update{
		ChannelPost: &models.Message{
			Chat: models.Chat{ID: -200, Type: models.ChatTypeChannel},
		},
	}
	uid, name, cid := extractIdentity(u)
	if uid != 0 || name != "" || cid != -200 {
		t.Errorf("channel post: got (%d,%s,%d), want (0,,-200)", uid, name, cid)
	}
}

func TestTelegramBotExtractIdentityChatMember(t *testing.T) {
	u := &models.Update{
		ChatMember: &models.ChatMemberUpdated{
			From: models.User{ID: 9, Username: "carol"},
			Chat: models.Chat{ID: -300, Type: models.ChatTypeSupergroup},
		},
	}
	uid, name, cid := extractIdentity(u)
	if uid != 9 || name != "carol" || cid != -300 {
		t.Errorf("chat_member: got (%d,%s,%d), want (9,carol,-300)", uid, name, cid)
	}
}

func TestFillExtraFromUser(t *testing.T) {
	extra := map[string]interface{}{}
	u := &models.User{
		ID: 42, Username: "alice", FirstName: "Alice", LastName: "Smith",
		LanguageCode: "en", IsBot: false, IsPremium: true,
	}
	fillExtraFromUser(extra, u)
	if extra["user_id"] != "42" || extra["user_username"] != "alice" ||
		extra["user_firstname"] != "Alice" || extra["user_lastname"] != "Smith" ||
		extra["user_language"] != "en" || extra["user_isbot"] != "false" ||
		extra["user_ispremium"] != "true" {
		t.Errorf("unexpected extra: %v", extra)
	}
}

func TestFillExtraFromUserNil(t *testing.T) {
	extra := map[string]interface{}{}
	fillExtraFromUser(extra, nil)
	if len(extra) != 0 {
		t.Errorf("expected no keys for nil user: %v", extra)
	}
}

func TestFillExtraFromChat(t *testing.T) {
	extra := map[string]interface{}{}
	c := models.Chat{ID: -100, Type: models.ChatTypeGroup, Title: "Tst", Username: "grp"}
	fillExtraFromChat(extra, c)
	if extra["chat_id"] != "-100" || extra["chat_type"] != "group" ||
		extra["chat_title"] != "Tst" || extra["chat_username"] != "grp" {
		t.Errorf("unexpected extra: %v", extra)
	}
}

func TestFillExtraFromMessageText(t *testing.T) {
	extra := map[string]interface{}{}
	m := &models.Message{ID: 7, Date: 1700000000, Text: "hello"}
	fillExtraFromMessage(extra, m, false)
	if extra["msg_id"] != "7" || extra["msg_timestamp"] != "1700000000" ||
		extra["text"] != "hello" || extra["msg_edited"] != "false" ||
		extra["msg_hasmedia"] != "false" {
		t.Errorf("unexpected extra: %v", extra)
	}
}

func TestFillExtraFromMessagePhoto(t *testing.T) {
	extra := map[string]interface{}{}
	m := &models.Message{
		ID: 8, Date: 1700000001,
		Photo: []models.PhotoSize{
			{FileID: "small", FileUniqueID: "us", Width: 90, Height: 90, FileSize: 100},
			{FileID: "big", FileUniqueID: "ub", Width: 800, Height: 600, FileSize: 5000},
		},
		Caption: "nice pic",
	}
	fillExtraFromMessage(extra, m, false)
	if extra["msg_hasmedia"] != "true" {
		t.Errorf("expected msg_hasmedia true: %v", extra)
	}
	if extra["msg_mediatype"] != "photo" {
		t.Errorf("expected msg_mediatype photo, got %v", extra["msg_mediatype"])
	}
	if extra["msg_file_id"] != "big" {
		t.Errorf("expected largest photo FileID 'big', got %v", extra["msg_file_id"])
	}
	if extra["msg_mediasize"] != "5000" {
		t.Errorf("expected size 5000, got %v", extra["msg_mediasize"])
	}
	if extra["msg_caption"] != "nice pic" {
		t.Errorf("expected caption, got %v", extra["msg_caption"])
	}
}

func TestFillExtraFromMessageDocument(t *testing.T) {
	extra := map[string]interface{}{}
	m := &models.Message{
		ID: 9, Date: 1700000002,
		Document: &models.Document{
			FileID: "doc1", FileUniqueID: "udoc1",
			FileName: "report.pdf", MimeType: "application/pdf", FileSize: 2048,
		},
	}
	fillExtraFromMessage(extra, m, true)
	if extra["msg_edited"] != "true" {
		t.Errorf("expected msg_edited true")
	}
	if extra["msg_mediatype"] != "document" ||
		extra["msg_medianame"] != "report.pdf" ||
		extra["msg_mediaext"] != ".pdf" ||
		extra["msg_mediasize"] != "2048" {
		t.Errorf("unexpected doc extras: %v", extra)
	}
}

func TestOnUpdatePropagatesMessage(t *testing.T) {
	tb, ch := newTestTelegramBot(t, map[string]string{})
	tb.onUpdate(context.Background(), nil, &models.Update{
		ID: 1,
		Message: &models.Message{
			ID: 10, Date: 1700000000, Text: "hi",
			From: &models.User{ID: 42, Username: "alice"},
			Chat: models.Chat{ID: -100, Type: models.ChatTypeGroup},
		},
	})
	select {
	case msg := <-ch:
		if msg.GetMessage() != "hi" {
			t.Errorf("expected main 'hi', got %v", msg.GetMessage())
		}
		x := msg.GetExtra()
		if x["type"] != "message" {
			t.Errorf("expected type=message, got %v", x["type"])
		}
		if x["user_id"] != "42" || x["chat_id"] != "-100" {
			t.Errorf("missing identity keys: %v", x)
		}
	case <-time.After(time.Second):
		t.Fatal("timed out waiting for propagated message")
	}
}

func TestOnUpdatePropagatesCommand(t *testing.T) {
	tb, ch := newTestTelegramBot(t, map[string]string{
		"telegrambot.commands": "/start",
	})
	tb.onUpdate(context.Background(), nil, &models.Update{
		Message: &models.Message{
			ID: 11, Date: 1700000000, Text: "/start hello world",
			From: &models.User{ID: 1, Username: "u"},
			Chat: models.Chat{ID: 1, Type: models.ChatTypePrivate},
		},
	})
	select {
	case msg := <-ch:
		x := msg.GetExtra()
		if x["type"] != "command" {
			t.Errorf("expected type=command, got %v", x["type"])
		}
		if x["command"] != "/start" {
			t.Errorf("expected command=/start, got %v", x["command"])
		}
		if x["command_args"] != "hello world" {
			t.Errorf("expected command_args='hello world', got %v", x["command_args"])
		}
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

func TestOnUpdateDropsDisabledEvent(t *testing.T) {
	tb, ch := newTestTelegramBot(t, map[string]string{
		"telegrambot.events": "command",
	})
	tb.onUpdate(context.Background(), nil, &models.Update{
		Message: &models.Message{
			Text: "hello",
			From: &models.User{ID: 1, Username: "u"},
			Chat: models.Chat{ID: 1},
		},
	})
	select {
	case msg := <-ch:
		t.Errorf("expected no propagation, got: %v", msg.GetExtra())
	case <-time.After(100 * time.Millisecond):
	}
}

func TestOnUpdateDropsDisallowedUser(t *testing.T) {
	tb, ch := newTestTelegramBot(t, map[string]string{
		"telegrambot.allowed_users": "999",
	})
	tb.onUpdate(context.Background(), nil, &models.Update{
		Message: &models.Message{
			Text: "hello",
			From: &models.User{ID: 1, Username: "u"},
			Chat: models.Chat{ID: 1},
		},
	})
	select {
	case msg := <-ch:
		t.Errorf("expected drop, got: %v", msg.GetExtra())
	case <-time.After(100 * time.Millisecond):
	}
}

func TestOnUpdatePropagatesCallbackQuery(t *testing.T) {
	tb, ch := newTestTelegramBot(t, map[string]string{})
	tb.onUpdate(context.Background(), nil, &models.Update{
		CallbackQuery: &models.CallbackQuery{
			ID:           "cbq1",
			From:         models.User{ID: 5, Username: "dave"},
			Data:         "action:yes",
			ChatInstance: "instx",
		},
	})
	select {
	case msg := <-ch:
		x := msg.GetExtra()
		if x["type"] != "callback_query" {
			t.Errorf("expected type=callback_query, got %v", x["type"])
		}
		if msg.GetMessage() != "action:yes" {
			t.Errorf("expected main 'action:yes', got %v", msg.GetMessage())
		}
		if x["callback_id"] != "cbq1" {
			t.Errorf("expected callback_id cbq1, got %v", x["callback_id"])
		}
	case <-time.After(time.Second):
		t.Fatal("timed out")
	}
}

func TestTelegramBotStartPollingWithInvalidToken(t *testing.T) {
	tb, _ := newTestTelegramBot(t, map[string]string{
		"telegrambot.token": "definitely-not-a-real-token",
	})
	tb.Start()
	time.Sleep(50 * time.Millisecond)
	tb.Stop()
}

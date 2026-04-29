---
title: "TelegramBot"
date: 2026-04-23T00:00:00+00:00
draft: false
---

## TelegramBot

This feeder consumes updates from a bot registered with [@BotFather](https://t.me/BotFather) using the [Telegram Bot API](https://core.telegram.org/bots/api).
Unlike the `telegram` feeder (MTProto user client), `telegrambot` is suited for interactive bots that expose commands or inline keyboards.

Based on [go-telegram/bot](https://github.com/go-telegram/bot).

### Parameters

| Parameter | Type | Default | Description |
|---|---|---|---|
| **token** | _STRING_ | empty | Bot token issued by @BotFather. Required. |
| **mode** | _STRING_ | `polling` | `polling` or `webhook`. |
| **addr** | _STRING_ | `:3000` | Listen address (webhook mode only). |
| **webhook_url** | _STRING_ | empty | Public URL registered with Telegram (required if `mode=webhook`). |
| **webhook_secret** | _STRING_ | empty | Telegram `secret_token` for incoming webhook verification. |
| **delete_webhook_on_stop** | _BOOL_ | `false` | Unregister the webhook from Telegram on shutdown. |
| **commands** | _STRING_ | empty | Comma-separated command list (e.g. `/start,/help`). Empty = auto-detect any `/xxx`. |
| **allowed_users** | _STRING_ | empty | Comma-separated user IDs or `@usernames`. Empty = all. Usernames matched case-insensitively. When set, the sender MUST be in the list — non-listed users are rejected regardless of chat origin. |
| **allowed_chats** | _STRING_ | empty | Comma-separated chat IDs. Empty = all. When `allowed_users` is also set, private DMs from listed users bypass this list (so listed users can always DM the bot). |
| **events** | _STRING_ | all | Comma-separated subset of `message,command,callback_query,edited_message,channel_post,edited_channel_post,chat_member,my_chat_member`. |
| **debug** | _BOOL_ | `false` | Enable library debug logging. |

{{< notice info "Polling example" >}}
`tgRule => <telegrambot: token="123:abc", commands="/start,/help", allowed_users="@alice,@bob"> | ...`
{{< /notice >}}

{{< notice info "Webhook example" >}}
`tgRule => <telegrambot: token="123:abc", mode="webhook", webhook_url="https://example.com/webhook", webhook_secret="sekret"> | ...`
{{< /notice >}}

### Output

Every propagated message carries a `type` extra field identifying the update kind.

Common fields (when applicable):

| Name | Description |
| --- | --- |
| type | Event kind. |
| update_id | Telegram update ID. |
| user_id, user_username, user_firstname, user_lastname, user_language, user_isbot, user_ispremium | Sender info. |
| chat_id, chat_type, chat_title, chat_username | Chat info. |

#### Event `message` / `command`

| Name | Description |
| --- | --- |
| text | Message text (also the `main` field). |
| msg_id, msg_timestamp, msg_date, msg_time | Timing. |
| msg_edited | Always `false` for this event. |
| msg_caption | Caption, if present. |
| msg_reply_to_id | Replied-to message ID, if any. |
| msg_forward_from | Forwarded-from user ID (only when forward origin is a user). |
| msg_forward_from_chat | Forwarded-from chat/channel ID. |
| msg_hasmedia | `true` when media attached. |
| msg_mediatype | `photo`, `document`, `video`, `audio`, `voice`, `sticker`, `animation`. |
| msg_medianame, msg_mediaext, msg_mediasize | Media metadata. |
| msg_file_id, msg_file_unique_id | Bot API file identifiers. |
| command | Command (e.g. `/start`). Only on `type=command`. |
| command_args | Text after the command. Only on `type=command`. |

#### Event `edited_message` / `edited_channel_post`

Same fields as above with `msg_edited=true` and an additional `msg_edit_date` field.

#### Event `channel_post` / `edited_channel_post`

Message fields as above. `user_*` may be absent (channel posts can lack a sender).

#### Event `callback_query`

| Name | Description |
| --- | --- |
| callback_id | Callback query ID. |
| callback_data | Button payload (also the `main` field). |
| callback_chatinstance | Telegram chat instance identifier. |
| chat_id, chat_type, chat_title, chat_username | Chat the keyboard message lives in (also populated when the original message is no longer accessible). Absent only for `inline_message_id`-only callbacks. |
| msg_id | ID of the message that carried the inline keyboard (useful with `edit_message`). |

#### Event `chat_member` / `my_chat_member`

| Name | Description |
| --- | --- |
| member_old_status | Previous member status. |
| member_new_status | New member status. |
| member_user_id, member_user_username | Affected user. |

### Examples

{{< notice info "React to /start command from authorized users only" >}}
`tgRule => <telegrambot: token="123:abc", commands="/start", allowed_users="@alice,12345"> | text(target="command", pattern="/start") | log(msg="start received from {{ .user_username }}")`
{{< /notice >}}

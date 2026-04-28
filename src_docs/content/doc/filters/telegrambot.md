---
title: "TelegramBot"
date: 2026-04-28T00:00:00+00:00
draft: false
---

## TelegramBot

This filter performs Bot API actions (send/edit/delete a message, answer callback queries, send media, download attachments) on top of the [`telegrambot`](../../feeders/telegrambot/) feeder. It uses the `*bot.Bot` instance the feeder publishes via the `_telegrambot_api` extra, so a single bot connection serves both directions of traffic.

Note: it can be used only in a rule fed by the `telegrambot` feeder.

Based on [go-telegram/bot](https://github.com/go-telegram/bot).

### Parameters

| Parameter | Type | Default | Description |
|---|---|---|---|
| **action** | _STRING_ | `send_message` | One of `send_message`, `edit_message`, `delete_message`, `answer_callback`, `send_media`, `download_file`. |
| **to_chat** | _STRING_ | empty | Target chat ID. Optional: when omitted, falls back to the incoming `chat_id` extra (reply-to-sender). Templated. |
| **text** | _STRING_ | empty | Message text or callback notification text. Templated. Required for `send_message`, `edit_message`, `answer_callback`. |
| **caption** | _STRING_ | empty | Caption for `send_media`. Templated. |
| **parse_mode** | _STRING_ | `HTML` | One of `HTML`, `Markdown`, `MarkdownV2`, or empty (plain). Templated. |
| **message_id** | _STRING_ | empty | Required for `edit_message` and `delete_message`. Templated. |
| **callback_id** | _STRING_ | empty | `answer_callback`: optional explicit callback query ID. Falls back to incoming `callback_id` extra. Templated. |
| **file_id** | _STRING_ | empty | `download_file`: optional explicit file ID. Falls back to incoming `msg_file_id` extra. Templated. |
| **filename** | _STRING_ | empty | `download_file`: output path on disk. Required. Templated. The basename of the Telegram-side `file_path` is exposed as `{{ .msg_filename }}` so it can be reused in the path. |
| **media** | _STRING_ | empty | `send_media`: source. URL (`http(s)://...`) → Telegram fetches it; existing local path → uploaded; otherwise treated as a `file_id` to re-send. Templated. Required. |
| **media_type** | _STRING_ | `document` | `send_media`: one of `photo`, `document`, `video`, `audio`, `voice`, `animation`, `sticker`. Templated. |
| **reply_markup** | _STRING_ | empty | Raw Bot API JSON for `InlineKeyboardMarkup`. Templated, then parsed. |

All string parameters support [templates](../../rules/templates/).

### Required parameters per action

| Action | Required |
|---|---|
| `send_message` | `text` |
| `edit_message` | `text`, `message_id` |
| `delete_message` | `message_id` |
| `answer_callback` | `text` |
| `send_media` | `media` |
| `download_file` | `filename` |

### Output

On success the filter mutates the incoming message and propagates it. Original extras are preserved.

| Action | Extras added |
|---|---|
| `send_message` | `main` overwritten with rendered text; `sent_message_id`, `sent_chat_id`, `sent_text`, `sent_date`. |
| `edit_message` | `edit_success="true"`, `edited_message_id`, `edited_chat_id`. |
| `delete_message` | `delete_success="true"`. |
| `answer_callback` | `callback_answered="true"`. |
| `send_media` | `main` overwritten with rendered caption; `sent_message_id`, `sent_chat_id`, `sent_text` (caption), `sent_date`. |
| `download_file` | `downloaded_path` (rendered filename), `msg_filename` (basename of Telegram `file_path`). |

On any Bot API error, missing required extra, or template/JSON failure the filter logs the error and returns `false`, stopping rule propagation. The pipeline keeps running.

### Examples

{{< notice info "Echo `/start` command" >}}
`tgRule => <telegrambot: token="123:abc", commands="/start"> | text(target="command", pattern="/start") | <telegrambot: text="welcome {{ .user_username }}!">`
{{< /notice >}}

{{< notice info "Reply with inline keyboard" >}}
`<telegrambot: text="Pick one", reply_markup="{\"inline_keyboard\":[[{\"text\":\"yes\",\"callback_data\":\"y\"},{\"text\":\"no\",\"callback_data\":\"n\"}]]}">`
{{< /notice >}}

{{< notice info "Answer a callback query" >}}
`text(target="type", pattern="callback_query") | <telegrambot: action="answer_callback", text="got it">`
{{< /notice >}}

{{< notice info "Send a photo from a URL" >}}
`<telegrambot: action="send_media", media_type="photo", media="https://example.com/img.jpg", caption="hi">`
{{< /notice >}}

{{< notice info "Edit a previously-sent message" >}}
`<telegrambot: text="Working..."> | <telegrambot: action="edit_message", text="Done!", message_id="{{ .sent_message_id }}", to_chat="{{ .sent_chat_id }}">`
{{< /notice >}}

{{< notice info "Download an attachment" >}}
`text(target="msg_hasmedia", pattern="true") | <telegrambot: action="download_file", filename="/data/{{ .chat_id }}/{{ .msg_filename }}">`
{{< /notice >}}

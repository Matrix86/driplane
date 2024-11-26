---
title: "Telegram"
date: 2024-02-10T23:02:24+02:00
draft: false
---

## Telegram

This filter allows you to download files received from the Telegram feeder or send messages. 
Note: it can be used only in a rule with the Telegram Feeder.

### Parameters

The following parameters are required from this filter:

| Parameter  | Type     | Default        | Description                                                                  |
|------------|----------|----------------|------------------------------------------------------------------------------|
| **action** | _STRING_ | "send_message" | action to perform: "send_message", "download_file"                           |
| **to**     | _STRING_ | ""             | used with action `send_message`, it has to contain the recipient of the message: username, @username, phone number with the country code (supports [Golang templates](https://golang.org/pkg/text/template/)) |
| **to_chatid** | _STRING_ | empty | the recipient of the message will be a chat |
| **filename** | _STRING_ | "" | specified the path of where to store the downloaded file: `msg_filename` in the extra will contain the name of the file (supports [Golang templates](https://golang.org/pkg/text/template/)) |
| **text** | _STRING_ | "" | the text of the file (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 
Each `action` can use different parameters:

#### action = send_message

| Parameter  | Type     | Default | Description                                                                                                                |
|------------|----------|---------|----------------------------------------------------------------------------------------------------------------------------|
| **to**     | _STRING_ | ""             | used with action `send_message`, it has to contain the recipient of the message: username, @username, phone number with the country code (supports [Golang templates](https://golang.org/pkg/text/template/)) |
| **to_chatid** | _STRING_ | empty | the recipient of the message will be a chat |
| **text** | _STRING_ | "" | the text of the file (supports [Golang templates](https://golang.org/pkg/text/template/)) |

#### action = download_file

| Parameter    | Type     | Default  | Description |
|--------------|----------|----------|-------------|
| **filename** | _STRING_ | ""       | specified the path of where to store the downloaded file: `msg_filename` in the extra will contain the name of the file (supports [Golang templates](https://golang.org/pkg/text/template/)) |

{{< notice info "Example" >}} 
`... | telegram(action="send_message", to="@username", text="the file '{{ .msg_filename }}' received from {{ .user_username }} has been downloaded.") | ...`
{{< /notice >}}

### Examples

{{< notice info "Download file from a specific channel and send a private message to the user @username" >}} 
`telegramRule => <telegram: app_id="xxx", app_hash="yyyy", phone_number="+1123654789", session_folder="/tmp/sessions"> | text(target="chan_id", pattern="123456") | telegram(action="download_file", filename="/tmp/{{ .msg_filename }}") | telegram(action="send_message", to="@username", text="the file '{{ .msg_filename }}' received from {{ .user_username }} has been downloaded.")`
{{< /notice >}}

{{< notice info "Reply to a private message" >}} 
`telegramRule => <telegram: app_id="xxx", app_hash="yyyy", phone_number="+1123654789", session_folder="/tmp/sessions"> | text(target="main", pattern="help") | telegram(action="send_message", to="{{ .user_username }}", text="Hi {{ .user_username }}! You wrote the following message: {{ .main }}.")`
{{< /notice >}}
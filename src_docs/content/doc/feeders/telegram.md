---
title: "Telegram"
date: 2024-02-10T22:55:47+02:00
draft: false
---

## Telegram

This feeder creates a stream using the [Telegram API](https://core.telegram.org/schema). 
In order to use this feeder, you need to [create a Telegram Application](https://core.telegram.org/api/obtaining_api_id) and get app_id and app_hash.
If you set the session folder the session will be stored and only the first run it will ask for the code received from the phone number specified in the config.

Based on [gotd/td](https://github.com/gotd/td)

### Parameters

| Parameter              | Type     | Default                  | Description                                                                     |
|------------------------|----------|--------------------------|---------------------------------------------------------------------------------|
| **app_id**             | _STRING_ | empty                    | app ID (see [this](https://core.telegram.org/api/obtaining_api_id))             |
| **app_hash**           | _STRING_ | empty                    | App hash (see [this](https://core.telegram.org/api/obtaining_api_id))           |
| **phone_number**       | _STRING_ | empty                    | Phone number of the account to use (it should contain the country code)         |
| **session_folder**     | _STRING_ | empty                    | Path of the folder where storing the sessions                                   |
 
 _*events = "channel_message","chat_message"_
 
{{< notice info "Example" >}} 
`... | <telegram: app_id="xxx", app_hash="yyyy", phone_number="+1123654789", session_folder="/tmp/sessions"> | ...`
{{< /notice >}}

### Output

The messages propagated by this feeder could contain different info and they depend from the type of the event received.

Every propagated Message will have a `type` extra field with the name of the events (see the above list).

#### Event chat_message

We received a message from a group or a user.
The `main` field of the Message contain the `text` of the message.

| Name | Description |
| --- | --- |
| type | it contain `channel_message` for this event |
| msg_edited | "true" if it is an edit event of a message |
| text | text of the message |
| msg_hasmedia | "true" if the message contains a media (photo or doc) |
| chat_callactive | "true" if there is a call active |
| chat_creator | "true" if the user is the creator |
| chat_deactivated | "true" if the chat has been deactivated |
| chat_id | ID of the chat |
| chat_partecipantscount | number of partecipants |
| chat_title | title of the chat |
| chat_version | version of the chat |
| user_bot | "true" if the user is a bot |
| user_isclosefriend | "true" if the user is a close friend |
| user_iscontact | "true" if the user is a contact |
| user_isdeleted | "true" if the user has been deleted |
| user_isfake | "true" if the user has been flagged as fake/spam |
| user_id | ID of the user |
| user_accesshash | access hash of the user |
| user_mutualcontact | "true" if we are a contact of the user and viceversa |
| user_premium | "true" if the user has a premium account |
| user_verified | "true" if the user is verified |
| user_firstname | name of the user |
| user_lastname | lastname of the user |
| user_username | username of the user |
| user_language | language of the user |
| user_phone | phone number of the user |

#### Event channel_message

We received a message from a group or a user.
The `main` field of the Message contain the `text` of the message.

| Name | Description |
| --- | --- |
| type | it contain `channel_message` for this event |
| msg_edited | "true" if it is an edit event of a message |
| text | text of the message |
| msg_hasmedia | "true" if the message contains a media (photo or doc) |
| chan_broadcast | "true" if it is a broadcast channel |
| chan_callactive | "true" if there is a call active |
| chan_creator | "true" if the user is the creator |
| chan_fake | "true" if the channel has been flagged as fake |
| chan_forum | "true" if the channel is a forum |
| chan_gigagroup | "true" if the channel is a gigagroup |
| chan_hasgeo | "true" if the channel has geoposition |
| chan_haslink | "true" if the channel has a link |
| chan_id | ID of the channel |
| chan_hasJoinRequest | "true" if the users has to be approved by admins  |
| chan_ismegagroup | "true" if the channel is a megagroup |
| chan_isrestricted | "true" if the channel is restricted |
| chan_title | title of the channel |
| chan_verified | "true" if the channel is verified |
| chan_partecipantscount | number of partecipants |
| chan_username | username of the channel |
| user_bot | "true" if the user is a bot |
| user_isclosefriend | "true" if the user is a close friend |
| user_iscontact | "true" if the user is a contact |
| user_isdeleted | "true" if the user has been deleted |
| user_isfake | "true" if the user has been flagged as fake/spam |
| user_id | ID of the user |
| user_accesshash | access hash of the user |
| user_mutualcontact | "true" if we are a contact of the user and viceversa |
| user_premium | "true" if the user has a premium account |
| user_verified | "true" if the user is verified |
| user_firstname | name of the user |
| user_lastname | lastname of the user |
| user_username | username of the user |
| user_language | language of the user |
| user_phone | phone number of the user |

### Examples

{{< notice info "Download file from a specific channel and send a private message to the user @username" >}} 
`telegramRule => <telegram: app_id="xxx", app_hash="yyyy", phone_number="+1123654789", session_folder="/tmp/sessions"> | text(target="chan_id", pattern="123456") | telegram(action="download_file", filename="/tmp/{{ .msg_filename }}") | telegram(action="send_message", to="@username", text="the file '{{ .msg_filename }}' received from {{ .user_username }} has been downloaded.")`
{{< /notice >}}
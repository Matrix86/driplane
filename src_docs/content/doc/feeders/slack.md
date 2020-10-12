---
weight: 5
title: "Slack"
date: 2020-10-06T22:05:07+02:00
draft: false
---

## Slack

This feeder creates a stream using the [Slack Events API](https://api.slack.com/events-api). 
In order to use this feeder, you need to [create a Slack App](https://api.slack.com/apps), define its Bot Token scopes and finally enable the Event Subscriptions.
The feeder will start a webserver to receive the events, so it should be reachable by slack. If you are behind a NAT or Firewall, you can enable [localtunnel](https://theboroer.github.io/localtunnel-www/).
Based on [slack-go/slack](https://github.com/slack-go/slack)

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **token** | _STRING_ | empty | token Bot (starting with xoxb-*) |
 | **verification_token** | _STRING_ | empty | token to verify the requests (you can find it in the Basic information section) |
 | **addr** | _STRING_ | ":3000" | address and port in the form IP:PORT |
 | **lt_enable** | _STRING_ | empty | enable localtunnel for the server |
 | **lt_subdomain** | _STRING_ | empty | specify a subdomain to use with localtunnel |
 | **events** | _STRING_ | *events | comma separated event list to handle |
 | **ignore_bot** | _BOOL_ | true | if true the bot will ignore mentions and messages created by another bot |
 
 _*events = "app_mention","app_home_opened","app_uninstalled","grid_migration_finished","grid_migration_started","link_shared","message","member_joined_channel","member_left_channel","pin_added","pin_removed","reaction_added","reaction_removed","tokens_revoked","file_shared"_
 
{{< notice info "Example" >}} 
`... | <slack: token="xoxb-xxx", verification_token="xxxx", lt_enable="true", lt_subdomain="domaintesttt"> | ...`
{{< /notice >}}

### Output

The messages propagated by this feeder could contain different info and they depend from the type of the event it received.

Every propagated Message will have a `type` extra field with the name of the events (see the above list).

#### Event app_mention

The bot has been mentioned by someone.
The `main` field of the Message contain the `text` of the event.

| Name | Description |
| --- | --- |
| type | it contain `app_mention` for this event |
| user | user who triggered the events |
| text | text of the message |
| timestamp | timestamp |
| threadtimestamp | timestamp of the thread |
| channel | channel name |
| eventtimestamp | timestamp of the event |
| userteam | filled when a message comes from a channel that is shared between workspaces |
| sourceteam | filled when a message comes from a channel that is shared between workspaces |
| botid | filled out when a bot triggers the app_mention event |

#### Event app_home_opened

User clicked into your App Home.
The `main` field of the Message contains the type name of the event.

| Name | Description |
| --- | --- |
| type | it contain `app_home_opened` for this event |
| user | user who triggered the events |
| channel | channel name |
| eventtimestamp | timestamp of the event |
| tab | filled when a message comes from a channel that is shared between workspaces |

#### Event app_uninstalled

Your Slack app was uninstalled.
The `main` field of the Message contains the type name of the event.

| Name | Description |
| --- | --- |
| type | it contain `app_uninstalled` for this event |

#### Event grid_migration_finished

An enterprise grid migration has finished on this workspace.
The `main` field of the Message contains the type name of the event.

| Name | Description |
| --- | --- |
| type | it contain `grid_migration_finished` for this event |
| enterpriseid | enterprise id |

#### Event grid_migration_started

An enterprise grid migration has started on this workspace.
The `main` field of the Message contains the type name of the event.

| Name | Description |
| --- | --- |
| type | it contain `grid_migration_started` for this event |
| enterpriseid | enterprise id |

#### Event link_shared

A message was posted containing one or more links relevant to your application.
The `main` field of the Message contains the shared URL.

| Name | Description |
| --- | --- |
| type | it contain `link_shared` for this event |
| timestamp | timestamp of the message |
| threadtimestamp | timestamp of the thread |
| domain | domain of the shared link |
| link | shared link |

#### Event file_shared

A message was posted containing one or more links relevant to your application.
The `main` field of the Message contains the text of the message used to share the file.

| Name | Description |
| --- | --- |
| id | id of the file upload |
|  created |  | 
|  timestamp | timestamp of the event | 
|  name | name of the file | 
|  mimetype | mimetype of the file | 
|  filetype | extension of the file | 
|  prettytype |  | 
|  user | user who triggered the event | 
|  size | size of the file | 
|  urlprivatedownload | url to download the file | 
|  imageexifrotation |  | 
|  originalw | width of the file if it is an image | 
|  originalh | height of the file if it is an imange | 
|  permalink |  | 
|  permalinkpublic |  |

#### Event message

A message was posted as direct message or in a channel/group.
The `main` field of the Message contains the text of the message.

| Name | Description |
| --- | --- |
| type | it contain `message` for this event |
| timestamp | timestamp of the message |
| threadtimestamp | timestamp of the thread |
| botid | filled if the message has been sent by a bot |
| channel | channel id |
| channeltype | type of the channel (im, group, mpim, channel) |
| clientmsgid | id of the message |
| subtype | sub type of the message |
| text | text of the message |
| user | user who sent the message |
| username | filled if it is a bot_message |
| userteam | filled when the message comes from a channel that is shared between workspaces |
| sourceteam | filled when the message comes from a channel that is shared between workspaces |


#### Event member_joined_channel

An user just joined in a channel or it has been invited.
The `main` field of the Message contains the type of the message.

| Name | Description |
| --- | --- |
| type | it contain `member_joined_channel` for this event |
| user | user who has been invited |
| channel | channel id |
| channeltype | type of the channel (im, group, mpim, channel) |
| inviter | user who invited |
| team | filled when the message comes from a channel that is shared between workspaces |

#### Event member_left_channel

An user just left the channel.
The `main` field of the Message contains the type of the message.

| Name | Description |
| --- | --- |
| type | it contain `member_left_channel` for this event |
| user | user who has been invited |
| channel | channel id |
| channeltype | type of the channel (im, group, mpim, channel) |
| team | filled when the message comes from a channel that is shared between workspaces |

#### Event pin_added / pin_removed

A pin has been added or removed.
The `main` field of the Message contains the type of the message.

| Name | Description |
| --- | --- |
| type | it contain `pin_added` or `pin_removed` for this event |
| user | user who triggered the event |
| channel | channel id |
| eventtimestamp | type of the channel (im, group, mpim, channel) |
| haspins |  |
| item | it contains the object `slackevents.Item` that can be used from the slack filter |


#### Event reaction_added/reaction_removed

A member has added an emoji reaction to an item.
The `main` field of the Message contains the type of the message.

| Name | Description |
| --- | --- |
| type | it contain `reaction_added` or `reaction_removed` for this event |
| user | user who triggered the event |
| itemuser |  |
| eventtimestamp | timestamp of the event |
| item | it contains the object `slackevents.Item` that can be used from the slack filter |


### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
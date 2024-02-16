---
title: "Imap"
date: 2023-11-16T20:08:45+02:00
draft: false
---

## Imap

This feeder creates a stream starting from emails received on the account read by an IMAP client. It is possible to define how often the email account should be checked.
Every time the email inbox is parsed a Message is sent down the lane. 

### Parameters

| Parameter                | Type                                                     | Default | Description                                                                                    |
|--------------------------|----------------------------------------------------------|---------|---------------------------------------------------------------------|
| **host**                 | _STRING_                                                 | empty   | Host of the IMAP server                                             |
| **port**                 | _STRING_                                                 | empty   | Port of the IMAP server                                             |
| **username**             | _STRING_                                                 | empty   | Username of the account                                             |
| **password**             | _STRING_                                                 | empty   | Password of the account                                             |
| **mailbox**              | _STRING_                                                 | "INBOX" | Name of the mailbox to read                                         |
| **freq**                 | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | "1m"    | how often the email account should be checked                       |
| **start_from_beginning** | _BOOL_                                                   | "true"  | if "true" it reads all the emails in the mailbox from the beginning |
| **get_attachments**      | _BOOL_                                                   | "false" | if "true" it reads also the attachments                             |
 
{{< notice info "Example" >}} 
`... | <imap: host="imap.gmail.com", port="993", username="test@gmail.com", password="xxxxx", get_attachments="true", freq="30m"> ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the email's subject.

#### Extra

| Name               | Description                                                        |
|--------------------|--------------------------------------------------------------------|
| from               | List of the senders in the form <email@mail.com> Name              |
| to                 | List of the recipients in the form <email@mail.com> Name           |
| reply_to           | List of address in the "Reply-To" header                           |
| in_reply_to        | Parent Message-id                                                  |
| cc                 | List of the CC Header Addresses in the form <email@mail.com> Name  |
| bcc                | List of the BCC Header Addresses in the form <email@mail.com> Name |
| sender             | Message sender                                                     |
| message_id         | Message-Id of the current email                                    |
| date               | Message Date                                                       |
| subject            | Subject of the email                                               |
| is_attachment      | It is "true" if has the following 2 fields                         |
| attachment_filename| Name of the attachment                                             |
| attachment_body    | Binary content of the attachment                                   |

{{< notice warning "ATTENTION" >}} 
Not all the Extra field could be filled. If the relative tag is not present on the feed it will be empty.
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
---
title: "RSS"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## RSS

This feeder creates a stream starting from a feed `RSS`, `ATOM` or `JSON`. 
It is based on [gofeed](https://github.com/mmcdole/gofeed) so you can refer to it for more info and supported formats.

### Parameters

| Parameter                | Type                                                     | Default | Description                                                                                                            |
|--------------------------|----------------------------------------------------------|---------|------------------------------------------------------------------------------------------------------------------------|
| **url**                  | _STRING_                                                 | empty   | URL of the feed                                                                                                        |
| **freq**                 | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 60s     | how often the feed should be parsed                                                                                    |
| **start_from_beginning** | _BOOL_                                                   | "false" | if "true" it starts to parse the feed from the beginning (the first time it will ignore the pubdate field of the feed) |
| **ignore_pubdate**       | _BOOL_                                                   | "false" | if "true" it ignores the pubdate and it returns all the feed content every time                                        |
 
{{< notice info "Example" >}} 
`... | <rss: url="https://example.rss", freq="5s", start_from_beginning="false"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the `item.Title` string of the `gofeed.Item` struct.

#### Extra

| Name           | Description                                                                                                 |
|----------------|-------------------------------------------------------------------------------------------------------------|
| feed_title     | title of the feed ([feed.Title](https://github.com/mmcdole/gofeed#default-mappings))                        |
| feed_feedlink  | feed url ([feed.FeedLink](https://github.com/mmcdole/gofeed#default-mappings))                              |
| feed_updated   | time of the last update ([feed.Updated](https://github.com/mmcdole/gofeed#default-mappings))                |
| feed_published | date of publication ([feed.Published](https://github.com/mmcdole/gofeed#default-mappings))                  |
| feed_author    | author in the form `name <e-mail>` ([feed.Author.Name](https://github.com/mmcdole/gofeed#default-mappings)) |
| feed_language  | language of the feed ([feed.Language](https://github.com/mmcdole/gofeed#default-mappings))                  |
| feed_copyright | copyright ([feed.Copyright](https://github.com/mmcdole/gofeed#default-mappings))                            |
| feed_generator | generator used to create the feed ([feed.Generator](https://github.com/mmcdole/gofeed#default-mappings))    |

In addition to the feed tags, the Extra will also contain the item's fields. Since they could be different from feed to feed and it is possible to configure custom tag, <ins>you will find all them in the extra with their name</ins>. 

{{< notice warning "ATTENTION" >}} 
Not all the Extra field could be filled. If the relative tag is not present on the feed it will be empty.
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
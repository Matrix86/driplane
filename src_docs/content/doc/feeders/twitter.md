---
weight: 3
title: "Twitter"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Twitter

This feeder creates a stream from tweets. It is possible to define the keywords, or the users to follow.
Based on [go-twitter](https://github.com/dghubble/go-twitter)

### Parameters

| Parameter           | Type     | Default | Description                                                                             |
|---------------------|----------|---------|-----------------------------------------------------------------------------------------|
| **consumerKey**     | _STRING_ | empty   | [Twitter Auth](https://developer.twitter.com/en/docs/twitter-api/getting-started/guide) |
| **consumerSecret**  | _STRING_ | empty   | [Twitter Auth](https://developer.twitter.com/en/docs/twitter-api/getting-started/guide) |
| **accessToken**     | _STRING_ | empty   | [Twitter Auth](https://developer.twitter.com/en/docs/twitter-api/getting-started/guide) |
| **accessSecret**    | _STRING_ | empty   | [Twitter Auth](https://developer.twitter.com/en/docs/twitter-api/getting-started/guide) |
| **keywords**        | _STRING_ | empty   | comma separated keywords that should match on the tweets                                |
| **users**           | _STRING_ | empty   | comma separated users list                                                              |
| **languages**       | _STRING_ | empty   | filter by language                                                                      |
| **disable_retweet** | _BOOL_   | "false" | don't include retweets in the stream                                                    |
| **disable_quoted**  | _BOOL_   | "false" | don't include quoted tweets in the stream                                               |
 
{{< notice info "Example" >}} 
`... | <twitter: users="goofy, mickeymouse",keywords="movie, cartoon"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain:

 * the text of the tweet if it is a normal tweet;
 * the text of the retweeted message if it is a retweet;
 * the text of the quote if it is a quote tweet;

#### Extra

| Name              | Description                                                       |
|-------------------|-------------------------------------------------------------------|
| link              | link to the tweet                                                 |
| language          | language used for the tweet                                       |
| username          | author of the tweet                                               |
| quoted            | "true" if the tweet if a quoted tweet, "false" otherwise          |
| retweet           | "true" if the tweet if a retweeted tweet, "false" otherwise       |
| original_username | author of the original tweet if it is a retweet or quoted tweet   |
| original_language | language of the original tweet if it is a retweet or quoted tweet |
| original_status   | text of the original tweet if it is a retweet or quoted tweet     |
| original_link     | link of the original tweet if it is a retweet or quoted tweet     |

{{< notice warning "ATTENTION" >}} 
In some circumstances the extra field could be empty.
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
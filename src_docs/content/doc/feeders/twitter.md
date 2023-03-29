---
weight: 3
title: "Twitter"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Twitter

This feeder creates a stream from tweets. It is possible to define the keywords, or the users to follow.
Based on [go-twitter](github.com/g8rswimmer/go-twitter/v2)

### Parameters

| Parameter           | Type     | Default | Description                                                                                                                        |
|---------------------|----------|---------|------------------------------------------------------------------------------------------------------------------------------------|
| **bearerToken**     | _STRING_ | empty   | [Twitter Auth](https://developer.twitter.com/en/docs/twitter-api/getting-started/guide)                                            |
| **keywords**        | _STRING_ | empty   | comma separated keywords that should match on the tweets                                                                           |
| **users**           | _STRING_ | empty   | comma separated users list                                                                                                         |
| **rules**           | _STRING_ | empty   | set multiple custom rules separated by the char | and in the form tag_name:rule [documentation](https://developer.twitter.com/en/docs/twitter-api/tweets/filtered-stream/integrate/build-a-rule) |
| **languages**       | _STRING_ | empty   | filter by language (comma separated languages)                                                                                     |
| **disable_retweet** | _BOOL_   | "false" | don't include retweets in the stream                                                                                               |
| **disable_quoted**  | _BOOL_   | "false" | don't include quoted tweets in the stream                                                                                          |
 
{{< notice info "Example" >}} 
`... | <twitter: users="goofy, mickeymouse",keywords="movie, cartoon", rules="rule1:movie OR cartoon|rule2:mickey mouse OR pluto"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain:

 * the text of the tweet if it is a normal tweet;
 * the text of the retweeted message if it is a retweet;
 * the text of the quote if it is a quote tweet;

#### Extra

| Name              | Description                                                            |
|-------------------|------------------------------------------------------------------------|
| link              | link to the tweet                                                      |
| language          | language used for the tweet                                            |
| username          | author of the tweet                                                    |
| author_id         | ID of the author of the tweet                                          |
| quoted            | "true" if the tweet is a quoted tweet, "false" otherwise               |
| retweet           | "true" if the tweet is a retweeted tweet, "false" otherwise            |
| response          | "true" if the tweet is a response for another tweet, "false" otherwise |
| reply_for_user    | it contains the userID if the tweet is a reply for a user              |
| original_link     | link of the original tweet if it is a retweet or quoted tweet          |
| original_username | username of the tweet linked to the current one                        |
| original_name     | name of the author of the tweet linked to the current one              |
| original_text     | text of the tweet linked to the current one                            |
| original_userid   | ID of the original author of the tweet if it is a retweet or a quote   |
| matched_rules     | list of the matched rules (tags) comma separated                       |

{{< notice warning "ATTENTION" >}} 
In some cases the extra field could be empty.
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
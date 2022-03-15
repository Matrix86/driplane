---
weight: 4
title: "Web"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Web

This feeder creates a stream starting from a web page. It is possible to define how often the page should be downloaded and parsed.
Every time the page is parsed a Message is sent down the lane. 

### Parameters

| Parameter     | Type                                                     | Default | Description                                                                                    |
|---------------|----------------------------------------------------------|---------|------------------------------------------------------------------------------------------------|
| **url**       | _STRING_                                                 | empty   | URL of the web page                                                                            |
| **freq**      | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 60s     | how often the page should be parsed                                                            |
| **text_only** | _BOOL_                                                   | "false" | if "true" it removes all the tags from the page                                                |
| **method**    | _STRING_                                                 | "GET"   | HTTP method to use on the requests                                                             |
| **headers**   | _JSON_                                                   | empty   | Headers to use in the request                                                                  |
| **data**      | _JSON_                                                   | empty   | POST fields to send with the requests (it's not possible to use in combination with `rawData`) |
| **rawData**   | _STRING_                                                 | empty   | raw body of the requests (it's not possible to use in combination with `data`)                 |
| **status**    | _STRING_                                                 | empty   | the filter will propagate the Message only if the returned status is this                      |
| **cookies**   | _STRING_                                                 | empty   | Path of the JSON file containing the cookies to use                                            |
 
{{< notice info "Example" >}} 
`... | <web: url="https://example.com", freq="30m", status="200", cookies="/path/to/exported.json"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the HTML source or the text of the website if the `text_only` parameter is set to true.

#### Extra

| Name        | Description            |
|-------------|------------------------|
| url         | URL of the web page    |
| title       | meta tag `title`       |
| description | meta tag `description` |
| image       | meta tag `image`       |
| sitename    | meta tag `sitename`    |

{{< notice warning "ATTENTION" >}} 
Not all the Extra field could be filled. If the relative tag is not present on the feed it will be empty.
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
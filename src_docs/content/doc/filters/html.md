---
title: "Html"
date: 2022-07-19T23:38:02+02:00
draft: false
---

## Html

This filter is used to extract information from an HTML page received through the Message.
Like jQuery for JS, we can set a selector to find in the page, extract text from the tags and the html content (we are using [goquery library](https://github.com/PuerkitoBio/goquery) under the hood).

### Parameters

| Parameter    | Type     | Default | Description                                                                                      |
|--------------|----------|---------|--------------------------------------------------------------------------------------------------|
| **target**   | _STRING_ | "main"  | the field of the Message that should be used for the filter (it could be main or an extra field) |
| **selector** | _STRING_ | ""      | the selector to find in the HTML page                                                            |
| **get**      | _STRING_ | "html"  | what do we want to retrieve on the tags found in the selected one: `html`, `text`, `attr`        |
| **attr**     | _STRING_ | ""      | if get is `attr` you can define what attr name it should extract                                 |


{{< notice info "Example" >}}
`... | html(selector=".link", get="attr", attr="href") | ...`
{{< /notice >}}

### Output

The filter will generate one or more Messages. It is possible to use more than 1 time this filter.
The field `fulltext` will contain the original `target` string.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
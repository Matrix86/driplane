---
title: "Json"
date: 2023-01-29T18:57:01+02:00
draft: false
---

## Json

This filter is used to extract information from a JSON doc received through the Message.
It is possible to use XPath query for JSON specifying a selector to search in the doc and extract data (it uses [jsonquery](https://github.com/antchfx/jsonquery)).

### Parameters

| Parameter    | Type     | Default | Description                                                                                      |
|--------------|----------|---------|--------------------------------------------------------------------------------------------------|
| **target**   | _STRING_ | "main"  | the field of the Message that should be used for the filter (it could be main or an extra field) |
| **selector** | _STRING_ | ""      | the selector to find the data in the JSON                                                        |


{{< notice info "Example" >}}
`... | html(selector="id", target="doc") | ...`
{{< /notice >}}

### Output

The filter will generate one or more Messages. It is possible to use more than 1 time this filter.
The field `fulltext` will contain the original `target` string.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
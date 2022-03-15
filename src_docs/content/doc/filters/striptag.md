---
title: "Striptag"
date: 2021-07-01T15:35:53+02:00
draft: false
---

## StripTag

This filter is used to remove all the HTML tags from a string.

### Parameters

| Parameter  | Type     | Default | Description                                                                                      |
|------------|----------|---------|--------------------------------------------------------------------------------------------------|
| **target** | _STRING_ | "main"  | the field of the Message that should be used for the filter (it could be main or an extra field) |

 
{{< notice info "Example" >}} 
`... | striptag(target="main") | ...`
{{< /notice >}}

### Output

The `main` of the output `Message` will have the text extracted from the `target` config, stripped by all the HTML tags. 

A new extra field is created: `fulltext` will contain the original `target` string.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
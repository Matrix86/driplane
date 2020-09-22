---
weight: 13
title: "MIME type"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## MIME type

This filter allows you to detect the MIME type of a file and its extension. 
_Based on the [gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) library._ 

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be the `main` or and extra field) |
 | **filename** | _STRING_ | empty | the filename of the file to detect (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 
{{< notice info "Example" >}} 
`... | mime(target="{{ .extra_field }}") | ...`
{{< /notice >}}

### Output

The propagated Message will contain the mimetype's string in the `main` field and the extension in the extra field `mimetype_ext`.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
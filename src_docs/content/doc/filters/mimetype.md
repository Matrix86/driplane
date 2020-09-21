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
 | **filename** | _STRING_ | "main" | the filename of the file to detect (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 
{{< notice info "Example" >}} 
`... | mime(filename="{{ .main }}") | ...`
{{< /notice >}}

### Output

The propagated Message will contain the mimetype's string in the `main` field and the extension in the extra field `mimetype_ext`.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
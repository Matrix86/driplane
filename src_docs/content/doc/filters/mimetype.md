---
weight: 13
title: "MIME type"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## MIME type

This filter allows you to detect the MIME type of a file and its extension. 
###### Based on the [gabriel-vasile/mimetype](https://github.com/gabriel-vasile/mimetype) library. 

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or and extra field) |
 
{{< notice info "Example" >}} 
`... | mime(target="main") | ...`
{{< /notice >}}

### Output

The propagated Message will contain the mimetype's string in the `main` field and the extension in the extra field `mimetype_ext`.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
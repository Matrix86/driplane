---
title: "Url"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Url

This filter is used to search or extract URLs from a Message.
Currently supported types of URLs are:
 * http/s
 * ftp

### Parameters

| Parameter   | Type     | Default | Description                                                                                      |
|-------------|----------|---------|--------------------------------------------------------------------------------------------------|
| **target**  | _STRING_ | "main"  | the field of the Message that should be used for the filter (it could be main or an extra field) |
| **http**    | _BOOL_   | "true"  | if "false", `http` scheme urls are ignored                                                       |
| **https**   | _BOOL_   | "true"  | if "false", `https` scheme urls are ignored                                                      |
| **ftp**     | _BOOL_   | "true"  | if "false", `ftp` scheme urls are ignored                                                        |
| **extract** | _BOOL_   | "true"  | if "true", the `main` field of the propagated Message will contain the found URL                 |

 
{{< notice info "Example" >}} 
`... | url(extract="true") | ...`
{{< /notice >}}

### Output

If the `extract` parameter is "false", the received Message will be propagated only if at least one URL is found in it. 
Otherwise, if `extract` is "true" and the Message contains one or more URLs, the `main` field of the propagated Message will contain only the extracted URLs.

If the `extract` parameter is "true" and the message is propagated, a new extra field is created: `fulltext` will contain the original `target` string.

{{< notice warning "ATTENTION" >}} 
If the targeted field contains multiple URLs, the filter will create and propagate multiple messages, one for each URL. 
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
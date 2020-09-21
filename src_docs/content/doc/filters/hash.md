---
weight: 5
title: "Hash"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Hash

This filter is used to search or extract hashes from a Message.
Supported types of hashes are:
 * MD5
 * SHA1
 * SHA256
 * SHA512

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or an extra field) |
 | **extract** | _BOOL_ | "false" | if `"true"` the main field of the output Message will be the extracted hash |
 | **md5** | _BOOL_ | "true" | if `"false"` md5 hashes will be ignored |
 | **sha1** | _BOOL_ | "true" | if `"false"` sha1 hashes will be ignored |
 | **sha256** | _BOOL_ | "true" | if `"false"` sha256 hashes will be ignored |
 | **sha512** | _BOOL_ | "true" | if `"false"` sha512 hashes will be ignored |

 
{{< notice info "Example" >}} 
`... | hash(target="description", extract="true") | ...`
{{< /notice >}}

### Output

If the `extract` parameter is "false", the received Message will be propagated only if at least a hash is found in it. 
Otherwise if `extract` is "true" and the Message contains one or more hashes, the `main` field of the propagated Message will contain only the extracted hash.

If the `extract` parameter is "true" and the message is propagated, a new extra field is created: `fulltext` will contain the original `target` string.

{{< notice warning "ATTENTION" >}} 
If the targeted field contains multiple hashes, the filter will create and propagate multiple messages, one for each hash. 
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
---
weight: 1
title: "Cache"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Cache

This filter introduces a simple cache mechanism in the rule. It is a TTL based cache and it can have a local visibility (cache is only visible to the current filter) or a global visibility (cache shared across **ALL** the rules).
If the target of the Message as input has been cached before, and his TTL is not expired, it will be dropped and not propagated to the next filter. 
Otherwise if the target of the Message is new to the cache, it is inserted in it and propagated to the next filter.

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or and extra field) |
 | **refresh_on_get** | _BOOL_ | "true" | the TTL is refreshed if the key has been looked up |
 | **ttl** | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 24h | how long after the key will be deleted |
 | **global** | _BOOL_ | "false" | make this cache global |
 
{{< notice info "Example" >}} 
`... | cache(ttl="24h", global="true") | ...`
{{< /notice >}}

### Output

The output is not being changed. This filter can only stop or not the propagation of the Message.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
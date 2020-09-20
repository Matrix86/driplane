---
weight: 2
title: "Changed"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Changed

This filter is similar to the cache. It can only stop the propagation of the Message across the lane, but only if the target of the received Message is different from the previous one.  

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or and extra field) |
 
{{< notice info "Example" >}} 
`... | changed(target="original_author") | ...`
{{< /notice >}}

### Output

The output is not be changed. This filter can only stop or not the propagation of the Message.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
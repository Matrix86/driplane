---
title: "Override"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Override

This filter allows you to change a field of a Message, before sending it to the next filter. Template can be used.

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **name** | _STRING_ | empty | name of the field to change (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 | **value** | _STRING_ | empty | new value to assign to the Message's field specified (supports [Golang templates](https://golang.org/pkg/text/template/)) |

 
{{< notice info "Example" >}} 
`... | override(name="description", value="{{ .title }}") | ...`
{{< /notice >}}

### Output

The propagated Message will be identical to the original, with only the specified field changed.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
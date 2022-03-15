---
title: "Echo"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Echo

This filter prints the Message on the logs. It is mostly used to debug the rules.  

### Parameters

| Parameter | Type   | Default | Description                     |
|-----------|--------|---------|---------------------------------|
| **extra** | _BOOL_ | "false" | print also all the extra fields |
 
{{< notice info "Example" >}} 
`... | echo(extra="false") | ...`
{{< /notice >}}

### Output

The output is not being changed. This filter prints the received Message and send it to the next filter in the rule.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
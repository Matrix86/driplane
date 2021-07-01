---
title: "System"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## System

This filter allows you to exec a command on the host machine. The received Message can be used to create the command to launch.
It supports [Golang templates](https://golang.org/pkg/text/template/) (only text.Template).

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **cmd** | _STRING_ | empty | command line to exec for each received Message (supports [Golang templates](https://golang.org/pkg/text/template/)) |

 
{{< notice info "Example" >}} 
`... | system(cmd="echo '{{ .author }} wrote {{ .main }}' >> logs.txt") | ...`
{{< /notice >}}

### Output

The propagated Message will contain the output of the command if it is provided, and it is not failed.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
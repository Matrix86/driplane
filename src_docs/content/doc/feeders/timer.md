---
weight: 0
title: "Timer"
date: 2021-02-20T18:06:02+02:00
draft: false
---

## Timer feeder

This feeder trigger a pipeline every time the timer is fired.

### Parameters

| Parameter | Type | Default | Description
 | --- | --- | --- | --- |
| **freq** | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 60s | The intervals (in duration) on how often to execute the pipeline |
 

{{< notice info "Example" >}}
`<timer: freq="30s"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain time in rfc3339 format.

#### Extra

| Name | Description |
| --- | --- |
| rfc3339 | time in rfc3339 format |
| timestamp | time in Unix timestamp |

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}  
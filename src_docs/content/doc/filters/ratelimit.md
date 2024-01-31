---
title: "RateLimit"
date: 2023-01-31T18:50:23+02:00
draft: false
---

## RateLimit

This filter allows you to set a rate limit on the messages that can go through it. So for example if we don't want to limit the number of messages in a pipe to 5 messages per second we just to set the parameter `rate` to 5.

### Parameters

| Parameter    | Type     | Default | Description                                                                                      |
|--------------|----------|---------|--------------------------------------------------------------------------------------------------|
| **rate**     | _STRING_ | "0"     | how many event per second you want to have as rate limiter                                       |


{{< notice info "Example" >}}
`... | ratelimit(rate="5") | ...`
{{< /notice >}}

### Output

The filter will slow down the rate of the messages as specified on the parameter `rate`.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
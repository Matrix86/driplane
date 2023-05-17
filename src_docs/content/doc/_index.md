---
title: "Driplane Documentation"
description: "Documentation of Driplane project"
date: 2020-01-11T14:09:21+09:00
---

# Introduction

`Driplane` allows you to create an automatic alerting system or start automated tasks triggered by events.
You can keep under control a stream source as Twitter, a file, a RSS feed or a website.
It includes a mini language that allows you to define the filtering pipelines (the rules), and it is extensible thanks to an integrated JS plugin system. 

With `driplane` you can create several rules starting from one or more streams, filter the content in the pipeline and launch tasks or send alerts if some event occurred.

The documentation can be found [HERE](https://matrix86.github.io/driplane/doc/)

## How it works

The user can define one or more rules. Usually a rule contains a source (`feeder`), which takes care of getting the information and sending updates (`Message`) through the pipeline, and one or more `filters`.
The filters' job is to choose whether to propagate or not the `Message` to the next filter in the pipeline relying on a _condition_, or change the `Message` received before to propagate it. The `Message` will be propagated only if it verifies the condition.

## Use cases

Using `driplane` it is possible to:

 * keep track of keywords or users on Twitter, receive the new tweets or quoted tweets from them, search for URLs or particular strings in them and send a Telegram or a Slack message through their webhooks.
 * keep track of a RSS feed or a website, and download and store on file all the new changes to them.
 * keep track of changes on a file, and launch alert if a particular condition is verified.
 
The rules and the JS plugins allow you to create very complex custom logics.
  
## Usage

```
Usage of ./bin/driplane:
  -config string
    	Set configuration file.
  -debug
    	Enable debug logs.
  -dry-run
    	Only test the rules syntax.
  -help
    	This help.
  -js string
    	Path of the js plugins.
  -rules string
    	Path of the rules' directory.
```

{{< alert theme="success" >}}
 \> driplane -config /path/to/config.yml
{{< /alert >}}


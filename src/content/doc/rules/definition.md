---
weight: 1
title: "First look"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## First look

`Driplane` include a simple language to define where get the stream and what operation exec, or how to filter the data.

In the rule could be defined 3 type of nodes:

* `FEEDER` : it is the node responsible for creating of a stream of data (read every changes on a file, get tweets from Twitter, etc..)

* `FILTER` : it receives data from a feeder or another filter and checks some conditions on them or makes some changes on them. It send the data to the next filter **ONLY** if the condition is verified.

* `RULE CALL` : every rule has a name, so you can define a rule with a preset feeder or a pipe of filters and connect to another filter/feeder.

It is possible to define custom parameters for feeders or filters. Each one of them has a different type of parameters that can change their behaviour and you can find a list of them in the related section.
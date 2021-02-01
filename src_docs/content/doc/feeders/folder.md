---
weight: 0
title: "Folder"
date: 2021-02-01T13:38:02+02:00
draft: false
---

## File feeder

This feeder can create the stream of fsnotify events for a given folder.

### Parameters

| Parameter | Type | Default | Description
 | --- | --- | --- | --- |
| **name** | _STRING_ | empty | the path of the folder that it has to keep track |

{{< notice info "Example" >}}
`... | <folder: name="/tmp"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the filename while the `op` extra the type of event.

#### Extra

| Name | Description |
| --- | --- |
| op | type of event |

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}  
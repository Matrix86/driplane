---
weight: 0
title: "Folder"
date: 2021-02-01T13:38:02+02:00
draft: false
---

## Folder feeder

This feeder can create the stream of fsnotify events for a given folder or for cloud platform storage like Amazon S3, 
Google Drive and Dropbox.

The feeder use [cloudwatcher](https://github.com/Matrix86/cloudwatcher) to keep track of changes on the chosen directory.

### Parameters

| Parameter | Type | Default | Description
 | --- | --- | --- | --- |
| **name** | _STRING_ | empty | the path of the folder that it has to keep track |
| **type** | _STRING_ | "local" | the type of service to use `local`, `dropbox`, `gdrive`, `s3` or `git` |
| **freq** | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 2s | how often the directory should be checked for updates |

{{< alert theme="warning" >}}
Some services like Gdrive, S3 and Dropbox require additional configurations (you can check them from [here](https://github.com/Matrix86/cloudwatcher/blob/main/README.md)).
You can pass them using the config file ([more here](https://matrix86.github.io/driplane/doc/configuration/)) OR
define them in the rule itself: `<folder: name="/", type="gdrive", client_id="xxx", client_secret="yyy", token="zzz">`
{{< /alert >}} 

{{< notice info "Example" >}}
`<folder: name="/tmp", type="local"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the filename while the `op` extra the type of event.

#### Extra

| Name | Description |
| --- | --- |
| op | type of event: `FileCreated`, `FileChanged`, `FileDeleted`, `TagsChanged` |
| size | for some events you can find the size of the file that triggered the event |

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}  
---
title: "File"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## File feeder

This feeder can create the stream starting from a file. Like the `tail -f` command it opens the specified file and propagates a data message if a line is being added to the file.

### Parameters

| Parameter    | Type     | Default | Description                                                            |
|--------------|----------|---------|------------------------------------------------------------------------|
| **filename** | _STRING_ | empty   | the path of the file that it has to keep track                         |
| **toend**    | _BOOL_   | "false" | the feeder will start to create data messages only for new added lines |
 
{{< notice info "Example" >}} 
`... | <file: filename="path/of/the/file.txt", toend="false"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the new read line.

#### Extra

| Name      | Description               |
|-----------|---------------------------|
| file_name | the name of the read file |

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}  
---
weight: 1
title: "File"
date: 2022-03-12T15:33:42+01:00
draft: false
---

## File

This filter get as input the path of a local file, read it and return the content back to the pipeline.

### Parameters

| Parameter  | Type     | Default | Description                                                                                           |
|------------|----------|---------|-------------------------------------------------------------------------------------------------------|
| **target** | _STRING_ | "main"  | the field of the Message that should be used for filter's output (it could be main or an extra field) |

{{< notice info "Example" >}}
`... | file(target="file_content") | ...`
{{< /notice >}}

### Output

The output will contain the content of the read file. Using the `target` parameter you can specify the output field's name.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
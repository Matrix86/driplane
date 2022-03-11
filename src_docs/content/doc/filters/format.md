---
title: "Format"
date: 2020-09-16T22:38:02+02:00 draft: false
---

## Format

This filter is used to format the received Message. It is based on
the [Golang templates](https://golang.org/pkg/text/template/) and it can load templates from the `template_path`
directory specified in the configuration file.

### Parameters

| Parameter | Type | Default | Description
 | --- | --- | --- | --- |
| **type** | _STRING_ | "text" | specify the type of template to use : `"text"` or `"html"` |
| **template** | _STRING_ | empty | a template could be specified directly here, instead of load it from file |
| **file** | _STRING_ | empty | load the template from file |
| **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or an extra field) |

In the template is allowed to use all the fields of the received Message: main or extra.

{{< notice info "Example" >}}
`... | format(type="html", template="main : {{.main}} extra : {{.file_name}}") | ...`
{{< /notice >}}

### Output

The new formatted text is sent to the next filter in the `main` field of the Message. The extra fields do not undergo
changes.

### Examples

{{< alert theme="warning" >}} Soon... {{< /alert >}} 
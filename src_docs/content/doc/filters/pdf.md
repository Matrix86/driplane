---
weight: 14
title: "Pdf"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Pdf

This filter allows you to extract plain text from a PDF file. 
_Based on the [ledongthuc/pdf](https://github.com/ledongthuc/pdf) library._

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be the `main` or and extra field) |
 | **filename** | _STRING_ | empty | the filename of the PDF file to parse (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 
{{< notice info "Example" >}} 
`... | pdf(target="{{ .extra_field }}") | ...`
{{< /notice >}}

{{< notice warning "ATTENTION" >}} 
The `filename` field override the `target`. They are mutually exclusive, so you can specify only one of them.
{{< /notice >}}

### Output

The propagated Message contains the plain text of the input PDF file (`fulltext` will be set to the file name received as input). 

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
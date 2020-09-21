---
weight: 14
title: "Pdf"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Pdf

This filter allows you to extract plain text from a PDF file. 
###### Based on the [ledongthuc/pdf](https://github.com/ledongthuc/pdf) library. 

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or and extra field) |
 
{{< notice info "Example" >}} 
`... | pdf(target="main") | ...`
{{< /notice >}}

### Output

The propagated Message contains the plain text of the input PDF file (`fulltext` will also set to the file name received as input). 

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
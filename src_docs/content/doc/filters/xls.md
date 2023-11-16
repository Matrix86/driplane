---
title: "XLS"
date: 2023-11-16T20:08:45+02:00
draft: false
---

## XLS

This filter allows you to extract all the rows from a Excel file. 
_Based on the [qax-os/excelize](https://github.com/qax-os/excelize) library._

### Parameters

| Parameter    | Type     | Default | Description                                                                                                |
|--------------|----------|---------|------------------------------------------------------------------------------------------------------------|
| **target**   | _STRING_ | "main"  | the field of the Message that should be used for the filter (it could be the `main` or and extra field)    |
| **filename** | _STRING_ | empty   | the filename of the XLS file to parse (supports [Golang templates](https://golang.org/pkg/text/template/)) |
 
{{< notice info "Example" >}} 
`... | xls(target="{{ .extra_field }}") | ...`
{{< /notice >}}

{{< notice warning "ATTENTION" >}} 
The `filename` field override the `target`. They are mutually exclusive, so you can specify only one of them.
{{< /notice >}}

### Output

The filter produces one Message for each row of the XLS file. 

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
---
title: "Elasticsearch"
date: 2022-03-12T15:40:23+01:00
draft: false
---

## ElasticSearch

This filter writes on ElasticSearch all the message it receives as input and return as output the document ID.
You can also specify what message's field it should use as input.

### Parameters

| Parameter  | Type     | Default | Description                                                                                           |
|------------|----------|---------|-------------------------------------------------------------------------------------------------------|
| **target** | _STRING_ | "main"  | the field of the Message that should be used for filter's output (it could be main or an extra field) |

{{< notice info "Example" >}}
`... | elasticsearch(target="input_field") | ...`
{{< /notice >}}

### Output

The `main` field will contain the docID returned by ElasticSearch.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
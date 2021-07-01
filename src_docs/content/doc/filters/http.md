---
title: "Http"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## HTTP

This filter allows you to send HTTP requests. When a `Message` arrives to the filter, we can decide if use the `main` field of the `Message` as URL on the request, or its content for the HTTP data.
This behaviour can be handled with the parameters.

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **url** | _STRING_ | empty | URL of the web page. It is possible use the [Golang templates](https://golang.org/pkg/text/template/) to use fields of the `Message` |
 | **download_to** | _STRING_ | empty | path of where to download the file. It is possible use the [Golang templates](https://golang.org/pkg/text/template/) to use fields of the `Message` | 
 | **text_only** | _BOOL_ | "false" | if "true" it removes all the tags from the body response |
 | **method** | _STRING_ | "GET" | HTTP method to use on the request |
 | **headers** | _JSON_ | empty | Headers to use in the request |
 | **data** | _JSON_ | empty | POST fields to send with the request (it's not possible to use in combination with `rawData`) |
 | **rawData** | _STRING_ | empty | raw body of the request (it's not possible to use in combination with `data`) |
 | **status** | _STRING_ | empty | the filter will propagate the Message only if the returned status has the specified value |
 | **cookies** | _STRING_ | empty | Path of the JSON file containing the cookies to use |

 
{{< notice info "Example" >}} 
`... | http(url="{{ .main }}", cookies="exported.json", headers="{\"Content-type\": \"application/json\"}") | ...`
{{< /notice >}}

### Output

If the request was successful, the output `Message` will have the `main` field set to the HTTP body response. If the `status` is set, and the response http status is different from it, the `Message` will be dropped.

{{< notice warning "ATTENTION" >}} 
The `Message` is dropped if the request is failed. 
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
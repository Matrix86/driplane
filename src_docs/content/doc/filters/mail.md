---
title: "Mail"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Mail

This filter allows you to send an e-mail. When a `Message` arrives to the filter, we can decide if use the `main` field of the `Message` as URL on the request, or its content for the HTTP data.
This behaviour can be handled with the parameters.

### Parameters

| Parameter    | Type     | Default | Description                                                                                              |
|--------------|----------|---------|----------------------------------------------------------------------------------------------------------|
| **body**     | _STRING_ | empty   | the body of the e-mail (supports [Golang templates](https://golang.org/pkg/text/template/))              |
| **username** | _STRING_ | empty   | username for the host authentication                                                                     |
| **password** | _STRING_ | empty   | password for the host authentication                                                                     |
| **host**     | _STRING_ | empty   | host server used to send the e-mail                                                                      |
| **port**     | _STRING_ | empty   | port of the host server                                                                                  |
| **fromAddr** | _STRING_ | empty   | source e-mail address                                                                                    |
| **fromName** | _STRING_ | empty   | source name address                                                                                      |
| **to**       | _STRING_ | empty   | destination e-mail address (supports multi-destination, comma separated)                                 |
| **subject**  | _STRING_ | empty   | subject field of the e-mail to send                                                                      |
| **use_auth** | _BOOL_   | "false" | if "true" the sendmail server will receive the credentials specified in `username` and `password` fields |
 
 
{{< notice info "Example" >}} 
`... | http(url="{{ .main }}", cookies="exported.json", headers="{\"Content-type\": \"application/json\"}") | ...`
{{< /notice >}}

{{< notice success "Remember" >}} 
Every default's value can be set in the configuration, creating a section with the name of the`Filter`/`Feeder`. 
[more info](../../configuration)
{{< /notice >}}

### Output

The input `Message` is always propagated to the next filter without changes.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
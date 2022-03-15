---
title: "Apt"
date: 2022-03-15T11:38:24+01:00
draft: false
---

## Apt feeder

This feeder can create the stream starting from an apt [repository](https://wiki.debian.org/it/DebianRepository). It supports also flat repositories.
It is possible to specify the frequency of the quering and receive a message every time a new package is published.

### Parameters

| Parameter    | Type                                                     | Default  | Description                                                                                 |
|--------------|----------------------------------------------------------|----------|---------------------------------------------------------------------------------------------|
| **url**      | _STRING_                                                 | empty    | URL of the apt repo                                                                         |
| **freq**     | _[DURATION](https://golang.org/pkg/time/#ParseDuration)_ | 60s      | how often the feed should be parsed                                                         |
| **suite**    | _STRING_                                                 | "stable" | suite of the repo to keep under control                                                     |
| **arch**     | _STRING_                                                 | empty"   | architecture of the repo, if empty the first arch returned by the Release file will be used |
| **index**    | _STRING_                                                 | empty    | URL of the Packages file (it overrides the url parameter)                                   |
| **insecure** | _BOOL_                                                   | false    | allow repository with insecure certificates                                                 |

{{< notice info "Example" >}}
`... | <apt: url="http://apt.modmyi.com/dists/stable/Release", insecure="true", freq="3h"> | ...`
{{< /notice >}}

### Output

#### Text

The `main` field of the Message will contain the filename of the package and all the other field will be present in extra fields.

#### Extra

List of the supported field that will be returned as extra field.

| Name           |
|----------------|
| Filename       |
| Size           |
| MD5sum         |
| SHA1           |
| SHA256         |
| DescriptionMD5 |
| Depends        |
| InstalledSize  |
| Package        |
| Architecture   |
| Version        |
| Section        |
| Maintainer     |
| Homepage       |
| Description    |
| Tag            |
| Author         |
| Name           |

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}  
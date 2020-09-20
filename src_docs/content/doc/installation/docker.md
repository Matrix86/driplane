---
weight: 2
title: "Docker"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Docker

`driplane` is containerized using a really lightweight Linux distribution called **Alpine Linux**.

To pull the latest image version:

{{< alert theme="success" >}}
docker pull matrix86/driplane
{{< /alert >}}

To run it:

{{< alert theme="success" >}}
docker run --rm -v config:/app/config -it matrix86/driplane:latest -config config/config.yaml
{{< /alert >}}

where the `config` directory contains the `config.yaml` file, the `rule` directory, the `js` directory and the `templates` directory.
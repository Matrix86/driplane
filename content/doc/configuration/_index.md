---
weight: 2
title: "Configuration"
date: 2020-09-16T22:38:03+02:00
draft: false
---

## Configuration

Driplane needs a yaml file to work. An example can be found [here](https://github.com/Matrix86/driplane/blob/master/config.yaml.example).

The configuration file could have several sections, but the mandatory one is the `general` one. It specifies all the paths needed by Driplane to find js files, templates and rules. 

> It contains the paths of rules, templates and javascript plugins.

```yaml
general:
  log_path: "" # if nothing is specified it prints logs on stdout
  rules_path: "rules" # path containing the rules
  js_path: "js" # path of the js plugins
  templates_path: "templates" # path of templates
  debug: false # if true enable the debug logs
```

<ins>In the configuration is it possible to define default params for _Feeders_ and _Filters_.</ins> In this way we don't need to specify that configuration in the rules.

> For the twitter feeder we can set the keys one time.
```yaml
twitter:
  consumerKey: "consumerKey",
  consumerSecret: "consumerSecret",
  accessToken: "accessToken",
  accessSecret: "accessSecret",
  keywords: "#italy #coding #malware something",
  stallWarnings: "true"
```  
> Each config will be visible __ONLY__ to the related filter or feeder.

We can also define `custom` configurations and they'll be available to all the feeders and filters.

```yaml
custom:
  customKey: "This is a custom configuration"
```
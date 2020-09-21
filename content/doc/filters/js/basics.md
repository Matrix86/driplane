---
weight: 1
title: "Basics"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## JS

This filter allows to extend the basic driplane's filter, defining Javascript scripts. It is based on [islazy/plugin](https://github.com/evilsocket/islazy) that is based on [robertkrimen/otto](https://github.com/robertkrimen/otto).
Defining a JS file with our custom logic, it is possible create a complex filter.

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **path** | _STRING_ | empty | path of the Javascript file (it can contains multiple functions) |
 | **function** | _STRING_ | empty | name of the function in the JS file to call when a `Message` is received |

{{< notice info "Example" >}} 
`... | js(path="script.js", function="MyFunction") | ...`
{{< /notice >}}

### Output

The output `Message` of this filter depends on the return value of the JS function itself. 

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}}
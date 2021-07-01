---
title: "Random"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Random

This filter is used to inject an extra field with a random number in the propagated `Message`.  

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **output** | _STRING_ | "main" | the field of the propagated `Message` that will contain the random number |
 | **min** | _STRING_ | "0" | the min value of the extracted number is `min` |
 | **max** | _STRING_ | "999999" | the max value of the extracted number is `max` |

{{< notice info "Example" >}} 
`... | random(output="random_field", min="9", max="90") | ...`
{{< /notice >}}

### Output

The output `Message` will be equal to the input, but it will also include the new random field.

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
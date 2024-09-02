---
title: "Number"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Number

This filter allows you to treat a string from the input Message as numeric value and apply some operator on it (for example, check if this field contains a number greater than X).

### Parameters

| Parameter   | Type     | Default | Description                                                                                                                                    |
|-------------|----------|---------|------------------------------------------------------------------------------------------------------------------------------------------------|
| **target**  | _STRING_ | "main"  | the field of the Message that should be used for the filter (it could be main or and extra field)                                              |
| **op**  | _STRING_   | ""   | Compare operator to use for the numeric value (">", ">=", "<", "<=", "!=", "==") |
| **value** | _STRING_   | "" | It has to be a numeric value and it is the number to use for the comparison      |

 
{{< notice info "Example" >}} 
`... | number(target="num_field", op=">=", value="44") | ...`
{{< /notice >}}

### Output

If the comparison is verified the received Message will be propagated. 

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
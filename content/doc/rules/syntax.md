---
weight: 2
title: "Syntax"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Syntax

The syntax of the driplane's rules is really simple. Like for the BASH every filter's output is sended to the next filter's input concatenating the two of them with the `|` character.
All the rules have to start with a name and end with a `;` char.

### Rule Name and Rule Call

Each rule have to start with an identifier follow by `=>`. This identifier identifies _rule name_ and it could be used in another rule to concatenate 2 rule together.

{{< notice warning "ATTENTION" >}} 
This name has to be **unique** in the rules.
{{< /notice >}}

> Example:
> `IDENTIFIER => ... ;`

It can be included in another rule (**rule call**) prepending the `@` to it.

> Example:
> `IDENTIFIER1 => ... ;`
> `IDENTIFIER2 => ... | @IDENTIFIER1 | ... ;`

### Feeder

The feeder creates the stream, so they don't accept inputs. For this they can be positioned **ONLY** to the beginning of a rule.

The feeder definition starts with a `<` char followed by an identifier. That's the type of the feeder we want to use.
 
After the type we found a `:` char followed by a list of **parameters** comma separated and a `>`.

{{< notice info "Parameters" >}} 
The parameters are in the form of key/value where the value is between double quotes `key="value"`.
{{< /notice >}}

> Example:
> `IDENTIFIER => <FEEDER_TYPE: param1="value1", param2="value2"> | ... ;` 

### Filter

The filter are the main operators of a rule, because they decide if a data is interesting and perform operations. 

The definition of a filter start with his name and it is followed by parameters contained between `(` and `)`.
According to the settings a Filter can change his behaviour and can modify the data passing through it.

> Example:
> `IDENTIFIER => ... | FILTER_TYPE( param1="value1", ... ) | ... ;`

{{< notice info "NOT" >}} 
The operator **NOT** `!` can be used on filter to negate his result (propagate the data if the condition is not verified).
It has to be put before the filter definition: `!FILTER_TYPE(...)`
{{< /notice >}}

{{< notice warning "ATTENTION" >}} 
All the parameters have to be enclosed in quotes. 
JSON requires **double quotes** to encode strings, so in order to define a JSON string you need to escape the quotes `\"`.  
{{< /notice >}}

### Data message and Extra

The data stream in driplane is based on text and the basic object that is part of it is the _Message_. 
The _Message_ is an object that contains the text that need to be filtered and extra.
The main string is identified as `text` in the filters, whereas the extra data are identified by a key.

There are fixed extra, created from driplane itself and other extra relative to a feeder or filter.

| Name | Description |
| --- | --- |
| source_feeder | the name of the feeder creates this Message |
| rule_name | the name of the rule that contains this filter |
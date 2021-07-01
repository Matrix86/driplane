---
title: "Text"
date: 2020-09-16T22:38:02+02:00
draft: false
---

## Text

This filter searches or extracts strings from the received Message. It can be used with a regular expression or a simple string.
If the string is found, the condition is matched and the Message is propagated to the next filter.

### Parameters

 | Parameter | Type | Default | Description 
 | --- | --- | --- | --- |
 | **target** | _STRING_ | "main" | the field of the Message that should be used for the filter (it could be main or and extra field) |
 | **regexp** | _BOOL_ | false | the pattern field is a regular expression |
 | **extract** | _BOOL_ | "false" | if "true" the `main` field of the propagated Message will contain the extracted string (it can be used only if `regexp` parameter is set true) |
 | **pattern** | _STRING_ | empty |specifies the pattern that should be matched on the Message to check the condition |

 
{{< notice info "Example" >}} 
`... | text(target="description", pattern="(#[^\\s]+)", regexp="true", extract="true") | ...`
{{< /notice >}}

### Output

If the `extract` parameter is "false", the received Message will be propagated only if the specified `pattern` is matched in the `target` field of the Message. 
Otherwise if `extract` is "true" (only `regexp` can be used in this case), and one or more strings matches with the pattern, the `main` field of the propagated Message will contain only the matched string.

If the `extract` parameter is "true" and the message is propagated, a new extra field is created: `fulltext` will contain the original `target` string.

{{< notice warning "ATTENTION" >}} 
If the targeted field contains multiple matches, the filter will create and propagate multiple Messages, one for each matched string. 
{{< /notice >}}

### Examples

{{< alert theme="warning" >}}
Soon...
{{< /alert >}} 
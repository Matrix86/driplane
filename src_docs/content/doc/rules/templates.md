---
weight: 3
title: "Templates"
date: 2026-04-17T00:00:00+00:00
draft: false
---

## Templates

Many filters in `driplane` support **Go templates** in their parameters. Templates allow you to dynamically compose strings using the fields of the current `Message` (both the `main` field and any `extra` fields).

Templates are based on the standard [Go text/template](https://golang.org/pkg/text/template/) package and are extended with all the functions from the [Sprig](https://masterminds.github.io/sprig/) library, giving you access to over 100 useful functions for string manipulation, math, date formatting, encoding, and more.

### Basic syntax

Inside a template, you can reference any field of the `Message` using the `{{ .field_name }}` syntax:

- `{{ .main }}` - the main text of the Message
- `{{ .file_name }}` - an extra field called `file_name`
- `{{ .user_username }}` - an extra field called `user_username`

{{< notice info "Example" >}}
`... | format(template="Author: {{ .author }} wrote: {{ .main }}") | ...`
{{< /notice >}}

### Sprig functions

All [Sprig functions](https://masterminds.github.io/sprig/) are available in templates. Functions can be called using the pipe `|` operator or the function call syntax.

#### String functions

| Function | Description | Example | Result |
|----------|-------------|---------|--------|
| `upper`  | Convert to uppercase | `{{ .main \| upper }}` | `HELLO` |
| `lower`  | Convert to lowercase | `{{ .main \| lower }}` | `hello` |
| `trim`   | Remove leading/trailing whitespace | `{{ .main \| trim }}` | `hello` |
| `trimPrefix` | Remove prefix | `{{ .main \| trimPrefix "http://" }}` | `example.com` |
| `trimSuffix` | Remove suffix | `{{ .main \| trimSuffix ".txt" }}` | `file` |
| `replace` | Replace occurrences | `{{ .main \| replace " " "_" }}` | `hello_world` |
| `contains` | Check if string contains substring | `{{ if contains "error" .main }}...{{ end }}` | |
| `hasPrefix` | Check prefix | `{{ if hasPrefix "http" .main }}...{{ end }}` | |
| `hasSuffix` | Check suffix | `{{ if hasSuffix ".pdf" .main }}...{{ end }}` | |
| `repeat` | Repeat a string N times | `{{ .main \| repeat 3 }}` | `hellohellohello` |
| `substr` | Extract substring | `{{ substr 0 5 .main }}` | `hello` |
| `nospace` | Remove all whitespace | `{{ .main \| nospace }}` | `helloworld` |
| `trunc` | Truncate to length | `{{ .main \| trunc 5 }}` | `hello` |
| `abbrev` | Abbreviate with ellipsis | `{{ .main \| abbrev 10 }}` | `hello w...` |
| `quote` | Wrap in double quotes | `{{ .main \| quote }}` | `"hello"` |
| `squote` | Wrap in single quotes | `{{ .main \| squote }}` | `'hello'` |
| `title` | Convert to title case | `{{ .main \| title }}` | `Hello World` |
| `snakecase` | Convert to snake_case | `{{ .main \| snakecase }}` | `hello_world` |
| `camelcase` | Convert to camelCase | `{{ .main \| camelcase }}` | `HelloWorld` |
| `kebabcase` | Convert to kebab-case | `{{ .main \| kebabcase }}` | `hello-world` |

#### Encoding functions

| Function | Description | Example |
|----------|-------------|---------|
| `b64enc` | Base64 encode | `{{ .main \| b64enc }}` |
| `b64dec` | Base64 decode | `{{ .main \| b64dec }}` |
| `sha256sum` | SHA256 hash | `{{ .main \| sha256sum }}` |
| `sha1sum` | SHA1 hash | `{{ .main \| sha1sum }}` |

#### Default and conditional functions

| Function | Description | Example |
|----------|-------------|---------|
| `default` | Provide a default value | `{{ .name \| default "unknown" }}` |
| `empty` | Check if value is empty | `{{ if empty .main }}no content{{ end }}` |
| `coalesce` | Return first non-empty value | `{{ coalesce .name .username "anonymous" }}` |
| `ternary` | Ternary operator | `{{ ternary "yes" "no" (contains "ok" .main) }}` |

#### List functions

| Function | Description | Example |
|----------|-------------|---------|
| `list` | Create a list | `{{ list "a" "b" "c" }}` |
| `join` | Join list elements | `{{ list "a" "b" "c" \| join "," }}` |
| `split` | Split string into map | `{{ (split "," .main)._0 }}` |
| `splitList` | Split string into list | `{{ .main \| splitList "," }}` |

#### Math functions

| Function | Description | Example |
|----------|-------------|---------|
| `add` | Addition | `{{ add 1 2 }}` |
| `sub` | Subtraction | `{{ sub 10 3 }}` |
| `mul` | Multiplication | `{{ mul 2 5 }}` |
| `div` | Division | `{{ div 10 2 }}` |
| `max` | Maximum value | `{{ max 1 5 3 }}` |
| `min` | Minimum value | `{{ min 1 5 3 }}` |

#### Date functions

| Function | Description | Example |
|----------|-------------|---------|
| `now` | Current time | `{{ now }}` |
| `date` | Format date | `{{ now \| date "2006-01-02" }}` |
| `dateModify` | Modify a date | `{{ now \| dateModify "-24h" }}` |

### Chaining functions

Functions can be chained using the pipe operator:

{{< notice info "Examples" >}}
`{{ .main | lower | trim | replace " " "_" }}` - lowercase, trim whitespace, replace spaces with underscores

`{{ .main | upper | trunc 50 }}` - uppercase then truncate to 50 characters

`{{ .name | default "anonymous" | upper }}` - use default if empty, then uppercase
{{< /notice >}}

### Conditionals

Go templates support conditional logic:

{{< notice info "Examples" >}}
`{{ if contains "error" .main }}ALERT: {{ .main }}{{ else }}{{ .main }}{{ end }}`

`{{ if and (not (empty .name)) (contains "@" .name) }}{{ .name }}{{ end }}`
{{< /notice >}}

### Filters supporting templates

The following filters accept Go templates (with Sprig functions) in one or more of their parameters:

| Filter | Parameters supporting templates |
|--------|--------------------------------|
| [Format]({{< ref "/doc/filters/format" >}}) | `template`, `file` |
| [Override]({{< ref "/doc/filters/override" >}}) | `name`, `value` |
| [System]({{< ref "/doc/filters/system" >}}) | `cmd` |
| [Http]({{< ref "/doc/filters/http" >}}) | `url`, `download_to`, `data` values, `rawData` |
| [Mail]({{< ref "/doc/filters/mail" >}}) | `body` |
| [Slack]({{< ref "/doc/filters/slack" >}}) | `to`, `text`, `filename`, `url` |
| [Telegram]({{< ref "/doc/filters/telegram" >}}) | `to`, `to_chatid`, `filename`, `text` |
| [LLM]({{< ref "/doc/filters/llm" >}}) | `prompt`, `system_prompt` |
| [Pdf]({{< ref "/doc/filters/pdf" >}}) | `filename` |
| [XLS]({{< ref "/doc/filters/xls" >}}) | `filename` |
| [MIME type]({{< ref "/doc/filters/mimetype" >}}) | `filename` |

### More information

For the complete list of Sprig functions, visit the [Sprig documentation](https://masterminds.github.io/sprig/).

For the full Go template syntax, see the [Go text/template documentation](https://golang.org/pkg/text/template/).

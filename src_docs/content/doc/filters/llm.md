---
title: "LLM"
date: 2026-03-06T00:00:00+00:00
draft: false
---

## LLM

This filter sends the received `Message` to a Large Language Model (LLM) and propagates the response. It uses [any-llm-go](https://github.com/mozilla-ai/any-llm-go) to support multiple LLM providers through a unified interface.

Both the `prompt` and `system_prompt` parameters support [Golang templates](https://golang.org/pkg/text/template/), allowing you to dynamically compose prompts using the `main` field and any extra fields of the incoming `Message`.

### Supported Providers

OpenAI, Anthropic, Ollama, DeepSeek, Groq, Mistral, Gemini, llama.cpp, Llamafile.

### Parameters

| Parameter         | Type     | Default  | Description                                                                                                                                          |
|-------------------|----------|----------|------------------------------------------------------------------------------------------------------------------------------------------------------|
| **model**         | _STRING_ |          | **(required)** The model name to use (e.g. `"gpt-4"`, `"claude-3-opus-20240229"`, `"llama3"`)                                                       |
| **prompt**        | _STRING_ |          | **(required)** The user prompt sent to the LLM. Supports [Golang templates](https://golang.org/pkg/text/template/) with `Message` fields             |
| **system_prompt** | _STRING_ | empty    | An optional system prompt. Supports [Golang templates](https://golang.org/pkg/text/template/) with `Message` fields                                  |
| **provider**      | _STRING_ | "openai" | The LLM provider to use: `openai`, `anthropic`, `ollama`, `deepseek`, `groq`, `mistral`, `gemini`, `llamacpp`, `llamafile`                           |
| **api_key**       | _STRING_ | empty    | API key for the provider (required for cloud providers like OpenAI, Anthropic, etc.)                                                                 |
| **api_url**       | _STRING_ | empty    | Custom base URL for the API endpoint (useful for proxies or self-hosted instances)                                                                   |
| **target**        | _STRING_ | "main"   | The field of the `Message` where the LLM response will be stored. Use `"main"` to replace the message content, or any other name to set an extra field |
| **temperature**   | _FLOAT_  | 0.7      | Sampling temperature for the model (higher values produce more random output)                                                                        |
| **max_tokens**    | _INT_    | 1024     | Maximum number of tokens to generate in the response                                                                                                 |

{{< notice info "Example" >}}
`... | llm(model="gpt-4", prompt="Summarize: {{ .main }}", api_key="sk-...", provider="openai") | ...`
{{< /notice >}}

### Output

The LLM response text is placed in the field specified by `target` (default: `main`). The following extra fields are set on the output `Message`:

| Extra Field              | Description                                  |
|--------------------------|----------------------------------------------|
| **llm_model**            | The model name returned by the provider      |
| **llm_prompt_tokens**    | Number of tokens in the prompt               |
| **llm_completion_tokens**| Number of tokens in the completion           |
| **llm_total_tokens**     | Total tokens used (prompt + completion)      |
| **llm_raw_response**     | The full JSON response from the provider     |

{{< notice warning "ATTENTION" >}}
The `Message` is dropped if the LLM request fails or returns no response choices.
{{< /notice >}}

### Examples

Using OpenAI to summarize text:

{{< notice info "Example" >}}
`... | llm(model="gpt-4", prompt="Summarize the following text:\n{{ .main }}", system_prompt="You are a helpful assistant.", api_key="sk-...", provider="openai") | ...`
{{< /notice >}}

Using a local Ollama instance:

{{< notice info "Example" >}}
`... | llm(model="llama3", prompt="{{ .main }}", provider="ollama", api_url="http://localhost:11434") | ...`
{{< /notice >}}

Using templates with extra fields and a custom target:

{{< notice info "Example" >}}
`... | llm(model="gpt-4", prompt="Translate '{{ .main }}' to {{ .language }}", target="translation", api_key="sk-...", provider="openai") | ...`
{{< /notice >}}

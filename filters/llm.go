package filters

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"text/template"

	"github.com/Matrix86/driplane/data"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers"
	"github.com/mozilla-ai/any-llm-go/providers/anthropic"
	"github.com/mozilla-ai/any-llm-go/providers/deepseek"
	"github.com/mozilla-ai/any-llm-go/providers/gemini"
	"github.com/mozilla-ai/any-llm-go/providers/groq"
	"github.com/mozilla-ai/any-llm-go/providers/llamacpp"
	"github.com/mozilla-ai/any-llm-go/providers/llamafile"
	"github.com/mozilla-ai/any-llm-go/providers/mistral"
	"github.com/mozilla-ai/any-llm-go/providers/ollama"
	"github.com/mozilla-ai/any-llm-go/providers/openai"

	"github.com/evilsocket/islazy/log"
)

// providerFactory is a function that creates a provider from config options
type providerFactory func(opts ...anyllm.Option) (providers.Provider, error)

// supportedProviders maps provider names to their constructors
var supportedProviders = map[string]providerFactory{
	"openai": func(opts ...anyllm.Option) (providers.Provider, error) {
		return openai.New(opts...)
	},
	"anthropic": func(opts ...anyllm.Option) (providers.Provider, error) {
		return anthropic.New(opts...)
	},
	"ollama": func(opts ...anyllm.Option) (providers.Provider, error) {
		return ollama.New(opts...)
	},
	"deepseek": func(opts ...anyllm.Option) (providers.Provider, error) {
		return deepseek.New(opts...)
	},
	"groq": func(opts ...anyllm.Option) (providers.Provider, error) {
		return groq.New(opts...)
	},
	"mistral": func(opts ...anyllm.Option) (providers.Provider, error) {
		return mistral.New(opts...)
	},
	"gemini": func(opts ...anyllm.Option) (providers.Provider, error) {
		return gemini.New(opts...)
	},
	"llamacpp": func(opts ...anyllm.Option) (providers.Provider, error) {
		return llamacpp.New(opts...)
	},
	"llamafile": func(opts ...anyllm.Option) (providers.Provider, error) {
		return llamafile.New(opts...)
	},
}

// LLM is a Filter that sends the input Message to a Large Language Model
// and propagates the response. It uses github.com/mozilla-ai/any-llm-go
// to support multiple LLM providers (OpenAI, Anthropic, Ollama, DeepSeek,
// Groq, Mistral, Gemini, llama.cpp, Llamafile).
type LLM struct {
	Base

	provider    providers.Provider
	model       string
	target      string
	temperature float64
	maxTokens   int

	prompt       *template.Template
	systemPrompt *template.Template

	params map[string]string
}

// NewLLMFilter is the registered method to instantiate an LLMFilter
func NewLLMFilter(p map[string]string) (Filter, error) {
	f := &LLM{
		params:      p,
		target:      "main",
		temperature: 0.7,
		maxTokens:   1024,
	}
	f.cbFilter = f.DoFilter

	// model (required)
	if v, ok := p["model"]; ok {
		f.model = v
	} else {
		return nil, fmt.Errorf("llmfilter: 'model' parameter is required")
	}

	// prompt: Go template for the user message (required)
	if v, ok := p["prompt"]; ok {
		t, err := template.New("llmFilterPrompt").Parse(v)
		if err != nil {
			return nil, fmt.Errorf("llmfilter: error parsing prompt template: %s", err)
		}
		f.prompt = t
	} else {
		return nil, fmt.Errorf("llmfilter: 'prompt' parameter is required")
	}

	// system_prompt: Go template for system message (optional)
	if v, ok := p["system_prompt"]; ok {
		t, err := template.New("llmFilterSystemPrompt").Parse(v)
		if err != nil {
			return nil, fmt.Errorf("llmfilter: error parsing system_prompt template: %s", err)
		}
		f.systemPrompt = t
	}

	// target: which field to set the response in (default "main")
	if v, ok := p["target"]; ok {
		f.target = v
	}

	// temperature
	if v, ok := p["temperature"]; ok {
		temp, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return nil, fmt.Errorf("llmfilter: invalid temperature value '%s': %s", v, err)
		}
		f.temperature = temp
	}

	// max_tokens
	if v, ok := p["max_tokens"]; ok {
		mt, err := strconv.Atoi(v)
		if err != nil {
			return nil, fmt.Errorf("llmfilter: invalid max_tokens value '%s': %s", v, err)
		}
		f.maxTokens = mt
	}

	// Build the any-llm-go provider
	providerName := "openai"
	if v, ok := p["provider"]; ok {
		providerName = v
	}

	if _, ok := supportedProviders[providerName]; !ok {
		names := make([]string, 0, len(supportedProviders))
		for k := range supportedProviders {
			names = append(names, k)
		}
		return nil, fmt.Errorf("llmfilter: unsupported provider '%s' (supported: %v)", providerName, names)
	}

	// Build config options for the provider
	var opts []anyllm.Option
	if v, ok := p["api_key"]; ok && v != "" {
		opts = append(opts, anyllm.WithAPIKey(v))
	}
	if v, ok := p["api_url"]; ok && v != "" {
		opts = append(opts, anyllm.WithBaseURL(v))
	}

	provider, err := supportedProviders[providerName](opts...)
	if err != nil {
		return nil, fmt.Errorf("llmfilter: error creating %s provider: %s", providerName, err)
	}
	f.provider = provider

	return f, nil
}

// DoFilter is the mandatory method used to "filter" the input data.Message
func (f *LLM) DoFilter(msg *data.Message) (bool, error) {
	// Render the prompt template
	promptText, err := msg.ApplyPlaceholder(f.prompt)
	if err != nil {
		return false, fmt.Errorf("llmfilter: error rendering prompt: %s", err)
	}

	// Render the system prompt template if present
	systemText := ""
	if f.systemPrompt != nil {
		systemText, err = msg.ApplyPlaceholder(f.systemPrompt)
		if err != nil {
			return false, fmt.Errorf("llmfilter: error rendering system_prompt: %s", err)
		}
	}

	log.Debug("[%s::%s] prompt: '%s', system: '%s'", f.Rule(), f.Name(), promptText, systemText)

	// Build messages
	messages := make([]anyllm.Message, 0, 2)
	if systemText != "" {
		messages = append(messages, anyllm.Message{
			Role:    anyllm.RoleSystem,
			Content: systemText,
		})
	}
	messages = append(messages, anyllm.Message{
		Role:    anyllm.RoleUser,
		Content: promptText,
	})

	// Build completion params
	temp := f.temperature
	maxTok := f.maxTokens
	params := anyllm.CompletionParams{
		Model:       f.model,
		Messages:    messages,
		Temperature: &temp,
		MaxTokens:   &maxTok,
	}

	// Call the LLM provider
	response, err := f.provider.Completion(context.Background(), params)
	if err != nil {
		return false, fmt.Errorf("llmfilter: completion error: %s", err)
	}

	if len(response.Choices) == 0 {
		return false, fmt.Errorf("llmfilter: no choices in response")
	}

	responseText := response.Choices[0].Message.ContentString()

	// Set extras
	msg.SetExtra("llm_model", response.Model)
	if response.Usage != nil {
		msg.SetExtra("llm_prompt_tokens", strconv.Itoa(response.Usage.PromptTokens))
		msg.SetExtra("llm_completion_tokens", strconv.Itoa(response.Usage.CompletionTokens))
		msg.SetExtra("llm_total_tokens", strconv.Itoa(response.Usage.TotalTokens))
	}

	// Serialize the full response as JSON for llm_raw_response
	if rawJSON, err := json.Marshal(response); err == nil {
		msg.SetExtra("llm_raw_response", string(rawJSON))
	}

	msg.SetTarget(f.target, responseText)

	return true, nil
}

// OnEvent is called when an event occurs
func (f *LLM) OnEvent(event *data.Event) {}

// init registers the filter
func init() {
	register("llm", NewLLMFilter)
}

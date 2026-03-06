package filters

import (
	"context"
	"fmt"
	"testing"

	"github.com/Matrix86/driplane/data"

	anyllm "github.com/mozilla-ai/any-llm-go"
	"github.com/mozilla-ai/any-llm-go/providers"
)

// mockProvider implements providers.Provider for testing
type mockProvider struct {
	name       string
	response   *providers.ChatCompletion
	err        error
	lastParams anyllm.CompletionParams
}

func (m *mockProvider) Name() string { return m.name }

func (m *mockProvider) Completion(_ context.Context, params anyllm.CompletionParams) (*providers.ChatCompletion, error) {
	m.lastParams = params
	return m.response, m.err
}

func (m *mockProvider) CompletionStream(_ context.Context, _ anyllm.CompletionParams) (<-chan providers.ChatCompletionChunk, <-chan error) {
	return nil, nil
}

// newTestLLMFilter creates an LLM filter and injects the mock provider.
// It uses "ollama" provider (which doesn't require an API key) then replaces it.
func newTestLLMFilter(params map[string]string, mock *mockProvider) (*LLM, error) {
	// Ensure provider is set to ollama (no API key required)
	if _, ok := params["provider"]; !ok {
		params["provider"] = "ollama"
	}
	filter, err := NewLLMFilter(params)
	if err != nil {
		return nil, err
	}
	f := filter.(*LLM)
	f.provider = mock
	return f, nil
}

func TestNewLLMFilterMissingModel(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"prompt":   "hello",
		"provider": "ollama",
	})
	if err == nil {
		t.Error("expected error when model is missing")
	}
}

func TestNewLLMFilterMissingPrompt(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"model":    "gpt-4",
		"provider": "ollama",
	})
	if err == nil {
		t.Error("expected error when prompt is missing")
	}
}

func TestNewLLMFilterInvalidProvider(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"model":    "gpt-4",
		"prompt":   "hello",
		"provider": "invalid",
	})
	if err == nil {
		t.Error("expected error for invalid provider")
	}
}

func TestNewLLMFilterInvalidTemperature(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"model":       "gpt-4",
		"prompt":      "hello",
		"provider":    "ollama",
		"temperature": "notanumber",
	})
	if err == nil {
		t.Error("expected error for invalid temperature")
	}
}

func TestNewLLMFilterInvalidMaxTokens(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"model":      "gpt-4",
		"prompt":     "hello",
		"provider":   "ollama",
		"max_tokens": "notanumber",
	})
	if err == nil {
		t.Error("expected error for invalid max_tokens")
	}
}

func TestNewLLMFilterInvalidPromptTemplate(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"model":    "gpt-4",
		"prompt":   "{{.invalid template",
		"provider": "ollama",
	})
	if err == nil {
		t.Error("expected error for invalid prompt template")
	}
}

func TestNewLLMFilterInvalidSystemPromptTemplate(t *testing.T) {
	_, err := NewLLMFilter(map[string]string{
		"model":         "gpt-4",
		"prompt":        "hello",
		"system_prompt": "{{.invalid template",
		"provider":      "ollama",
	})
	if err == nil {
		t.Error("expected error for invalid system_prompt template")
	}
}

func TestNewLLMFilterDefaults(t *testing.T) {
	filter, err := NewLLMFilter(map[string]string{
		"model":    "llama3",
		"prompt":   "hello",
		"provider": "ollama",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f, ok := filter.(*LLM)
	if !ok {
		t.Fatal("cannot cast to *LLM")
	}
	if f.target != "main" {
		t.Errorf("expected default target 'main', got '%s'", f.target)
	}
	if f.temperature != 0.7 {
		t.Errorf("expected default temperature 0.7, got %f", f.temperature)
	}
	if f.maxTokens != 1024 {
		t.Errorf("expected default maxTokens 1024, got %d", f.maxTokens)
	}
	if f.provider == nil {
		t.Error("expected provider to be set")
	}
}

func TestNewLLMFilterCustomParams(t *testing.T) {
	filter, err := NewLLMFilter(map[string]string{
		"model":       "llama3",
		"prompt":      "hello",
		"provider":    "ollama",
		"target":      "response",
		"temperature": "0.5",
		"max_tokens":  "2048",
	})
	if err != nil {
		t.Fatalf("constructor returned error: %s", err)
	}

	f := filter.(*LLM)
	if f.target != "response" {
		t.Errorf("expected target 'response', got '%s'", f.target)
	}
	if f.temperature != 0.5 {
		t.Errorf("expected temperature 0.5, got %f", f.temperature)
	}
	if f.maxTokens != 2048 {
		t.Errorf("expected maxTokens 2048, got %d", f.maxTokens)
	}
}

func TestLLMDoFilterBasic(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			ID:    "chatcmpl-123",
			Model: "gpt-4",
			Choices: []providers.Choice{
				{
					Index:        0,
					Message:      providers.Message{Role: "assistant", Content: "Hi there!"},
					FinishReason: "stop",
				},
			},
			Usage: &providers.Usage{
				PromptTokens:     10,
				CompletionTokens: 5,
				TotalTokens:      15,
			},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":         "gpt-4",
		"prompt":        "{{ .main }}",
		"system_prompt": "You are helpful",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("Hello world")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter returned error: %s", err)
	}
	if !ok {
		t.Error("DoFilter returned false")
	}
	if msg.GetMessage() != "Hi there!" {
		t.Errorf("expected 'Hi there!', got '%s'", msg.GetMessage())
	}

	// Verify messages sent to provider
	if len(mock.lastParams.Messages) != 2 {
		t.Fatalf("expected 2 messages (system + user), got %d", len(mock.lastParams.Messages))
	}
	if mock.lastParams.Messages[0].Role != "system" {
		t.Errorf("expected system role, got '%s'", mock.lastParams.Messages[0].Role)
	}
	if mock.lastParams.Messages[0].ContentString() != "You are helpful" {
		t.Errorf("expected system content 'You are helpful', got '%s'", mock.lastParams.Messages[0].ContentString())
	}
	if mock.lastParams.Messages[1].Role != "user" {
		t.Errorf("expected user role, got '%s'", mock.lastParams.Messages[1].Role)
	}
	if mock.lastParams.Messages[1].ContentString() != "Hello world" {
		t.Errorf("expected user content 'Hello world', got '%s'", mock.lastParams.Messages[1].ContentString())
	}

	// Verify extras
	extras := msg.GetExtra()
	if extras["llm_model"] != "gpt-4" {
		t.Errorf("expected llm_model 'gpt-4', got '%v'", extras["llm_model"])
	}
	if extras["llm_prompt_tokens"] != "10" {
		t.Errorf("expected llm_prompt_tokens '10', got '%v'", extras["llm_prompt_tokens"])
	}
	if extras["llm_completion_tokens"] != "5" {
		t.Errorf("expected llm_completion_tokens '5', got '%v'", extras["llm_completion_tokens"])
	}
	if extras["llm_total_tokens"] != "15" {
		t.Errorf("expected llm_total_tokens '15', got '%v'", extras["llm_total_tokens"])
	}
	if extras["llm_raw_response"] == nil || extras["llm_raw_response"] == "" {
		t.Error("expected llm_raw_response to be set")
	}
}

func TestLLMDoFilterNoSystemPrompt(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model: "gpt-4",
			Choices: []providers.Choice{
				{Message: providers.Message{Role: "assistant", Content: "response"}},
			},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4",
		"prompt": "hello",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter error: %s", err)
	}
	if !ok {
		t.Error("DoFilter returned false")
	}

	// Should only have user message, no system
	if len(mock.lastParams.Messages) != 1 {
		t.Errorf("expected 1 message (user only), got %d", len(mock.lastParams.Messages))
	}
	if mock.lastParams.Messages[0].Role != "user" {
		t.Errorf("expected user role, got '%s'", mock.lastParams.Messages[0].Role)
	}
}

func TestLLMDoFilterWithTarget(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model: "gpt-4",
			Choices: []providers.Choice{
				{Message: providers.Message{Role: "assistant", Content: "llm reply"}},
			},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4",
		"prompt": "{{ .main }}",
		"target": "llm_response",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("original text")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter error: %s", err)
	}
	if !ok {
		t.Error("DoFilter returned false")
	}

	// main should still be the original
	if msg.GetMessage() != "original text" {
		t.Errorf("expected main to be 'original text', got '%s'", msg.GetMessage())
	}
	// target should have the LLM reply
	extras := msg.GetExtra()
	if extras["llm_response"] != "llm reply" {
		t.Errorf("expected llm_response extra 'llm reply', got '%v'", extras["llm_response"])
	}
}

func TestLLMDoFilterTemplateWithExtras(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model:   "gpt-4",
			Choices: []providers.Choice{{Message: providers.Message{Content: "hola"}}},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4",
		"prompt": "translate: {{ .main }} to {{ .language }}",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessageWithExtra("hello", map[string]interface{}{"language": "Spanish"})
	_, err = f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter error: %s", err)
	}

	// Verify template rendered correctly
	expectedPrompt := "translate: hello to Spanish"
	if mock.lastParams.Messages[0].ContentString() != expectedPrompt {
		t.Errorf("expected prompt '%s', got '%s'", expectedPrompt, mock.lastParams.Messages[0].ContentString())
	}
}

func TestLLMDoFilterCompletionError(t *testing.T) {
	mock := &mockProvider{
		name:     "mock",
		response: nil,
		err:      fmt.Errorf("API error: rate limited"),
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4",
		"prompt": "hello",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err == nil {
		t.Error("expected error from completion failure")
	}
	if ok {
		t.Error("expected false for error response")
	}
}

func TestLLMDoFilterNoChoices(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model:   "gpt-4",
			Choices: []providers.Choice{},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4",
		"prompt": "hello",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("test")
	_, err = f.DoFilter(msg)
	if err == nil {
		t.Error("expected error when no choices returned")
	}
}

func TestLLMDoFilterNilUsage(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model: "gpt-4",
			Choices: []providers.Choice{
				{Message: providers.Message{Role: "assistant", Content: "ok"}},
			},
			Usage: nil, // Some providers may not return usage
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4",
		"prompt": "hello",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("test")
	ok, err := f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter error: %s", err)
	}
	if !ok {
		t.Error("DoFilter returned false")
	}

	extras := msg.GetExtra()
	// Usage extras should not be set when usage is nil
	if _, exists := extras["llm_prompt_tokens"]; exists {
		t.Error("expected no llm_prompt_tokens when usage is nil")
	}
}

func TestLLMDoFilterTemperatureAndMaxTokens(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model:   "gpt-4",
			Choices: []providers.Choice{{Message: providers.Message{Content: "ok"}}},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":       "gpt-4",
		"prompt":      "hello",
		"temperature": "0.3",
		"max_tokens":  "512",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("test")
	_, err = f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter error: %s", err)
	}

	if mock.lastParams.Temperature == nil || *mock.lastParams.Temperature != 0.3 {
		t.Errorf("expected temperature 0.3, got %v", mock.lastParams.Temperature)
	}
	if mock.lastParams.MaxTokens == nil || *mock.lastParams.MaxTokens != 512 {
		t.Errorf("expected max_tokens 512, got %v", mock.lastParams.MaxTokens)
	}
}

func TestLLMDoFilterModelInParams(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model:   "gpt-4",
			Choices: []providers.Choice{{Message: providers.Message{Content: "ok"}}},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "gpt-4-turbo",
		"prompt": "hello",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}

	msg := data.NewMessage("test")
	_, err = f.DoFilter(msg)
	if err != nil {
		t.Fatalf("DoFilter error: %s", err)
	}

	if mock.lastParams.Model != "gpt-4-turbo" {
		t.Errorf("expected model 'gpt-4-turbo', got '%s'", mock.lastParams.Model)
	}
}

func TestLLMSupportedProviders(t *testing.T) {
	expected := []string{"openai", "anthropic", "ollama", "deepseek", "groq", "mistral", "gemini", "llamacpp", "llamafile"}
	for _, name := range expected {
		if _, ok := supportedProviders[name]; !ok {
			t.Errorf("expected provider '%s' to be in supportedProviders map", name)
		}
	}
}

func TestLLMOnEvent(t *testing.T) {
	mock := &mockProvider{
		name: "mock",
		response: &providers.ChatCompletion{
			Model:   "test",
			Choices: []providers.Choice{{Message: providers.Message{Content: "ok"}}},
		},
	}

	f, err := newTestLLMFilter(map[string]string{
		"model":  "test",
		"prompt": "hello",
	}, mock)
	if err != nil {
		t.Fatalf("constructor error: %s", err)
	}
	// OnEvent should not panic
	f.OnEvent(nil)
}

package config

import (
	"os"
	"testing"
)

func TestNew_Defaults(t *testing.T) {
	cfg := New()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.MistralURL != "https://api.mistral.ai" {
		t.Errorf("expected MistralURL https://api.mistral.ai, got %s", cfg.MistralURL)
	}
	if cfg.GeminiURL != "https://generativelanguage.googleapis.com" {
		t.Errorf("expected GeminiURL https://generativelanguage.googleapis.com, got %s", cfg.GeminiURL)
	}
	if cfg.SystemPrompt == "" {
		t.Errorf("expected non-empty SystemPrompt")
	}
}

func TestNew_MistralDefaults(t *testing.T) {
	// Ensure Mistral is the default provider when nothing is configured.
	os.Unsetenv("PROVIDER")
	cfg := New()

	if cfg.Provider != "mistral" {
		t.Errorf("expected default provider mistral, got %s", cfg.Provider)
	}
	if cfg.STTModel != "voxtral-mini-latest" {
		t.Errorf("expected STTModel voxtral-mini-latest, got %s", cfg.STTModel)
	}
	if cfg.LLMModel != "mistral-small-latest" {
		t.Errorf("expected LLMModel mistral-small-latest, got %s", cfg.LLMModel)
	}
}

func TestNew_GeminiDefaults(t *testing.T) {
	cfg := New()

	if cfg.GeminiSTTModel != "gemini-3-flash" {
		t.Errorf("expected GeminiSTTModel gemini-3-flash, got %s", cfg.GeminiSTTModel)
	}
	if cfg.GeminiLLMModel != "gemini-3-flash" {
		t.Errorf("expected GeminiLLMModel gemini-3-flash, got %s", cfg.GeminiLLMModel)
	}
}

func TestNew_ProviderEnvOverride(t *testing.T) {
	t.Setenv("PROVIDER", "gemini")

	cfg := New()

	if cfg.Provider != "gemini" {
		t.Errorf("expected provider gemini from env, got %s", cfg.Provider)
	}
}

func TestNew_MistralEnvOverrides(t *testing.T) {
	t.Setenv("MISTRAL_API_KEY", "test-mistral-key")
	t.Setenv("MISTRAL_STT_MODEL", "custom-stt")
	t.Setenv("MISTRAL_LLM_MODEL", "custom-llm")

	cfg := New()

	if cfg.MistralKey != "test-mistral-key" {
		t.Errorf("expected MistralKey test-mistral-key, got %s", cfg.MistralKey)
	}
	if cfg.STTModel != "custom-stt" {
		t.Errorf("expected STTModel custom-stt, got %s", cfg.STTModel)
	}
	if cfg.LLMModel != "custom-llm" {
		t.Errorf("expected LLMModel custom-llm, got %s", cfg.LLMModel)
	}
}

func TestNew_GeminiEnvOverrides(t *testing.T) {
	t.Setenv("GEMINI_API_KEY", "test-gemini-key")
	t.Setenv("GEMINI_STT_MODEL", "gemini-custom-stt")
	t.Setenv("GEMINI_LLM_MODEL", "gemini-custom-llm")

	cfg := New()

	if cfg.GeminiKey != "test-gemini-key" {
		t.Errorf("expected GeminiKey test-gemini-key, got %s", cfg.GeminiKey)
	}
	if cfg.GeminiSTTModel != "gemini-custom-stt" {
		t.Errorf("expected GeminiSTTModel gemini-custom-stt, got %s", cfg.GeminiSTTModel)
	}
	if cfg.GeminiLLMModel != "gemini-custom-llm" {
		t.Errorf("expected GeminiLLMModel gemini-custom-llm, got %s", cfg.GeminiLLMModel)
	}
}

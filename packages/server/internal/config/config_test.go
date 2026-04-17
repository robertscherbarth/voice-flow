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
	if cfg.LMStudioURL != "http://localhost:1234/v1" {
		t.Errorf("expected LMStudioURL http://localhost:1234/v1, got %s", cfg.LMStudioURL)
	}
	if cfg.SystemPrompt == "" {
		t.Errorf("expected non-empty SystemPrompt")
	}
}

func TestNew_MistralProviderModels(t *testing.T) {
	// Force mistral provider explicitly to test its model defaults.
	t.Setenv("PROVIDER", "mistral")
	os.Unsetenv("MISTRAL_STT_MODEL")
	os.Unsetenv("MISTRAL_LLM_MODEL")

	cfg := New()

	if cfg.Provider != "mistral" {
		t.Errorf("expected provider mistral, got %s", cfg.Provider)
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

	if cfg.GeminiSTTModel != "gemini-2.5-flash-lite" {
		t.Errorf("expected GeminiSTTModel gemini-2.5-flash-lite, got %s", cfg.GeminiSTTModel)
	}
	if cfg.GeminiLLMModel != "gemini-2.5-flash-lite" {
		t.Errorf("expected GeminiLLMModel gemini-2.5-flash-lite, got %s", cfg.GeminiLLMModel)
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
	t.Setenv("PROVIDER", "mistral") // pin provider so active models come from Mistral
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

func TestNew_LocalProvider(t *testing.T) {
	// "local" provider: STT uses Mistral model, LLM uses LM Studio model.
	t.Setenv("PROVIDER", "local")
	os.Unsetenv("MISTRAL_STT_MODEL")
	os.Unsetenv("LMSTUDIO_MODEL")

	cfg := New()

	if cfg.Provider != "local" {
		t.Errorf("expected provider local, got %s", cfg.Provider)
	}
	// STT must still route through Mistral's default model
	if cfg.STTModel != "voxtral-mini-latest" {
		t.Errorf("expected STTModel voxtral-mini-latest for local provider, got %s", cfg.STTModel)
	}
	// LLM must route to LM Studio's Gemma 4 model
	if cfg.LLMModel != "google/gemma-4-26b-a4b" {
		t.Errorf("expected LLMModel google/gemma-4-26b-a4b for local provider, got %s", cfg.LLMModel)
	}
	if cfg.LMStudioURL != "http://localhost:1234/v1" {
		t.Errorf("expected LMStudioURL http://localhost:1234/v1, got %s", cfg.LMStudioURL)
	}
}

func TestNew_LocalProviderEnvOverrides(t *testing.T) {
	t.Setenv("PROVIDER", "local")
	t.Setenv("LMSTUDIO_URL", "http://192.168.1.100:1234/v1")
	t.Setenv("LMSTUDIO_MODEL", "gemma-4-custom")
	t.Setenv("MISTRAL_STT_MODEL", "voxtral-large")

	cfg := New()

	if cfg.LMStudioURL != "http://192.168.1.100:1234/v1" {
		t.Errorf("expected LMStudioURL http://192.168.1.100:1234/v1, got %s", cfg.LMStudioURL)
	}
	if cfg.LMStudioModel != "gemma-4-custom" {
		t.Errorf("expected LMStudioModel gemma-4-custom, got %s", cfg.LMStudioModel)
	}
	// Active STT model should reflect the override
	if cfg.STTModel != "voxtral-large" {
		t.Errorf("expected STTModel voxtral-large, got %s", cfg.STTModel)
	}
	// Active LLM model should reflect the LM Studio override
	if cfg.LLMModel != "gemma-4-custom" {
		t.Errorf("expected LLMModel gemma-4-custom, got %s", cfg.LLMModel)
	}
}

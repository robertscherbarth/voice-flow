package config

import (
	"testing"
)

func TestNew(t *testing.T) {
	cfg := New()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.MistralURL != "https://api.mistral.ai" {
		t.Errorf("expected url https://api.mistral.ai, got %s", cfg.MistralURL)
	}
	if cfg.STTModel != "voxtral-mini-latest" {
		t.Errorf("expected STTModel voxtral-mini-latest, got %s", cfg.STTModel)
	}
	if cfg.LLMModel != "mistral-small-latest" {
		t.Errorf("expected LLMModel mistral-small-latest, got %s", cfg.LLMModel)
	}
	if cfg.SystemPrompt == "" {
		t.Errorf("expected non-empty SystemPrompt")
	}
}

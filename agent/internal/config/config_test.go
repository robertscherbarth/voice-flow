package config

import (
	"testing"
)

func TestNew(t *testing.T) {
	cfg := New()

	if cfg.Port != "8080" {
		t.Errorf("expected port 8080, got %s", cfg.Port)
	}
	if cfg.OllamaURL != "http://localhost:11434" {
		t.Errorf("expected url http://localhost:11434, got %s", cfg.OllamaURL)
	}
	if cfg.WhisperURL != "http://localhost:8081" {
		t.Errorf("expected WhisperURL http://localhost:8081, got %s", cfg.WhisperURL)
	}
	if cfg.STTModel != "base" {
		t.Errorf("expected STTModel base, got %s", cfg.STTModel)
	}
	if cfg.LLMModel != "mistral" {
		t.Errorf("expected LLMModel mistral, got %s", cfg.LLMModel)
	}
	if cfg.SystemPrompt == "" {
		t.Errorf("expected non-empty SystemPrompt")
	}
}

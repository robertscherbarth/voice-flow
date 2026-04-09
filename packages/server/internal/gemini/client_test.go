package gemini

import (
	"context"
	"os"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	c, err := NewClient(context.Background(), "test-key")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestTranscribeAudio_NoApiKey(t *testing.T) {
	// The SDK might not error on client creation with empty key, but will error on request
	c, err := NewClient(context.Background(), "")
	if err != nil {
		t.Logf("NewClient errored as expected (optional): %v", err)
		return
	}
	if c == nil {
		t.Fatal("expected non-nil client or error")
	}
	_, err = c.TranscribeAudio(context.Background(), []byte("audio"), "test.wav", "gemini-3-flash-preview")
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
}

func TestImproveText_NoApiKey(t *testing.T) {
	c, err := NewClient(context.Background(), "")
	if err != nil {
		t.Logf("NewClient errored as expected (optional): %v", err)
		return
	}
	if c == nil {
		t.Fatal("expected non-nil client or error")
	}
	_, err = c.ImproveText(context.Background(), "hello", "gemini-3-flash-preview", "system prompt")
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
}

func TestAudioMIMEType(t *testing.T) {
	tests := []struct {
		filename string
		expected string
	}{
		{"audio.m4a", "audio/mp4"},
		{"audio.mp3", "audio/mpeg"},
		{"audio.wav", "audio/wav"},
		{"audio.ogg", "audio/ogg"},
		{"audio.aac", "audio/aac"},
		{"audio.webm", "audio/webm"},
		{"audio.flac", "audio/flac"},
		{"audio.unknown", "audio/wav"},
		{"", "audio/wav"},
	}
	for _, tt := range tests {
		t.Run(tt.filename, func(t *testing.T) {
			got := audioMIMEType(tt.filename)
			if got != tt.expected {
				t.Errorf("audioMIMEType(%q) = %q, want %q", tt.filename, got, tt.expected)
			}
		})
	}
}

// Integration test - only runs if GEMINI_API_KEY is set
func TestIntegration_Gemini(t *testing.T) {
	apiKey := os.Getenv("GEMINI_API_KEY")
	if apiKey == "" {
		t.Skip("Skipping integration test: GEMINI_API_KEY not set")
	}

	ctx := context.Background()
	client, err := NewClient(ctx, apiKey)
	if err != nil {
		t.Fatalf("Failed to create client: %v", err)
	}

	t.Run("ImproveText", func(t *testing.T) {
		improved, err := client.ImproveText(ctx, "hello world", "gemini-3-flash-preview", "Make it more formal")
		if err != nil {
			t.Fatalf("ImproveText failed: %v", err)
		}
		if improved == "" {
			t.Error("Got empty response from Gemini")
		}
		if !strings.Contains(strings.ToLower(improved), "hello") {
			t.Errorf("Expected response to contain 'hello', got: %s", improved)
		}
	})
}

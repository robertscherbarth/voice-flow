package gemini

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestNewClient(t *testing.T) {
	c := NewClient("https://example.com", "test-key")
	if c == nil {
		t.Fatal("expected non-nil client")
	}
}

func TestTranscribeAudio_MissingKey(t *testing.T) {
	c := NewClient("https://example.com", "")
	_, err := c.TranscribeAudio(context.Background(), []byte("audio"), "test.wav", "gemini-2.0-flash")
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("expected API key error, got: %v", err)
	}
}

func TestImproveText_MissingKey(t *testing.T) {
	c := NewClient("https://example.com", "")
	_, err := c.ImproveText(context.Background(), "hello", "gemini-2.0-flash", "system prompt")
	if err == nil {
		t.Fatal("expected error when API key is empty")
	}
	if !strings.Contains(err.Error(), "API key") {
		t.Errorf("expected API key error, got: %v", err)
	}
}

func TestTranscribeAudio_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify path contains the model name
		if !strings.Contains(r.URL.Path, "gemini-2.0-flash") {
			t.Errorf("expected model in path, got: %s", r.URL.Path)
		}
		// Verify API key is passed as query param
		if r.URL.Query().Get("key") != "test-key" {
			t.Errorf("expected key=test-key, got: %s", r.URL.Query().Get("key"))
		}
		// Verify the request body contains inline audio data
		var req generateContentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		if len(req.Contents) == 0 || len(req.Contents[0].Parts) < 2 {
			t.Errorf("expected contents with at least 2 parts (audio + text prompt)")
		}
		if req.Contents[0].Parts[0].InlineData == nil {
			t.Errorf("expected inline_data in first part")
		}

		resp := generateContentResponse{
			Candidates: []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			}{
				{Content: struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				}{Parts: []struct {
					Text string `json:"text"`
				}{{Text: "hello world"}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	text, err := c.TranscribeAudio(context.Background(), []byte("fakeaudio"), "test.wav", "gemini-2.0-flash")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "hello world" {
		t.Errorf("expected 'hello world', got %q", text)
	}
}

func TestImproveText_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var req generateContentRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("failed to decode request: %v", err)
		}
		// Verify system instruction is set
		if req.SystemInstruction == nil || len(req.SystemInstruction.Parts) == 0 {
			t.Errorf("expected system_instruction to be set")
		}
		// Verify temperature
		if req.GenerationConfig == nil || req.GenerationConfig.Temperature != 0.2 {
			t.Errorf("expected temperature 0.2")
		}

		resp := generateContentResponse{
			Candidates: []struct {
				Content struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				} `json:"content"`
			}{
				{Content: struct {
					Parts []struct {
						Text string `json:"text"`
					} `json:"parts"`
				}{Parts: []struct {
					Text string `json:"text"`
				}{{Text: "improved text"}}}},
			},
		}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	text, err := c.ImproveText(context.Background(), "raw text", "gemini-2.0-flash", "system prompt")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if text != "improved text" {
		t.Errorf("expected 'improved text', got %q", text)
	}
}

func TestTranscribeAudio_HTTPError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Error(w, "unauthorized", http.StatusUnauthorized)
	}))
	defer server.Close()

	c := NewClient(server.URL, "bad-key")
	_, err := c.TranscribeAudio(context.Background(), []byte("audio"), "test.wav", "gemini-2.0-flash")
	if err == nil {
		t.Fatal("expected error on HTTP 401")
	}
	if !strings.Contains(err.Error(), "401") {
		t.Errorf("expected 401 in error, got: %v", err)
	}
}

func TestImproveText_EmptyCandidates(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return empty candidates list
		resp := generateContentResponse{}
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(resp)
	}))
	defer server.Close()

	c := NewClient(server.URL, "test-key")
	_, err := c.ImproveText(context.Background(), "text", "gemini-2.0-flash", "prompt")
	if err == nil {
		t.Fatal("expected error on empty candidates")
	}
	if !strings.Contains(err.Error(), "no content") {
		t.Errorf("expected 'no content' error, got: %v", err)
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

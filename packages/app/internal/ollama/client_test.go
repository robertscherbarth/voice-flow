package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestTranscribeAudio(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody map[string]interface{}
		expectedText string
		expectError  bool
		errContains  string
	}{
		{
			name:         "success",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{"text": "hello world"},
			expectedText: "hello world",
			expectError:  false,
		},
		{
			name:         "http error",
			responseCode: http.StatusInternalServerError,
			responseBody: map[string]interface{}{"error": "server error"},
			expectError:  true,
			errContains:  "ollama returned status 500",
		},
		{
			name:         "missing text in response",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{"something_else": "foo"},
			expectError:  true,
			errContains:  "transcription text not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/v1/audio/transcriptions" {
					t.Errorf("expected path /v1/audio/transcriptions, got %s", r.URL.Path)
				}
				if r.Method != http.MethodPost {
					t.Errorf("expected method POST, got %s", r.Method)
				}

				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, server.URL)
			ctx := context.Background()

			text, err := client.TranscribeAudio(ctx, []byte("fake audio data"), "whisper")

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if text != tt.expectedText {
					t.Errorf("expected text %q, got %q", tt.expectedText, text)
				}
			}
		})
	}
}

func TestImproveText(t *testing.T) {
	tests := []struct {
		name         string
		responseCode int
		responseBody map[string]interface{}
		expectedText string
		expectError  bool
		errContains  string
	}{
		{
			name:         "success",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{"response": "Hello, world."},
			expectedText: "Hello, world.",
			expectError:  false,
		},
		{
			name:         "http error",
			responseCode: http.StatusBadRequest,
			responseBody: map[string]interface{}{"error": "bad request"},
			expectError:  true,
			errContains:  "ollama returned status 400",
		},
		{
			name:         "missing response field",
			responseCode: http.StatusOK,
			responseBody: map[string]interface{}{"text": "wrong field"},
			expectError:  true,
			errContains:  "response text not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/api/generate" {
					t.Errorf("expected path /api/generate, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.responseCode)
				json.NewEncoder(w).Encode(tt.responseBody)
			}))
			defer server.Close()

			client := NewClient(server.URL, server.URL)
			ctx := context.Background()

			text, err := client.ImproveText(ctx, "hello world umm", "mistral", "fix this")

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected error containing %q, got nil", tt.errContains)
				}
				if !strings.Contains(err.Error(), tt.errContains) {
					t.Errorf("expected error to contain %q, got %q", tt.errContains, err.Error())
				}
			} else {
				if err != nil {
					t.Fatalf("expected no error, got %v", err)
				}
				if text != tt.expectedText {
					t.Errorf("expected text %q, got %q", tt.expectedText, text)
				}
			}
		})
	}
}

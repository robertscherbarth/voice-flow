package ollama

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

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

			client := NewClient(server.URL)
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

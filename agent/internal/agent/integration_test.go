//go:build integration

package agent_test

import (
	"bytes"
	"context"
	"encoding/json"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"voice-agent/internal/agent"
	"voice-agent/internal/config"
	"voice-agent/internal/mistral"
)

func TestIntegration_ProcessAudio(t *testing.T) {
	// Setup real configuration
	cfg := config.New()

	if cfg.MistralKey == "" {
		t.Skip("Skipping integration test: MISTRAL_API_KEY is not set")
	}

	// Create real Mistral client for both STT and LLM
	mistralClient := mistral.NewClient(cfg.MistralURL, cfg.MistralKey)

	// Create the handler
	handler := agent.NewHandler(mistralClient, mistralClient, cfg)

	// Read the real M4A file
	audioPath := filepath.Join("..", "..", "..", "scripts", "test-sentence.m4a")
	audioData, err := os.ReadFile(audioPath)
	if err != nil {
		t.Fatalf("Failed to read test audio file at %s: %v", audioPath, err)
	}

	// Prepare multipart form request
	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	// Add audio file
	part, err := writer.CreateFormFile("audio", "test-sentence.m4a")
	if err != nil {
		t.Fatalf("Failed to create form file: %v", err)
	}
	part.Write(audioData)

	// We no longer override the system prompt so we can test the real one.
	writer.Close()

	// Create HTTP request
	req := httptest.NewRequest(http.MethodPost, "/process", &requestBody)
	req.Header.Set("Content-Type", writer.FormDataContentType())

	// We need a context with a timeout just in case
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
	defer cancel()
	req = req.WithContext(ctx)

	// Record response
	rr := httptest.NewRecorder()

	t.Log("Sending request to Mistral for STT and LLM. This might take a few seconds...")
	startTime := time.Now()
	handler.ServeHTTP(rr, req)

	t.Logf("Processing took %v", time.Since(startTime))

	if rr.Code != http.StatusOK {
		t.Fatalf("Expected status OK, got %d. Body: %s", rr.Code, rr.Body.String())
	}

	var resp map[string]string
	if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
		t.Fatalf("Failed to decode response: %v", err)
	}

	resultText := strings.TrimSpace(resp["text"])

	if resultText == "" {
		t.Errorf("Expected a non-empty string, got %q", resp["text"])
	} else {
		t.Logf("Success! Pipeline replied with: %q", resp["text"])
	}
}

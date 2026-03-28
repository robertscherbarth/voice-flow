package agent

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"mime/multipart"
	"net/http"
	"net/http/httptest"
	"testing"

	"voice-agent/internal/config"
)

// mockLLMClient implements LLMClient for testing
type mockLLMClient struct {
	improveTextFunc func(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
}

func (m *mockLLMClient) ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
	if m.improveTextFunc != nil {
		return m.improveTextFunc(ctx, transcript, modelName, systemPrompt)
	}
	return "Mock improved text.", nil
}

// mockSTTClient implements STTClient for testing
type mockSTTClient struct {
	transcribeAudioFunc func(ctx context.Context, audioData []byte, filename, modelName string) (string, error)
}

func (m *mockSTTClient) TranscribeAudio(ctx context.Context, audioData []byte, filename, modelName string) (string, error) {
	if m.transcribeAudioFunc != nil {
		return m.transcribeAudioFunc(ctx, audioData, filename, modelName)
	}
	return "mock raw text", nil
}

func TestProcessHandler(t *testing.T) {
	tests := []struct {
		name           string
		method         string
		sttError       error
		llmError       error
		expectedStatus int
		expectedText   string
		isMissingAudio bool
	}{
		{
			name:           "successful processing",
			method:         http.MethodPost,
			expectedStatus: http.StatusOK,
			expectedText:   "Mock improved text.",
		},
		{
			name:           "wrong method",
			method:         http.MethodGet,
			expectedStatus: http.StatusMethodNotAllowed,
		},
		{
			name:           "missing audio file",
			method:         http.MethodPost,
			isMissingAudio: true,
			expectedStatus: http.StatusBadRequest,
		},
		{
			name:           "stt failure",
			method:         http.MethodPost,
			sttError:       errors.New("stt failure"),
			expectedStatus: http.StatusInternalServerError,
		},
		{
			name:           "llm failure",
			method:         http.MethodPost,
			llmError:       errors.New("llm failure"),
			expectedStatus: http.StatusInternalServerError,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mockLLM := &mockLLMClient{
				improveTextFunc: func(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
					if tt.llmError != nil {
						return "", tt.llmError
					}
					return "Mock improved text.", nil
				},
			}

			mockSTT := &mockSTTClient{
				transcribeAudioFunc: func(ctx context.Context, audioData []byte, filename, modelName string) (string, error) {
					if tt.sttError != nil {
						return "", tt.sttError
					}
					return "raw text", nil
				},
			}

			cfg := config.New()
			handler := NewHandler(mockLLM, mockSTT, cfg)

			var requestBody bytes.Buffer
			var contentType string

			if tt.method == http.MethodPost && !tt.isMissingAudio {
				writer := multipart.NewWriter(&requestBody)
				part, _ := writer.CreateFormFile("audio", "test.wav")
				part.Write([]byte("fake audio bytes"))
				writer.Close()
				contentType = writer.FormDataContentType()
			} else {
				// Just a plain text body if GET or missing audio
				requestBody.WriteString("invalid body")
				contentType = "text/plain"
			}

			req := httptest.NewRequest(tt.method, "/process", &requestBody)
			if tt.method == http.MethodPost {
				req.Header.Set("Content-Type", contentType)
			}

			rr := httptest.NewRecorder()
			handler.ServeHTTP(rr, req)

			if rr.Code != tt.expectedStatus {
				t.Errorf("expected status %d, got %d", tt.expectedStatus, rr.Code)
			}

			if rr.Code == http.StatusOK {
				var resp map[string]string
				if err := json.NewDecoder(rr.Body).Decode(&resp); err != nil {
					t.Fatalf("failed to decode response: %v", err)
				}
				if resp["text"] != tt.expectedText {
					t.Errorf("expected text %q, got %q", tt.expectedText, resp["text"])
				}
			}
		})
	}
}

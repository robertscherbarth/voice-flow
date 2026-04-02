package mistral

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"mime/multipart"
	"net/http"
	"time"
)

// Client defines the interface for communicating with Mistral STT and LLM models.
type Client interface {
	TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error)
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
}

// clientImpl is the concrete implementation of the Mistral Client.
type clientImpl struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a new Client pointing to the given Mistral API.
func NewClient(apiURL, apiKey string) Client {
	return &clientImpl{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

// TranscribeAudio sends an audio file to the Mistral API endpoint.
func (c *clientImpl) TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("mistral API key is not configured")
	}

	url := fmt.Sprintf("%s/v1/audio/transcriptions", c.apiURL)

	var requestBody bytes.Buffer
	writer := multipart.NewWriter(&requestBody)

	if filename == "" {
		filename = "audio.wav"
	}

	part, err := writer.CreateFormFile("file", filename)
	if err != nil {
		return "", fmt.Errorf("create form file: %w", err)
	}
	if _, err := part.Write(audioData); err != nil {
		return "", fmt.Errorf("write audio data: %w", err)
	}

	if modelName != "" {
		if err := writer.WriteField("model", modelName); err != nil {
			return "", fmt.Errorf("write model field: %w", err)
		}
	}

	if err := writer.Close(); err != nil {
		return "", fmt.Errorf("close writer: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, &requestBody)
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("stt request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mistral returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("Mistral STT raw response: %s", string(bodyBytes))

	var result struct {
		Text string `json:"text"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Text, nil
}

// ImproveText sends the raw transcript to Mistral LLM for improvement.
func (c *clientImpl) ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("mistral API key is not configured")
	}

	url := fmt.Sprintf("%s/v1/chat/completions", c.apiURL)

	payload := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": transcript},
		},
		"temperature": 0.2, // low temperature for consistent text improvement
	}

	requestBody, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(requestBody))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+c.apiKey)

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("mistral returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Choices []struct {
			Message struct {
				Content string `json:"content"`
			} `json:"message"`
		} `json:"choices"`
	}
	if err := json.Unmarshal(bodyBytes, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no response choices returned from Mistral")
	}

	return result.Choices[0].Message.Content, nil
}

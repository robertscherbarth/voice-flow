package ollama

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// Client defines the interface for communicating with Ollama models.
type Client interface {
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
}

// clientImpl is the concrete implementation of the Ollama Client.
type clientImpl struct {
	ollamaURL  string
	httpClient *http.Client
}

// NewClient returns a new Client pointing to the given base URL.
func NewClient(ollamaURL string) Client {
	return &clientImpl{
		ollamaURL:  ollamaURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

// ImproveText sends the raw transcript to the local Ollama LLM for improvement.
func (c *clientImpl) ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
	url := fmt.Sprintf("%s/api/generate", c.ollamaURL)

	payload := map[string]interface{}{
		"model":  modelName,
		"prompt": transcript,
		"system": systemPrompt,
		"stream": false,
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

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("llm request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		bodyBytes, _ := io.ReadAll(resp.Body)
		return "", fmt.Errorf("ollama returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	var result struct {
		Response string `json:"response"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	return result.Response, nil
}

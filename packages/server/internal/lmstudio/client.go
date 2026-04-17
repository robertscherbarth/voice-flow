package lmstudio

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client defines the interface for communicating with an LM Studio LLM.
type Client interface {
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
	ImproveTextStream(ctx context.Context, transcript, modelName, systemPrompt string, onChunk func(string)) error
}

// clientImpl is the concrete implementation of the LM Studio Client.
type clientImpl struct {
	apiURL     string
	httpClient *http.Client
}

// NewClient returns a new Client pointing to the LM Studio OpenAI-compatible API.
// No API key is required; LM Studio runs locally without authentication.
func NewClient(apiURL string) Client {
	return &clientImpl{
		apiURL: apiURL,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

// ImproveText sends the raw transcript to the LM Studio LLM for text improvement.
// It uses the OpenAI-compatible /v1/chat/completions endpoint.
func (c *clientImpl) ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
	url := fmt.Sprintf("%s/chat/completions", c.apiURL)

	payload := map[string]interface{}{
		"model": modelName,
		"messages": []map[string]string{
			{"role": "system", "content": systemPrompt},
			{"role": "user", "content": transcript},
		},
		"temperature": 0.2, // low temperature for consistent text improvement
		"thinking": map[string]string{
			"type": "disabled", // disable chain-of-thought to avoid long latency
		},
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

	log.Printf("LM Studio request: model=%s url=%s", modelName, url)

	llmStart := time.Now()
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("lm studio request failed: %w", err)
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("lm studio returned status %d: %s", resp.StatusCode, string(bodyBytes))
	}

	log.Printf("LM Studio LLM latency: %v", time.Since(llmStart))

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
		return "", fmt.Errorf("no response choices returned from LM Studio")
	}

	return result.Choices[0].Message.Content, nil
}

// ImproveTextStream is a placeholder that calls ImproveText in a single chunk for now.
func (c *clientImpl) ImproveTextStream(ctx context.Context, transcript, modelName, systemPrompt string, onChunk func(string)) error {
	text, err := c.ImproveText(ctx, transcript, modelName, systemPrompt)
	if err != nil {
		return err
	}
	onChunk(text)
	return nil
}

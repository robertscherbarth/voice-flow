package gemini

import (
	"bytes"
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"time"
)

// Client defines the interface for communicating with Gemini STT and LLM models.
type Client interface {
	TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error)
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
}

// clientImpl is the concrete implementation of the Gemini Client.
type clientImpl struct {
	apiURL     string
	apiKey     string
	httpClient *http.Client
}

// NewClient returns a new Client pointing to the given Gemini API.
func NewClient(apiURL, apiKey string) Client {
	return &clientImpl{
		apiURL: apiURL,
		apiKey: apiKey,
		httpClient: &http.Client{
			Timeout: 3 * time.Minute,
		},
	}
}

// generateContent is the shared request structure for all Gemini generateContent calls.
type generateContentRequest struct {
	SystemInstruction *systemInstruction `json:"system_instruction,omitempty"`
	Contents          []content          `json:"contents"`
	GenerationConfig  *generationConfig  `json:"generation_config,omitempty"`
}

type systemInstruction struct {
	Parts []part `json:"parts"`
}

type content struct {
	Parts []part `json:"parts"`
}

type part struct {
	Text       string      `json:"text,omitempty"`
	InlineData *inlineData `json:"inline_data,omitempty"`
}

type inlineData struct {
	MIMEType string `json:"mime_type"`
	Data     string `json:"data"` // base64-encoded
}

type generationConfig struct {
	Temperature float64 `json:"temperature"`
}

// generateContentResponse is the top-level Gemini API response.
type generateContentResponse struct {
	Candidates []struct {
		Content struct {
			Parts []struct {
				Text string `json:"text"`
			} `json:"parts"`
		} `json:"content"`
	} `json:"candidates"`
}

// doGenerateContent executes a POST to the generateContent endpoint for the given model.
func (c *clientImpl) doGenerateContent(ctx context.Context, modelName string, reqBody generateContentRequest) (string, error) {
	url := fmt.Sprintf("%s/v1beta/models/%s:generateContent?key=%s", c.apiURL, modelName, c.apiKey)

	bodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return "", fmt.Errorf("marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return "", fmt.Errorf("create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	respBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("read response body: %w", err)
	}

	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("gemini returned status %d: %s", resp.StatusCode, string(respBytes))
	}

	var result generateContentResponse
	if err := json.Unmarshal(respBytes, &result); err != nil {
		return "", fmt.Errorf("decode response: %w", err)
	}

	if len(result.Candidates) == 0 || len(result.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("no content returned from Gemini")
	}

	return result.Candidates[0].Content.Parts[0].Text, nil
}

// TranscribeAudio sends an audio file to Gemini via multimodal inline data.
func (c *clientImpl) TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("gemini API key is not configured")
	}

	if modelName == "" {
		modelName = "gemini-3.1-flash"
	}

	mimeType := audioMIMEType(filename)
	encoded := base64.StdEncoding.EncodeToString(audioData)

	reqBody := generateContentRequest{
		Contents: []content{
			{
				Parts: []part{
					{
						InlineData: &inlineData{
							MIMEType: mimeType,
							Data:     encoded,
						},
					},
					{
						Text: "Transcribe this audio exactly as spoken. Output only the transcription text, nothing else. Do not add explanations, labels, or punctuation beyond what was spoken.",
					},
				},
			},
		},
	}

	text, err := c.doGenerateContent(ctx, modelName, reqBody)
	if err != nil {
		return "", err
	}

	log.Printf("Gemini STT response: %s", text)
	return text, nil
}

// ImproveText sends the raw transcript to Gemini for post-processing.
func (c *clientImpl) ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
	if c.apiKey == "" {
		return "", fmt.Errorf("gemini API key is not configured")
	}

	if modelName == "" {
		modelName = "gemini-3.1-flash"
	}

	reqBody := generateContentRequest{
		SystemInstruction: &systemInstruction{
			Parts: []part{{Text: systemPrompt}},
		},
		Contents: []content{
			{
				Parts: []part{{Text: transcript}},
			},
		},
		GenerationConfig: &generationConfig{
			Temperature: 0.2, // low temperature for consistent text improvement
		},
	}

	return c.doGenerateContent(ctx, modelName, reqBody)
}

// audioMIMEType returns a best-guess MIME type based on the audio filename extension.
func audioMIMEType(filename string) string {
	if len(filename) > 4 {
		switch filename[len(filename)-4:] {
		case ".m4a":
			return "audio/mp4"
		case ".mp3":
			return "audio/mpeg"
		case ".ogg":
			return "audio/ogg"
		case ".wav":
			return "audio/wav"
		case ".aac":
			return "audio/aac"
		}
	}
	if len(filename) > 5 {
		switch filename[len(filename)-5:] {
		case ".webm":
			return "audio/webm"
		case ".flac":
			return "audio/flac"
		}
	}
	// Default to wav as the Swift frontend primarily records in that format
	return "audio/wav"
}

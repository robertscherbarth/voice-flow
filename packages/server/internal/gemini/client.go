package gemini

import (
	"context"
	"fmt"
	"log"
	"path/filepath"
	"time"

	"google.golang.org/genai"
)

// Client defines the interface for communicating with Gemini STT and LLM models.
type Client interface {
	TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error)
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
	ImproveTextStream(ctx context.Context, transcript, modelName, systemPrompt string, onChunk func(string)) error
}

// clientImpl is the concrete implementation of the Gemini Client using the official SDK.
type clientImpl struct {
	client *genai.Client
}

// NewClient returns a new Client pointing to the given Gemini API.
func NewClient(ctx context.Context, apiKey string) (Client, error) {
	config := &genai.ClientConfig{
		APIKey: apiKey,
	}
	client, err := genai.NewClient(ctx, config)
	if err != nil {
		return nil, fmt.Errorf("create genai client: %w", err)
	}

	return &clientImpl{
		client: client,
	}, nil
}

// TranscribeAudio sends an audio file to Gemini using inline data for minimal latency.
func (c *clientImpl) TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("genai client not initialized")
	}
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}

	// 1. Prepare transcription prompt
	prompt := "Transcribe this audio exactly as spoken. Output only the transcription text, nothing else. Do not add explanations, labels, or punctuation beyond what was spoken."

	contents := []*genai.Content{
		{
			Role: "user",
			Parts: []*genai.Part{
				{Text: prompt},
				{
					InlineData: &genai.Blob{
						MIMEType: audioMIMEType(filename),
						Data:     audioData,
					},
				},
			},
		},
	}

	// 2. Generate transcription using inline data
	generateStart := time.Now()
	resp, err := c.client.Models.GenerateContent(ctx, modelName, contents, nil)
	if err != nil {
		return "", fmt.Errorf("generate transcription: %w", err)
	}
	log.Printf("Gemini STT generate latency: %v", time.Since(generateStart))

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	transcript := resp.Candidates[0].Content.Parts[0].Text
	log.Printf("Gemini STT response: %s", transcript)
	return transcript, nil
}

// ImproveText sends the raw transcript to Gemini for post-processing using a system instruction.
func (c *clientImpl) ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error) {
	if c.client == nil {
		return "", fmt.Errorf("genai client not initialized")
	}
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Temperature: ptr(float32(1)),
	}

	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: transcript}},
		},
	}

	llmStart := time.Now()
	resp, err := c.client.Models.GenerateContent(ctx, modelName, contents, config)
	if err != nil {
		return "", fmt.Errorf("improve text: %w", err)
	}
	log.Printf("Gemini LLM latency: %v", time.Since(llmStart))

	if len(resp.Candidates) == 0 || len(resp.Candidates[0].Content.Parts) == 0 {
		return "", fmt.Errorf("empty response from Gemini")
	}

	return resp.Candidates[0].Content.Parts[0].Text, nil
}

// ImproveTextStream sends the transcript to Gemini and streams the improved text back via a callback.
func (c *clientImpl) ImproveTextStream(ctx context.Context, transcript, modelName, systemPrompt string, onChunk func(string)) error {
	if c.client == nil {
		return fmt.Errorf("genai client not initialized")
	}
	if modelName == "" {
		modelName = "gemini-2.0-flash"
	}

	config := &genai.GenerateContentConfig{
		SystemInstruction: &genai.Content{
			Parts: []*genai.Part{{Text: systemPrompt}},
		},
		Temperature: ptr(float32(1)),
	}

	contents := []*genai.Content{
		{
			Role:  "user",
			Parts: []*genai.Part{{Text: transcript}},
		},
	}

	iter := c.client.Models.GenerateContentStream(ctx, modelName, contents, config)
	for resp, err := range iter {
		if err != nil {
			return fmt.Errorf("improve text stream: %w", err)
		}

		if len(resp.Candidates) > 0 && len(resp.Candidates[0].Content.Parts) > 0 {
			chunk := resp.Candidates[0].Content.Parts[0].Text
			onChunk(chunk)
		}
	}
	return nil
}

func ptr[T any](v T) *T {
	return &v
}

// audioMIMEType returns a best-guess MIME type based on the audio filename extension.
func audioMIMEType(filename string) string {
	ext := filepath.Ext(filename)
	switch ext {
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
	case ".webm":
		return "audio/webm"
	case ".flac":
		return "audio/flac"
	default:
		return "audio/wav"
	}
}

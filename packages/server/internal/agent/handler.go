package agent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"time"

	"voice-agent/internal/config"
)

// Define interfaces locally to decouple from specific providers
type LLMClient interface {
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
	ImproveTextStream(ctx context.Context, transcript, modelName, systemPrompt string, onChunk func(string)) error
}

type STTClient interface {
	TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error)
}

// EvalRecord represents a single evaluation request for quality assessment.
type EvalRecord struct {
	Timestamp    time.Time `json:"timestamp"`
	STTModel     string    `json:"stt_model"`
	LLMModel     string    `json:"llm_model"`
	SystemPrompt string    `json:"system_prompt"`
	Transcript   string    `json:"transcript"`
	ImprovedText string    `json:"improved_text"`
}

// Handler handles agent HTTP requests.
type Handler struct {
	llmClient LLMClient
	sttClient STTClient
	cfg       *config.Config
}

// NewHandler creates a new HTTP Handler injected with dependencies.
func NewHandler(llmClient LLMClient, sttClient STTClient, cfg *config.Config) *Handler {
	return &Handler{
		llmClient: llmClient,
		sttClient: sttClient,
		cfg:       cfg,
	}
}

// ServeHTTP handles the /process endpoint logic.
func (h *Handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	// Parse multipart form (max 32 MB)
	if err := r.ParseMultipartForm(32 << 20); err != nil {
		http.Error(w, "Error parsing form", http.StatusBadRequest)
		return
	}

	// Model selection is the server's responsibility based on the configured provider.
	// The frontend does not send model names; we always use the active provider's defaults.
	sttModel := h.cfg.STTModel
	llmModel := h.cfg.LLMModel

	systemPrompt := r.FormValue("system_prompt")
	// Fall back to the server's configured system prompt if none provided
	if systemPrompt == "" {
		systemPrompt = h.cfg.SystemPrompt
	}

	file, header, err := r.FormFile("audio")
	if err != nil {
		http.Error(w, "Error retrieving audio file", http.StatusBadRequest)
		return
	}
	defer file.Close()

	audioData, err := io.ReadAll(file)
	if err != nil {
		http.Error(w, "Error reading audio file", http.StatusInternalServerError)
		return
	}

	log.Printf("Received audio file: %s (%d bytes)", header.Filename, len(audioData))
	log.Printf("Config: STT=%s, LLM=%s", sttModel, llmModel)

	// 1. Send audioData to STT model
	log.Println("Transcribing audio...")
	sttStart := time.Now()
	transcript, err := h.sttClient.TranscribeAudio(r.Context(), audioData, header.Filename, sttModel)
	if err != nil {
		log.Printf("STT Error: %v", err)
		http.Error(w, "Failed to transcribe audio", http.StatusInternalServerError)
		return
	}
	log.Printf("STT latency: %v", time.Since(sttStart))
	log.Printf("Raw Transcript: %s", transcript)

	var improvedText string
	if transcript != "" {
		// 2. Send transcript to LLM model with systemPrompt using streaming
		log.Println("Improving text...")
		llmStart := time.Now()

		// Set headers for streaming
		w.Header().Set("Content-Type", "text/event-stream")
		w.Header().Set("Cache-Control", "no-cache")
		w.Header().Set("Connection", "keep-alive")
		w.Header().Set("Access-Control-Allow-Origin", "*")

		var fullImprovedText string
		err = h.llmClient.ImproveTextStream(r.Context(), transcript, llmModel, systemPrompt, func(chunk string) {
			fullImprovedText += chunk
			// Write chunk as SSE
			fmt.Fprintf(w, "data: %s\n\n", chunk)
			if f, ok := w.(http.Flusher); ok {
				f.Flush()
			}
		})

		if err != nil {
			log.Printf("LLM Streaming Error: %v", err)
			// Since we've already started the stream, we can't http.Error.
			// We can send an error event if the frontend expects it.
			fmt.Fprintf(w, "event: error\ndata: %v\n\n", err)
			return
		}

		improvedText = fullImprovedText
		log.Printf("LLM latency (full stream): %v", time.Since(llmStart))
		log.Printf("Improved Text: %s", improvedText)

		// 3. Save evaluation data if in DevMode
		if h.cfg.DevMode {
			record := EvalRecord{
				Timestamp:    time.Now(),
				STTModel:     sttModel,
				LLMModel:     llmModel,
				SystemPrompt: systemPrompt,
				Transcript:   transcript,
				ImprovedText: improvedText,
			}
			go h.saveEvalData(record)
		}
	} else {
		log.Println("Transcript is empty, skipping LLM improvement.")
		// Return empty JSON if we didn't start a stream
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]string{"text": ""})
	}
}

func (h *Handler) saveEvalData(record EvalRecord) {
	data, err := json.Marshal(record)
	if err != nil {
		log.Printf("Error marshaling eval record: %v", err)
		return
	}

	// Use absolute path for reliability
	absPath, err := filepath.Abs(h.cfg.EvalDataPath)
	if err != nil {
		log.Printf("Error getting absolute path for %s: %v", h.cfg.EvalDataPath, err)
		absPath = h.cfg.EvalDataPath
	}

	// Ensure parent directory exists
	dir := filepath.Dir(absPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		log.Printf("Error creating eval data directory %s: %v", dir, err)
		return
	}

	// Append to file
	f, err := os.OpenFile(absPath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		log.Printf("Error opening eval data file %s: %v", absPath, err)
		return
	}
	defer f.Close()

	if _, err := f.Write(append(data, '\n')); err != nil {
		log.Printf("Error writing to eval data file %s: %v", absPath, err)
	} else {
		log.Printf("Evaluation data saved to %s", absPath)
	}
}

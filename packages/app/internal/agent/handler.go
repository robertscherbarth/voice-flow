package agent

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http"

	"voice-agent/internal/config"
)

// Define interfaces locally to decouple from specific providers
type LLMClient interface {
	ImproveText(ctx context.Context, transcript, modelName, systemPrompt string) (string, error)
}

type STTClient interface {
	TranscribeAudio(ctx context.Context, audioData []byte, filename string, modelName string) (string, error)
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

	sttModel := r.FormValue("stt_model")
	// If the frontend sent the old Ollama model or nothing, use the Mistral default
	if sttModel == "" || sttModel == "karanchopda333/whisper" || sttModel == "base" {
		sttModel = h.cfg.STTModel
	}

	llmModel := r.FormValue("llm_model")
	// Use the configured Mistral default if not provided
	if llmModel == "" || llmModel == "mistral" {
		llmModel = h.cfg.LLMModel
	}

	systemPrompt := r.FormValue("system_prompt")
	// If it's empty or the old prompt, use the new Mistral-tuned prompt
	if systemPrompt == "" || systemPrompt == "You are an assistant that improves the grammar, structure, and clarity of spoken text. Return ONLY the improved text, without conversational filler or preambles." {
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
	transcript, err := h.sttClient.TranscribeAudio(r.Context(), audioData, header.Filename, sttModel)
	if err != nil {
		log.Printf("STT Error: %v", err)
		http.Error(w, "Failed to transcribe audio", http.StatusInternalServerError)
		return
	}
	log.Printf("Raw Transcript: %s", transcript)

	var improvedText string
	if transcript != "" {
		// 2. Send transcript to LLM model with systemPrompt
		log.Println("Improving text...")
		improvedText, err = h.llmClient.ImproveText(r.Context(), transcript, llmModel, systemPrompt)
		if err != nil {
			log.Printf("LLM Error: %v", err)
			http.Error(w, "Failed to improve text", http.StatusInternalServerError)
			return
		}
		log.Printf("Improved Text: %s", improvedText)
	} else {
		log.Println("Transcript is empty, skipping LLM improvement.")
	}

	// Return the final text as JSON
	w.Header().Set("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(map[string]string{"text": improvedText}); err != nil {
		log.Printf("Error encoding response: %v", err)
	}
}

package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"voice-agent/internal/agent"
	"voice-agent/internal/config"
	"voice-agent/internal/gemini"
	"voice-agent/internal/lmstudio"
	"voice-agent/internal/mistral"
)

func main() {
	if err := run(); err != nil {
		log.Fatalf("Agent failed: %v", err)
	}
}

func run() error {
	// Initialize configuration
	cfg := config.New()

	log.Printf("Starting with provider: %s", cfg.Provider)

	// Initialize the provider client for both STT and LLM based on config
	var llmClient agent.LLMClient
	var sttClient agent.STTClient

	switch cfg.Provider {
	case "gemini":
		if cfg.GeminiKey == "" {
			log.Println("WARNING: GEMINI_API_KEY is not set. Transcription and LLM tasks will fail.")
		}
		geminiClient, err := gemini.NewClient(context.Background(), cfg.GeminiKey)
		if err != nil {
			return fmt.Errorf("create gemini client: %w", err)
		}
		llmClient = geminiClient
		sttClient = geminiClient
	case "local":
		// Hybrid: Mistral handles STT (cloud), LM Studio handles text improvement (local).
		if cfg.MistralKey == "" {
			log.Println("WARNING: MISTRAL_API_KEY is not set. Transcription (STT) will fail.")
		}
		log.Printf("LM Studio URL: %s, model: %s", cfg.LMStudioURL, cfg.LMStudioModel)
		sttClient = mistral.NewClient(cfg.MistralURL, cfg.MistralKey)
		llmClient = lmstudio.NewClient(cfg.LMStudioURL)
	case "mistral", "":
		if cfg.MistralKey == "" {
			log.Println("WARNING: MISTRAL_API_KEY is not set. Transcription and LLM tasks will fail.")
		}
		mistralClient := mistral.NewClient(cfg.MistralURL, cfg.MistralKey)
		llmClient = mistralClient
		sttClient = mistralClient
	default:
		return fmt.Errorf("unknown provider %q: must be \"mistral\", \"gemini\", or \"local\"", cfg.Provider)
	}

	// Initialize agent handler
	handler := agent.NewHandler(llmClient, sttClient, cfg)

	// Configure routing
	mux := http.NewServeMux()
	mux.Handle("/process", handler)

	// Configure server
	srv := &http.Server{
		Addr:    ":" + cfg.Port,
		Handler: mux,
	}

	// Channel to listen for errors coming from the listener
	serverErrors := make(chan error, 1)

	// Start the server
	go func() {
		log.Printf("Starting voice-agent server on %s...", srv.Addr)
		serverErrors <- srv.ListenAndServe()
	}()

	// Channel to listen for an interrupt or terminate signal from the OS
	shutdown := make(chan os.Signal, 1)
	signal.Notify(shutdown, os.Interrupt, syscall.SIGTERM)

	// Blocking main and waiting for shutdown
	select {
	case err := <-serverErrors:
		return fmt.Errorf("server error: %w", err)

	case sig := <-shutdown:
		log.Printf("Start shutdown... Signal: %v", sig)

		// Create context for shutdown
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		// Asking listener to shut down and shed load
		if err := srv.Shutdown(ctx); err != nil {
			if err := srv.Close(); err != nil {
				return fmt.Errorf("could not stop server gracefully: %w", err)
			}
			return fmt.Errorf("could not stop server gracefully: %w", err)
		}
	}

	log.Println("Server stopped gracefully")
	return nil
}

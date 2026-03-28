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

	if cfg.MistralKey == "" {
		log.Println("WARNING: MISTRAL_API_KEY is not set. Transcription and LLM tasks will fail.")
	}

	// Initialize Mistral client for both STT and LLM
	mistralClient := mistral.NewClient(cfg.MistralURL, cfg.MistralKey)

	// Initialize agent handler
	handler := agent.NewHandler(mistralClient, mistralClient, cfg)

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

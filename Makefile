-include .env
export

.PHONY: all build-agent run-agent build-app run-app clean test test-integration

# Default target
all: build-app

# Build the Go agent
build-agent:
	@echo "Building Go agent..."
	cd agent && go build -o voice-agent cmd/voice-agent/main.go

# Run the Go agent independently (useful for debugging)
run-agent:
	@echo "Starting Go agent on :8080..."
	cd agent && go run cmd/voice-agent/main.go

# Run unit tests for Go agent
test:
	@echo "Running Go unit tests..."
	cd agent && go test -v -race ./...

# Run integration tests (requires Ollama to be running locally)
test-integration:
	@echo "Running Go integration tests..."
	cd agent && go test -v -tags=integration ./internal/agent/

# Build the Swift macOS application (this also builds and bundles the Go agent)
build-app:
	@echo "Building macOS application..."
	cd SwiftApp && ./build.sh

# Run the Swift macOS application
run-app: build-app
	@echo "Opening VoiceAgent..."
	open SwiftApp/VoiceAgent.app

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf SwiftApp/VoiceAgent.app
	rm -f agent/voice-agent
	@echo "Clean complete."

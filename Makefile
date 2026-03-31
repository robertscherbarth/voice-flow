-include .env
export

.PHONY: all build-server run-server build-desktop run-desktop clean test test-integration

# Default target
all: build-desktop

# Build the Go server
build-server:
	@echo "Building Go server..."
	cd packages/server && ./build.sh

# Run the Go server independently (useful for debugging)
run-server: build-server
	@echo "Starting Go server on :8080..."
	cd build/server && ./server

# Run unit tests for Go server
test:
	@echo "Running Go unit tests..."
	cd packages/server && go test -v -race ./...

# Run integration tests (requires Ollama to be running locally)
test-integration:
	@echo "Running Go integration tests..."
	cd packages/server && go test -v -tags=integration ./internal/agent/

# Build the Swift macOS application (requires the Go server to be built first)
build-desktop: build-server
	@echo "Building macOS desktop application..."
	cd packages/desktop && ./build.sh

# Run the Swift macOS desktop application
run-desktop:
	@echo "Opening VoiceAgent..."
	open build/VoiceAgent.app

# Clean build artifacts
clean:
	@echo "Cleaning up..."
	rm -rf build/
	rm -f packages/server/voice-agent
	@echo "Clean complete."

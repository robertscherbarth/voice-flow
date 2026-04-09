#!/bin/bash
set -e

BUILD_DIR="../../build/server"

echo "Building Go server..."

# Create build directory
mkdir -p "$BUILD_DIR"
mkdir -p "$BUILD_DIR/prompt"

# Build binary
go build -o "$BUILD_DIR/server" cmd/voice-agent/main.go

# Copy prompt config
cp prompt/optimize.yaml "$BUILD_DIR/prompt/"

# Copy app config
cp config.yaml "$BUILD_DIR/"

echo "Server built successfully at $BUILD_DIR"

#!/bin/bash
set -e

APP_NAME="VoiceAgent"
BUNDLE_DIR="$APP_NAME.app"
CONTENTS_DIR="$BUNDLE_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

echo "Building Go agent..."
cd ../app
go build -o voice-agent cmd/voice-agent/main.go
cd ../desktop

echo "Building $APP_NAME..."

# Create Bundle structure
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Copy Go agent
cp ../app/voice-agent "$RESOURCES_DIR/"

# Compile Swift code
swiftc -o "$MACOS_DIR/$APP_NAME" Sources/VoiceAgent/*.swift \
    -framework Cocoa \
    -framework AVFoundation \
    -framework SwiftUI

# Create Info.plist
cat > "$CONTENTS_DIR/Info.plist" <<EOF
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN" "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleExecutable</key>
    <string>$APP_NAME</string>
    <key>CFBundleIdentifier</key>
    <string>com.example.VoiceAgent</string>
    <key>CFBundleName</key>
    <string>$APP_NAME</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>CFBundleShortVersionString</key>
    <string>1.0</string>
    <key>LSUIElement</key>
    <true/>
    <key>NSMicrophoneUsageDescription</key>
    <string>VoiceAgent needs microphone access to record your voice.</string>
</dict>
</plist>
EOF

echo "Build successful: $BUNDLE_DIR"

#!/bin/bash
set -e

APP_NAME="VoiceAgent"
BUILD_DIR="../../build"
BUNDLE_DIR="$BUILD_DIR/$APP_NAME.app"
CONTENTS_DIR="$BUNDLE_DIR/Contents"
MACOS_DIR="$CONTENTS_DIR/MacOS"
RESOURCES_DIR="$CONTENTS_DIR/Resources"

echo "Building $APP_NAME..."

# Remove old bundle if it exists to avoid stale files
rm -rf "$BUNDLE_DIR"

# Create Bundle structure
mkdir -p "$MACOS_DIR"
mkdir -p "$RESOURCES_DIR"

# Copy App Icon
cp AppIcon.icns "$RESOURCES_DIR/"

# Copy Go server binary and prompt config (must be built before running this script)
if [ ! -f "../../build/server/server" ]; then
    echo "Error: Go server binary not found at ../../build/server/server"
    echo "Please build the server first using: make build-server"
    exit 1
fi
cp "../../build/server/server" "$RESOURCES_DIR/"
cp -R "../../build/server/prompt" "$RESOURCES_DIR/"

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
    <key>CFBundleIconFile</key>
    <string>AppIcon</string>
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

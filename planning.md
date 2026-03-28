# Local "Wispr Flow" Clone - Project Plan

## Overview
A native macOS menu bar application designed as a fast, private dictation tool. It uses a split-architecture approach: a native Swift/AppKit frontend for flawless OS integration (audio capture, hotkeys, pasting), and a headless Go server backend for ultra-fast, offline AI orchestration using Ollama.

## Architecture & Tech Stack

### 1. The macOS Native Frontend (Swift / AppKit)
This is the user-facing application that lives in the macOS menu bar.
*   **Language**: Swift (AppKit/SwiftUI hybrid for preferences).
*   **UI**: Menu bar item (`NSStatusItem`) showing current state (Idle ⚪️, Recording 🔴, Processing ⏳). Native preferences window for configuration.
*   **Input**: Push-to-Talk via the `Fn` key, managed via `NSEvent.addGlobalMonitorForEvents(matching: .flagsChanged)`.
*   **Audio Capture**: Apple's `AVFoundation` to record high-quality audio directly to a temporary `.wav` file.
*   **Output / OS Integration**:
    *   `NSPasteboard` (Writes the final text to the macOS clipboard).
    *   Accessibility APIs (`CGEvent` or `osascript`) to simulate `Cmd + V` for auto-pasting.

### 2. The Go Agent Backend (Golang)
A lightweight background server that is launched and managed by the Swift frontend.
*   **Language**: Go (Golang).
*   **Role**: Stateless HTTP server running locally (e.g., port 8080).
*   **AI Pipeline (Ollama)**: Connects to a local Ollama instance at `http://localhost:11434`.
    *   **STT (Speech-to-Text)**: Forwards audio to an Ollama Whisper model (`karanchopda333/whisper`).
    *   **Text Improvement**: Sends raw transcripts to a local LLM (e.g., `mistral`) with a strict system prompt to remove conversational filler and fix grammar.

## Data Flow
1. User presses and holds the `Fn` key (Push-to-Talk).
2. Swift app starts recording via `AVFoundation` to a temp `.wav` file.
3. User releases the `Fn` key to stop recording.
4. Swift app sends the `.wav` file and user preferences (models, prompt) to the Go backend via HTTP `POST /process`.
5. Go backend sends audio to Ollama Whisper endpoint -> Receives raw transcript.
6. Go backend sends raw transcript to Ollama LLM endpoint with prompt -> Receives improved text.
7. Go backend returns improved text to Swift frontend.
8. Swift app copies text to `NSPasteboard` and triggers macOS `Cmd + V`.
9. Text instantly appears in the user's active application window.

## Security & Permissions
macOS will require the following permissions, handled natively by the Swift frontend:
*   **Microphone**: To capture audio via `AVFoundation`.
*   **Accessibility**: To listen for global `Fn` key events and inject `Cmd + V` keystrokes.
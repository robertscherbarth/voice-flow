# Local "Wispr Flow" Clone - System Design

## Overview
A split-architecture, local-first dictation tool designed to run entirely offline with zero API costs. It mimics the native experience of Wispr Flow by using a Swift macOS frontend for flawless OS integration, and a Go backend server for fast AI orchestration via local Ollama models.

## Architecture

**1. The macOS Native Frontend (Swift / AppKit)**
This is the user-facing application. It runs as a menu bar app (`NSStatusItem`) without a dock icon. Its responsibilities are strictly tied to the OS:
*   **Lifecycle Management:** When the user launches the app, it quietly spins up the Go backend server as a background subprocess and ensures it shuts down when the app quits.
*   **Input Handling:** Listens for a global push-to-talk hotkey (press and hold `Fn` key) to toggle recording.
*   **Audio Capture:** Uses Apple's native `AVFoundation` to record high-quality microphone audio directly to a temporary `.wav` file.
*   **Output Handling:** Receives the finalized text from the Go server, copies it to the macOS clipboard (`NSPasteboard`), and uses Accessibility APIs to simulate a `Cmd+V` keystroke, injecting the text into whatever app the user is currently using.

**2. The Go Agent Backend (Golang)**
This is a lightweight, headless HTTP server running on `localhost` (e.g., port 8080). It acts as the brain of the operation.
*   **API Endpoint:** Exposes a `POST /process` endpoint that accepts the audio file and configuration settings from the Swift frontend.
*   **AI Orchestration:** Acts as the bridge to the local Ollama instance (`http://localhost:11434`). It first sends the audio to the local Whisper model for raw Speech-to-Text.
*   **Text Refinement:** It takes the raw transcript and sends it to a local LLM (e.g., `mistral`) with a strict system prompt to fix grammar, remove filler words, and format the text properly.
*   **Response:** Returns the polished text as an HTTP response back to the Swift app.

## Data Flow & User Experience

1. **Push-to-Talk:** You are typing in any app. You press and hold the `Fn` key.
2. **Recording State:** The Swift app intercepts the hotkey. The menu bar icon changes from an Idle state (⚪️) to a Recording state (🔴). `AVFoundation` begins writing audio from the default microphone to a temporary `.wav` file.
3. **Completion:** You release the `Fn` key. The icon changes to a Processing state (⏳). Recording stops and the file is closed.
4. **Handoff to Go:** The Swift app reads the `.wav` file and sends it as a multipart form data request (with config parameters) to the Go backend at `http://localhost:8080/process`.
5. **AI Processing (Go):** 
   - Go forwards the audio to the local Ollama Whisper endpoint for transcription.
   - Go immediately wraps the raw text in a system prompt and sends it to the local Ollama LLM.
6. **Delivery & Injection:** The Go server responds to Swift with the final polished text. The Swift app copies this text to the `NSPasteboard` (the clipboard). It then uses Accessibility APIs (`CGEvent` or `osascript`) to simulate the `Cmd + V` keystroke, instantly pasting the text into your active window. The tray icon returns to Idle (⚪️).

## Configuration & Error Handling

*   **Preferences Window:** A native macOS settings window where the user can define their preferred Ollama STT model, LLM model, and custom system prompts. These settings are passed dynamically to the Go server on each request.
*   **Permissions:** On first launch, the app explicitly requests Microphone (for recording) and Accessibility (for pasting) permissions, guiding the user to macOS System Settings if denied.
*   **Error Handling:** If Ollama is down, a request times out, or the transcription fails, the Go server returns an error code. The Swift app catches this, plays a system beep, displays a macOS notification, and safely resets the menu bar state so it doesn't get stuck.
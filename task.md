# Local "Wispr Flow" Clone - Concrete Tasks

## Phase 1: Go Backend Foundation
- [x] Initialize the Go module (`go mod init voice-agent`).
- [x] Create a basic HTTP server listening on `localhost:8080`.
- [x] Implement a `POST /process` endpoint that accepts:
    - [x] A `.wav` file (multipart form data).
    - [x] Configuration parameters (stt_model, llm_model, system_prompt).
- [x] Implement the Ollama STT client:
    - [x] Send the `.wav` file to the local `karanchopda333/whisper` model endpoint in Ollama.
    - [x] Parse the JSON response to extract the raw text transcription.
- [x] Implement the Ollama LLM client:
    - [x] Send the raw text to a local text model (e.g., `mistral`).
    - [x] Parse the JSON response to extract the finalized, polished text.
- [x] Return the final text as an HTTP response to the caller.

## Phase 2: Swift Frontend Foundation
- [x] Initialize an Xcode Project (macOS App, Swift/AppKit).
- [x] Set up the `NSStatusItem` for the macOS menu bar (hide the dock icon in `Info.plist` via `LSUIElement`).
- [x] Create and assign simple 16x16 icon states (Idle ⚪️, Recording 🔴, Processing ⏳) to update the tray dynamically.
- [x] Build a simple native macOS Preferences window (SwiftUI/AppKit) to configure:
    - [x] STT Model name.
    - [x] LLM Model name.
    - [x] Custom system prompt for the text improvement phase.

## Phase 3: Audio Capture & Input Handling (Swift)
- [x] Implement `NSEvent.addGlobalMonitorForEvents(matching: .flagsChanged)` to detect when the `Fn` key is pressed and released (Push-to-Talk).
- [x] Ensure the app correctly prompts the user for Accessibility permissions on first launch.
- [x] Integrate `AVFoundation` to capture microphone input when the `Fn` key is held.
    - [x] Prompt for Microphone permissions.
    - [x] Stream captured audio into a temporary `.wav` file in `NSTemporaryDirectory`.
    - [x] Stop recording and close the file when the `Fn` key is released.

## Phase 4: Integration & OS Output
- [x] Write the Swift networking code to construct a `multipart/form-data` request with the `.wav` file and Preferences data.
- [x] Send the request to the local Go server (`http://localhost:8080/process`) when the `Fn` key is released.
- [x] Handle error states from the Go server gracefully (e.g., play a system beep, reset menu bar icon, show a notification).
- [x] Upon success, write the received improved text to the macOS clipboard (`NSPasteboard.general`).
- [x] Implement the AppleScript or `CGEvent` auto-paste mechanism (`Cmd + V`).

## Phase 5: Packaging & Polish
- [x] Bundle the Go binary (`voice-agent`) inside the Swift `.app` bundle (e.g., in the `Resources` folder).
- [x] Implement lifecycle management in the Swift `AppDelegate`:
    - [x] Launch the bundled Go binary as an `NSTask`/`Process` when the Swift app starts.
    - [x] Terminate the Go process cleanly when the Swift app quits.
- [x] Test the full push-to-talk flow across different active applications (Notes, browser, terminal).
- [x] Document the required macOS System Preferences in a `README.md`.
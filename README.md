# VoiceAgent

A local "Wispr Flow" clone. It lives in your macOS menu bar, listens for a global hotkey (`Fn`), records your voice, and uses Ollama to transcribe and improve the text, then automatically pastes it into your active window.

## How to Build and Run

1. Make sure you have Go and Xcode Command Line Tools installed.
2. Ensure Ollama is running locally (`http://localhost:11434`) with your desired models installed (e.g. `karanchopda333/whisper` and `mistral`).
3. Run the build script to compile the Go backend and package the Swift frontend:
   ```bash
   cd SwiftApp
   ./build.sh
   ```
4. Open the built application:
   ```bash
   open VoiceAgent.app
   ```

## Permissions

On first launch, the app requires:
1. **Accessibility**: To listen for the `Fn` key globally and to simulate the `Cmd+V` paste keystroke. You may need to go to System Settings -> Privacy & Security -> Accessibility and add the app.
2. **Microphone**: To record your voice. You will be prompted.

## Usage

1. Press and hold the **`Fn`** key.
2. Speak your thought. (Menu bar icon turns 🔴).
3. Release the **`Fn`** key. (Menu bar icon turns ⏳).
4. The text will be polished by Ollama and automatically pasted into whatever app you currently have active!

## Configuration

Click on the ⚪️ icon in your menu bar and select **Preferences...** to configure your Ollama models and system prompt.

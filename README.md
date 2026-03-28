# VoiceAgent

VoiceAgent is a macOS menu bar application that acts as a local, privacy-respecting "Wispr Flow" clone. It listens for a global hotkey, records your voice, and uses the Mistral API to accurately transcribe and polish the text before automatically pasting it into your active window.

The project is structured into two main components:
*   **Frontend (`packages/desktop`)**: A Swift-based macOS menu bar application handling the UI, hotkeys, and system permissions.
*   **Backend (`packages/app`)**: A Go server bundled inside the macOS app that processes audio and communicates with Mistral's Speech-to-Text (STT) and Large Language Model (LLM) endpoints.

---

## Tutorial: Getting Started

This tutorial will guide you through setting up your environment, building the application, and running it for the first time.

### Prerequisites

Ensure you have the following installed on your macOS system:
*   **Go** (for compiling the backend)
*   **Xcode Command Line Tools** (for packaging the Swift application)
*   **Mistral API Key** (for transcription and text processing)

### Step 1: Configure Your Environment

VoiceAgent uses a `.env` file to manage secrets. Create a `.env` file in the root directory of the project and add your Mistral API key:

```bash
echo "MISTRAL_API_KEY=your_actual_api_key_here" > .env
```

### Step 2: Build the Application

Use the provided `Makefile` to compile the Go backend and package the Swift frontend into a macOS `.app` bundle. The `Makefile` automatically reads your `.env` file.

```bash
make build-app
```

### Step 3: Run VoiceAgent

Launch the compiled application using the `Makefile` command to ensure the environment variables are correctly inherited:

```bash
make run-app
```

On your first launch, macOS will prompt you for specific system permissions (Accessibility and Microphone).

---

## How-to Guide: Using VoiceAgent

Once VoiceAgent is running in your menu bar and permissions are granted, you can dictate text into any application.

1. **Focus a window:** Click into the text field or document where you want the text to appear.
2. **Start recording:** Press and hold the **`Fn`** key. The menu bar icon will turn 🔴 to indicate it is recording your voice.
3. **Speak your thought:** Dictate your sentence or paragraph.
4. **Process and paste:** Release the **`Fn`** key. The menu bar icon will turn ⏳ while the Mistral API transcribes and polishes the text. The final result will be automatically pasted into your active window.

---

## How-to Guide: Configuration

You can customize the Mistral models VoiceAgent uses and the system prompt that dictates how your text is formatted.

1. Click on the VoiceAgent icon (⚪️) in your macOS menu bar.
2. Select **Preferences...** from the dropdown menu.
3. Adjust the **STT Model** (defaults to `voxtral-mini-latest`) and **LLM Model** (defaults to `mistral-small-latest`).
4. Modify the **System Prompt** to change how the AI edits your speech (e.g., instructing it to translate to another language or adopt a specific writing style).

---

## Reference: Permissions Details

VoiceAgent requires specific macOS security permissions to function correctly. If the app fails to respond to your hotkey or paste text, ensure these are enabled in **System Settings -> Privacy & Security**.

*   **Accessibility:** Required to globally listen for the `Fn` key press (even when VoiceAgent is not the active app) and to simulate the `Cmd+V` keystroke used to automatically paste the final text.
*   **Microphone:** Required to record your voice while the hotkey is held down. You will be prompted for this permission during your first recording attempt.

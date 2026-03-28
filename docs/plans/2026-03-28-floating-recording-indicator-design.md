# Floating Recording Indicator Design

## Overview
A persistent, always-visible macOS floating utility window that indicates the current recording state (Recording, Paused, Waiting). It uses an audio/voice-focused iconography approach (Option B) to clearly communicate to the user when the application is actively listening.

## Architecture

### State Management
A central `@Observable` (or `ObservableObject`) view model manages a `RecordingState` enum:
- `recording`: Actively capturing audio.
- `paused`: Recording is temporarily halted.
- `waiting`: Processing input or waiting to begin.

### Window Presentation (AppKit Bridging)
To achieve an "always-on-top", borderless, and draggable window on macOS, the application utilizes AppKit's `NSPanel` rather than a standard SwiftUI `WindowGroup`.
- **Level:** `.floating` (stays above other desktop windows).
- **StyleMask:** `[.borderless, .nonactivatingPanel]` (no title bar, does not steal keyboard focus).
- **Behavior:** `isMovableByWindowBackground = true` allows clicking and dragging anywhere on the view.
- **Hosting:** The SwiftUI view is embedded using an `NSHostingController`.

## Visual Design & Animations

### Appearance
- **Size:** Compact, ~44x44 points.
- **Shape:** Circular pill shape.
- **Background:** `.regularMaterial` with a subtle `.shadow(radius: 4)` for a native macOS frosted glass depth effect.

### Iconography (Option B - Audio Focus)
The indicator utilizes SF Symbols to represent states:
- **Recording:** `mic.fill` (Tint: `.red`). Features a subtle `.symbolEffect(.pulse.byLayer)` for a "live" feel.
- **Paused:** `mic.slash.fill` (Tint: `.secondary` / Gray). Static.
- **Waiting:** `waveform` (Tint: `.orange` or `.blue`). Features `.symbolEffect(.variableColor)` to indicate active processing.

### Transitions
State changes are animated using `.contentTransition(.symbolEffect(.replace))`, ensuring the microphone smoothly morphs into the slash or waveform.

### Interaction
- **Click & Drag:** Moves the floating window around the screen.
- **Tap/Click:** Toggles or cycles the recording state.

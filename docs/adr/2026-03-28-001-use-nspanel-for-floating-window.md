# ADR 001: Use NSPanel for Always-On-Top Floating Window

## Date
2026-03-28

## Status
Accepted

## Context
The application requires a persistent, always-visible floating indicator for the
recording state on macOS. It must stay above all other desktop windows and be
draggable without stealing keyboard focus or displaying a standard window title bar.

Pure SwiftUI's `WindowGroup` lacks the necessary modifiers to set a window level to
`.floating`, make it borderless, and allow it to be movable by clicking its background
natively without relying on private or complex introspection modifiers.

The indicator needs to communicate three distinct states to the user:
- `recording` — actively capturing audio
- `waiting` — processing input or waiting for the server response
- `idle` — no active session

The desired visual design is a compact (~44×44 pt) circular pill with a
`.regularMaterial` background (frosted glass) and a subtle shadow, using SF Symbols to
represent each state:
- `mic.fill` in red with a `.pulse.byLayer` symbol effect for recording
- `waveform` in orange/blue with `.variableColor` for waiting/processing
- `mic.slash.fill` in gray (static) for idle/paused

State transitions animate using `.contentTransition(.symbolEffect(.replace))` so the
icon morphs smoothly between states.

## Decision
Use AppKit's `NSPanel` configured with `.nonactivatingPanel` and `.borderless` style
masks, hosted via `NSHostingController`, to contain the SwiftUI indicator view.

Key configuration:
- `panel.level = .floating` — stays above all other windows
- `panel.styleMask = [.borderless, .nonactivatingPanel]` — no title bar, no focus steal
- `panel.isMovableByWindowBackground = true` — click-and-drag anywhere on the view
- SwiftUI state is driven by a shared `FloatingIndicatorManager` (`ObservableObject`)
  with a `@Published var state: RecordingState`

## Consequences

**Positive:**
- Achieves the exact desired UX: always on top, draggable anywhere, borderless.
- Prevents the window from stealing keyboard focus, allowing users to continue typing
  in other applications while the indicator is visible.
- Fully supported and documented Apple macOS API.
- SwiftUI handles all rendering and animation; AppKit is only used for the window shell.

**Negative:**
- Introduces an AppKit dependency into a predominantly SwiftUI application.
- Requires manual window lifecycle management (instantiating and configuring the
  `NSPanel`) instead of a purely declarative `WindowGroup` approach.

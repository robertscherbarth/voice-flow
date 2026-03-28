# ADR 001: Use NSPanel for Always-On-Top Floating Window

## Date
2026-03-28

## Status
Accepted

## Context
The application requires a persistent, always-visible floating indicator for the recording state on macOS. It must stay above all other desktop windows and be draggable without stealing keyboard focus or displaying a standard window title bar.

Pure SwiftUI's `WindowGroup` lacks the necessary modifiers to set a window level to `.floating`, make it borderless, and allow it to be movable by clicking its background natively without relying on private or complex introspection modifiers.

## Decision
We will use AppKit's `NSPanel` wrapped via `NSHostingController` to host the SwiftUI view.

## Consequences

**Positive:**
- Achieves the exact desired UX (always on top, draggable anywhere, borderless).
- Prevents the window from stealing keyboard focus (using `.nonactivatingPanel`), allowing users to continue typing in other applications while interacting with the recording indicator.
- Fully supported and documented Apple macOS API.

**Negative:**
- Introduces an AppKit dependency into a predominantly SwiftUI application.
- Requires manual window lifecycle management (instantiating and configuring the `NSPanel`) instead of a purely declarative data-driven window approach (like `WindowGroup`).

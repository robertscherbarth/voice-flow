# ADR 003: Split Swift Frontend / Embedded Go Backend Architecture

## Date
2026-03-28

## Status
Accepted

## Context
A macOS menu-bar dictation app needs two very different capabilities that pull in
opposite directions:

1. **Deep OS integration** — global hotkeys, audio capture via `AVFoundation`,
   clipboard writes, simulated keypresses via `CGEvent`, `NSStatusItem` menu bar
   presence, and subprocess lifecycle management. These require Swift and AppKit; they
   cannot be done from a generic backend language.

2. **AI orchestration** — multipart HTTP requests to a Speech-to-Text API (Mistral
   `voxtral-mini-latest`), LLM text improvement (Mistral `mistral-small-latest`),
   response parsing, system-prompt management, evaluation data collection, and config
   loading. These are pure business logic with no OS dependency and benefit from fast
   iteration, easy testing, and a rich standard library.

Implementing both in Swift would mean writing all HTTP and AI logic against Foundation
networking with no testing story and slow compile cycles. Implementing both in Go would
require CGo or inter-process bridging just to capture audio or interact with the
pasteboard.

## Decision
Separate the application into two processes with a clean HTTP boundary:

- **Swift frontend** (`packages/desktop/`) handles all OS integration exclusively.
  It has no AI logic, no prompt management, and no direct knowledge of which models
  are in use. It records audio to a temp file, POSTs it to `localhost:8080/process`,
  and pastes the returned text into the active window.

- **Go backend** (`packages/server/`) is a lightweight `net/http` server that handles
  all AI orchestration, HTTP routing, config loading, and evaluation data. It exposes
  a single `POST /process` endpoint.

- The Go binary is compiled into `build/server/server` and copied into the `.app`
  bundle at `Contents/Resources/` by `make build-desktop`. `AgentManager.swift`
  launches it as a subprocess on app start and terminates it on quit. The subprocess's
  working directory is set to `resourcePath` so all relative paths (e.g. to
  `prompt/optimize.yaml`) resolve correctly at runtime.

- The LLM system prompt lives in `packages/server/prompt/optimize.yaml` and is loaded
  at server startup. It can be edited without recompiling either binary.

## Consequences

**Positive:**
- Each layer can be developed and tested independently. The Go server runs standalone
  via `make run-server` without launching the Swift app.
- Go's standard library and table-driven test tooling make the AI/HTTP layer easy to
  unit-test with mocks (`mockLLMClient`, `mockSTTClient`).
- Swapping AI providers or models requires only Go changes; the Swift app is unaffected.
- The external prompt file (`optimize.yaml`) allows prompt iteration without any
  recompilation.
- Clean separation enforces a discipline: if a new feature requires OS access, it goes
  in Swift; if it requires AI or HTTP logic, it goes in Go.

**Negative:**
- The app bundle must contain two compiled binaries, increasing build complexity.
- Subprocess lifecycle management (launch, health-check, teardown) adds code to the
  Swift layer (`AgentManager.swift`).
- The HTTP boundary introduces latency (negligible on localhost) and a failure mode
  (server not yet ready) that requires the Swift app to handle gracefully.
- No Xcode project means no Swift Package Manager, no Swift test target, and no code
  signing configuration out of the box.

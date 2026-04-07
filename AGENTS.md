# AGENTS.md — Coding Agent Guidelines for voice-flow

This file documents build commands, test commands, and code style conventions for
agentic coding assistants working in this repository.

---

## Project Overview

**voice-flow** is a macOS menu-bar dictation app (similar to Wispr Flow). It has a
split architecture:

- `packages/desktop/` — Swift macOS frontend (AppKit + SwiftUI + AVFoundation),
  compiled directly with `swiftc` (no Xcode project, no Swift Package Manager)
- `packages/server/` — Go HTTP backend (`net/http` only, no framework) that handles
  audio transcription and LLM text improvement via the Mistral API
- The Go binary is embedded in the `.app` bundle under `Resources/` and launched as a
  subprocess by the Swift app at runtime

---

## Build Commands

All commands are run from the **repo root** via `make`:

```bash
make build-server        # Compile Go binary → build/server/server
make build-desktop       # Compile Go, then package Swift → build/VoiceAgent.app
make run-server          # Build + run Go server standalone on :8080
make run-dev             # Build + run Go server with DEV_MODE=true (saves eval data)
make run-desktop         # open build/VoiceAgent.app
make clean               # Remove build/ and packages/server/voice-agent
```

To build individual packages directly:

```bash
# Go server (from packages/server/)
go build -o ../../build/server/server cmd/voice-agent/main.go

# Swift desktop (from packages/desktop/) — see build.sh for full swiftc invocation
bash build.sh
```

---

## Test Commands

### Go — Unit Tests

```bash
make test
# equivalent to (from packages/server/):
go test -v -race ./...
```

### Go — Run a Single Test Function

```bash
# From packages/server/
go test -v -run TestProcessHandler ./internal/agent/
go test -v -run TestNew ./internal/config/
go test -v -run TestImproveText ./internal/ollama/

# Run a specific subtest
go test -v -run "TestProcessHandler/successful_processing" ./internal/agent/
```

### Go — Integration Tests

Integration tests require a real `MISTRAL_API_KEY` and are gated by a build tag:

```bash
make test-integration
# equivalent to (from packages/server/):
MISTRAL_API_KEY=<key> go test -v -tags=integration ./internal/agent/

# Single integration test
MISTRAL_API_KEY=<key> go test -v -tags=integration -run TestIntegration_ProcessAudio ./internal/agent/
```

Integration test files begin with `//go:build integration` and live in
`package agent_test` (external test package). They call `t.Skip` when
`MISTRAL_API_KEY` is unset.

### Swift — No Tests

There is currently no Swift test target. The desktop package has no `Tests/`
directory and no `Package.swift`.

---

## Go Code Style

### Imports

Group imports in two blocks separated by a blank line: stdlib first, then internal
module paths. `goimports` / `gofmt` ordering (alphabetical within each block).

```go
import (
    "context"
    "encoding/json"
    "fmt"
    "log"
    "net/http"

    "voice-agent/internal/config"
    "voice-agent/internal/mistral"
)
```

### Formatting

- Standard `gofmt` / `goimports` — tabs for indentation, no line-length limit
- Run `gofmt -w .` before committing

### Naming

| Element | Convention | Example |
|---|---|---|
| Package | short, lowercase, one word | `agent`, `config`, `mistral` |
| Exported type | PascalCase | `Handler`, `Config`, `EvalRecord` |
| Unexported impl type | suffix `Impl` | `clientImpl` |
| Constructor | `New` or `NewXxx`, returns interface | `func NewClient(...) Client` |
| Unexported func/var | camelCase | `loadSystemPrompt`, `sttModel` |
| Acronyms | fully uppercased | `STTModel`, `LLMClient`, `MistralURL` |
| Test mocks | `mock` prefix + interface name | `mockLLMClient`, `mockSTTClient` |
| Test table var | `tt` (conventional) | `for _, tt := range tests` |

### Interfaces

Define interfaces in the **consuming** package, not in the implementing package.
This allows mocking without circular imports.

```go
// Defined in internal/agent/handler.go — not in internal/mistral/
type LLMClient interface {
    ImproveText(ctx context.Context, text string) (string, error)
}
```

### Error Handling

- Wrap errors with a short verb phrase: `fmt.Errorf("marshal request: %w", err)`
- HTTP handler errors: `http.Error(w, message, statusCode)` then `return`
- Non-fatal operational errors: `log.Printf("Error saving eval data: %v", err)` and
  continue (graceful degradation)
- Fatal startup errors: `log.Fatalf(...)` in `main()`
- Prefer `if err := ...; err != nil` single-line check pattern

### Tests

- All unit tests use table-driven style: `[]struct{ name string; ... }` + `t.Run`
- Mock structs embed a function field per method so each test case can override
  behavior inline: `improveTextFunc func(ctx context.Context, text string) (string, error)`
- Unit tests live in `package agent` (internal); integration tests in `package agent_test`

---

## Swift Code Style

### Imports

List imports at the top with no blank lines between them, most fundamental first:

```swift
import Cocoa        // AppKit base
import SwiftUI      // SwiftUI layer
import AVFoundation // domain-specific frameworks after
```

Import only what a file actually uses.

### Formatting

- **4 spaces** for indentation (not tabs)
- One blank line between method definitions within a type
- Closing braces on their own lines
- No trailing whitespace

### Types

- `class` for stateful managers and AppKit-adjacent objects
- `struct` for SwiftUI views and simple data models
- `static let shared = TypeName()` for singletons
- `enum` with camelCase cases for state: `idle`, `recording`, `waiting`
- `@AppStorage` for persisted preferences; `@Published` + `ObservableObject` for UI state
- `Result<T, Error>` for async callback return types

### Naming

| Element | Convention | Example |
|---|---|---|
| Type | PascalCase | `AgentClient`, `FloatingIndicatorManager` |
| Variable / function | camelCase | `statusItem`, `startRecordingFlow()` |
| File | one type per file, named after it | `AgentClient.swift` |
| ObjC-visible selectors | `@objc` prefix | `@objc func showPreferences()` |

### Memory Management

- Always use `[weak self]` in closures that capture `self`
- Follow with `guard let self = self else { return }` before using `self`

### Error Handling

- `do { ... } catch { print("Failed to ...") }` for throwing calls
- `guard let x = optional else { ...; return }` for unwrapping with early return
- Switch over `Result`: `case .success(let value):` / `case .failure(let error):`
- `print(...)` for diagnostics — no structured logging, no `os.log`
- `NSSound.beep()` for user-visible errors in the menu-bar context

### Comments

- Plain `//` line comments only — no doc comments (`///`)
- Comments explain **why**, not what: `// Wait for pasteboard to sync before typing`
- Inline state labels are acceptable: `updateIcon("⚪️") // Idle`

---

## Architecture Conventions

**Strict layer separation:** The Swift frontend handles only OS integration — hotkeys,
audio capture, permissions, clipboard writes, simulated keypresses, and subprocess
lifecycle. All AI orchestration, HTTP handling, and business logic lives in Go.

**External system prompt:** `packages/server/prompt/optimize.yaml` holds the LLM
system prompt. Edit it without recompiling. The server probes multiple relative paths
(`../../prompt/`, `../prompt/`, etc.) to find the file from different working
directories (binary vs. test runner).

**DEV_MODE:** Set `DEV_MODE=true` to enable saving evaluation data (JSONL) alongside
audio. Use `make run-dev` locally. Never enable in production.

**Go server embedding:** `make build-desktop` copies the Go binary into
`build/VoiceAgent.app/Contents/Resources/`. `AgentManager.swift` launches it as a
subprocess with its working directory set to `resourcePath` so relative paths resolve.

---

## Documentation Conventions

- **README.md** follows the Diátaxis structure: Tutorial → How-to Guide → Reference
- **All design and architecture documentation lives in `docs/adr/`** as Architecture
  Decision Records. There is no separate plans or design directory.
- ADR filename format: `YYYY-MM-DD-NNN-slug.md` (e.g. `2026-03-28-003-split-swift-go-architecture.md`)
- ADR sections: Date / Status / Context / Decision / Consequences
- Write an ADR in `docs/adr/` before making any significant architectural change

# ADR 004: Multi-Provider LLM Support (Mistral + Gemini)

**Date:** 2026-04-09

**Status:** Proposed

## Context

The voice-flow server currently hardcodes Mistral as the sole provider for both
speech-to-text (Voxtral) and LLM text improvement. We want to support Google
Gemini as an alternative provider, using its multimodal capabilities for both STT
(audio-in) and text improvement. The active provider should be selectable via
configuration, with Mistral remaining the default.

## Decision

### Provider Abstraction

The `agent` package already defines narrow `LLMClient` and `STTClient` interfaces.
We add a new `internal/gemini/` package whose `clientImpl` satisfies both interfaces,
mirroring the existing `internal/mistral/` package. The handler remains unchanged —
it programs against interfaces, not concrete types.

### Configuration

Introduce a new `packages/server/config.yaml` file:

```yaml
provider: mistral

mistral:
  api_url: "https://api.mistral.ai"
  stt_model: "voxtral-mini-latest"
  llm_model: "mistral-small-latest"

gemini:
  api_url: "https://generativelanguage.googleapis.com"
  stt_model: "gemini-2.0-flash"
  llm_model: "gemini-2.0-flash"
```

API keys stay in `.env` as `MISTRAL_API_KEY` and `GEMINI_API_KEY`. Environment
variables override YAML values (e.g., `PROVIDER=gemini` overrides the YAML default).

The existing `prompt/optimize.yaml` is unchanged and shared by all providers.

### Gemini API Integration

- **STT:** Use Gemini's multimodal `generateContent` endpoint with inline audio
  data and a transcription-only prompt. This avoids requiring a separate Google
  Cloud Speech-to-Text setup.
- **LLM:** Use the same `generateContent` endpoint with `systemInstruction` for
  the system prompt and `temperature: 0.2`.
- **Auth:** API key passed as query parameter (`?key=`), per Gemini convention.

### Wiring (main.go)

A switch on `cfg.Provider` creates the appropriate client and injects it into
`agent.NewHandler` as both `LLMClient` and `STTClient`. No mixed-provider
configurations for now — the selected provider handles everything.

## Consequences

- **Backward compatible:** Without any config changes, Mistral is used (same as
  today). Existing `.env` files continue to work.
- **New dependency surface:** Gemini REST API adds a second external dependency,
  but with the same isolation pattern as Mistral.
- **config.yaml must be found at runtime:** The same path-probing strategy used for
  `prompt/optimize.yaml` is reused for `config.yaml`. The build script must copy it
  into the app bundle.
- **Future providers:** The pattern extends naturally — add a new package, add a
  YAML section, add a case to the switch.
- **No mixed providers:** A single provider handles both STT and LLM. If a future
  need arises (e.g., Gemini STT + Mistral LLM), the config can be extended then.

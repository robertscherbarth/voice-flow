# ADR 002: Save Evaluation Data in Development Mode

## Date
2026-04-01

## Status
Accepted

## Context
When running the voice-agent server locally there is no way to systematically review
the output quality of the AI models. To improve the system prompt and evaluate the
performance of different models (STT and LLM), every request's raw transcript and
improved text needs to be captured for later analysis.

This capability must have zero impact on production: no extra disk I/O, no behaviour
change, no risk of leaking user audio data.

## Decision
Introduce a `DEV_MODE` boolean configuration flag (read from the `DEV_MODE`
environment variable). When enabled, the `/process` HTTP handler asynchronously
appends a structured record to a local **JSON Lines (JSONL)** file after each
successful request.

The record type (`EvalRecord`) captures:
```go
type EvalRecord struct {
    Timestamp    time.Time `json:"timestamp"`
    STTModel     string    `json:"stt_model"`
    LLMModel     string    `json:"llm_model"`
    SystemPrompt string    `json:"system_prompt"`
    Transcript   string    `json:"transcript"`
    ImprovedText string    `json:"improved_text"`
}
```

Records are appended to `test-data/evaluation_data.jsonl` (configurable via
`EVAL_DATA_PATH`). The parent directory is created with `os.MkdirAll` if absent. File
I/O errors are logged but never interrupt the HTTP response flow. `make run-dev` sets
`DEV_MODE=true` for local development.

## Consequences

**Positive:**
- Generating large evaluation datasets requires only normal usage of the app locally.
- JSONL format is trivially consumable by bulk processing pipelines and LLM fine-tuning
  workflows.
- The `DEV_MODE` flag guarantees zero performance or disk-IO impact in production.
- Each field in `EvalRecord` is independently useful: model names enable A/B
  comparisons; the system prompt snapshot captures the exact prompt version used.

**Negative:**
- Requires handling concurrent file append safely (`O_APPEND` is atomic at the OS
  level for small writes, but care is needed for large records).
- Test suite must explicitly set `DevMode: false` (the zero value) to avoid writing
  files during unit test runs.

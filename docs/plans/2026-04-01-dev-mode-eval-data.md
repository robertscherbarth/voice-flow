# Implementation Plan: Developer Mode Evaluation Data

## 1. Configuration Updates (`packages/server/internal/config/config.go`)
- Add `DevMode bool` and `EvalDataPath string` to the `Config` struct.
- Read the `DEV_MODE` environment variable (parse boolean).
- Read the `EVAL_DATA_PATH` environment variable, defaulting to `test-data/evaluation_data.jsonl`.

## 2. Data Structure (`packages/server/internal/agent/handler.go`)
- Define the `EvalRecord` struct to serialize the data:
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

## 3. Handler Logic (`packages/server/internal/agent/handler.go`)
- In `ServeHTTP`, after the LLM successfully returns the `improvedText`:
  - Check `if h.cfg.DevMode`.
  - If true, spawn a goroutine or handle synchronously (synchronous is fine for local dev) to save the record.
  - Ensure the parent directory (`test-data/`) exists using `os.MkdirAll`.
  - Open the file at `h.cfg.EvalDataPath` using `os.OpenFile` with flags `os.O_APPEND|os.O_CREATE|os.O_WRONLY`.
  - Marshal the `EvalRecord` to JSON and append it with a newline character `\n`.
  - Log any I/O errors, ensuring they do not crash the HTTP response flow.

## 4. Test Updates (`packages/server/internal/agent/handler_test.go`)
- Update all instances where `NewHandler` or `&Handler{}` is initialized in the test suite to include the updated `Config` struct.
- Ensure `DevMode` defaults to `false` during standard testing so tests do not write arbitrary files.

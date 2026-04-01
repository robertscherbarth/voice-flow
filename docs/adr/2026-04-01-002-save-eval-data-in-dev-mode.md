# ADR: Save Evaluation Data in Development Mode

## Context
When running the voice-agent server locally, there is currently no way to systematically review the output quality of the AI models. To improve the system prompt and evaluate the performance of different models (STT and LLM), we need a way to capture the raw transcripts alongside the improved text for every request. 

## Decision
We will introduce a new `DEV_MODE` configuration flag. When the server is running with this flag enabled, the `/process` HTTP handler will asynchronously save the raw transcript, the improved text, the models used, and the system prompt to a local file.

We have chosen to append these records to a single file in **JSON Lines (JSONL)** format at `test-data/evaluation_data.jsonl`. 

## Consequences
- **Positive:** We can easily generate large datasets of test data by simply using the application locally.
- **Positive:** JSONL format is ideal for bulk processing, automated evaluations, and LLM fine-tuning pipelines.
- **Positive:** Wrapping this logic behind a `DEV_MODE` flag ensures zero performance or disk-io impact in production environments.
- **Negative:** Requires handling file I/O operations and ensuring concurrent requests append safely to the same file.

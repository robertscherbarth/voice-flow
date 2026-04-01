package config

import (
	"log"
	"os"

	"gopkg.in/yaml.v3"
)

// PromptConfig holds the YAML prompt configuration.
type PromptConfig struct {
	SystemPrompt string `yaml:"system_prompt"`
}

// Config holds the application configuration.
type Config struct {
	Port         string
	MistralURL   string
	MistralKey   string
	STTModel     string
	LLMModel     string
	SystemPrompt string
	DevMode      bool
	EvalDataPath string
}

// New returns a new Config with default values and environment overrides.
func New() *Config {
	mistralKey := os.Getenv("MISTRAL_API_KEY")

	sttModel := os.Getenv("MISTRAL_STT_MODEL")
	if sttModel == "" {
		sttModel = "voxtral-mini-latest"
	}

	llmModel := os.Getenv("MISTRAL_LLM_MODEL")
	if llmModel == "" {
		llmModel = "mistral-small-latest"
	}

	devMode := os.Getenv("DEV_MODE") == "true"
	evalDataPath := os.Getenv("EVAL_DATA_PATH")
	if evalDataPath == "" {
		evalDataPath = "test-data/evaluation_data.jsonl"
	}

	systemPrompt := loadSystemPrompt()

	return &Config{
		Port:         "8080",
		MistralURL:   "https://api.mistral.ai",
		MistralKey:   mistralKey,
		STTModel:     sttModel,
		LLMModel:     llmModel,
		SystemPrompt: systemPrompt,
		DevMode:      devMode,
		EvalDataPath: evalDataPath,
	}
}

// loadSystemPrompt reads the system prompt from prompt/optimize.yaml if it exists.
func loadSystemPrompt() string {
	defaultPrompt := "You are a precise audio transcription editor. Your ONLY job is to remove filler words (like 'um', 'uh', 'like') and fix obvious grammatical errors from the provided spoken transcript.\n\nCRITICAL RULES:\n1. DO NOT change the perspective or pronouns (e.g., if the user says 'you', keep it as 'you'; do NOT change it to 'I').\n2. DO NOT rewrite the sentence to sound better if it changes the original meaning or tone.\n3. If the text is already clear, return it exactly as provided.\n4. Return ONLY the final text. Do not add quotes, explanations, or conversational filler."

	// Find the project root relative to the current working directory
	// In production, the working directory is usually where the binary is run
	// In tests, the working directory is the package directory
	path := "prompt/optimize.yaml"

	// If we're running tests from subdirectories, the prompt directory is in the root of the server module
	if _, err := os.Stat("../../prompt/optimize.yaml"); err == nil {
		path = "../../prompt/optimize.yaml"
	} else if _, err := os.Stat("../prompt/optimize.yaml"); err == nil {
		path = "../prompt/optimize.yaml"
	} else if _, err := os.Stat("../../../prompt/optimize.yaml"); err == nil {
		path = "../../../prompt/optimize.yaml"
	}

	data, err := os.ReadFile(path)
	if err != nil {
		log.Printf("Warning: failed to read %s, using default prompt: %v", path, err)
		return defaultPrompt
	}

	var promptCfg PromptConfig
	if err := yaml.Unmarshal(data, &promptCfg); err != nil {
		log.Printf("Warning: failed to parse %s, using default prompt: %v", path, err)
		return defaultPrompt
	}

	if promptCfg.SystemPrompt == "" {
		log.Printf("Warning: empty system_prompt in yaml, using default prompt")
		return defaultPrompt
	}

	log.Printf("Successfully loaded system prompt from %s", path)

	return promptCfg.SystemPrompt
}

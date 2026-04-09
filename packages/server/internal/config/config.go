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

// appProviderConfig holds per-provider settings from config.yaml.
type appProviderConfig struct {
	APIURL   string `yaml:"api_url"`
	STTModel string `yaml:"stt_model"`
	LLMModel string `yaml:"llm_model"`
}

// appConfig is the top-level structure of config.yaml.
type appConfig struct {
	Provider string            `yaml:"provider"`
	Mistral  appProviderConfig `yaml:"mistral"`
	Gemini   appProviderConfig `yaml:"gemini"`
}

// Config holds the application configuration.
type Config struct {
	Port           string
	Provider       string // "mistral" or "gemini"
	MistralURL     string
	MistralKey     string
	STTModel       string
	LLMModel       string
	GeminiURL      string
	GeminiKey      string
	GeminiSTTModel string
	GeminiLLMModel string
	SystemPrompt   string
	DevMode        bool
	EvalDataPath   string
}

// New returns a new Config loaded from config.yaml with environment variable overrides.
func New() *Config {
	app := loadAppConfig()

	// Provider: YAML default, then env override
	provider := app.Provider
	if p := os.Getenv("PROVIDER"); p != "" {
		provider = p
	}
	if provider == "" {
		provider = "mistral"
	}

	// Mistral settings: YAML → env override
	mistralURL := app.Mistral.APIURL
	if mistralURL == "" {
		mistralURL = "https://api.mistral.ai"
	}
	mistralKey := os.Getenv("MISTRAL_API_KEY")

	sttModel := app.Mistral.STTModel
	if v := os.Getenv("MISTRAL_STT_MODEL"); v != "" {
		sttModel = v
	}
	if sttModel == "" {
		sttModel = "voxtral-mini-latest"
	}

	llmModel := app.Mistral.LLMModel
	if v := os.Getenv("MISTRAL_LLM_MODEL"); v != "" {
		llmModel = v
	}
	if llmModel == "" {
		llmModel = "mistral-small-latest"
	}

	// Gemini settings: YAML → env override
	geminiURL := app.Gemini.APIURL
	if geminiURL == "" {
		geminiURL = "https://generativelanguage.googleapis.com"
	}
	geminiKey := os.Getenv("GEMINI_API_KEY")

	geminiSTTModel := app.Gemini.STTModel
	if v := os.Getenv("GEMINI_STT_MODEL"); v != "" {
		geminiSTTModel = v
	}
	if geminiSTTModel == "" {
		geminiSTTModel = "gemini-3-flash-preview"
	}

	geminiLLMModel := app.Gemini.LLMModel
	if v := os.Getenv("GEMINI_LLM_MODEL"); v != "" {
		geminiLLMModel = v
	}
	if geminiLLMModel == "" {
		geminiLLMModel = "gemini-3-flash-preview"
	}

	devMode := os.Getenv("DEV_MODE") == "true"
	evalDataPath := os.Getenv("EVAL_DATA_PATH")
	if evalDataPath == "" {
		evalDataPath = "test-data/evaluation_data.jsonl"
	}

	systemPrompt := loadSystemPrompt()

	// STTModel and LLMModel are the active provider's model defaults used by the handler.
	activeSSTModel := sttModel
	activeLLMModel := llmModel
	if provider == "gemini" {
		activeSSTModel = geminiSTTModel
		activeLLMModel = geminiLLMModel
	}

	return &Config{
		Port:           "8080",
		Provider:       provider,
		MistralURL:     mistralURL,
		MistralKey:     mistralKey,
		STTModel:       activeSSTModel,
		LLMModel:       activeLLMModel,
		GeminiURL:      geminiURL,
		GeminiKey:      geminiKey,
		GeminiSTTModel: geminiSTTModel,
		GeminiLLMModel: geminiLLMModel,
		SystemPrompt:   systemPrompt,
		DevMode:        devMode,
		EvalDataPath:   evalDataPath,
	}
}

// loadAppConfig reads config.yaml from one of several candidate paths and returns
// the parsed appConfig. Missing or malformed files return an empty config (all fields
// default to zero values; callers fall back to hardcoded defaults).
func loadAppConfig() appConfig {
	candidates := []string{
		"config.yaml",
		"../../config.yaml",
		"../config.yaml",
		"../../../config.yaml",
	}

	for _, p := range candidates {
		data, err := os.ReadFile(p)
		if err != nil {
			continue
		}
		var cfg appConfig
		if err := yaml.Unmarshal(data, &cfg); err != nil {
			log.Printf("Warning: failed to parse %s: %v", p, err)
			return appConfig{}
		}
		log.Printf("Loaded app config from %s", p)
		return cfg
	}

	log.Printf("Warning: config.yaml not found, using built-in defaults")
	return appConfig{}
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

package config

import (
	"fmt"
	"os"
	"time"

	"github.com/joho/godotenv"
)

// Config holds all configuration for the application
type Config struct {
	// AdGuard Home settings
	AdGuardURL      string
	AdGuardUser     string
	AdGuardPassword string

	// Application settings
	PollInterval time.Duration
	DBPath       string
	LogLevel     string

	// LLM settings
	LLMEnabled  bool
	LLMProvider string // gemini, ollama, openai, anthropic
	LLMTimeout  time.Duration

	// Gemini settings
	GeminiAPIKey string
	GeminiModel  string

	// OpenAI settings
	OpenAIAPIKey string
	OpenAIModel  string

	// Anthropic settings
	AnthropicAPIKey string
	AnthropicModel  string

	// Ollama settings
	OllamaURL   string
	OllamaModel string
}

// Load reads configuration from environment variables
func Load() (*Config, error) {
	// Load .env file if it exists (ignore error if file doesn't exist)
	_ = godotenv.Load()

	cfg := &Config{
		AdGuardURL:      getEnv("AGH_URL", ""),
		AdGuardUser:     getEnv("AGH_USER", ""),
		AdGuardPassword: getEnv("AGH_PASS", ""),
		DBPath:          getEnv("DB_PATH", "./data/guardian.db"),
		LogLevel:        getEnv("LOG_LEVEL", "info"),

		// LLM settings
		LLMEnabled:  getBoolEnv("LLM_ENABLE", false),
		LLMProvider: getEnv("LLM_PROVIDER", "gemini"),

		// Gemini settings
		GeminiAPIKey: getEnv("GEMINI_API_KEY", ""),
		GeminiModel:  getEnv("GEMINI_MODEL", "gemini-1.5-flash"),

		// OpenAI settings
		OpenAIAPIKey: getEnv("OPENAI_API_KEY", ""),
		OpenAIModel:  getEnv("OPENAI_MODEL", "gpt-4o-mini"),

		// Anthropic settings
		AnthropicAPIKey: getEnv("ANTHROPIC_API_KEY", ""),
		AnthropicModel:  getEnv("ANTHROPIC_MODEL", "claude-3-5-sonnet-20241022"),

		// Ollama settings
		OllamaURL:   getEnv("OLLAMA_URL", "http://localhost:11434"),
		OllamaModel: getEnv("OLLAMA_MODEL", "llama3"),
	}

	// Parse poll interval
	pollIntervalStr := getEnv("POLL_INTERVAL", "10s")
	pollInterval, err := time.ParseDuration(pollIntervalStr)
	if err != nil {
		return nil, fmt.Errorf("invalid POLL_INTERVAL: %w", err)
	}
	cfg.PollInterval = pollInterval

	// Parse LLM timeout
	llmTimeoutStr := getEnv("LLM_TIMEOUT", "30s")
	llmTimeout, err := time.ParseDuration(llmTimeoutStr)
	if err != nil {
		return nil, fmt.Errorf("invalid LLM_TIMEOUT: %w", err)
	}
	cfg.LLMTimeout = llmTimeout

	// Validate required fields
	if err := cfg.Validate(); err != nil {
		return nil, err
	}

	return cfg, nil
}

// Validate checks that all required configuration is present
func (c *Config) Validate() error {
	if c.AdGuardURL == "" {
		return fmt.Errorf("AGH_URL is required")
	}
	if c.AdGuardUser == "" {
		return fmt.Errorf("AGH_USER is required")
	}
	if c.AdGuardPassword == "" {
		return fmt.Errorf("AGH_PASS is required")
	}
	if c.PollInterval < time.Second {
		return fmt.Errorf("POLL_INTERVAL must be at least 1 second")
	}

	// Validate LLM configuration if enabled
	if c.LLMEnabled {
		switch c.LLMProvider {
		case "gemini":
			if c.GeminiAPIKey == "" {
				return fmt.Errorf("GEMINI_API_KEY is required when LLM_PROVIDER=gemini")
			}
		case "openai":
			if c.OpenAIAPIKey == "" {
				return fmt.Errorf("OPENAI_API_KEY is required when LLM_PROVIDER=openai")
			}
		case "anthropic":
			if c.AnthropicAPIKey == "" {
				return fmt.Errorf("ANTHROPIC_API_KEY is required when LLM_PROVIDER=anthropic")
			}
		case "ollama":
			if c.OllamaURL == "" {
				return fmt.Errorf("OLLAMA_URL is required when LLM_PROVIDER=ollama")
			}
		default:
			return fmt.Errorf("invalid LLM_PROVIDER: %s (must be gemini, openai, anthropic, or ollama)", c.LLMProvider)
		}
	}

	return nil
}

// getEnv retrieves an environment variable or returns a default value
func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// getBoolEnv retrieves a boolean environment variable or returns a default value
func getBoolEnv(key string, defaultValue bool) bool {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value == "true" || value == "1" || value == "yes"
}

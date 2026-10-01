package nimble

import (
	"os"
	"strings"
	"time"
)

// Defaults used when configuration is not supplied.
const (
	// DefaultBaseURL is the conventional local Ollama endpoint. It is only a
	// default: callers must be able to point at a remote backend via
	// OLLAMA_BASE_URL, so no code should assume the server is local.
	DefaultBaseURL = "http://localhost:11434"
	// DefaultModel is the Nimble decision model name.
	DefaultModel = "nimble"
	// DefaultTimeout bounds a single Nimble/Ollama call.
	DefaultTimeout = 30 * time.Second
)

// Config holds the Nimble provider's connection settings.
type Config struct {
	// BaseURL is the Ollama-compatible backend root, for example
	// "http://ollama.internal:11434".
	BaseURL string
	// Model is the model served by the backend.
	Model string
	// Timeout bounds a single request.
	Timeout time.Duration
}

// WithDefaults returns a copy of cfg with zero fields replaced by defaults.
func (c Config) WithDefaults() Config {
	if strings.TrimSpace(c.BaseURL) == "" {
		c.BaseURL = DefaultBaseURL
	}
	if strings.TrimSpace(c.Model) == "" {
		c.Model = DefaultModel
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	return c
}

// ConfigFromEnv reads Nimble configuration from the environment:
//
//	OLLAMA_BASE_URL  backend root URL (default http://localhost:11434)
//	NIMBLE_MODEL     model name          (default "nimble")
//	NIMBLE_TIMEOUT   request timeout     (default 30s, Go duration syntax)
func ConfigFromEnv() Config {
	return Config{
		BaseURL: envOr("OLLAMA_BASE_URL", DefaultBaseURL),
		Model:   envOr("NIMBLE_MODEL", DefaultModel),
		Timeout: envDurationOr("NIMBLE_TIMEOUT", DefaultTimeout),
	}.WithDefaults()
}

func envOr(key, fallback string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return fallback
}

func envDurationOr(key string, fallback time.Duration) time.Duration {
	v := strings.TrimSpace(os.Getenv(key))
	if v == "" {
		return fallback
	}
	d, err := time.ParseDuration(v)
	if err != nil || d <= 0 {
		return fallback
	}
	return d
}

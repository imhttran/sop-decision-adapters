package julia

import (
	"os"
	"strings"
	"time"
)

// Defaults used when configuration is not supplied.
const (
	// DefaultModel is the Julia decision model identifier.
	DefaultModel = "julia"
	// DefaultTimeout bounds a single inference call.
	DefaultTimeout = 30 * time.Second
)

// Config holds the Julia provider's settings.
type Config struct {
	// Model is the model identifier reported in DecisionResult.Model.
	Model string
	// ModelPath is the path to the ONNX model. It is forwarded to the inference
	// command as JULIA_MODEL_PATH.
	ModelPath string
	// InferenceCmd is the external inference command argv. An empty command
	// means the runner is unavailable.
	InferenceCmd []string
	// Timeout bounds a single inference call.
	Timeout time.Duration
}

// WithDefaults returns a copy of cfg with zero fields replaced by defaults.
func (c Config) WithDefaults() Config {
	if strings.TrimSpace(c.Model) == "" {
		c.Model = DefaultModel
	}
	if c.Timeout <= 0 {
		c.Timeout = DefaultTimeout
	}
	return c
}

// ConfigFromEnv reads Julia configuration from the environment:
//
//	JULIA_MODEL          model identifier                     (default "julia")
//	JULIA_MODEL_PATH     path to the ONNX model (forwarded as JULIA_MODEL_PATH)
//	JULIA_INFERENCE_CMD  external inference command, whitespace-split into argv
//	                     (empty => runner is unavailable)
//	JULIA_TIMEOUT        inference timeout                   (default 30s, Go duration syntax)
func ConfigFromEnv() Config {
	return Config{
		Model:        envOr("JULIA_MODEL", DefaultModel),
		ModelPath:    strings.TrimSpace(os.Getenv("JULIA_MODEL_PATH")),
		InferenceCmd: splitCommand(os.Getenv("JULIA_INFERENCE_CMD")),
		Timeout:      envDurationOr("JULIA_TIMEOUT", DefaultTimeout),
	}.WithDefaults()
}

// splitCommand whitespace-splits a command string into argv. No shell quoting
// rules are applied: an argument containing spaces is unsupported.
func splitCommand(raw string) []string {
	fields := strings.Fields(raw)
	if len(fields) == 0 {
		return nil
	}
	return fields
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

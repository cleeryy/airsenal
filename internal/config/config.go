package config

import (
	"os"
	"strconv"
)

// Config holds runtime configuration sourced from environment variables.
type Config struct {
	Port      string
	CheatsDir string
	EnableMCP bool
}

// Load returns configuration populated from environment variables with sensible defaults.
func Load() *Config {
	return &Config{
		Port:      envOr("AIRSENAL_PORT", "8080"),
		CheatsDir: envOr("AIRSENAL_CHEATS_DIR", "./cheats"),
		EnableMCP: envBool("AIRSENAL_ENABLE_MCP", false),
	}
}

func envOr(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func envBool(key string, fallback bool) bool {
	v := os.Getenv(key)
	if v == "" {
		return fallback
	}
	b, err := strconv.ParseBool(v)
	if err != nil {
		return fallback
	}
	return b
}

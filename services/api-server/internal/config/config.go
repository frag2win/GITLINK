// Package config handles loading and validation of application
// configuration from environment variables.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all configuration values for the api-server.
type Config struct {
	// Port is the HTTP port the server listens on (default "3000").
	Port string

	// DBUrl is the postgres connection string.
	DBUrl string

	// ReposPath is the root directory where bare Git repositories are stored.
	ReposPath string

	// GitSocketPath is the Unix domain socket path for the git-server sidecar.
	GitSocketPath string

	// P2PSocketPath is the Unix domain socket path for the libp2p-node sidecar.
	P2PSocketPath string

	// LogLevel controls log verbosity: "debug", "info", "warn", "error".
	LogLevel string
}

// Load reads configuration from environment variables and applies defaults.
func Load() (*Config, error) {
	cfg := &Config{
		Port:          envOrDefault("API_PORT", "3000"),
		DBUrl:         envOrDefault("API_DB_URL", ""),
		ReposPath:     envOrDefault("API_REPOS_PATH", "./repos"),
		GitSocketPath: envOrDefault("API_GIT_SOCKET", "/tmp/git-server.sock"),
		P2PSocketPath: envOrDefault("API_P2P_SOCKET", "/tmp/libp2p-node.sock"),
		LogLevel:      envOrDefault("API_LOG_LEVEL", "info"),
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation: %w", err)
	}

	return cfg, nil
}

// validate checks that mandatory config values are present and sane.
func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("port must not be empty")
	}
	if c.DBUrl == "" {
		return fmt.Errorf("database URL must not be empty")
	}
	if !isValidLogLevel(c.LogLevel) {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	return nil
}

// isValidLogLevel returns true when level is one of the accepted values.
func isValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

// envOrDefault reads an environment variable; if unset or empty it returns
// the provided fallback.
func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

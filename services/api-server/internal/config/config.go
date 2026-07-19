package config

import (
	"fmt"
	"os"
	"strings"
)

type Config struct {
	Port           string
	DBUrl          string
	ReposPath      string
	SSHPort        string
	GitIPCNetwork  string
	GitIPCAddress  string
	P2PSocketPath  string
	LogLevel       string
	JWTSecret      []byte
	AuthSocketPath string
	DevMode        bool
}

func Load() (*Config, error) {
	jwtSecretStr := os.Getenv("JWT_SECRET")

	cfg := &Config{
		Port:           envOrDefault("API_PORT", "3000"),
		DBUrl:          os.Getenv("API_DB_URL"),
		ReposPath:      envOrDefault("API_REPOS_PATH", "./repos"),
		SSHPort:        envOrDefault("API_SSH_PORT", "2222"),
		GitIPCNetwork:  envOrDefault("API_GIT_IPC_NETWORK", "unix"),
		GitIPCAddress:  envOrDefault("API_GIT_IPC_ADDRESS", "/tmp/git-server.sock"),
		P2PSocketPath:  envOrDefault("API_P2P_SOCKET", "/tmp/libp2p-node.sock"),
		LogLevel:       envOrDefault("API_LOG_LEVEL", "info"),
		JWTSecret:      []byte(jwtSecretStr),
		AuthSocketPath: envOrDefault("API_AUTH_SOCKET_PATH", "/run/git/auth.sock"),
		DevMode:        os.Getenv("API_DEV_MODE") == "true",
	}

	if err := cfg.validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	return cfg, nil
}

func (c *Config) validate() error {
	if c.Port == "" {
		return fmt.Errorf("port must not be empty")
	}
	if c.DBUrl == "" {
		return fmt.Errorf("API_DB_URL must be provided")
	}
	if len(c.JWTSecret) == 0 {
		return fmt.Errorf("JWT_SECRET must be provided and not empty")
	}
	if c.GitIPCNetwork == "" {
		return fmt.Errorf("API_GIT_IPC_NETWORK must be provided")
	}
	if c.GitIPCAddress == "" {
		return fmt.Errorf("API_GIT_IPC_ADDRESS must be provided")
	}
	if c.ReposPath == "" {
		return fmt.Errorf("API_REPOS_PATH must be provided")
	}
	if c.AuthSocketPath == "" {
		return fmt.Errorf("API_AUTH_SOCKET_PATH must be provided")
	}
	if !isValidLogLevel(c.LogLevel) {
		return fmt.Errorf("invalid log level: %s", c.LogLevel)
	}
	return nil
}

func isValidLogLevel(level string) bool {
	switch strings.ToLower(level) {
	case "debug", "info", "warn", "error":
		return true
	default:
		return false
	}
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

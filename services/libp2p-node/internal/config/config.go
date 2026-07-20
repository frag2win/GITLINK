// Package config handles configuration for the libp2p-node.
package config

import (
	"fmt"
	"os"
	"strings"
)

// Config holds all settings for the libp2p-node service.
type Config struct {
	// IdentityKeyPath is the file path to the Ed25519 private key.
	// If the file does not exist, a new key pair is generated and saved.
	IdentityKeyPath string

	// ListenAddrs is a list of multiaddrs the libp2p host should listen on.
	// Example: ["/ip4/0.0.0.0/tcp/4001", "/ip6/::/tcp/4001"]
	ListenAddrs []string

	// BootstrapPeers is a list of multiaddrs for bootstrap nodes used
	// by the Kademlia DHT (Phase 2).
	BootstrapPeers []string

	// SocketPath is the Unix domain socket path used to communicate
	// with the api-server sidecar.
	SocketPath string

	ProxySocket string
	QueueDir    string
}

// Load reads configuration from environment variables.
func Load() (*Config, error) {
	cfg := &Config{
		IdentityKeyPath: envOrDefault("P2P_IDENTITY_KEY", "./identity/peer.key"),
		ListenAddrs:     splitCSV(envOrDefault("P2P_LISTEN_ADDRS", "/ip4/0.0.0.0/tcp/4001")),
		BootstrapPeers:  splitCSV(os.Getenv("P2P_BOOTSTRAP_PEERS")),
		SocketPath:      envOrDefault("P2P_API_SOCKET", "/tmp/libp2p-node.sock"),
		ProxySocket:     envOrDefault("PROXY_SOCKET", "/tmp/git-proxy.sock"),
		QueueDir:        envOrDefault("P2P_QUEUE_DIR", "/app/queue"),
	}

	if err := cfg.validate(); err != nil {
		return nil, err
	}
	return cfg, nil
}

func (c *Config) validate() error {
	if c.IdentityKeyPath == "" {
		return fmt.Errorf("identity key path must not be empty")
	}
	if len(c.ListenAddrs) == 0 {
		return fmt.Errorf("at least one listen address is required")
	}
	return nil
}

func envOrDefault(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if t := strings.TrimSpace(p); t != "" {
			out = append(out, t)
		}
	}
	return out
}

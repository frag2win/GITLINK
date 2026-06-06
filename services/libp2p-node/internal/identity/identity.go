// Package identity handles loading and generating Ed25519 peer identity keys.
package identity

import (
	"crypto/rand"
	"fmt"
	"os"
	"path/filepath"

	"github.com/libp2p/go-libp2p/core/crypto"
	"github.com/libp2p/go-libp2p/core/peer"
)

// LoadOrGenerate loads an Ed25519 private key from the given file path.
// If the file does not exist, a new key pair is generated, saved to disk,
// and returned. The key is stored in libp2p's protobuf-serialised format.
func LoadOrGenerate(keyPath string) (crypto.PrivKey, error) {
	// Attempt to load existing key
	data, err := os.ReadFile(keyPath)
	if err == nil {
		return unmarshalKey(data)
	}

	if !os.IsNotExist(err) {
		return nil, fmt.Errorf("read identity key: %w", err)
	}

	// Generate a new Ed25519 key pair
	privKey, _, err := crypto.GenerateEd25519Key(rand.Reader)
	if err != nil {
		return nil, fmt.Errorf("generate Ed25519 key: %w", err)
	}

	// Persist the key for future runs
	if err := saveKey(keyPath, privKey); err != nil {
		return nil, err
	}

	return privKey, nil
}

// unmarshalKey deserialises a protobuf-encoded private key.
func unmarshalKey(data []byte) (crypto.PrivKey, error) {
	key, err := crypto.UnmarshalPrivateKey(data)
	if err != nil {
		return nil, fmt.Errorf("unmarshal identity key: %w", err)
	}
	return key, nil
}

// saveKey serialises the private key and writes it to disk with
// restrictive permissions (owner-read-only).
func saveKey(keyPath string, key crypto.PrivKey) error {
	dir := filepath.Dir(keyPath)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return fmt.Errorf("create identity directory: %w", err)
	}

	data, err := crypto.MarshalPrivateKey(key)
	if err != nil {
		return fmt.Errorf("marshal identity key: %w", err)
	}

	if err := os.WriteFile(keyPath, data, 0o400); err != nil {
		return fmt.Errorf("write identity key: %w", err)
	}

	return nil
}

// PeerIDFromKey extracts the libp2p peer ID from a private key.
func PeerIDFromKey(key crypto.PrivKey) (string, error) {
	id, err := peer.IDFromPrivateKey(key)
	if err != nil {
		return "", fmt.Errorf("extract peer ID: %w", err)
	}
	return id.String(), nil
}

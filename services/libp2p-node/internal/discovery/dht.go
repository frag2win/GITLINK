package discovery

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
)

// ===========================================================
// Phase 2 — Kademlia DHT for internet-wide peer discovery.
//
// This module is a stub. In Phase 2 it will:
//   - Join the Kademlia DHT using bootstrap peers
//   - Provide (announce) repository CIDs to the DHT
//   - Discover peers hosting specific repositories
//   - Maintain a routing table for efficient lookups
// ===========================================================

// DHT wraps the Kademlia DHT for content-addressed peer discovery.
type DHT struct {
	host host.Host
	// dht  *dht.IpfsDHT  // uncomment in Phase 2
}

// NewDHT creates and bootstraps a Kademlia DHT instance.
// Phase 2: will accept bootstrap peer multiaddrs from config.
func NewDHT(ctx context.Context, h host.Host, bootstrapPeers []string) (*DHT, error) {
	// TODO [Phase 2]: Create a new Kademlia DHT in server mode.
	// TODO [Phase 2]: Connect to bootstrap peers.
	// TODO [Phase 2]: Bootstrap the routing table.

	return nil, fmt.Errorf("DHT not implemented — Phase 2 feature")
}

// Provide announces that this node can serve the given repository.
func (d *DHT) Provide(ctx context.Context, repoName string) error {
	// TODO [Phase 2]: Convert repoName to a CID.
	// TODO [Phase 2]: Call dht.Provide() to announce availability.

	return fmt.Errorf("DHT.Provide not implemented — Phase 2 feature")
}

// FindProviders looks up peers that can serve the given repository.
func (d *DHT) FindProviders(ctx context.Context, repoName string) ([]string, error) {
	// TODO [Phase 2]: Convert repoName to a CID.
	// TODO [Phase 2]: Call dht.FindProviders() to discover peers.

	return nil, fmt.Errorf("DHT.FindProviders not implemented — Phase 2 feature")
}

// Close shuts down the DHT.
func (d *DHT) Close() error {
	// TODO [Phase 2]: Close the DHT instance.
	return nil
}

// Package relay provides Circuit Relay v2 support for the libp2p-node.
package relay

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ===========================================================
// Phase 2 — Circuit Relay v2 fallback.
//
// When direct connections and hole punching both fail, traffic
// is routed through a public relay node. Circuit Relay v2 is
// time- and bandwidth-limited to prevent abuse.
//
// This module will:
//   - Discover and connect to public relay nodes
//   - Reserve a relay slot for this peer
//   - Use relayed connections as a signalling channel for DCUtR
//   - Fall back to relayed transport for Git pack data if needed
// ===========================================================

// Relay manages Circuit Relay v2 connections.
type Relay struct {
	host       host.Host
	relayAddrs []string
}

// NewRelay creates a Circuit Relay v2 client.
func NewRelay(ctx context.Context, h host.Host, relayAddrs []string) (*Relay, error) {
	// TODO [Phase 2]: Enable relay client via libp2p.EnableAutoRelayWithStaticRelays().
	// TODO [Phase 2]: Connect to known relay nodes.
	// TODO [Phase 2]: Reserve relay slots.

	return nil, fmt.Errorf("Relay not implemented — Phase 2 feature")
}

// ConnectViaRelay establishes a relayed connection to the target peer.
func (r *Relay) ConnectViaRelay(ctx context.Context, target peer.ID) error {
	// TODO [Phase 2]: Build a circuit relay multiaddr for the target.
	// TODO [Phase 2]: Dial the target through the relay.

	return fmt.Errorf("Relay.ConnectViaRelay not implemented — Phase 2 feature")
}

// Close releases relay reservations.
func (r *Relay) Close() error {
	// TODO [Phase 2]: Clean up relay connections.
	return nil
}

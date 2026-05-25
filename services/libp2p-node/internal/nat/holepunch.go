package nat

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
)

// ===========================================================
// Phase 2 — DCUtR (Direct Connection Upgrade through Relay)
//         hole punching coordinator.
//
// When two peers are both behind NATs, DCUtR coordinates a
// simultaneous-open to establish a direct connection, using a
// relay as the signalling channel.
//
// Prerequisites:
//   - AutoNAT must detect both peers as NATPrivate
//   - At least one Circuit Relay v2 node must be reachable
// ===========================================================

// HolePuncher coordinates DCUtR hole punching attempts.
type HolePuncher struct {
	host host.Host
}

// NewHolePuncher creates a DCUtR hole punching coordinator.
func NewHolePuncher(ctx context.Context, h host.Host) (*HolePuncher, error) {
	// TODO [Phase 2]: Enable hole punching via libp2p.EnableHolePunching() option.
	// TODO [Phase 2]: This is typically configured at host creation time;
	//                 this constructor may just wrap the existing host capability.

	return nil, fmt.Errorf("HolePuncher not implemented — Phase 2 feature")
}

// Punch attempts a direct connection to target via hole punching.
// It assumes a relayed connection already exists.
func (hp *HolePuncher) Punch(ctx context.Context, target peer.ID) error {
	// TODO [Phase 2]: Trigger a DCUtR upgrade for the relayed connection to target.

	return fmt.Errorf("HolePuncher.Punch not implemented — Phase 2 feature")
}

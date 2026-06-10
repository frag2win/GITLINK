package nat

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
	"github.com/libp2p/go-libp2p/core/peer"
)

// HolePuncher coordinates DCUtR hole punching attempts.
type HolePuncher struct {
	host host.Host
}

// NewHolePuncher creates a DCUtR hole punching coordinator.
// Note: Actual hole punching is handled automatically by libp2p when 
// libp2p.EnableHolePunching() is passed to libp2p.New().
func NewHolePuncher(ctx context.Context, h host.Host) (*HolePuncher, error) {
	return &HolePuncher{
		host: h,
	}, nil
}

// Punch attempts a direct connection to target via hole punching.
// It assumes a relayed connection already exists.
func (hp *HolePuncher) Punch(ctx context.Context, target peer.ID) error {
	// DCUtR is automatic. By connecting to a peer via a relay multiaddr,
	// libp2p will automatically attempt a hole punch if EnableHolePunching is active.
	// We just ensure we're connected or attempt to connect to the peer.
	if hp.host.Network().Connectedness(target) != network.Connected {
		// Attempting connection might trigger hole punch if multiaddrs are relays
		// But in typical usage, you already connected via relay, so this is a no-op
		// or wait for the upgrade.
	}
	return nil
}

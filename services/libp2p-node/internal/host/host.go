// Package host creates and configures the go-libp2p host instance.
package host

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p"
	"github.com/libp2p/go-libp2p/core/crypto"
	libp2phost "github.com/libp2p/go-libp2p/core/host"
	"github.com/multiformats/go-multiaddr"
)

// New creates a libp2p host configured with:
//   - TCP transport
//   - Noise security protocol
//   - Yamux stream multiplexer
//   - The provided Ed25519 identity key
//   - The provided listen multiaddrs
func New(ctx context.Context, privKey crypto.PrivKey, listenAddrs []string) (libp2phost.Host, error) {
	// Parse listen multiaddrs
	addrs := make([]multiaddr.Multiaddr, 0, len(listenAddrs))
	for _, s := range listenAddrs {
		ma, err := multiaddr.NewMultiaddr(s)
		if err != nil {
			return nil, fmt.Errorf("invalid listen address %q: %w", s, err)
		}
		addrs = append(addrs, ma)
	}

	// Build the host with sensible defaults.
	// go-libp2p v0.33+ enables TCP+Noise+Yamux by default when you
	// supply Identity and ListenAddrs.
	h, err := libp2p.New(
		libp2p.Identity(privKey),
		libp2p.ListenAddrs(addrs...),

		// Explicitly request the transports and security we want:
		// libp2p.Transport(tcp.NewTCPTransport),
		// libp2p.Security(noise.ID, noise.New),
		// libp2p.Muxer(yamux.ID, yamux.DefaultTransport),

		// TODO: Add NATPortMap for UPnP/NAT-PMP (Phase 2).
		// TODO: Add EnableAutoRelay (Phase 2).
		// TODO: Add EnableHolePunching (Phase 2).
	)
	if err != nil {
		return nil, fmt.Errorf("create libp2p host: %w", err)
	}

	return h, nil
}

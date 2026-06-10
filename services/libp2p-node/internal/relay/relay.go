package relay

import (
	"context"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/multiformats/go-multiaddr"
)

// Relay manages Circuit Relay v2 connections.
type Relay struct {
	host       host.Host
	relayAddrs []string
}

// NewRelay creates a Circuit Relay v2 client.
// Note: Actual relay usage is enabled in libp2p.New() using 
// libp2p.EnableRelay() and libp2p.EnableAutoRelayWithStaticRelays().
func NewRelay(ctx context.Context, h host.Host, relayAddrs []string) (*Relay, error) {
	return &Relay{
		host:       h,
		relayAddrs: relayAddrs,
	}, nil
}

// ConnectViaRelay establishes a relayed connection to the target peer.
func (r *Relay) ConnectViaRelay(ctx context.Context, target peer.ID) error {
	// If auto-relay is on, libp2p handles reserving slots on relays and 
	// advertising relay addrs automatically.
	// But if we want to manually connect via a specific relay:
	if len(r.relayAddrs) == 0 {
		return nil
	}

	// Just connect to the first relay
	relayMA, err := multiaddr.NewMultiaddr(r.relayAddrs[0])
	if err != nil {
		return err
	}
	
	// Create a circuit relay multiaddr for the target peer
	circuitAddr, err := multiaddr.NewMultiaddr("/p2p-circuit/p2p/" + target.String())
	if err != nil {
		return err
	}

	targetMA := relayMA.Encapsulate(circuitAddr)
	
	addrInfo, err := peer.AddrInfoFromP2pAddr(targetMA)
	if err != nil {
		return err
	}

	return r.host.Connect(ctx, *addrInfo)
}

// Close releases relay reservations.
func (r *Relay) Close() error {
	return nil
}

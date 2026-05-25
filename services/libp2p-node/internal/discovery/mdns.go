// Package discovery provides peer discovery mechanisms for the libp2p-node.
package discovery

import (
	"context"
	"log"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	"github.com/libp2p/go-libp2p/p2p/discovery/mdns"
)

const (
	// MDNSServiceTag is the mDNS service name used for local discovery.
	MDNSServiceTag = "localrepo.p2p._tcp"
)

// MDNS wraps the go-libp2p mDNS discovery service for finding peers
// on the same local network without any internet connectivity.
type MDNS struct {
	service mdns.Service
	host    host.Host
}

// discoveryNotifee receives notifications when new peers are discovered.
type discoveryNotifee struct {
	h host.Host
}

// HandlePeerFound is called by the mDNS service when a new peer is seen.
func (n *discoveryNotifee) HandlePeerFound(pi peer.AddrInfo) {
	if pi.ID == n.h.ID() {
		return // ignore ourselves
	}

	log.Printf("mDNS: discovered peer %s", pi.ID.String())

	// Attempt to connect to the discovered peer.
	if err := n.h.Connect(context.Background(), pi); err != nil {
		log.Printf("mDNS: failed to connect to %s: %v", pi.ID, err)
	} else {
		log.Printf("mDNS: connected to %s", pi.ID)
	}
}

// NewMDNS starts the mDNS discovery service on the given host.
func NewMDNS(ctx context.Context, h host.Host) (*MDNS, error) {
	notifee := &discoveryNotifee{h: h}

	service := mdns.NewMdnsService(h, MDNSServiceTag, notifee)
	if err := service.Start(); err != nil {
		return nil, err
	}

	return &MDNS{
		service: service,
		host:    h,
	}, nil
}

// Close stops the mDNS discovery service.
func (m *MDNS) Close() error {
	return m.service.Close()
}

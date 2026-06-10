package discovery

import (
	"context"
	"fmt"
	"log"

	"github.com/ipfs/go-cid"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/peer"
	dht "github.com/libp2p/go-libp2p-kad-dht"
	"github.com/multiformats/go-multiaddr"
	"github.com/multiformats/go-multihash"
)

// DHT wraps the Kademlia DHT for content-addressed peer discovery.
type DHT struct {
	host host.Host
	dht  *dht.IpfsDHT
}

// NewDHT creates and bootstraps a Kademlia DHT instance.
func NewDHT(ctx context.Context, h host.Host, bootstrapPeers []string) (*DHT, error) {
	opts := []dht.Option{
		dht.Mode(dht.ModeServer),
	}

	kdht, err := dht.New(ctx, h, opts...)
	if err != nil {
		return nil, fmt.Errorf("failed to create DHT: %w", err)
	}

	if err := kdht.Bootstrap(ctx); err != nil {
		return nil, fmt.Errorf("failed to bootstrap DHT: %w", err)
	}

	var parsedPeers []peer.AddrInfo
	for _, addrStr := range bootstrapPeers {
		ma, err := multiaddr.NewMultiaddr(addrStr)
		if err != nil {
			log.Printf("invalid bootstrap peer multiaddr %s: %v", addrStr, err)
			continue
		}
		pi, err := peer.AddrInfoFromP2pAddr(ma)
		if err != nil {
			log.Printf("invalid bootstrap peer info %s: %v", addrStr, err)
			continue
		}
		parsedPeers = append(parsedPeers, *pi)
	}

	for _, pi := range parsedPeers {
		if err := h.Connect(ctx, pi); err != nil {
			log.Printf("failed to connect to bootstrap peer %s: %v", pi.ID, err)
		} else {
			log.Printf("connected to bootstrap peer %s", pi.ID)
		}
	}

	return &DHT{host: h, dht: kdht}, nil
}

// repoNameToCID converts a repository name to a standard CID v1.
func repoNameToCID(repoName string) (cid.Cid, error) {
	hash, err := multihash.Sum([]byte(repoName), multihash.SHA2_256, -1)
	if err != nil {
		return cid.Undef, err
	}
	return cid.NewCidV1(cid.Raw, hash), nil
}

// Provide announces that this node can serve the given repository.
func (d *DHT) Provide(ctx context.Context, repoName string) error {
	c, err := repoNameToCID(repoName)
	if err != nil {
		return fmt.Errorf("failed to convert repo name to CID: %w", err)
	}
	log.Printf("providing repo %s with CID %s", repoName, c.String())
	return d.dht.Provide(ctx, c, true)
}

// FindProviders looks up peers that can serve the given repository.
func (d *DHT) FindProviders(ctx context.Context, repoName string) ([]string, error) {
	c, err := repoNameToCID(repoName)
	if err != nil {
		return nil, fmt.Errorf("failed to convert repo name to CID: %w", err)
	}

	peers, err := d.dht.FindProviders(ctx, c)
	if err != nil {
		return nil, fmt.Errorf("failed to find providers: %w", err)
	}

	var peerIDs []string
	for _, p := range peers {
		if p.ID == d.host.ID() {
			continue // skip self
		}
		peerIDs = append(peerIDs, p.ID.String())
	}
	return peerIDs, nil
}

// Close shuts down the DHT.
func (d *DHT) Close() error {
	return d.dht.Close()
}

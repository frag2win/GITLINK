// Package main is the entry point for the libp2p-node service.
//
// It loads the peer identity key, creates a go-libp2p host with TCP
// transport, Noise security, and Yamux multiplexing, starts mDNS for
// local peer discovery, and registers stream handlers for the Git
// protocol.
package main

import (
	"context"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/localrepo/libp2p-node/internal/config"
	"github.com/localrepo/libp2p-node/internal/discovery"
	"github.com/localrepo/libp2p-node/internal/host"
	"github.com/localrepo/libp2p-node/internal/identity"
	"github.com/localrepo/libp2p-node/internal/protocol"
	"github.com/localrepo/libp2p-node/internal/socket"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// ---- Load configuration ----
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	// ---- Load or generate peer identity ----
	privKey, err := identity.LoadOrGenerate(cfg.IdentityKeyPath)
	if err != nil {
		log.Fatalf("failed to load identity: %v", err)
	}

	// ---- Create libp2p host ----
	h, err := host.New(ctx, privKey, cfg.ListenAddrs)
	if err != nil {
		log.Fatalf("failed to create libp2p host: %v", err)
	}
	defer h.Close()

	log.Printf("libp2p host started: %s", h.ID())
	for _, addr := range h.Addrs() {
		log.Printf("  listening on: %s/p2p/%s", addr, h.ID())
	}

	// ---- API socket client (to communicate with api-server) ----
	apiClient := socket.NewAPIClient(cfg.SocketPath)

	// ---- Register Git protocol stream handler ----
	protoHandler := protocol.NewHandler(apiClient)
	h.SetStreamHandler(protocol.GitProtocolID, protoHandler.HandleStream)

	// ---- Start mDNS discovery ----
	mdns, err := discovery.NewMDNS(ctx, h)
	if err != nil {
		log.Printf("WARNING: mDNS discovery failed to start: %v", err)
	} else {
		defer mdns.Close()
		log.Println("mDNS discovery active")
	}

	// ---- Phase 2 stubs: DHT, AutoNAT, Hole Punching, Relay ----
	// These will be activated in Phase 2 for internet-wide connectivity.
	// See internal/discovery/dht.go, internal/nat/, internal/relay/.

	// ---- Wait for shutdown signal ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down libp2p-node…")
	cancel()
}

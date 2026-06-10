// Package main is the entry point for the libp2p-node service.
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
	"github.com/localrepo/libp2p-node/internal/nat"
	"github.com/localrepo/libp2p-node/internal/protocol"
	"github.com/localrepo/libp2p-node/internal/proxy"
	"github.com/localrepo/libp2p-node/internal/queue"
	"github.com/localrepo/libp2p-node/internal/relay"
	"github.com/localrepo/libp2p-node/internal/socket"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	queue.Init()

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

	// ---- Register Git protocol stream handler (Handles INCOMING git clones) ----
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

	// ---- Phase 2: Kademlia DHT ----
	// Provide the repository CIDs to the DHT (example bootstrap nodes would go here)
	dhtService, err := discovery.NewDHT(ctx, h, []string{
		// "/ip4/104.131.131.82/tcp/4001/p2p/QmaCpDMGvV2BGHeYERUEnRQAwe3N8SzbUtfsmvsqQLuvuJ", // standard IPFS bootstrapper
	})
	if err != nil {
		log.Printf("WARNING: DHT failed to initialize: %v", err)
	} else {
		defer dhtService.Close()
		log.Println("DHT active")
	}

	// ---- Phase 2: NAT Traversal (AutoNAT, HolePunch, Relay) ----
	autoNATService, _ := nat.NewAutoNAT(ctx, h)
	_ = autoNATService // runs in background
	
	holePuncher, _ := nat.NewHolePuncher(ctx, h)
	_ = holePuncher

	relayService, _ := relay.NewRelay(ctx, h, []string{})
	defer relayService.Close()

	// ---- Start Local Git Proxy (Handles OUTGOING git clones) ----
	proxyServer := proxy.NewServer(h)
	go func() {
		proxyPort := os.Getenv("PROXY_PORT")
		if proxyPort == "" {
			proxyPort = "4000"
		}
		if err := proxyServer.Start("0.0.0.0:" + proxyPort); err != nil {
			log.Printf("local proxy server failed: %v", err)
		}
	}()

	// ---- Wait for shutdown signal ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	log.Println("shutting down libp2p-node...")
	cancel()
}

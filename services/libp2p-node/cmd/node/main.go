// Package main is the entry point for the libp2p-node service.
package main

import (
	"context"
	"log/slog"
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

	// ---- Load configuration ----
	cfg, err := config.Load()
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	queue.Init(cfg.QueueDir)

	// ---- Load or generate peer identity ----
	privKey, err := identity.LoadOrGenerate(cfg.IdentityKeyPath)
	if err != nil {
		slog.Error("failed to load identity", "error", err)
		os.Exit(1)
	}

	// ---- Create libp2p host ----
	h, err := host.New(ctx, privKey, cfg.ListenAddrs)
	if err != nil {
		slog.Error("failed to create libp2p host", "error", err)
		os.Exit(1)
	}
	defer h.Close()

	slog.Info("libp2p host started", "id", h.ID())
	for _, addr := range h.Addrs() {
		slog.Info("listening on", "addr", addr.String()+"/p2p/"+h.ID().String())
	}

	// ---- API socket client (to communicate with api-server) ----
	apiClient := socket.NewAPIClient(cfg.SocketPath)

	// ---- Register Git protocol stream handler (Handles INCOMING git clones) ----
	protoHandler := protocol.NewHandler(apiClient)
	h.SetStreamHandler(protocol.GitProtocolID, protoHandler.HandleStream)

	// ---- Start mDNS discovery ----
	mdns, err := discovery.NewMDNS(ctx, h)
	if err != nil {
		slog.Warn("mDNS discovery failed to start", "error", err)
	} else {
		defer mdns.Close()
		slog.Info("mDNS discovery active")
	}

	// ---- Phase 2: Kademlia DHT ----
	// Provide the repository CIDs to the DHT
	dhtService, err := discovery.NewDHT(ctx, h, cfg.BootstrapPeers)
	if err != nil {
		slog.Warn("DHT failed to initialize", "error", err)
	} else {
		defer dhtService.Close()
		slog.Info("DHT active")
	}

	// ---- Phase 2: NAT Traversal (AutoNAT, HolePunch, Relay) ----
	autoNATService, _ := nat.NewAutoNAT(ctx, h)
	_ = autoNATService // runs in background
	
	holePuncher, _ := nat.NewHolePuncher(ctx, h)
	_ = holePuncher

	relayService, _ := relay.NewRelay(ctx, h, []string{})
	defer relayService.Close()

	// ---- Start Local Git Proxy (Handles OUTGOING git clones) via UDS ----
	proxyServer := proxy.NewServer(h)
	go func() {
		if err := proxyServer.Start(cfg.ProxySocket); err != nil {
			slog.Error("local proxy server failed", "error", err)
		}
	}()

	// ---- Start IPC Socket Server (Handles INCOMING commands from api-server) ----
	socketServer := socket.NewServer(h, cfg.SocketPath)
	go func() {
		if err := socketServer.Start(ctx); err != nil {
			slog.Error("socket server failed", "error", err)
		}
	}()

	// ---- Wait for shutdown signal ----
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	slog.Info("shutting down libp2p-node...")
	cancel()
}

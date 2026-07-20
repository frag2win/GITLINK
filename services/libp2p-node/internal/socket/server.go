package socket

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"net"
	"os"
	"sync"
	"time"

	"github.com/libp2p/go-libp2p/core/host"
	"github.com/localrepo/libp2p-node/internal/queue"
)

type Server struct {
	host       host.Host
	socketPath string
	mu         sync.Mutex
}

func NewServer(h host.Host, socketPath string) *Server {
	return &Server{
		host:       h,
		socketPath: socketPath,
	}
}

type IncomingCommand struct {
	Command       string `json:"command"`
	TaskUUID      string `json:"task_uuid"`
	RepoName      string `json:"repo_name"`
	TargetPeerID  string `json:"target_peer_id"`
	Priority      int    `json:"priority"`
	CorrelationID string `json:"correlation_id"`
}

type PeerInfoResponse struct {
	ID        string   `json:"id"`
	Addrs     []string `json:"addrs"`
	LatencyMs int64    `json:"latency_ms"`
}

func (s *Server) Start(ctx context.Context) error {
	s.mu.Lock()
	_ = os.Remove(s.socketPath)
	listener, err := net.Listen("unix", s.socketPath)
	if err != nil {
		s.mu.Unlock()
		return fmt.Errorf("failed to listen on UDS socket %s: %w", s.socketPath, err)
	}
	_ = os.Chmod(s.socketPath, 0666)
	s.mu.Unlock()

	slog.Info("p2p socket server listening", "path", s.socketPath)

	go func() {
		<-ctx.Done()
		_ = listener.Close()
		_ = os.Remove(s.socketPath)
	}()

	for {
		conn, err := listener.Accept()
		if err != nil {
			select {
			case <-ctx.Done():
				return nil
			default:
				slog.Error("p2p socket server accept error", "error", err)
				time.Sleep(100 * time.Millisecond)
				continue
			}
		}

		go s.handleConn(ctx, conn)
	}
}

func (s *Server) handleConn(ctx context.Context, conn net.Conn) {
	defer conn.Close()
	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	var cmd IncomingCommand
	if err := json.NewDecoder(conn).Decode(&cmd); err != nil {
		slog.Error("p2p socket server decode error", "error", err)
		return
	}

	logger := slog.With("command", cmd.Command, "task_uuid", cmd.TaskUUID, "correlation_id", cmd.CorrelationID)
	logger.Debug("handling incoming UDS command")

	switch cmd.Command {
	case "SYNC_REPO":
		res := queue.ExecuteSync(ctx, s.host, cmd.TaskUUID, cmd.RepoName, cmd.TargetPeerID, cmd.CorrelationID)
		_ = json.NewEncoder(conn).Encode(res)

	case "LIST_PEERS":
		peers := make([]PeerInfoResponse, 0)
		for _, conn := range s.host.Network().Conns() {
			pID := conn.RemotePeer().String()
			addrs := []string{conn.RemoteMultiaddr().String()}
			peers = append(peers, PeerInfoResponse{
				ID:        pID,
				Addrs:     addrs,
				LatencyMs: 15,
			})
		}
		_ = json.NewEncoder(conn).Encode(peers)

	default:
		slog.Warn("unknown command received on p2p socket", "command", cmd.Command)
		_ = json.NewEncoder(conn).Encode(map[string]string{"error": "unknown command"})
	}
}

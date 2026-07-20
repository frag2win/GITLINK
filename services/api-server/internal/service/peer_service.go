package service

import (
	"context"
	"fmt"

	"github.com/localrepo/api-server/internal/ipc"
	"github.com/localrepo/api-server/internal/models"
)

type PeerService interface {
	GetConnectedPeers(ctx context.Context) ([]ipc.PeerInfo, error)
	DispatchSync(ctx context.Context, task *models.SyncTask) (*ipc.P2PSyncResponse, error)
}

type peerService struct {
	p2pClient *ipc.P2PClient
}

func NewPeerService(p2pClient *ipc.P2PClient) PeerService {
	return &peerService{p2pClient: p2pClient}
}

func (s *peerService) GetConnectedPeers(ctx context.Context) ([]ipc.PeerInfo, error) {
	return s.p2pClient.GetPeers(ctx)
}

func (s *peerService) DispatchSync(ctx context.Context, task *models.SyncTask) (*ipc.P2PSyncResponse, error) {
	if task == nil {
		return nil, fmt.Errorf("task cannot be nil")
	}

	req := &ipc.P2PSyncRequest{
		Command:       "SYNC_REPO",
		TaskUUID:      task.TaskUUID,
		RepoName:      task.RepoName,
		TargetPeerID:  task.TargetPeerID,
		Priority:      task.Priority,
		CorrelationID: task.CorrelationID,
	}

	return s.p2pClient.TriggerSync(ctx, req)
}

// Package nat provides NAT traversal utilities for the libp2p-node.
package nat

import (
	"context"
	"fmt"

	"github.com/libp2p/go-libp2p/core/host"
)

// ===========================================================
// Phase 2 — AutoNAT service for NAT type detection.
//
// AutoNAT uses external peers to determine whether this node is
// behind a NAT and what type of NAT it is (full-cone, symmetric,
// etc.). This information is used to decide whether hole punching
// is possible or a relay is needed.
// ===========================================================

// NATStatus represents the detected NAT type.
type NATStatus int

const (
	NATUnknown NATStatus = iota
	NATPublic
	NATPrivate
)

// String returns a human-readable NAT status label.
func (s NATStatus) String() string {
	switch s {
	case NATPublic:
		return "public"
	case NATPrivate:
		return "private (behind NAT)"
	default:
		return "unknown"
	}
}

// AutoNAT wraps the libp2p AutoNAT service.
type AutoNAT struct {
	host   host.Host
	status NATStatus
}

// NewAutoNAT creates and starts the AutoNAT service.
func NewAutoNAT(ctx context.Context, h host.Host) (*AutoNAT, error) {
	// TODO [Phase 2]: Enable AutoNAT via libp2p.EnableNATService() option.
	// TODO [Phase 2]: Subscribe to reachability change events.
	// TODO [Phase 2]: Update a.status when events fire.

	return nil, fmt.Errorf("AutoNAT not implemented — Phase 2 feature")
}

// Status returns the currently detected NAT status.
func (a *AutoNAT) Status() NATStatus {
	if a == nil {
		return NATUnknown
	}
	return a.status
}

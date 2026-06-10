package nat

import (
	"context"
	"fmt"
	"log"

	"github.com/libp2p/go-libp2p/core/event"
	"github.com/libp2p/go-libp2p/core/host"
	"github.com/libp2p/go-libp2p/core/network"
)

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

// AutoNAT wraps the libp2p AutoNAT service reachability subscription.
type AutoNAT struct {
	host   host.Host
	status NATStatus
}

// NewAutoNAT creates and starts listening for NAT reachability events.
// The actual AutoNAT service is enabled in the libp2p host constructor.
func NewAutoNAT(ctx context.Context, h host.Host) (*AutoNAT, error) {
	sub, err := h.EventBus().Subscribe(new(event.EvtLocalReachabilityChanged))
	if err != nil {
		return nil, fmt.Errorf("failed to subscribe to reachability events: %w", err)
	}

	an := &AutoNAT{
		host:   h,
		status: NATUnknown,
	}

	go func() {
		for {
			select {
			case <-ctx.Done():
				return
			case e, ok := <-sub.Out():
				if !ok {
					return
				}
				evt := e.(event.EvtLocalReachabilityChanged)
				switch evt.Reachability {
				case network.ReachabilityPublic:
					an.status = NATPublic
					log.Println("AutoNAT: Node is reachable from the public internet (Public NAT).")
				case network.ReachabilityPrivate:
					an.status = NATPrivate
					log.Println("AutoNAT: Node is NOT reachable from the public internet (Private NAT).")
				case network.ReachabilityUnknown:
					an.status = NATUnknown
					log.Println("AutoNAT: Node reachability is unknown.")
				}
			}
		}
	}()

	return an, nil
}

// Status returns the currently detected NAT status.
func (a *AutoNAT) Status() NATStatus {
	if a == nil {
		return NATUnknown
	}
	return a.status
}

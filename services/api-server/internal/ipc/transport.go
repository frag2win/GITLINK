package ipc

import (
	"context"
	"fmt"
	"net"
	"time"
)

// Transport defines the interface for dialing the IPC server.
type Transport interface {
	DialContext(ctx context.Context) (net.Conn, error)
}

type UnixSocketTransport struct {
	socketPath string
	timeout    time.Duration
}

func NewUnixSocketTransport(socketPath string, timeout time.Duration) *UnixSocketTransport {
	return &UnixSocketTransport{
		socketPath: socketPath,
		timeout:    timeout,
	}
}

func (t *UnixSocketTransport) DialContext(ctx context.Context) (net.Conn, error) {
	dialer := net.Dialer{Timeout: t.timeout}
	return dialer.DialContext(ctx, "unix", t.socketPath)
}

type TcpLoopbackTransport struct {
	address string
	timeout time.Duration
}

func NewTcpLoopbackTransport(address string, timeout time.Duration) *TcpLoopbackTransport {
	return &TcpLoopbackTransport{
		address: address,
		timeout: timeout,
	}
}

func (t *TcpLoopbackTransport) DialContext(ctx context.Context) (net.Conn, error) {
	dialer := net.Dialer{Timeout: t.timeout}
	return dialer.DialContext(ctx, "tcp", t.address)
}

// NewTransport creates the appropriate Transport based on the network type.
func NewTransport(network string, address string, timeout time.Duration) (Transport, error) {
	switch network {
	case "unix":
		return NewUnixSocketTransport(address, timeout), nil
	case "tcp":
		return NewTcpLoopbackTransport(address, timeout), nil
	default:
		return nil, fmt.Errorf("unsupported ipc transport network: %s", network)
	}
}

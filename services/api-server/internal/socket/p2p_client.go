package socket

import (
	"encoding/json"
	"fmt"
	"log"
	"net"
	"time"
)

// P2PClient listens on a Unix domain socket for commands sent by the
// libp2p-node sidecar. The libp2p-node forwards remote peer requests
// (clone, push, pull) to the api-server through this socket.
type P2PClient struct {
	socketPath string
	listener   net.Listener
}

// NewP2PClient creates a P2PClient bound to the given socket path.
func NewP2PClient(socketPath string) *P2PClient {
	return &P2PClient{
		socketPath: socketPath,
	}
}

// Listen starts accepting connections on the Unix domain socket.
// It blocks and should be called in a goroutine.
func (pc *P2PClient) Listen(handler func(req *Request) *Response) error {
	var err error
	pc.listener, err = net.Listen("unix", pc.socketPath)
	if err != nil {
		return fmt.Errorf("listen on %s: %w", pc.socketPath, err)
	}
	defer pc.listener.Close()

	log.Printf("p2p socket listening on %s", pc.socketPath)

	for {
		conn, err := pc.listener.Accept()
		if err != nil {
			// Listener closed — exit gracefully.
			return nil
		}

		go pc.handleConnection(conn, handler)
	}
}

// handleConnection reads a single request, passes it to the handler,
// and writes the response back.
func (pc *P2PClient) handleConnection(conn net.Conn, handler func(*Request) *Response) {
	defer conn.Close()

	_ = conn.SetDeadline(time.Now().Add(30 * time.Second))

	var req Request
	if err := json.NewDecoder(conn).Decode(&req); err != nil {
		log.Printf("p2p socket decode error: %v", err)
		return
	}

	resp := handler(&req)
	if resp == nil {
		resp = &Response{
			Success: false,
			Error:   "handler returned nil response",
		}
	}

	if err := json.NewEncoder(conn).Encode(resp); err != nil {
		log.Printf("p2p socket encode error: %v", err)
	}
}

// Close stops the listener.
func (pc *P2PClient) Close() error {
	if pc.listener != nil {
		return pc.listener.Close()
	}
	return nil
}

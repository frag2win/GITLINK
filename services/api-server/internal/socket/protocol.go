package socket

// Action constants define the operations supported over the socket
// protocol. Both the git-server and libp2p-node use these same action
// names to ensure a consistent vocabulary.
const (
	// Git operations
	ActionInitRepo     = "init-repo"
	ActionDeleteRepo   = "delete-repo"
	ActionListBranches = "list-branches"
	ActionGetBranch    = "get-branch"
	ActionCreateBranch = "create-branch"
	ActionDeleteBranch = "delete-branch"
	ActionLog          = "log"
	ActionShow         = "show"
	ActionLsTree       = "ls-tree"
	ActionCatFile      = "cat-file"

	// P2P operations
	ActionPeerClone     = "peer-clone"
	ActionPeerPush      = "peer-push"
	ActionPeerPull      = "peer-pull"
	ActionPeerHandshake = "peer-handshake"
)

// Request is the JSON message sent over the Unix domain socket.
type Request struct {
	// Action is one of the Action* constants above.
	Action string `json:"action"`

	// Params carries action-specific key-value arguments.
	Params map[string]string `json:"params,omitempty"`

	// PeerID identifies the remote peer that originated the request
	// (set by the libp2p-node when forwarding).
	PeerID string `json:"peerID,omitempty"`

	// Payload is an optional binary or large-text payload encoded as
	// base64 for transport over JSON.
	Payload string `json:"payload,omitempty"`
}

// Response is the JSON message returned over the Unix domain socket.
type Response struct {
	// Success indicates whether the requested action completed without error.
	Success bool `json:"success"`

	// Error contains a human-readable error message when Success is false.
	Error string `json:"error,omitempty"`

	// Data carries the action-specific result. The shape depends on the
	// action that was requested.
	Data interface{} `json:"data,omitempty"`
}

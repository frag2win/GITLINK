# p2p-protocol

## Local HTTP Proxy Design Decision

### Why http://localhost:4000/p2p/<peer-id>/repo instead of git-remote-p2p://

Two approaches exist for Git-over-P2P:

**Option A — Custom Git remote helper binary:**
- User installs git-remote-p2p binary to PATH
- Runs: git clone p2p://12D3KooW.../my-repo
- Git delegates to the binary for transport

**Option B — Local HTTP proxy (chosen approach):**
- libp2p-node daemon exposes HTTP on localhost:4000
- Runs: git clone http://localhost:4000/p2p/12D3KooW.../my-repo
- Standard git HTTP transport, no custom binary needed

**Why Option B:**
- Zero installation friction — standard git works unchanged
- Works on any OS without PATH manipulation
- The daemon is already running — the proxy costs nothing extra
- HTTP transport is debuggable with curl during development
- Identical wire format to Phase 1 HTTP — proven before P2P added

**How it works:**
proxy/proxy.go parses the Peer ID from the URL path, dials a
libp2p stream to that peer using the /localrepo/git/1.0.0 protocol,
then pipes raw HTTP bytes bidirectionally between the git client
and the remote libp2p stream. The remote peer's protocol/handler.go
receives the stream and reverse-proxies it to its local api-server:3000.

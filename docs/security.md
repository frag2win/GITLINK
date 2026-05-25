# Security Architecture — P2P Git Hosting Platform

## Overview

The platform is designed with **defense in depth**. No single compromised component can grant an attacker full access to repository data or the host machine. The architecture relies on three core security mechanisms:

1. **Container isolation** with restricted network access
2. **Noise protocol** for authenticated, encrypted peer-to-peer transport
3. **Per-repository access control lists (ACLs)** for fine-grained authorization

---

## Threat Model Summary

### In Scope

| Threat                                  | Mitigation                                         |
|-----------------------------------------|-----------------------------------------------------|
| Remote code execution in git-server     | `network: none` — no outbound exfiltration possible |
| Malicious pack-file injection           | Object validation before writing to disk            |
| Man-in-the-middle on peer connections   | Noise protocol with mutual PeerID authentication    |
| Unauthorized repository access          | Per-repo ACLs checked on every operation            |
| API server compromise → git data theft  | API has no direct filesystem access to repo objects  |
| Rogue peer sending crafted messages     | libp2p protocol validation + message size limits    |
| Denial of service via resource exhaustion | Rate limiting on API + connection limits on libp2p |

### Out of Scope (Trust Assumptions)

- The **host operating system** is trusted (root access = game over for any local app).
- **Docker daemon** is trusted and properly configured.
- The user's **local network** is not assumed to be safe (all peer traffic is encrypted).
- **Physical access** to the machine is not defended against.

---

## Container Isolation

### Network Segmentation

```
┌─────────────────────────────────────────────────┐
│                   Host Network                   │
│                                                   │
│   ┌───────────────┐                               │
│   │  libp2p-node  │  ← Only container on host net │
│   │  (Go)         │     Inbound: TCP 4001          │
│   └───────┬───────┘     Outbound: Any (p2p)        │
│           │                                         │
│           │ Unix sockets only                       │
│           ▼                                         │
│   ┌───────────────┐         ┌───────────────────┐  │
│   │  api-server   │         │   git-server      │  │
│   │  (bridge net) │         │   (network: none) │  │
│   │  Port 3000    │         │   NO TCP/UDP      │  │
│   └───────────────┘         └───────────────────┘  │
└─────────────────────────────────────────────────────┘
```

| Container    | Network Mode | Can reach internet? | Can reach LAN? | Can receive inbound? |
|-------------|-------------|---------------------|----------------|---------------------|
| libp2p-node | host        | ✅ Yes               | ✅ Yes          | ✅ Yes (port 4001)   |
| api-server  | bridge      | ❌ No (internal)     | ❌ No           | ✅ Mapped port 3000  |
| git-server  | none        | ❌ No                | ❌ No           | ❌ No                |

### Why `network: none` for git-server?

The git server is the component most likely to process untrusted input (pack-files from remote peers). By removing its network stack entirely:

- **Exfiltration is impossible** — Even with arbitrary code execution, there is no network interface to send data through.
- **Inbound attacks are blocked** — No listening sockets exist; the only entry point is the Unix domain socket.
- **Blast radius is minimized** — A compromised git-server can only affect the `/repos` directory and the Unix socket.

### Container Capabilities

All containers run with a **minimal capability set**:

```yaml
# Applied to all services
security_opt:
  - no-new-privileges:true
cap_drop:
  - ALL
cap_add:
  - DAC_OVERRIDE   # Required for git-server to manage repo files
```

---

## Noise Protocol — Peer-to-Peer Encryption

All libp2p connections use the **Noise XX handshake pattern**, providing:

| Property              | Details                                         |
|----------------------|--------------------------------------------------|
| Key exchange         | X25519 Diffie-Hellman                            |
| Cipher               | ChaChaPoly (ChaCha20-Poly1305)                  |
| Authentication       | Mutual — both peers prove their PeerID           |
| Forward secrecy      | ✅ Yes — ephemeral keys per session               |
| Identity binding     | PeerID = hash of public key (self-certifying)    |

### Handshake Flow

```
Initiator (A)                         Responder (B)
    │                                       │
    │──── e, s ────────────────────────────►│  A sends ephemeral + static key
    │                                       │
    │◄─── e, ee, se, s, es ────────────────│  B responds with keys + proof
    │                                       │
    │──── es ──────────────────────────────►│  A completes handshake
    │                                       │
    │◄════ Encrypted channel established ═══►│
```

After the handshake, both peers have authenticated each other's PeerID and established an encrypted channel. All subsequent data (git pack-files, ref advertisements, control messages) is encrypted.

### PeerID Verification

- Each node generates an **Ed25519 key pair** on first run, stored in the `peer-identity` volume.
- The **PeerID** is the multihash of the public key.
- When connecting to a known peer, the initiator verifies that the responder's PeerID matches the expected value.
- Unknown peers can be accepted or rejected based on the user's trust settings.

---

## Per-Repository Access Control

### ACL Model

Each repository has an associated ACL stored in the API server's SQLite database:

```json
{
  "repo": "my-project",
  "owner": "12D3KooWA1b2c3d4...",
  "acl": [
    {
      "peer_id": "12D3KooWX9y8z7w6...",
      "permissions": ["read", "write"]
    },
    {
      "peer_id": "12D3KooWM5n6o7p8...",
      "permissions": ["read"]
    }
  ],
  "default_policy": "deny"
}
```

### Permission Levels

| Permission | Allows                                      |
|-----------|----------------------------------------------|
| `read`    | Clone, fetch, view refs                      |
| `write`   | Push, create/delete branches and tags        |
| `admin`   | Modify ACL, delete repository, transfer ownership |

### Enforcement Points

1. **libp2p-node** — Checks ACL before proxying a peer's request to git-server. Rejects unauthorized requests at the protocol level.
2. **api-server** — Checks ACL before executing any repository operation via the REST API.
3. **git-server** — (Defense in depth) Validates that the request came from an authorized Unix socket client.

### Default Policies

| Setting          | Behavior                                             |
|-----------------|------------------------------------------------------|
| `deny` (default) | Only explicitly listed peers can access the repo    |
| `read-public`    | Any peer can clone/fetch; push requires ACL entry   |
| `allow-all`      | Any peer can read and write (use with caution)      |

---

## Input Validation

### Git Pack-File Validation

Before writing any pack-file to disk, the git-server performs:

1. **Header validation** — Verify the pack-file magic bytes (`PACK`), version (2 or 3), and object count.
2. **Object type checking** — Ensure each object is a valid Git type (commit, tree, blob, tag).
3. **SHA-1/SHA-256 integrity** — Verify the trailing checksum of the entire pack-file.
4. **Size limits** — Reject pack-files exceeding the configured maximum size (default: 1 GiB).
5. **Delta chain depth** — Limit delta chain depth to prevent decompression bombs.

### API Input Validation

- All API endpoints validate request bodies against JSON schemas.
- Repository names are restricted to `[a-zA-Z0-9._-]` with a maximum length of 128 characters.
- Peer IDs are validated as legitimate base58-encoded multihashes.
- Rate limiting is applied per-client (10 requests/second default).

### libp2p Message Validation

- Maximum message size: 4 MiB for control messages, streaming for pack data.
- Protocol version negotiation rejects unsupported versions.
- Malformed messages cause stream reset without crashing the node.

---

## Secrets Management

| Secret           | Storage Location        | Access Control                    |
|-----------------|------------------------|-----------------------------------|
| Peer private key | `peer-identity` volume | Read-only mount in libp2p-node    |
| SQLite database  | Host bind mount        | Accessible only by api-server     |
| `.env` file      | Host filesystem        | Never committed to git (in .gitignore) |

### Key Rotation

- Rotating the peer key changes the node's PeerID, effectively creating a new identity.
- To rotate: stop services, delete the `peer-identity` volume, restart. A new key is generated automatically.
- Peers that had the old PeerID in their ACLs will need to be updated.

---

## Logging & Auditing

All security-relevant events are logged with structured JSON:

```json
{
  "level": "warn",
  "service": "libp2p-node",
  "event": "auth_rejected",
  "peer_id": "12D3KooWX9y8z7w6...",
  "repo": "my-project",
  "reason": "not_in_acl",
  "timestamp": "2024-01-15T10:30:00Z"
}
```

### Logged Events

- Peer connection established / disconnected
- Authentication success / failure
- ACL check pass / deny
- Repository created / deleted
- Pack-file received / rejected
- API rate limit triggered

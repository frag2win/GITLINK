# Full A–Z File Structure

> Every file. Every directory. With its exact purpose.
> Build order: git-server (Rust) → api-server (Go) → libp2p-node (Go) → ui (React)

---

## Legend

```
[Rust]   = Rust source file
[Go]     = Go source file
[TS]     = TypeScript source file
[Config] = Configuration file
[SQL]    = SQL migration file
[Proto]  = Protobuf schema
[Shell]  = Shell script
[Docker] = Docker related file
[YAML]   = YAML configuration
[MD]     = Markdown documentation
```

---

## Complete Tree

```
[project-name]/
│
├── .env.example                          [Config]  template for all env vars — copy to .env, never commit .env
├── .gitignore                            [Config]  ignores: .env, target/, dist/, *.sock, node_modules/
├── .dockerignore                         [Config]  excludes node_modules, .git, test files from Docker context
├── Makefile                              [Config]  dev/build/test commands — single entry point for all ops
├── README.md                             [MD]      project overview, quick start, architecture summary
│
├── docker-compose.yml                    [YAML]    production: 3 containers, volumes, network modes, resource limits
├── docker-compose.test.yml               [YAML]    test topology: isolated network, toxiproxy sidecar, temp volumes
├── docker-compose.dev.yml                [YAML]    dev overrides: hot reload mounts, relaxed resource limits
│
│
├── proto/                                          shared socket protocol — Go and Rust both generated from here
│   ├── git_commands.proto                [Proto]   defines all commands: CreateRepo, Push, Pull, ListCommits, etc
│   └── generated/
│       ├── git_commands.pb.go            [Go]      auto-generated — never edit manually
│       └── git_commands_pb.rs            [Rust]    auto-generated — never edit manually
│
│
├── config/                                         environment variable files — one place for all services
│   ├── development.env                   [Config]  dev values: DEBUG=true, log level verbose, local socket paths
│   ├── test.env                          [Config]  test values: in-memory DB, temp dirs, deterministic ports
│   └── production.env                    [Config]  production defaults: log level warn, resource paths
│
│
├── docs/                                           project documentation — everything needed to understand the system
│   ├── architecture.md                   [MD]      container topology, data flow diagrams, design decisions
│   ├── security.md                       [MD]      threat model, attack surface, Docker least-privilege rationale
│   ├── development.md                    [MD]      how to set up dev environment, run tests, add features
│   ├── testing.md                        [MD]      all test layers: unit → integration → NAT simulation → 2-machine
│   ├── p2p-protocol.md                   [MD]      libp2p internals: DHT, hole punching, relay, Noise encryption
│   └── api.md                            [MD]      REST API reference: all endpoints, request/response shapes
│
│
├── scripts/                                        automation — manual steps must not exist in docs
│   ├── setup.sh                          [Shell]   install deps, generate proto, create .env from example
│   ├── build.sh                          [Shell]   build all three service binaries + React static build
│   ├── test.sh                           [Shell]   run full test suite: unit → integration → Docker tests
│   ├── test-unit.sh                      [Shell]   run only unit tests across all services (fast, no Docker)
│   ├── test-integration.sh               [Shell]   run integration tests with Docker compose test topology
│   ├── two-machine-test.sh               [Shell]   real NAT test: run on Machine B, targets Machine A peer ID
│   ├── generate-proto.sh                 [Shell]   runs protoc to regenerate Go and Rust from .proto file
│   └── revoke-contributor.sh             [Shell]   remove SSH key + peer ID from DB, log revocation to audit
│
│
├── services/
│   │
│   │
│   ├── git-server/                                 [Rust] network_mode: none — touches /repos, nothing else
│   │   │
│   │   ├── Cargo.toml                    [Config]  deps: git2, tokio, serde, serde_json, rusqlite, prost, thiserror
│   │   ├── Cargo.lock                    [Config]  exact versions — committed for reproducible builds
│   │   ├── Dockerfile                    [Docker]  multi-stage: cargo build → copy binary to distroless image
│   │   ├── .dockerignore                 [Config]  excludes target/ from Docker context — huge directory
│   │   │
│   │   └── src/
│   │       │
│   │       ├── main.rs                   [Rust]    entry: parse config, bind unix socket, start tokio runtime
│   │       ├── config.rs                 [Rust]    reads env vars: socket path, repos path, log level
│   │       ├── error.rs                  [Rust]    unified error type via thiserror — all errors flow through here
│   │       │
│   │       ├── git/                                pure git operations — no socket knowledge, no HTTP
│   │       │   ├── mod.rs                [Rust]    re-exports all git submodules
│   │       │   ├── repository.rs         [Rust]    create/open/delete bare repos via git2 — path jailing here
│   │       │   ├── commits.rs            [Rust]    walk commit history, format commit objects, diff between refs
│   │       │   ├── branches.rs           [Rust]    list/create/delete branches, resolve refs to commit hashes
│   │       │   ├── objects.rs            [Rust]    read blob/tree/tag objects — powers file browser
│   │       │   ├── pack.rs               [Rust]    receive and apply pack files — Rust borrow checker critical here
│   │       │   └── sanitize.rs           [Rust]    validate repo names, block path traversal, enforce naming rules
│   │       │
│   │       ├── socket/                             unix socket server — receives commands from api-server
│   │       │   ├── mod.rs                [Rust]    re-exports socket submodules
│   │       │   ├── server.rs             [Rust]    tokio unix socket listener, accept loop, spawn handler per conn
│   │       │   ├── protocol.rs           [Rust]    deserialise protobuf commands, serialise responses
│   │       │   └── connection.rs         [Rust]    per-connection state, read framing, write framing
│   │       │
│   │       ├── handlers/                           route commands to git operations
│   │       │   ├── mod.rs                [Rust]    match command type → call correct handler
│   │       │   ├── repo_handler.rs       [Rust]    handles CreateRepo, DeleteRepo, ListRepos commands
│   │       │   ├── push_handler.rs       [Rust]    handles Push command — receive pack, apply, update refs
│   │       │   ├── pull_handler.rs       [Rust]    handles Pull/Clone — build pack file, stream to caller
│   │       │   ├── commit_handler.rs     [Rust]    handles ListCommits, GetCommit, GetDiff commands
│   │       │   └── object_handler.rs     [Rust]    handles GetFile, GetTree — powers UI file browser
│   │       │
│   │       └── tests/                              integration tests — require actual git repos on disk
│   │           ├── mod.rs                [Rust]    test helpers: temp repo dir, socket client stub
│   │           ├── repo_test.rs          [Rust]    create repo, verify bare structure, delete, verify gone
│   │           ├── push_pull_test.rs     [Rust]    push a commit, pull it back, assert identical pack
│   │           ├── path_traversal_test.rs[Rust]    attempt ../../etc/passwd names — must all return errors
│   │           ├── pack_fuzz_test.rs     [Rust]    malformed pack files — must error, never panic
│   │           └── concurrent_push_test.rs[Rust]   3 goroutines push simultaneously — no corruption
│   │
│   │
│   ├── api-server/                                 [Go + Fiber] bridge: internal — REST API + serves React UI
│   │   │
│   │   ├── go.mod                        [Config]  module: github.com/[name]/api-server
│   │   ├── go.sum                        [Config]  exact dependency checksums
│   │   ├── Dockerfile                    [Docker]  multi-stage: go build → copy binary + embedded UI to minimal image
│   │   │
│   │   ├── cmd/
│   │   │   └── server/
│   │   │       └── main.go               [Go]      entry: load config, init DB, open sockets, start Fiber app
│   │   │
│   │   └── internal/
│   │       │
│   │       ├── config/
│   │       │   └── config.go             [Go]      reads all env vars into a typed Config struct, validates required fields
│   │       │
│   │       ├── router/
│   │       │   └── router.go             [Go]      all route definitions in one place — maps paths to handler funcs
│   │       │
│   │       ├── handlers/                           one file per resource — HTTP in, HTTP out, no business logic
│   │       │   ├── repos.go              [Go]      GET/POST/DELETE /repos — list, create, delete repos
│   │       │   ├── repos_test.go         [Go]      unit tests: mock socket, assert HTTP responses
│   │       │   ├── commits.go            [Go]      GET /repos/:name/commits — commit log with pagination
│   │       │   ├── commits_test.go       [Go]      unit tests for commit list handler
│   │       │   ├── branches.go           [Go]      GET/POST/DELETE /repos/:name/branches
│   │       │   ├── branches_test.go      [Go]      unit tests for branch handlers
│   │       │   ├── files.go              [Go]      GET /repos/:name/tree/:ref/:path — file browser content
│   │       │   ├── files_test.go         [Go]      unit tests for file tree handler
│   │       │   ├── contributors.go       [Go]      GET/POST/DELETE /contributors — manage SSH keys + peer IDs
│   │       │   ├── contributors_test.go  [Go]      unit tests for contributor management
│   │       │   ├── peers.go              [Go]      GET /peers — connection status, queue depth, peer list
│   │       │   ├── peers_test.go         [Go]      unit tests for peer status handler
│   │       │   ├── audit.go              [Go]      GET /audit — paginated audit log query
│   │       │   └── audit_test.go         [Go]      unit tests for audit log handler
│   │       │
│   │       ├── middleware/
│   │       │   ├── auth.go               [Go]      verify SSH key or peer ID on every non-public request
│   │       │   ├── auth_test.go          [Go]      valid key passes, invalid key 401, missing key 401
│   │       │   ├── logger.go             [Go]      structured request logging: method, path, status, latency
│   │       │   ├── audit.go              [Go]      write audit log entry for every mutating operation
│   │       │   └── ratelimit.go          [Go]      per-peer rate limiting — prevent hammering from one contributor
│   │       │
│   │       ├── models/
│   │       │   ├── repo.go               [Go]      Repo struct: name, description, created_at, contributor_count
│   │       │   ├── commit.go             [Go]      Commit struct: hash, author, message, timestamp, parent_hash
│   │       │   ├── contributor.go        [Go]      Contributor struct: peer_id, ssh_key, name, added_at, repos
│   │       │   ├── audit_entry.go        [Go]      AuditEntry: timestamp, peer_id, operation, repo, result
│   │       │   └── peer_status.go        [Go]      PeerStatus: peer_id, online, last_seen, queue_depth
│   │       │
│   │       ├── database/
│   │       │   ├── db.go                 [Go]      open SQLite connection, run migrations on startup
│   │       │   ├── db_test.go            [Go]      test migrations run clean on fresh DB
│   │       │   ├── repos.go              [Go]      repo CRUD queries
│   │       │   ├── contributors.go       [Go]      contributor CRUD + key lookup queries
│   │       │   ├── audit.go              [Go]      insert and query audit log entries
│   │       │   └── migrations/
│   │       │       ├── 001_initial.sql   [SQL]     create repos, contributors, audit_log tables
│   │       │       ├── 002_contributors.sql[SQL]   add per-repo access control join table
│   │       │       ├── 003_audit_log.sql [SQL]     add indexes on audit_log for time-range queries
│   │       │       └── 004_peer_status.sql[SQL]    add peer_status table for connection tracking
│   │       │
│   │       ├── socket/
│   │       │   ├── git_client.go         [Go]      connects to Rust git-server unix socket, sends proto commands
│   │       │   ├── git_client_test.go    [Go]      mock socket server, verify correct proto messages sent
│   │       │   ├── p2p_client.go         [Go]      connects to libp2p-node unix socket, forwards Git streams
│   │       │   └── p2p_client_test.go    [Go]      mock socket, verify stream forwarding behaviour
│   │       │
│   │       └── services/                           business logic layer — handlers call services, not DB directly
│   │           ├── repo_service.go       [Go]      create repo: validate name, write DB, send socket command to Rust
│   │           ├── repo_service_test.go  [Go]      mock socket + DB, test full repo creation flow
│   │           ├── auth_service.go       [Go]      validate SSH key format, check contributor DB, return identity
│   │           ├── audit_service.go      [Go]      write structured audit entries with context propagation
│   │           └── queue_service.go      [Go]      manage offline commit queue: enqueue, flush on reconnect
│   │
│   │   integration/
│   │       ├── push_test.go              [Go]      full e2e: HTTP push → Go socket → Rust → /repos
│   │       ├── clone_test.go             [Go]      full e2e: HTTP clone → Rust → pack file → Git client
│   │       └── auth_test.go              [Go]      valid key reaches handler, invalid key blocked at middleware
│   │
│   │
│   └── libp2p-node/                                [Go] network_mode: host — P2P only, no filesystem access
│       │
│       ├── go.mod                        [Config]  module: github.com/[name]/libp2p-node
│       ├── go.sum                        [Config]  exact dependency checksums
│       ├── Dockerfile                    [Docker]  multi-stage: go build → minimal image, only identity volume
│       │
│       ├── cmd/
│       │   └── node/
│       │       └── main.go               [Go]      entry: load identity, build host, start discovery, bind bridge
│       │
│       └── internal/
│           │
│           ├── config/
│           │   └── config.go             [Go]      socket path, bootstrap peers, DHT mode, relay config from env
│           │
│           ├── identity/
│           │   ├── keypair.go            [Go]      generate Ed25519 key on first run, load from disk on restart
│           │   └── keypair_test.go       [Go]      same key loaded twice gives same Peer ID — deterministic
│           │
│           ├── host/
│           │   ├── host.go               [Go]      build go-libp2p host: TCP transport, Noise, yamux, all options
│           │   └── host_test.go          [Go]      host starts, has valid Peer ID, accepts connections
│           │
│           ├── discovery/
│           │   ├── mdns.go               [Go]      mDNS service for same-network peer discovery — zero config
│           │   ├── mdns_test.go          [Go]      two nodes on loopback find each other within 5 seconds
│           │   ├── dht.go                [Go]      Kademlia DHT for cross-network discovery, bootstrap peers
│           │   └── dht_test.go           [Go]      node announces to DHT, second node finds it by Peer ID
│           │
│           ├── nat/
│           │   ├── autonat.go            [Go]      detect NAT type: full cone / restricted / symmetric
│           │   ├── autonat_test.go       [Go]      mock peers report reachability, assert correct NAT type detected
│           │   ├── dcutr.go              [Go]      DCUtR protocol: coordinate simultaneous hole punch via relay
│           │   └── dcutr_test.go         [Go]      two nodes behind simulated NAT connect directly after punch
│           │
│           ├── relay/
│           │   ├── relay.go              [Go]      circuit relay v2 client: find relay, register, use as fallback
│           │   └── relay_test.go         [Go]      symmetric NAT scenario: connection succeeds via relay
│           │
│           ├── protocol/
│           │   ├── git_protocol.go       [Go]      /git/1.0.0 protocol handler: receive Git streams over libp2p
│           │   ├── git_protocol_test.go  [Go]      open /git/1.0.0 stream, send pack data, verify received
│           │   ├── auth_protocol.go      [Go]      /auth/1.0.0 — verify contributor Peer ID before Git access
│           │   └── queue_protocol.go     [Go]      /queue/1.0.0 — receive queued commits when host comes online
│           │
│           └── bridge/
│               ├── socket.go             [Go]      forward incoming libp2p Git streams → api-server unix socket
│               ├── socket_test.go        [Go]      mock api-server socket, verify stream correctly forwarded
│               └── queue.go             [Go]      local queue: store commits when host offline, flush on reconnect
│
│
└── ui/                                             [React + Vite + TypeScript] static build — served by Go/Fiber
    │
    ├── package.json                      [Config]  deps: react, vite, tailwind, zustand, react-router
    ├── tsconfig.json                     [Config]  strict TypeScript config
    ├── vite.config.ts                    [TS]      build to /dist, proxy /api to Go in dev mode
    ├── tailwind.config.ts                [TS]      tailwind theme, dark mode class strategy
    ├── index.html                        [Config]  single HTML entry point — Vite injects script tag
    │
    └── src/
        │
        ├── main.tsx                      [TS]      React root: mount App, wrap with router and store providers
        ├── App.tsx                       [TS]      router config: maps URL paths to page components
        │
        ├── api/
        │   ├── client.ts                 [TS]      typed fetch wrapper: base URL, auth header, error handling
        │   ├── repos.ts                  [TS]      API calls: listRepos, createRepo, deleteRepo
        │   ├── commits.ts                [TS]      API calls: listCommits, getCommit, getDiff
        │   ├── branches.ts               [TS]      API calls: listBranches, createBranch, deleteBranch
        │   ├── contributors.ts           [TS]      API calls: listContributors, addContributor, revoke
        │   ├── files.ts                  [TS]      API calls: getTree, getFileContent
        │   └── peers.ts                  [TS]      API calls: getPeerStatus, getQueueDepth
        │
        ├── store/
        │   ├── connectionStore.ts        [TS]      Zustand: online/offline/syncing, queue count, last seen
        │   ├── repoStore.ts              [TS]      Zustand: current repo, active branch, selected file path
        │   └── authStore.ts              [TS]      Zustand: current contributor identity, SSH key fingerprint
        │
        ├── hooks/
        │   ├── useRepo.ts                [TS]      fetch repo data, loading state, refetch on change
        │   ├── useCommits.ts             [TS]      paginated commit list with infinite scroll
        │   ├── useBranches.ts            [TS]      branch list, current branch, switch handler
        │   ├── useConnection.ts          [TS]      subscribe to connection store, poll /peers endpoint
        │   └── useFileTree.ts            [TS]      recursive file tree fetch, expand/collapse state
        │
        ├── pages/
        │   ├── Dashboard.tsx             [TS]      home: repo list, connection status badge, quick actions
        │   ├── RepoDetail.tsx            [TS]      repo view: file browser + commit history side by side
        │   ├── CommitHistory.tsx         [TS]      full commit log with search and branch filter
        │   ├── BranchList.tsx            [TS]      all branches, create modal, delete with confirmation
        │   ├── Contributors.tsx          [TS]      contributor list, add SSH key form, revoke button
        │   ├── AuditLog.tsx              [TS]      time-ordered audit entries with filter by user/repo
        │   └── Settings.tsx              [TS]      Peer ID display, backup mirror config, app settings
        │
        ├── components/
        │   │
        │   ├── layout/
        │   │   ├── Layout.tsx            [TS]      outer shell: sidebar + top bar + main content area
        │   │   ├── Sidebar.tsx           [TS]      repo list nav, connection status indicator, settings link
        │   │   └── TopBar.tsx            [TS]      current repo name, branch selector, search
        │   │
        │   ├── repo/
        │   │   ├── RepoCard.tsx          [TS]      repo list item: name, last commit, contributor count
        │   │   ├── CreateRepoModal.tsx   [TS]      modal: repo name input, description, create button
        │   │   └── RepoHeader.tsx        [TS]      repo page header: name, clone URL, branch selector
        │   │
        │   ├── git/
        │   │   ├── FileBrowser.tsx       [TS]      tree view of repo files, click to view content
        │   │   ├── FileViewer.tsx        [TS]      raw file content display with syntax highlighting
        │   │   ├── CommitList.tsx        [TS]      scrollable commit list with hash, author, message, time
        │   │   ├── CommitDetail.tsx      [TS]      single commit: metadata + unified diff view
        │   │   ├── DiffViewer.tsx        [TS]      unified diff with added/removed line highlighting
        │   │   └── BranchSelector.tsx    [TS]      dropdown: switch branch, shows current branch name
        │   │
        │   └── common/
        │       ├── StatusBadge.tsx       [TS]      online/offline/syncing/queued — colour coded badge
        │       ├── QueueIndicator.tsx    [TS]      shows count of queued commits when host offline
        │       ├── Modal.tsx             [TS]      generic modal wrapper with backdrop and close handler
        │       ├── Spinner.tsx           [TS]      loading spinner for async operations
        │       ├── EmptyState.tsx        [TS]      empty list placeholder with icon and action button
        │       ├── ErrorBoundary.tsx     [TS]      catches render errors, shows fallback with retry
        │       ├── CopyButton.tsx        [TS]      copy text to clipboard with confirmation tick
        │       └── ConfirmDialog.tsx     [TS]      "are you sure?" dialog for destructive actions
        │
        ├── types/
        │   ├── repo.ts                   [TS]      Repo, CreateRepoRequest, RepoListResponse interfaces
        │   ├── commit.ts                 [TS]      Commit, CommitDiff, DiffHunk interfaces
        │   ├── contributor.ts            [TS]      Contributor, AddContributorRequest interfaces
        │   ├── peer.ts                   [TS]      PeerStatus, ConnectionState, QueueEntry interfaces
        │   └── api.ts                    [TS]      ApiResponse<T>, ApiError, PaginatedResponse<T> generics
        │
        ├── constants/
        │   ├── endpoints.ts              [TS]      all API paths as typed constants — no magic strings anywhere
        │   └── config.ts                 [TS]      poll intervals, timeout values, max file size for viewer
        │
        └── utils/
            ├── formatters.ts             [TS]      format commit hash (short), relative time, file size
            ├── classnames.ts             [TS]      merge Tailwind class strings safely
            └── errors.ts                 [TS]      parse API error responses into user-friendly messages
```

---

## File Count Summary

| Service | Files | Language |
|---------|-------|----------|
| `services/git-server` | 28 | Rust |
| `services/api-server` | 52 | Go |
| `services/libp2p-node` | 26 | Go |
| `ui` | 48 | TypeScript / React |
| Root + proto + config + docs + scripts | 30 | Mixed |
| **Total** | **184** | |

---

## Build Order (Which File to Write First)

```
Step 1 — proto/git_commands.proto
         └── defines the Go↔Rust contract before either side is written

Step 2 — services/git-server/src/git/sanitize.rs
         └── safety first: path validation before any repo ops

Step 3 — services/git-server/src/git/repository.rs
         └── create and open bare repos

Step 4 — services/git-server/src/socket/server.rs
         └── expose git ops over unix socket

Step 5 — services/api-server/internal/socket/git_client.go
         └── Go side of the socket — verify proto contract works

Step 6 — services/api-server/internal/handlers/repos.go
         └── first real HTTP endpoint — testable with curl

Step 7 — services/api-server/internal/database/migrations/001_initial.sql
         └── schema before any DB code

Step 8 — services/libp2p-node/internal/identity/keypair.go
         └── Peer ID before any networking

Step 9 — services/libp2p-node/internal/host/host.go
         └── libp2p host with TCP + Noise

Step 10 — services/libp2p-node/internal/discovery/mdns.go
          └── local network works before cross-network

Step 11 — ui/src/api/client.ts
          └── typed API client before any components

Step 12 — ui/src/pages/Dashboard.tsx
          └── first page — repo list
```

---

## Critical Files (Must Be Correct Before Others)

| File | Why Critical |
|------|-------------|
| `proto/git_commands.proto` | Both Go and Rust depend on this — wrong here breaks everything |
| `services/git-server/src/git/sanitize.rs` | Path traversal lives here — must be correct before any repo ops |
| `services/git-server/src/git/pack.rs` | Rust borrow checker guards the most dangerous parsing surface |
| `services/api-server/internal/middleware/auth.go` | Wrong here means unauthenticated access to all repos |
| `services/api-server/internal/database/migrations/001_initial.sql` | Schema is the foundation — wrong here cascades everywhere |
| `services/libp2p-node/internal/identity/keypair.go` | Peer ID is permanent identity — must be deterministic and stable |
| `services/libp2p-node/internal/nat/dcutr.go` | Hole punching timing — wrong here means cross-network fails silently |
| `docker-compose.yml` | Wrong network_mode kills P2P or exposes repos — must be exact |

---

*184 files total. Every file has exactly one job.*

# Phase 5: Operations & Real-Time Infrastructure

## Architecture Overview

Phase 5 equips GITLINK with enterprise operational infrastructure: live real-time WebSockets streaming with missed event replay, Dead-Letter Queue (DLQ) audit-validated task replay, tiered health diagnostics, categorized operational metrics, and a Conflict Analysis & Diagnostics Engine.

```
+-----------------------------------------------------------------------+
|                             GITLINK Web UI                            |
+-----------------------------------------------------------------------+
        ^                                               ^
        | WebSocket (/ws/notifications)                 | REST API
        v                                               v
+-----------------------------------------------------------------------+
|                        services/api-server                            |
|                                                                       |
|  +-----------------+    +-------------------+    +-----------------+  |
|  | WebSocketHub    |<---| Domain EventBus   |--->| NotificationSvc |  |
|  +-----------------+    +-------------------+    +-----------------+  |
|                                                                       |
|  +---------------------+    +-------------------+  +---------------+  |
|  | ConflictAnalysisSvc |    | Admin Ops & DLQ   |  | Tiered Health |  |
|  +---------------------+    +-------------------+  +---------------+  |
+-----------------------------------------------------------------------+
```

---

## Technical Core Components

### 1. Real-Time WebSockets Engine with Reconnect Resilience
- Connection Hub (`ws_hub.go`): Thread-safe WebSocket connection pool mapped to authenticated user IDs.
- Fiber Handler (`ws_handler.go`): Endpoint `/ws/notifications` using header-based JWT authentication (`Authorization: Bearer <token>`).
- Reconnect Replay: Supports `last_event_id` query parameter to query `NotificationService` and stream missed notifications received during offline period.

### 2. Dead-Letter Queue (DLQ) & Admin Operations API
- DLQ Inspection (`GET /api/v1/sync/dlq`): Lists failed tasks exceeding maximum retries (`retry_count >= max_retries`).
- Authorized Replay (`POST /api/v1/sync/dlq/:id/replay`): Enforces Admin authorization (`RequireAdminRole`), validates schema, resets `retry_count = 0`, updates status to `pending`, and logs entry to `audit_logs`.
- Admin Operations API (`admin_handler.go`): `/api/v1/admin/workers`, `/api/v1/admin/peers`, `/api/v1/admin/sync/restart`.

### 3. Conflict Analysis & Diagnostics Engine
- Diagnostic Engine (`conflict_service.go`): Computes 3-way merge base OID, analyzes file-level diff hunks, and outputs a diagnostic `ConflictReport`.
- Detailed Conflict Model (`conflict_report.go`): Stores `MergeBaseSHA`, `BaseCommit`, `HeadCommit`, and `Files[]` with `Hunks[]` (`StartLine`, `EndLine`, `Reason`, `GitConflictType`).

### 4. Tiered Health Endpoints & Categorized Metrics
- Health Levels:
  - `GET /health` -> Liveness (`200 OK`)
  - `GET /ready` -> Readiness Check
  - `GET /api/v1/health/deep` -> Deep Diagnostics returning DB, Git Server IPC, libp2p, sync_worker, queue_depth, pending_dlq, and uptime_seconds.
- Categorized Operational Metrics (`GET /api/v1/metrics`): Returns structured JSON (`repository_metrics`, `sync_metrics`, `peer_metrics`, `queue_metrics`, `review_metrics`, `system_metrics`).

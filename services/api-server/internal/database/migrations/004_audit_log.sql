-- ---- v4: Audit log table ----
CREATE TABLE IF NOT EXISTS audit_logs (
    id        TEXT PRIMARY KEY,
    timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
    peer_id   TEXT NOT NULL,
    operation TEXT NOT NULL,
    repo_name TEXT DEFAULT '',
    details   TEXT DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_audit_logs_peer ON audit_logs(peer_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_time ON audit_logs(timestamp);

-- ---- v3: Permissions join table ----
CREATE TABLE IF NOT EXISTS permissions (
    repo_id TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
    peer_id TEXT NOT NULL REFERENCES contributors(peer_id) ON DELETE CASCADE,
    role    TEXT NOT NULL DEFAULT 'read' CHECK(role IN ('read', 'write', 'admin')),
    PRIMARY KEY (repo_id, peer_id)
);

CREATE INDEX IF NOT EXISTS idx_permissions_peer ON permissions(peer_id);

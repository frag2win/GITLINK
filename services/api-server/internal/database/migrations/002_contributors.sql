-- ---- v2: Contributors table ----
CREATE TABLE IF NOT EXISTS contributors (
    id         TEXT PRIMARY KEY,
    name       TEXT NOT NULL DEFAULT '',
    public_key TEXT NOT NULL UNIQUE,
    peer_id    TEXT NOT NULL UNIQUE,
    created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
);

package database

import "fmt"

// migrations is the ordered list of SQL statements to bring the database
// schema up to date. Each migration is idempotent (uses IF NOT EXISTS).
var migrations = []string{
	// ---- v1: Repositories table ----
	`CREATE TABLE IF NOT EXISTS repos (
		id          TEXT PRIMARY KEY,
		name        TEXT NOT NULL UNIQUE,
		description TEXT NOT NULL DEFAULT '',
		owner       TEXT NOT NULL,
		is_private  INTEGER NOT NULL DEFAULT 0,
		created_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		updated_at  DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`,

	// ---- v2: Contributors table ----
	`CREATE TABLE IF NOT EXISTS contributors (
		id         TEXT PRIMARY KEY,
		name       TEXT NOT NULL DEFAULT '',
		public_key TEXT NOT NULL UNIQUE,
		peer_id    TEXT NOT NULL UNIQUE,
		created_at DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP
	);`,

	// ---- v3: Permissions join table ----
	`CREATE TABLE IF NOT EXISTS permissions (
		repo_id TEXT NOT NULL REFERENCES repos(id) ON DELETE CASCADE,
		peer_id TEXT NOT NULL REFERENCES contributors(peer_id) ON DELETE CASCADE,
		role    TEXT NOT NULL DEFAULT 'read' CHECK(role IN ('read', 'write', 'admin')),
		PRIMARY KEY (repo_id, peer_id)
	);`,

	// ---- v4: Audit log table ----
	`CREATE TABLE IF NOT EXISTS audit_logs (
		id        TEXT PRIMARY KEY,
		timestamp DATETIME NOT NULL DEFAULT CURRENT_TIMESTAMP,
		peer_id   TEXT NOT NULL,
		operation TEXT NOT NULL,
		repo_name TEXT DEFAULT '',
		details   TEXT DEFAULT ''
	);`,

	// ---- Indexes ----
	`CREATE INDEX IF NOT EXISTS idx_repos_owner       ON repos(owner);`,
	`CREATE INDEX IF NOT EXISTS idx_permissions_peer   ON permissions(peer_id);`,
	`CREATE INDEX IF NOT EXISTS idx_audit_logs_peer    ON audit_logs(peer_id);`,
	`CREATE INDEX IF NOT EXISTS idx_audit_logs_time    ON audit_logs(timestamp);`,
}

// Migrate applies all pending database migrations in order.
func (db *DB) Migrate() error {
	for i, stmt := range migrations {
		if _, err := db.conn.Exec(stmt); err != nil {
			return fmt.Errorf("migration %d failed: %w", i+1, err)
		}
	}
	return nil
}

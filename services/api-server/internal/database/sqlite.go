// Package database provides SQLite storage for the api-server.
//
// It uses modernc.org/sqlite (a pure-Go SQLite driver) so the binary
// can be statically compiled without CGO.
package database

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"

	_ "modernc.org/sqlite"
)

// DB wraps the standard sql.DB connection and provides domain-specific
// CRUD helpers.
type DB struct {
	mu   sync.RWMutex
	conn *sql.DB
}

// New opens (or creates) a SQLite database at the given path.
// The parent directory is created if it does not exist.
func New(dbPath string) (*DB, error) {
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return nil, fmt.Errorf("create db directory: %w", err)
	}

	conn, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("open sqlite: %w", err)
	}

	// Enable WAL mode for better concurrent read performance.
	if _, err := conn.Exec("PRAGMA journal_mode=WAL"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("set WAL mode: %w", err)
	}

	// Enable foreign keys.
	if _, err := conn.Exec("PRAGMA foreign_keys=ON"); err != nil {
		conn.Close()
		return nil, fmt.Errorf("enable foreign keys: %w", err)
	}

	return &DB{conn: conn}, nil
}

// Close closes the underlying database connection.
func (db *DB) Close() error {
	return db.conn.Close()
}

// Conn returns the raw *sql.DB for advanced queries.
func (db *DB) Conn() *sql.DB {
	return db.conn
}

// ---------- Repository CRUD ----------

// CreateRepo inserts a new repository record.
func (db *DB) CreateRepo(name, description, owner string, isPrivate bool) (string, error) {
	db.mu.Lock()
	defer db.mu.Unlock()

	// TODO: Generate UUID for ID.
	// TODO: INSERT INTO repos (id, name, description, owner, is_private, created_at, updated_at) VALUES (?, ?, ?, ?, ?, ?, ?).
	// TODO: Return the generated ID.

	return "", fmt.Errorf("CreateRepo not implemented")
}

// GetRepo retrieves a single repo by ID.
func (db *DB) GetRepo(id string) (*sql.Row, error) {
	// TODO: SELECT * FROM repos WHERE id = ?.
	return nil, fmt.Errorf("GetRepo not implemented")
}

// ListRepos returns all repos visible to a given peer.
func (db *DB) ListRepos(peerID string, page, limit int) (*sql.Rows, error) {
	// TODO: Query repos the peer owns or is a contributor of.
	// TODO: Apply pagination with LIMIT and OFFSET.
	return nil, fmt.Errorf("ListRepos not implemented")
}

// DeleteRepo removes a repo record by ID.
func (db *DB) DeleteRepo(id string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// TODO: DELETE FROM repos WHERE id = ?.
	return fmt.Errorf("DeleteRepo not implemented")
}

// ---------- Contributor CRUD ----------

// AddContributor links a peer to a repository with the given role.
func (db *DB) AddContributor(repoID, peerID, publicKey, role string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// TODO: INSERT INTO contributors and permissions tables.
	return fmt.Errorf("AddContributor not implemented")
}

// RemoveContributor unlinks a peer from a repository.
func (db *DB) RemoveContributor(repoID, peerID string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// TODO: DELETE FROM permissions WHERE repo_id = ? AND peer_id = ?.
	return fmt.Errorf("RemoveContributor not implemented")
}

// ---------- Audit Log ----------

// InsertAuditLog records an audit event.
func (db *DB) InsertAuditLog(peerID, operation, repoName, details string) error {
	db.mu.Lock()
	defer db.mu.Unlock()

	// TODO: INSERT INTO audit_logs (id, timestamp, peer_id, operation, repo_name, details) VALUES (...).
	return fmt.Errorf("InsertAuditLog not implemented")
}

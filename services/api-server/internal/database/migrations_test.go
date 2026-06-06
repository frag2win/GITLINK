package database

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMigrations_RunClean(t *testing.T) {
	// Open an in-memory SQLite DB
	db, err := New(":memory:")
	require.NoError(t, err)
	defer db.Close()

	// Run migrations
	err = db.Migrate()
	require.NoError(t, err)

	// Verify tables exist
	tables := []string{"repos", "contributors", "audit_logs", "permissions"}
	for _, table := range tables {
		var name string
		row := db.conn.QueryRow(
			"SELECT name FROM sqlite_master WHERE type='table' AND name=?", table,
		)
		err := row.Scan(&name)
		assert.NoError(t, err, "table %s should exist", table)
	}

	// Run migrations again — must be idempotent
	err = db.Migrate()
	assert.NoError(t, err, "migrations must be idempotent")
}

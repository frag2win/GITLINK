package database

import (
	"fmt"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

// DB wraps the gorm.DB connection.
type DB struct {
	Conn *gorm.DB
}

// New connects to PostgreSQL using the provided DSN (Data Source Name).
func New(dsn string) (*DB, error) {
	if dsn == "" {
		return nil, fmt.Errorf("dsn is empty")
	}

	conn, err := gorm.Open(postgres.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to open postgres: %w", err)
	}

	return &DB{Conn: conn}, nil
}

// Close is a no-op for GORM (connection pooling is handled automatically).
func (db *DB) Close() error {
	sqlDB, err := db.Conn.DB()
	if err != nil {
		return err
	}
	return sqlDB.Close()
}

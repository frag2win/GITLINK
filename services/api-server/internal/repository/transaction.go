package repository

import (
	"context"

	"gorm.io/gorm"
)

type txKey struct{}

// TransactionManager handles atomic database operations.
type TransactionManager interface {
	WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error
}

type transactionManager struct {
	db *gorm.DB
}

// NewTransactionManager creates a new TransactionManager.
func NewTransactionManager(db *gorm.DB) TransactionManager {
	return &transactionManager{db: db}
}

// WithTransaction executes the given function within a database transaction.
// The transaction is injected into the context passed to the function.
func (tm *transactionManager) WithTransaction(ctx context.Context, fn func(txCtx context.Context) error) error {
	return tm.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		txCtx := context.WithValue(ctx, txKey{}, tx)
		return fn(txCtx)
	})
}

// getDB extracts the transaction from the context if it exists, otherwise returns the default DB.
func getDB(ctx context.Context, db *gorm.DB) *gorm.DB {
	if tx, ok := ctx.Value(txKey{}).(*gorm.DB); ok {
		return tx
	}
	return db.WithContext(ctx)
}

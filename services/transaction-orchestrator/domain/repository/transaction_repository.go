package repository

import (
	"context"
	"micro-mart/services/transaction-orchestrator/domain/aggregate"
	"micro-mart/services/transaction-orchestrator/domain/entity"
)

// TransactionRepository defines the interface for transaction persistence
type TransactionRepository interface {
	// SaveTransaction persists a new transaction
	SaveTransaction(ctx context.Context, tx *aggregate.Transaction) error

	// UpdateTransaction updates an existing transaction
	UpdateTransaction(ctx context.Context, tx *aggregate.Transaction) error

	// GetTransaction retrieves a transaction by ID
	GetTransaction(ctx context.Context, id string) (*aggregate.Transaction, error)

	// FindTransactionsByStatus retrieves transactions by status
	FindTransactionsByStatus(ctx context.Context, statuses []entity.TransactionStatus) ([]*aggregate.Transaction, error)

	// DeleteTransaction deletes a transaction
	DeleteTransaction(ctx context.Context, id string) error
}

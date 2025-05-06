package db_impl

import (
	"context"
	"micro-mart/pkg/db"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/user/infrastructure/ent_impl/ent"

	_ "github.com/lib/pq"
)

// txKey is a type used as a key for storing transaction in context
type txKey struct{}

// NewTxKey creates a new txKey instance
func NewTxKey() txKey {
	return txKey{}
}

// EntDB implements the db.Database interface using ent ORM
type EntDB struct {
	client *ent.Client
}

var _ db.Database = (*EntDB)(nil)

// NewEntDb creates and initializes a new EntDB instance
//
// It reads database configuration, establishes a connection to the database,
// sets up connection pool, and optionally performs auto migration.
//
// Returns:
//   - db.Database: An interface that can be used to interact with the database
//
// Panics if it fails to connect to the database or create schema resources (when auto-migrate is enabled)
func NewEntDb(client *ent.Client) db.Database {
	return &EntDB{client: client}
}

func (e *EntDB) GetConn(ctx context.Context) any {
	return e.client
}

func (e *EntDB) GetTx(ctx context.Context) any {
	return ctx.Value(txKey{})
}

func (e *EntDB) GetClient(ctx context.Context) any {
	if tx, ok := e.GetTx(ctx).(*ent.Tx); ok {
		return tx.Client()
	} else {
		return e.GetConn(ctx).(*ent.Client)
	}
}

func (e *EntDB) Begin(ctx context.Context) (context.Context, *mmerror.MmError) {
	// Check if ent client is initialized
	if e.client == nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "ent client not found", nil)
		mmotel.Error(ctx, mmErr.Error())
		return nil, mmErr
	}

	tx, err := e.client.Tx(ctx)
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to start transaction", err)
		mmotel.Error(ctx, mmErr.Error())
		return nil, mmErr
	}
	return context.WithValue(ctx, txKey{}, tx), nil
}

func (e *EntDB) Commit(ctx context.Context) (context.Context, *mmerror.MmError) {
	tx, ok := ctx.Value(txKey{}).(*ent.Tx)
	if !ok {
		mmErr := mmerror.New(mmerror.InternalServerError, "transaction not found in context", nil)
		mmotel.Error(ctx, mmErr.Error())
		return ctx, mmErr
	}

	if err := tx.Commit(); err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to commit transaction", err)
		mmotel.Error(ctx, mmErr.Error())
		return ctx, mmErr
	}

	return context.WithValue(ctx, txKey{}, nil), nil
}

func (e *EntDB) Rollback(ctx context.Context) (context.Context, *mmerror.MmError) {
	tx, ok := ctx.Value(txKey{}).(*ent.Tx)
	if !ok {
		mmErr := mmerror.New(mmerror.InternalServerError, "transaction not found in context", nil)
		mmotel.Error(ctx, mmErr.Error())
		return ctx, mmErr
	}

	if err := tx.Rollback(); err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to rollback transaction", err)
		mmotel.Error(ctx, mmErr.Error())
		return ctx, mmErr
	}

	return context.WithValue(ctx, txKey{}, nil), nil
}

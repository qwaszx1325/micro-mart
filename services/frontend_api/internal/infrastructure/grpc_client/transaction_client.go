package grpc_client

import (
	"context"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/frontend_api/internal/config"
	"micro-mart/services/frontend_api/internal/infrastructure/grpc_client/transaction"

	"go.uber.org/fx"
	"google.golang.org/grpc"
)

// NewTransactionClient creates a new TransactionClient
func NewTransactionClient(lc fx.Lifecycle, cfg *config.Config) *transaction.TransactionClient {
	// Create a connection to the transaction orchestrator service
	conn, err := mmotel.DialWithTracing(
		context.Background(),
		cfg.TransactionOrchestratorAddress,
		grpc.WithInsecure(),
	)
	if err != nil {
		mmotel.Error(nil, "Failed to connect to transaction orchestrator service: "+err.Error())
		// Return a mock client as fallback
		return transaction.NewTransactionClient(nil)
	}

	// Create a client with the connection
	client := transaction.NewTransactionClient(conn)

	// Register lifecycle hooks
	lc.Append(fx.Hook{
		OnStop: func(ctx context.Context) error {
			mmotel.Info(ctx, "Closing connection to transaction orchestrator service")
			return conn.Close()
		},
	})

	// Log the creation
	mmotel.Info(nil, "Created TransactionClient connected to "+cfg.TransactionOrchestratorAddress)

	return client
}

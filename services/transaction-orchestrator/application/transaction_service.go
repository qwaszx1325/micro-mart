package application

import (
	"context"

	"micro-mart/pkg/mmotel"
	"micro-mart/services/transaction-orchestrator/domain/repository"
	"micro-mart/services/transaction-orchestrator/domain/service"
)

// TransactionServiceImpl implements the application layer transaction service
type TransactionServiceImpl struct {
	transactionSvc *service.TransactionService
}

// NewTransactionService creates a new TransactionServiceImpl
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	userClient service.UserServiceClient,
	authClient service.AuthServiceClient,
) *TransactionServiceImpl {
	// Create domain service
	domainService := service.NewTransactionService(
		transactionRepo,
		userClient,
		authClient,
	)

	return &TransactionServiceImpl{
		transactionSvc: domainService,
	}
}

// RegisterUser handles user registration with transaction orchestration
func (s *TransactionServiceImpl) RegisterUser(ctx context.Context, req *service.RegisterRequest) (*service.RegisterUserResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "RegisterUser")
	defer span.End()

	// Delegate to domain service
	return s.transactionSvc.RegisterUserTransaction(ctx, req)
}

// GetTransactionService returns the domain transaction service
func (s *TransactionServiceImpl) GetTransactionService() *service.TransactionService {
	return s.transactionSvc
}
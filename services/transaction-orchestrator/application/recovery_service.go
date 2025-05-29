package application

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"micro-mart/pkg/mmotel"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/domain/model"
	"micro-mart/services/transaction-orchestrator/domain/repository"
	"micro-mart/services/transaction-orchestrator/domain/service"
)

// RecoveryService handles recovery of pending transactions
type RecoveryService struct {
	transactionRepo repository.TransactionRepository
	transactionSvc  *service.TransactionService
	config          *config.Config
}

// NewRecoveryService creates a new RecoveryService
func NewRecoveryService(
	transactionRepo repository.TransactionRepository,
	transactionSvc *service.TransactionService,
	config *config.Config,
) *RecoveryService {
	return &RecoveryService{
		transactionRepo: transactionRepo,
		transactionSvc:  transactionSvc,
		config:          config,
	}
}

// StartRecoveryProcess starts the recovery process in a separate goroutine
func (s *RecoveryService) StartRecoveryProcess(ctx context.Context) {
	go func() {
		for {
			// Run recovery process
			err := s.RecoverPendingTransactions(ctx)
			if err != nil {
				log.Printf("Error in recovery process: %v", err)
			}

			// Sleep for the configured interval
			time.Sleep(time.Duration(s.config.RecoveryInterval) * time.Second)
		}
	}()
}

// RecoverPendingTransactions recovers pending transactions
func (s *RecoveryService) RecoverPendingTransactions(ctx context.Context) error {
	ctx, span := mmotel.StartSpan(ctx, "RecoverPendingTransactions")
	defer span.End()

	// Find transactions that need recovery
	statuses := []model.TransactionStatus{
		model.StatusPending,
		model.StatusRollingBack,
	}

	transactions, err := s.transactionRepo.FindTransactionsByStatus(ctx, statuses)
	if err != nil {
		mmotel.Error(ctx, "Failed to find pending transactions: "+err.Error())
		return err
	}

	log.Printf("Found %d transactions to recover", len(transactions))

	// Process each transaction
	for _, tx := range transactions {
		err := s.recoverTransaction(ctx, tx)
		if err != nil {
			mmotel.Error(ctx, "Failed to recover transaction: "+err.Error())
			// Continue with other transactions even if one fails
		}
	}

	return nil
}

// recoverTransaction recovers a single transaction
func (s *RecoveryService) recoverTransaction(ctx context.Context, tx *model.Transaction) error {
	ctx, span := mmotel.StartSpan(ctx, "recoverTransaction")
	defer span.End()

	log.Printf("Recovering transaction %s of type %s with status %s", tx.ID, tx.Type, tx.Status)

	// Handle different transaction types
	switch tx.Type {
	case "USER_REGISTRATION":
		return s.recoverUserRegistration(ctx, tx)
	// Add other transaction types as needed
	default:
		log.Printf("Unknown transaction type: %s", tx.Type)
		return nil
	}
}

// recoverUserRegistration recovers a user registration transaction
func (s *RecoveryService) recoverUserRegistration(ctx context.Context, tx *model.Transaction) error {
	ctx, span := mmotel.StartSpan(ctx, "recoverUserRegistration")
	defer span.End()

	// Check transaction status
	switch tx.Status {
	case model.StatusPending:
		// Check which steps are completed
		if len(tx.Steps) < 2 {
			// Invalid transaction, mark as failed
			tx.Status = model.StatusFailed
			tx.UpdatedAt = time.Now()
			return s.transactionRepo.UpdateTransaction(ctx, tx)
		}

		// Check first step (user registration)
		if tx.Steps[0].Status == "PENDING" {
			// First step not completed, can't do much, mark as failed
			tx.Status = model.StatusFailed
			tx.UpdatedAt = time.Now()
			return s.transactionRepo.UpdateTransaction(ctx, tx)
		}

		// Check second step (token generation)
		if tx.Steps[0].Status == "COMPLETED" && tx.Steps[1].Status == "PENDING" {
			// First step completed, second step pending
			// Extract user info from first step result
			var userResp service.RegisterResponse
			err := json.Unmarshal(tx.Steps[0].Result, &userResp)
			if err != nil {
				mmotel.Error(ctx, "Failed to unmarshal user response: "+err.Error())
				return err
			}

			// Retry token generation
			authResp, err := s.transactionSvc.GetAuthClient().GenerateTokens(
				ctx, userResp.UserId, userResp.Username, userResp.Email, userResp.Role)
			if err != nil {
				// Still failing, mark transaction for rollback
				tx.Status = model.StatusRollingBack
				tx.Steps[1].Status = "FAILED"
				tx.UpdatedAt = time.Now()
				tx.Steps[1].UpdatedAt = time.Now()
				return s.transactionRepo.UpdateTransaction(ctx, tx)
			}

			// Token generation succeeded
			tx.Steps[1].Status = "COMPLETED"
			tx.Steps[1].Result = serializeToJSON(authResp)
			tx.Status = model.StatusCompleted
			tx.UpdatedAt = time.Now()
			tx.Steps[1].UpdatedAt = time.Now()
			return s.transactionRepo.UpdateTransaction(ctx, tx)
		}

	case model.StatusRollingBack:
		// Transaction is being rolled back
		// Check if user was created
		if tx.Steps[0].Status == "COMPLETED" {
			// Extract user ID from first step result
			var userResp service.RegisterResponse
			err := json.Unmarshal(tx.Steps[0].Result, &userResp)
			if err != nil {
				mmotel.Error(ctx, "Failed to unmarshal user response: "+err.Error())
				return err
			}

			// Delete user
			err = s.transactionSvc.GetUserClient().DeleteUser(ctx, userResp.UserId)
			if err != nil {
				mmotel.Error(ctx, "Failed to delete user during rollback: "+err.Error())
				// Continue anyway, mark as rolled back
			}
		}

		// Mark transaction as rolled back
		tx.Status = model.StatusRolledBack
		tx.UpdatedAt = time.Now()
		return s.transactionRepo.UpdateTransaction(ctx, tx)
	}

	return nil
}

// Helper function to serialize data to JSON
func serializeToJSON(data interface{}) []byte {
	bytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}
	}
	return bytes
}

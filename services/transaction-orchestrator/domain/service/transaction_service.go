package service

import (
	"context"
	"encoding/json"
	"time"

	"github.com/google/uuid"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/transaction-orchestrator/domain/model"
	"micro-mart/services/transaction-orchestrator/domain/repository"
)

// TransactionService orchestrates distributed transactions
type TransactionService struct {
	transactionRepo repository.TransactionRepository
	userClient      UserServiceClient
	authClient      AuthServiceClient
	// Add other service clients as needed
}

// GetUserClient returns the user service client
func (s *TransactionService) GetUserClient() UserServiceClient {
	return s.userClient
}

// GetAuthClient returns the auth service client
func (s *TransactionService) GetAuthClient() AuthServiceClient {
	return s.authClient
}

// NewTransactionService creates a new TransactionService
func NewTransactionService(
	transactionRepo repository.TransactionRepository,
	userClient UserServiceClient,
	authClient AuthServiceClient,
) *TransactionService {
	return &TransactionService{
		transactionRepo: transactionRepo,
		userClient:      userClient,
		authClient:      authClient,
	}
}

// UserServiceClient is the interface for the user service
type UserServiceClient interface {
	Register(ctx context.Context, req *RegisterRequest) (*RegisterResponse, error)
	DeleteUser(ctx context.Context, userID string) error
}

// AuthServiceClient is the interface for the auth service
type AuthServiceClient interface {
	GenerateTokens(ctx context.Context, userID, username, email, role string) (*AuthTokenResponse, error)
	RevokeTokens(ctx context.Context, userID string) error
}

// RegisterRequest represents a user registration request
type RegisterRequest struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterResponse represents a user registration response
type RegisterResponse struct {
	UserId   string `json:"user_id"`
	Username string `json:"username"`
	Email    string `json:"email"`
	Role     string `json:"role"`
}

// AuthTokenResponse represents an auth token response
type AuthTokenResponse struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

// RegisterUserTransaction orchestrates the user registration transaction
func (s *TransactionService) RegisterUserTransaction(ctx context.Context, req *RegisterRequest) (*RegisterUserResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "RegisterUserTransaction")
	defer span.End()

	// Generate a unique transaction ID
	txID := uuid.New().String()

	// Create transaction record
	tx := &model.Transaction{
		ID:     txID,
		Type:   "USER_REGISTRATION",
		Status: model.StatusPending,
		Steps: []model.TransactionStep{
			{
				ID:          uuid.New().String(),
				TransactionID: txID,
				ServiceName: "user",
				Operation:   "Register",
				Status:      "PENDING",
				Payload:     serializeToJSON(req),
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
			{
				ID:          uuid.New().String(),
				TransactionID: txID,
				ServiceName: "auth",
				Operation:   "GenerateTokens",
				Status:      "PENDING",
				CreatedAt:   time.Now(),
				UpdatedAt:   time.Now(),
			},
		},
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}

	// Save transaction record
	err := s.transactionRepo.SaveTransaction(ctx, tx)
	if err != nil {
		mmotel.Error(ctx, "Failed to save transaction: "+err.Error())
		return nil, err
	}

	// Execute first step: user registration
	userResp, err := s.userClient.Register(ctx, req)
	if err != nil {
		// Update transaction status to failed
		tx.Status = model.StatusFailed
		tx.Steps[0].Status = "FAILED"
		tx.UpdatedAt = time.Now()
		tx.Steps[0].UpdatedAt = time.Now()
		s.transactionRepo.UpdateTransaction(ctx, tx)

		mmotel.Error(ctx, "User registration failed: "+err.Error())
		return nil, err
	}

	// Update first step status
	tx.Steps[0].Status = "COMPLETED"
	tx.Steps[0].Result = serializeToJSON(userResp)
	tx.Steps[0].UpdatedAt = time.Now()

	// Prepare payload for second step
	authReq := struct {
		UserID   string `json:"user_id"`
		Username string `json:"username"`
		Email    string `json:"email"`
		Role     string `json:"role"`
	}{
		UserID:   userResp.UserId,
		Username: userResp.Username,
		Email:    userResp.Email,
		Role:     userResp.Role,
	}

	tx.Steps[1].Payload = serializeToJSON(authReq)
	tx.Steps[1].UpdatedAt = time.Now()
	tx.UpdatedAt = time.Now()

	err = s.transactionRepo.UpdateTransaction(ctx, tx)
	if err != nil {
		mmotel.Error(ctx, "Failed to update transaction: "+err.Error())
	}

	// Execute second step: generate tokens
	authResp, err := s.authClient.GenerateTokens(ctx, userResp.UserId, userResp.Username, userResp.Email, userResp.Role)
	if err != nil {
		// Update transaction status to rolling back
		tx.Status = model.StatusRollingBack
		tx.Steps[1].Status = "FAILED"
		tx.UpdatedAt = time.Now()
		tx.Steps[1].UpdatedAt = time.Now()

		err = s.transactionRepo.UpdateTransaction(ctx, tx)
		if err != nil {
			mmotel.Error(ctx, "Failed to update transaction: "+err.Error())
		}

		// Execute compensation operation: delete user
		compensationErr := s.userClient.DeleteUser(ctx, userResp.UserId)
		if compensationErr != nil {
			// Record compensation operation failure
			mmotel.Error(ctx, "Compensation failed: "+compensationErr.Error())
		}

		// Update transaction status to rolled back
		tx.Status = model.StatusRolledBack
		tx.UpdatedAt = time.Now()

		err = s.transactionRepo.UpdateTransaction(ctx, tx)
		if err != nil {
			mmotel.Error(ctx, "Failed to update transaction: "+err.Error())
		}

		return nil, err
	}

	// Update second step status
	tx.Steps[1].Status = "COMPLETED"
	tx.Steps[1].Result = serializeToJSON(authResp)
	tx.Status = model.StatusCompleted
	tx.UpdatedAt = time.Now()
	tx.Steps[1].UpdatedAt = time.Now()

	err = s.transactionRepo.UpdateTransaction(ctx, tx)
	if err != nil {
		mmotel.Error(ctx, "Failed to update transaction: "+err.Error())
	}

	// Return complete result
	return &RegisterUserResponse{
		UserInfo: userResp,
		AuthInfo: authResp,
	}, nil
}

// RegisterUserResponse combines the responses from user and auth services
type RegisterUserResponse struct {
	UserInfo *RegisterResponse  `json:"user_info"`
	AuthInfo *AuthTokenResponse `json:"auth_info"`
}

// Helper function to serialize data to JSON
func serializeToJSON(data interface{}) []byte {
	bytes, err := json.Marshal(data)
	if err != nil {
		return []byte{}
	}
	return bytes
}

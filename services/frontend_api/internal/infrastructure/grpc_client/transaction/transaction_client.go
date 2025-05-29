package transaction

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	pb "micro-mart/pkg/pb/gen/transaction"
	"micro-mart/services/frontend_api/internal/model/request"

	"google.golang.org/grpc"
)

// TransactionClient provides methods to interact with the transaction orchestrator service
type TransactionClient struct {
	conn *grpc.ClientConn
}

// NewTransactionClient creates a new TransactionClient
func NewTransactionClient(conn *grpc.ClientConn) *TransactionClient {
	return &TransactionClient{
		conn: conn,
	}
}

// RegisterUserResponse represents the response from the transaction orchestrator
type RegisterUserResponse struct {
	UserID       string
	Username     string
	Email        string
	Role         string
	AccessToken  string
	RefreshToken string
}

// RegisterUserTransaction orchestrates the user registration transaction
func (c *TransactionClient) RegisterUserTransaction(ctx context.Context, req *request.RegisterRequest) (*RegisterUserResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "TransactionClient.RegisterUserTransaction")
	defer span.End()

	// If the connection is nil, return a mock response
	if c.conn == nil {
		mmotel.Warn(ctx, "Using mock implementation of RegisterUserTransaction because connection is nil")
		return &RegisterUserResponse{
			UserID:       "mock-user-id",
			Username:     req.UserName,
			Email:        req.Email,
			Role:         "user",
			AccessToken:  "mock-access-token",
			RefreshToken: "mock-refresh-token",
		}, nil
	}

	// Create a client for the transaction service
	client := pb.NewTransactionServiceClient(c.conn)

	// Convert the request to the protobuf format
	pbReq := &pb.RegisterUserRequest{
		Username: req.UserName,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call the transaction orchestrator service
	pbResp, err := client.RegisterUser(ctx, pbReq)
	if err != nil {
		mmotel.Error(ctx, "Failed to call transaction orchestrator service: "+err.Error())
		return nil, mmerror.New(mmerror.InternalServerError, "Failed to register user", err)
	}

	// Convert the response to our format
	return &RegisterUserResponse{
		UserID:       pbResp.UserInfo.UserId,
		Username:     pbResp.UserInfo.Username,
		Email:        pbResp.UserInfo.Email,
		Role:         pbResp.UserInfo.Role,
		AccessToken:  pbResp.AuthInfo.AccessToken,
		RefreshToken: pbResp.AuthInfo.RefreshToken,
	}, nil
}

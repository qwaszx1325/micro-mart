package grpc_impl

import (
	"context"

	"google.golang.org/grpc"
	"micro-mart/pkg/mmotel"
	pb "micro-mart/pkg/pb/gen/transaction"
	"micro-mart/services/transaction-orchestrator/application"
	"micro-mart/services/transaction-orchestrator/domain/service"
)

// TransactionServer implements the TransactionService gRPC server
type TransactionServer struct {
	pb.UnimplementedTransactionServiceServer
	transactionService *application.TransactionServiceImpl
}

// NewTransactionServer creates a new TransactionServer
func NewTransactionServer(transactionService *application.TransactionServiceImpl) *TransactionServer {
	return &TransactionServer{
		transactionService: transactionService,
	}
}

// RegisterServer registers the TransactionServer with the gRPC server
func RegisterServer(server *grpc.Server, transactionService *application.TransactionServiceImpl) {
	pb.RegisterTransactionServiceServer(server, NewTransactionServer(transactionService))
}

// RegisterUser implements the RegisterUser RPC method
func (s *TransactionServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "TransactionServer.RegisterUser")
	defer span.End()

	// Convert protobuf request to domain request
	domainReq := &service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call application service
	result, err := s.transactionService.RegisterUser(ctx, domainReq)
	if err != nil {
		mmotel.Error(ctx, "Failed to register user: "+err.Error())
		return nil, err
	}

	// Convert domain response to protobuf response
	return &pb.RegisterUserResponse{
		UserInfo: &pb.UserInfo{
			UserId:   result.UserInfo.UserId,
			Username: result.UserInfo.Username,
			Email:    result.UserInfo.Email,
			Role:     result.UserInfo.Role,
		},
		AuthInfo: &pb.AuthInfo{
			AccessToken:  result.AuthInfo.AccessToken,
			RefreshToken: result.AuthInfo.RefreshToken,
		},
	}, nil
}
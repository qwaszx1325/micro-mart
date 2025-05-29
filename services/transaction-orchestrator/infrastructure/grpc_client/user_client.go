package grpc_client

import (
	"context"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/user"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/domain/service"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// UserClient implements the UserServiceClient interface
type UserClient struct {
	conn           *grpc.ClientConn
	userGrpcClient user.UserServiceClient
}

// Ensure UserClient implements the UserServiceClient interface
var _ service.UserServiceClient = (*UserClient)(nil)

// NewUserClient creates a new UserClient
func NewUserClient(cfg *config.Config) (*UserClient, error) {
	// Get address from config
	grpcAddr := cfg.UserUrl

	// New grpc client with tracing middleware
	conn, err := grpc.Dial(
		grpcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)),
		)),
	)
	if err != nil {
		return nil, err
	}
	userGrpc := user.NewUserServiceClient(conn)

	return &UserClient{
		conn:           conn,
		userGrpcClient: userGrpc,
	}, nil
}

// Register implements the Register method of the UserServiceClient interface
func (c *UserClient) Register(ctx context.Context, req *service.RegisterRequest) (*service.RegisterResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "UserClient.Register")
	defer span.End()

	// Convert domain request to gRPC request
	grpcReq := &user.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call the gRPC service
	response, err := c.userGrpcClient.Register(ctx, grpcReq)
	if err != nil {
		mmotel.Error(ctx, "Failed to register user", mmotel.NewField("error", err))
		return nil, err
	}

	// Convert gRPC response to domain response
	return &service.RegisterResponse{
		UserId:   response.UserId,
		Username: response.Username,
		Email:    response.Email,
		Role:     response.Role,
	}, nil
}

// DeleteUser implements the DeleteUser method of the UserServiceClient interface
func (c *UserClient) DeleteUser(ctx context.Context, userID string) error {
	ctx, span := mmotel.StartSpan(ctx, "UserClient.DeleteUser")
	defer span.End()

	// This is a placeholder. In a real implementation, you would call the DeleteUser method
	// of the user service. Since it's not implemented in the example, we'll just log it.
	mmotel.Info(ctx, "DeleteUser called", mmotel.NewField("userID", userID))
	
	// In a real implementation, you would do something like:
	// _, err := c.userGrpcClient.DeleteUser(ctx, &user.DeleteUserRequest{UserId: userID})
	// return err
	
	return nil
}
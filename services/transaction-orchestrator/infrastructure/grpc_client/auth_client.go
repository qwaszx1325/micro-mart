package grpc_client

import (
	"context"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/auth"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/domain/service"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// AuthClient implements the AuthServiceClient interface
type AuthClient struct {
	conn           *grpc.ClientConn
	authGrpcClient auth.AuthServiceClient
}

// Ensure AuthClient implements the AuthServiceClient interface
var _ service.AuthServiceClient = (*AuthClient)(nil)

// NewAuthClient creates a new AuthClient
func NewAuthClient(cfg *config.Config) (*AuthClient, error) {
	// Get address from config
	grpcAddr := cfg.AuthUrl

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
	authGrpc := auth.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:           conn,
		authGrpcClient: authGrpc,
	}, nil
}

// GenerateTokens implements the GenerateTokens method of the AuthServiceClient interface
func (c *AuthClient) GenerateTokens(ctx context.Context, userID, username, email, role string) (*service.AuthTokenResponse, error) {
	ctx, span := mmotel.StartSpan(ctx, "AuthClient.GenerateTokens")
	defer span.End()

	// Convert domain request to gRPC request
	grpcReq := &auth.GenerateTokenRequest{
		UserId:   userID,
		Username: username,
		Email:    email,
		Role:     role,
	}

	// Call the gRPC service
	response, err := c.authGrpcClient.GenerateTokens(ctx, grpcReq)
	if err != nil {
		mmotel.Error(ctx, "Failed to generate tokens", mmotel.NewField("error", err))
		return nil, err
	}

	// Convert gRPC response to domain response
	return &service.AuthTokenResponse{
		AccessToken:  response.AccessToken,
		RefreshToken: response.RefreshToken,
	}, nil
}

// RevokeTokens implements the RevokeTokens method of the AuthServiceClient interface
func (c *AuthClient) RevokeTokens(ctx context.Context, userID string) error {
	ctx, span := mmotel.StartSpan(ctx, "AuthClient.RevokeTokens")
	defer span.End()

	// This is a placeholder. In a real implementation, you would call the RevokeTokens method
	// of the auth service. Since it's not implemented in the example, we'll just log it.
	mmotel.Info(ctx, "RevokeTokens called", mmotel.NewField("userID", userID))
	
	// In a real implementation, you would do something like:
	// _, err := c.authGrpcClient.RevokeTokens(ctx, &auth.RevokeTokensRequest{UserId: userID})
	// return err
	
	return nil
}
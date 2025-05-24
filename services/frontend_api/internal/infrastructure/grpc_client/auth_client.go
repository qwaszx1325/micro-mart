package grpc_client

import (
	"context"
	"fmt"
	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/auth"
	"micro-mart/services/frontend_api/internal/config"
)

type AuthClient struct {
	conn           *grpc.ClientConn
	authGrpcClient auth.AuthServiceClient
}

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
		return &AuthClient{}, err
	}
	authGrpc := auth.NewAuthServiceClient(conn)

	return &AuthClient{
		conn:           conn,
		authGrpcClient: authGrpc,
	}, nil
}

// 取得access token 和 refresh token
func (c *AuthClient) GenerateTokens(ctx context.Context, userId string, username string, email string, role string) (*auth.AuthTokenResponse, *mmerror.MmError) {

	grpcReq := &auth.GenerateTokenRequest{
		UserId:   userId,
		Username: username,
		Email:    email,
		Role:     role,
	}

	fmt.Println("hi")

	response, grpcErr := c.authGrpcClient.GenerateTokens(ctx, grpcReq)
	if grpcErr != nil {
		errConverted, ok := mmerror.FromGrpcErr(grpcErr)
		if ok {
			mmotel.Error(ctx, errConverted.Error())
			return nil, errConverted
		}
		err := mmerror.New(mmerror.InternalServerError, "can't found the mmErr from grpcErr", grpcErr)
		mmotel.Error(ctx, err.Error())
		return nil, err
	}
	return response, nil

}

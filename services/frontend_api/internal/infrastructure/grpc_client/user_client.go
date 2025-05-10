package grpc_client

import (
	"context"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/user"
	"micro-mart/services/frontend_api/internal/model/request"
	"micro-mart/services/user/config"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type UserClient struct {
	conn           *grpc.ClientConn
	userGrpcClient user.UserServiceClient
}

func NewUserClient(cfg *config.Config) (*UserClient, error) {
	// Get address from config
	gprcAddr := cfg.UserUrl

	// New grpc client with tracing middleware
	conn, err := grpc.Dial(
		gprcAddr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithStatsHandler(otelgrpc.NewClientHandler(
			otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(
				propagation.TraceContext{},
				propagation.Baggage{},
			)),
		)),
	)
	if err != nil {
		return &UserClient{}, err
	}
	userGrpc := user.NewUserServiceClient(conn)

	return &UserClient{
		conn:           conn,
		userGrpcClient: userGrpc,
	}, nil
}

// Register 用戶註冊
func (c *UserClient) Register(ctx context.Context, req *request.RegisterRequest) (*user.AuthResponse, error) {
	// 使用 mmotel 包裝 span
	var response *user.AuthResponse
	var err error

	err = mmotel.WithSpan(ctx, "UserClient.Register", func(ctx context.Context) error {
		// 將前端請求轉換為 gRPC 請求
		grpcReq := &user.RegisterRequest{
			Username: req.UserName,
			Email:    req.Email,
			Password: req.Password,
		}

		// 調用 gRPC 服務
		response, err = c.userGrpcClient.Register(ctx, grpcReq)
		return err
	})

	return response, err
}

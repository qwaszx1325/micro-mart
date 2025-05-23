package grpc_client

import (
	"context"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/pb/gen/user"
	"micro-mart/services/frontend_api/internal/config"
	"micro-mart/services/frontend_api/internal/model/request"

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
func (c *UserClient) Register(ctx context.Context, req *request.RegisterRequest) (*user.AuthResponse, *mmerror.MmError) {

	grpcReq := &user.RegisterRequest{
		Username: req.UserName,
		Email:    req.Email,
		Password: req.Password,
	}
	// create user as auth identity
	response, grpcErr := c.userGrpcClient.Register(ctx, grpcReq)

	ctx, span := mmotel.StartSpan(ctx, "Register")
	defer span.End()
	if grpcErr != nil {
		errConverted, ok := mmerror.FromGrpcErr(grpcErr)
		if ok {
			mmotel.Error(ctx, errConverted.Error())
			return nil, errConverted
		}
		err := mmerror.New(mmerror.InternalServerError, "can't found the kgsErr from grpcErr", grpcErr)
		mmotel.Error(ctx, err.Error())
		return nil, err
	}

	return response, nil
}

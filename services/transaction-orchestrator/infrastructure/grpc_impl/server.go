package grpc_impl

import (
	"context"
	"fmt"
	"log"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"micro-mart/pkg/mmotel"
	pb "micro-mart/pkg/pb/gen/transaction"
	"micro-mart/services/transaction-orchestrator/application"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/domain/service"
)

// TransactionServer implements the TransactionService gRPC server
type TransactionServer struct {
	pb.UnimplementedTransactionServiceServer
	transactionService *application.TransactionServiceImpl
}

// RegisterUser implements the RegisterUser gRPC method
func (s *TransactionServer) RegisterUser(ctx context.Context, req *pb.RegisterUserRequest) (*pb.RegisterUserResponse, error) {
	// Convert gRPC request to domain request
	domainReq := &service.RegisterRequest{
		Username: req.Username,
		Email:    req.Email,
		Password: req.Password,
	}

	// Call application service
	result, err := s.transactionService.RegisterUser(ctx, domainReq)
	if err != nil {
		return nil, err
	}

	// Convert domain response to gRPC response
	response := &pb.RegisterUserResponse{
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
	}

	return response, nil
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

// NewGrpcServer creates a new gRPC server with the transaction service registered
func NewGrpcServer(lc fx.Lifecycle, transactionService *application.TransactionServiceImpl, recoveryService *application.RecoveryService) *grpc.Server {
	cfg := config.GetConfig()

	// 提早初始化 Tracer
	shutdown := mmotel.InitTracer(cfg.ServiceName,
		mmotel.WithJaegerExporter(cfg.OtelUrl),
		mmotel.WithSamplingRatio(1.0),
	)

	// 建立 gRPC server，此時 Tracer 已初始化完成
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mmotel.UnaryTraceInterceptor(),
			mmotel.ErrorLoggingInterceptor(),
		),
		grpc.ChainStreamInterceptor(
			mmotel.StreamTraceInterceptor(),
			mmotel.StreamErrorLoggingInterceptor(),
		),
	)

	// lifecycle 管理
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 註冊 gRPC 服務
			RegisterServer(s, transactionService)

			// 監聽
			lis, err := net.Listen("tcp", cfg.ServiceUrl)
			if err != nil {
				log.Fatalf("監聽失敗: %v", err)
				return err
			}

			go func() {
				ctx := context.WithValue(context.Background(), "service_name", cfg.Host.ServiceName)
				if err := s.Serve(lis); err != nil {
					mmotel.Error(ctx, "服務啟動失敗", mmotel.NewField("error", err))
				}
			}()

			// 啟動恢復服務
			go recoveryService.StartRecoveryProcess(ctx)

			mmotel.Info(ctx, fmt.Sprintf("gRPC服務已啟動，監聽於 %s", cfg.ServiceUrl))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			s.GracefulStop()
			shutdown()
			mmotel.Info(ctx, "gRPC服務已優雅停止")
			return nil
		},
	})

	return s
}

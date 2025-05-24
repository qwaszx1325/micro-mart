package grpc_impl

import (
	"context"
	"fmt"
	"log"
	"micro-mart/pkg/pb/gen/auth"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"micro-mart/pkg/mmotel" // 使用您的 mmotel 包
	"micro-mart/services/auth/application"
	"micro-mart/services/auth/config"
)

func NewGrpcServer(lc fx.Lifecycle, authService *application.AuthService) *grpc.Server {
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
			// 監聽
			lis, err := net.Listen("tcp", cfg.ServiceUrl)
			if err != nil {
				log.Fatalf("監聽失敗: %v", err)
				return err
			}

			auth.RegisterAuthServiceServer(s, authService)

			go func() {
				ctx := context.WithValue(context.Background(), "service_name", cfg.Host.ServiceName)
				if err := s.Serve(lis); err != nil {
					mmotel.Error(ctx, "服務啟動失敗", mmotel.NewField("error", err))
				}
			}()

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

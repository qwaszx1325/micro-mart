package grpc_impl

import (
	"context"
	"fmt"
	"log"
	"net"

	"go.uber.org/fx"
	"google.golang.org/grpc"

	"micro-mart/pkg/mmotel" // 使用您的 mmotel 包
	"micro-mart/pkg/pb/gen/user"
	"micro-mart/services/user/application"
	"micro-mart/services/user/config"
)

func NewGrpcServer(lc fx.Lifecycle, userService *application.UserService) *grpc.Server {
	// 獲取配置
	cfg := config.GetConfig()

	// 創建新的grpc伺服器
	s := grpc.NewServer(
		grpc.ChainUnaryInterceptor(
			mmotel.UnaryTraceInterceptor(),
			mmotel.ErrorLoggingInterceptor(), // 添加錯誤日誌攔截器
		),
		grpc.ChainStreamInterceptor(
			mmotel.StreamTraceInterceptor(),
			mmotel.StreamErrorLoggingInterceptor(), // 添加流式錯誤日誌攔截器
		),
	)

	var shutdown func()
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// 初始化追蹤器
			shutdown = mmotel.InitTracer(cfg.ServiceName,
				mmotel.WithJaegerExporter(cfg.OtelUrl), // 假設有 Jaeger 端點配置
			)

			// 監聽端口
			lis, err := net.Listen("tcp", cfg.ServiceUrl)
			if err != nil {
				log.Fatalf("監聽失敗: %v", err)
				return err
			}

			// 註冊服務
			user.RegisterUserServiceServer(s, userService)

			// 啟動服務
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
			// 優雅停止服務
			s.GracefulStop()

			// 關閉追蹤器
			if shutdown != nil {
				shutdown()
			}

			mmotel.Info(ctx, "gRPC服務已優雅停止")
			return nil
		},
	})

	return s
}

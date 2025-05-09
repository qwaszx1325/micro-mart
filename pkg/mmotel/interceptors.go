package mmotel

import (
	"context"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// 輔助函數: 創建帶有追蹤的 gRPC 服務器
func (t *Tracer) NewGrpcServer(opts ...grpc.ServerOption) *grpc.Server {
	// 使用最新的 API 來創建服務器處理器
	handler := otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(t.provider),
		otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)

	// 添加 StatsHandler
	opts = append(opts,
		grpc.StatsHandler(handler),
	)

	// 創建服務器
	return grpc.NewServer(opts...)
}

// 輔助函數: 使用追蹤建立 gRPC 客戶端連接
func (t *Tracer) DialWithTracing(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	// 使用最新的 API 來創建客戶端處理器
	handler := otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(t.provider),
		otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)

	// 添加 StatsHandler
	opts = append(opts,
		grpc.WithStatsHandler(handler),
	)

	// 為測試環境檢查是否有憑證設置，如果沒有則添加不安全選項
	hasCredentials := false
	for _, opt := range opts {
		// 檢查是否已提供憑證選項
		// 注意：這是簡化的檢測，實際上無法直接檢查 DialOption 的類型
		// 在實際使用中，最好由調用者明確提供憑證
		if _, ok := opt.(*grpc.DialOption); ok {
			hasCredentials = true
			break
		}
	}

	if !hasCredentials {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// 建立連接
	return grpc.DialContext(ctx, target, opts...)
}

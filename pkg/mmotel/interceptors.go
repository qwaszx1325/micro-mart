// pkg/mmotel/interceptors.go
package mmotel

import (
	"context"
	"log"

	"go.opentelemetry.io/contrib/instrumentation/google.golang.org/grpc/otelgrpc"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/status"
)

// UnaryTraceInterceptor 返回用於追蹤 gRPC 一元調用的攔截器
func UnaryTraceInterceptor() grpc.UnaryServerInterceptor {
	if globalTracer == nil {
		log.Println("警告: 追蹤器未初始化，使用 NoOp 攔截器")
		return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
			return handler(ctx, req)
		}
	}

	// 這是一個包裝攔截器，它會調用 OTel 的 StatsHandler，然後進行額外的處理
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 創建一個子 span
		newCtx, span := StartSpan(ctx, info.FullMethod)
		defer span.End()

		// 調用原始處理器
		resp, err := handler(newCtx, req)

		// 處理錯誤
		if err != nil {
			span.RecordError(err)
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return resp, err
	}
}

// StreamTraceInterceptor 返回用於追蹤 gRPC 流調用的攔截器
func StreamTraceInterceptor() grpc.StreamServerInterceptor {
	if globalTracer == nil {
		log.Println("警告: 追蹤器未初始化，使用 NoOp 攔截器")
		return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
			return handler(srv, ss)
		}
	}

	// 這是一個包裝攔截器
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 創建一個子 span
		ctx := ss.Context()
		newCtx, span := StartSpan(ctx, info.FullMethod)
		defer span.End()

		// 創建包裝的 ServerStream 來傳遞新的 context
		wrappedStream := &wrappedServerStream{
			ServerStream: ss,
			ctx:          newCtx,
		}

		// 調用原始處理器
		err := handler(srv, wrappedStream)

		// 處理錯誤
		if err != nil {
			span.RecordError(err)
			s, _ := status.FromError(err)
			span.SetStatus(codes.Error, s.Message())
		} else {
			span.SetStatus(codes.Ok, "")
		}

		return err
	}
}

// wrappedServerStream 包裝 grpc.ServerStream 以使用自定義 context
type wrappedServerStream struct {
	grpc.ServerStream
	ctx context.Context
}

// Context 返回包裝的 context
func (w *wrappedServerStream) Context() context.Context {
	return w.ctx
}

// ErrorLoggingInterceptor 返回一個 gRPC 一元攔截器，用於記錄錯誤
func ErrorLoggingInterceptor() grpc.UnaryServerInterceptor {
	return func(ctx context.Context, req interface{}, info *grpc.UnaryServerInfo, handler grpc.UnaryHandler) (interface{}, error) {
		// 調用原始處理器
		resp, err := handler(ctx, req)

		// 如果有錯誤，記錄它
		if err != nil {
			Error(ctx, "gRPC 調用錯誤", NewField("method", info.FullMethod), NewField("error", err))
		}

		return resp, err
	}
}

// StreamErrorLoggingInterceptor 返回一個 gRPC 流攔截器，用於記錄錯誤
func StreamErrorLoggingInterceptor() grpc.StreamServerInterceptor {
	return func(srv interface{}, ss grpc.ServerStream, info *grpc.StreamServerInfo, handler grpc.StreamHandler) error {
		// 調用原始處理器
		err := handler(srv, ss)

		// 如果有錯誤，記錄它
		if err != nil {
			Error(ss.Context(), "gRPC 流調用錯誤", NewField("method", info.FullMethod), NewField("error", err))
		}

		return err
	}
}

// DialWithTracing 使用追蹤創建 gRPC 客戶端連接
func DialWithTracing(ctx context.Context, target string, opts ...grpc.DialOption) (*grpc.ClientConn, error) {
	if globalTracer == nil {
		log.Println("警告: 追蹤器未初始化，使用無追蹤連接")
		return grpc.DialContext(ctx, target, opts...)
	}

	// 使用 OpenTelemetry 官方的 gRPC 客戶端處理器
	handler := otelgrpc.NewClientHandler(
		otelgrpc.WithTracerProvider(globalTracer.provider),
		otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)

	// 添加 StatsHandler
	opts = append(opts,
		grpc.WithStatsHandler(handler),
	)

	// 檢查是否已經提供了憑證
	hasCredentials := false
	for _, opt := range opts {
		// 由於無法直接檢查 DialOption 類型，我們簡化處理
		// 在實際使用中，最好明確提供憑證選項
		if opt == grpc.WithTransportCredentials(insecure.NewCredentials()) ||
			opt == grpc.WithInsecure() { // 雖然已棄用，但為兼容性檢查
			hasCredentials = true
			break
		}
	}

	// 如果沒有提供憑證，使用不安全連接（僅用於開發/測試）
	if !hasCredentials {
		opts = append(opts, grpc.WithTransportCredentials(insecure.NewCredentials()))
	}

	// TODO: grpc.DialContext 已被標記為棄用，但目前沒有明確的替代方案
	// 當 gRPC 庫提供明確替代方案時，應更新此代碼
	return grpc.DialContext(ctx, target, opts...)
}

// NewGrpcServer 創建一個帶有追蹤功能的 gRPC 服務器
func NewGrpcServer(opts ...grpc.ServerOption) *grpc.Server {
	if globalTracer == nil {
		log.Println("警告: 追蹤器未初始化，創建無追蹤服務器")
		return grpc.NewServer(opts...)
	}

	// 使用 OpenTelemetry 官方的 gRPC 服務器處理器
	handler := otelgrpc.NewServerHandler(
		otelgrpc.WithTracerProvider(globalTracer.provider),
		otelgrpc.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
	)

	// 添加 StatsHandler
	opts = append(opts, grpc.StatsHandler(handler))

	// 創建服務器
	return grpc.NewServer(opts...)
}

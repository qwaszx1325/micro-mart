// pkg/mmotel/tracer.go
package mmotel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

type Span interface {
	End()
}

type otelSpan struct {
	span trace.Span
}

func (s otelSpan) End() {
	s.span.End()
}

// StartTrace 開始一個新的追蹤 span
func StartTrace(ctx context.Context) (context.Context, Span) {
	// 從上下文中獲取 service name 或使用預設值
	serviceName := ctx.Value("service_name")
	if serviceName == nil {
		serviceName = "micro-mart"
	}

	tracer := otel.Tracer(fmt.Sprintf("%v", serviceName))

	// 從上下文中獲取最後一個函數名稱或使用預設值
	functionName := ctx.Value("function_name")
	if functionName == nil {
		functionName = "unknown"
	}

	ctx, span := tracer.Start(ctx, fmt.Sprintf("%v", functionName))
	return ctx, otelSpan{span}
}

// Error 記錄一個錯誤到當前 span
func Error(ctx context.Context, msg string) {
	span := trace.SpanFromContext(ctx)
	span.SetStatus(codes.Error, msg)
	span.RecordError(fmt.Errorf(msg))
}

// AddAttribute 添加一個屬性到當前 span
func AddAttribute(ctx context.Context, key string, value interface{}) {
	span := trace.SpanFromContext(ctx)

	switch v := value.(type) {
	case string:
		span.SetAttributes(attribute.String(key, v))
	case int:
		span.SetAttributes(attribute.Int(key, v))
	case int64:
		span.SetAttributes(attribute.Int64(key, v))
	case float64:
		span.SetAttributes(attribute.Float64(key, v))
	case bool:
		span.SetAttributes(attribute.Bool(key, v))
	default:
		span.SetAttributes(attribute.String(key, fmt.Sprintf("%v", v)))
	}
}

// InitTracer 初始化 OpenTelemetry 追蹤器並連接到 Jaeger
func InitTracer(serviceName string) func() {
	ctx := context.Background()

	// 使用新的推薦方法建立 OTLP exporter
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("jaeger:4317"), // 使用 Docker 服務名稱
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		fmt.Printf("Failed to create trace exporter: %v\n", err)
		return func() {}
	}

	// 創建一個資源，描述您的服務
	res, err := resource.New(
		ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(serviceName),
			semconv.ServiceVersionKey.String("1.0.0"),
		),
	)
	if err != nil {
		fmt.Printf("Failed to create resource: %v\n", err)
		return func() {}
	}

	// 創建一個批處理 span 處理器
	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	// 創建並設置 tracer provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)

	// 設置全局的傳播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 返回一個清理函數
	return func() {
		// 關閉 tracer provider
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(cctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}
}

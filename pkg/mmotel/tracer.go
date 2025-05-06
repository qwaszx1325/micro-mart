package mmotel

import (
	"context"
	"fmt"
	"runtime"
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
	"go.uber.org/zap"
)

// Field 定義一個鍵值對字段結構
type Field struct {
	Key   string
	Value interface{}
}

// NewField 創建一個新的字段
func NewField(key string, value interface{}) Field {
	return Field{
		Key:   key,
		Value: value,
	}
}

// Span 定義一個追蹤Span接口
type Span interface {
	End()
}

// otelSpan 實現Span接口
type otelSpan struct {
	span trace.Span
}

func (s otelSpan) End() {
	s.span.End()
}

// StartTrace 開始一個新的追蹤Span
func StartTrace(ctx context.Context) (context.Context, Span) {
	// 從上下文中獲取service name或使用預設值
	serviceName := ctx.Value("service_name")
	if serviceName == nil {
		serviceName = "micro-mart"
	}

	tracer := otel.Tracer(fmt.Sprintf("%v", serviceName))

	// 自動獲取調用者信息
	caller, funcName := getCaller(2)

	ctx, span := tracer.Start(ctx, funcName)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()

	// 設置基本屬性
	attributes := []attribute.KeyValue{
		attribute.String("traceID", traceID),
		attribute.String("spanID", spanID),
		attribute.String("caller", caller),
		attribute.String("funcName", funcName),
	}

	span.SetAttributes(attributes...)

	return ctx, otelSpan{span}
}

// Info 記錄信息級別的日誌並添加到追蹤
func Info(ctx context.Context, message string, fields ...Field) {
	span, zapFields := setSpanAttrsAndZapFields(ctx, fields...)
	span.AddEvent(message)
	zap.L().Info(message, zapFields...)
}

// Warn 記錄警告級別的日誌並添加到追蹤
func Warn(ctx context.Context, message string, fields ...Field) {
	span, zapFields := setSpanAttrsAndZapFields(ctx, fields...)
	span.AddEvent(message)
	span.SetStatus(codes.Error, message)
	zap.L().Warn(message, zapFields...)
}

// Error 記錄錯誤級別的日誌並添加到追蹤
func Error(ctx context.Context, message string, fields ...Field) {
	span, zapFields := setSpanAttrsAndZapFields(ctx, fields...)
	span.AddEvent(message)
	span.SetStatus(codes.Error, message)
	span.RecordError(fmt.Errorf(message))
	zap.L().Error(message, zapFields...)
}

// AddAttribute 添加一個屬性到當前span
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

// setSpanAttrsAndZapFields 設置Span屬性並創建Zap字段
func setSpanAttrsAndZapFields(ctx context.Context, fields ...Field) (span trace.Span, zapFields []zap.Field) {
	span = trace.SpanFromContext(ctx)
	traceID := span.SpanContext().TraceID().String()
	spanID := span.SpanContext().SpanID().String()
	caller, funcName := getCaller(3)

	// 創建Span屬性和Zap日誌字段
	attributes := []attribute.KeyValue{
		attribute.String("traceID", traceID),
		attribute.String("spanID", spanID),
		attribute.String("caller", caller),
		attribute.String("funcName", funcName),
	}

	zapFields = []zap.Field{
		zap.String("traceID", traceID),
		zap.String("spanID", spanID),
		zap.String("caller", caller),
		zap.String("funcName", funcName),
	}

	for _, field := range fields {
		attributes = append(attributes, attribute.String(field.Key, fmt.Sprintf("%v", field.Value)))
		zapFields = append(zapFields, zap.Any(field.Key, field.Value))
	}
	span.SetAttributes(attributes...)

	return span, zapFields
}

// getCaller 獲取調用者信息
func getCaller(skip int) (caller string, funcName string) {
	pc, file, line, ok := runtime.Caller(skip)
	if !ok {
		return "unknown", "unknown"
	}
	fn := runtime.FuncForPC(pc)
	return fmt.Sprintf("%s:%d", file, line), fn.Name()
}

// InitTracer 初始化OpenTelemetry追蹤器並連接到Jaeger
func InitTracer(serviceName string) func() {
	ctx := context.Background()

	// 使用OTLP exporter
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint("jaeger:4317"),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		fmt.Printf("Failed to create trace exporter: %v\n", err)
		return func() {}
	}

	// 創建服務資源
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

	// 創建批處理Span處理器
	bsp := sdktrace.NewBatchSpanProcessor(exporter)

	// 創建並設置追蹤提供者
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithSampler(sdktrace.AlwaysSample()),
		sdktrace.WithResource(res),
		sdktrace.WithSpanProcessor(bsp),
	)
	otel.SetTracerProvider(tp)

	// 設置全局傳播器
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 返回清理函數
	return func() {
		cctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if err := tp.Shutdown(cctx); err != nil {
			fmt.Printf("Error shutting down tracer provider: %v\n", err)
		}
	}
}

package mmotel

import (
	"context"
	"fmt"
	"time"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracegrpc"
	"go.opentelemetry.io/otel/exporters/stdout/stdouttrace"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
)

// Tracer 是我們自定義的追蹤器
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	options  *Options
}

// 創建新的追蹤器
func NewTracer(serviceName string, opts ...Option) (*Tracer, error) {
	options := defaultOptions()
	options.ServiceName = serviceName

	// 應用自定義選項
	for _, opt := range opts {
		opt(options)
	}

	// 創建 trace provider
	tp, err := initTracerProvider(options)
	if err != nil {
		return nil, fmt.Errorf("初始化追蹤器失敗: %w", err)
	}

	// 設置全局 trace provider 和 propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// 創建 tracer
	tracer := tp.Tracer(serviceName)

	return &Tracer{
		provider: tp,
		tracer:   tracer,
		options:  options,
	}, nil
}

// 初始化 trace provider
func initTracerProvider(options *Options) (*sdktrace.TracerProvider, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	// 創建資源
	res, err := resource.New(ctx,
		resource.WithAttributes(
			semconv.ServiceNameKey.String(options.ServiceName),
			semconv.ServiceVersionKey.String(options.ServiceVersion),
			attribute.String("environment", options.Environment),
		),
		resource.WithSchemaURL(semconv.SchemaURL),
	)
	if err != nil {
		return nil, fmt.Errorf("創建資源失敗: %w", err)
	}

	// 選擇導出器
	var exporter sdktrace.SpanExporter
	switch options.ExporterType {
	case ExporterTypeJaeger:
		exp, err := createJaegerExporter(ctx, options.JaegerEndpoint)
		if err != nil {
			return nil, err
		}
		exporter = exp
	case ExporterTypeStdout:
		exp, err := stdouttrace.New(
			stdouttrace.WithPrettyPrint(),
		)
		if err != nil {
			return nil, fmt.Errorf("創建標準輸出導出器失敗: %w", err)
		}
		exporter = exp
	default:
		return nil, fmt.Errorf("不支持的導出器類型: %s", options.ExporterType)
	}

	// 創建 trace provider
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(
			sdktrace.ParentBased(
				sdktrace.TraceIDRatioBased(options.SamplingRatio),
			),
		),
	)

	return tp, nil
}

// 創建 Jaeger OTLP 導出器
func createJaegerExporter(ctx context.Context, endpoint string) (sdktrace.SpanExporter, error) {
	if endpoint == "" {
		return nil, fmt.Errorf("Jaeger 端點不能為空")
	}

	// 使用最新的 API
	exporter, err := otlptracegrpc.New(ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
	)
	if err != nil {
		return nil, fmt.Errorf("創建 OTLP 導出器失敗: %w", err)
	}

	return exporter, nil
}

// 關閉追蹤器並釋放資源
func (t *Tracer) Shutdown(ctx context.Context) error {
	return t.provider.Shutdown(ctx)
}

// StartSpan 開始一個新的 span
func (t *Tracer) StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	return t.tracer.Start(ctx, name, opts...)
}

// 輔助方法: WithSpan 包裝一個函數調用，創建 span 並自動結束
func (t *Tracer) WithSpan(ctx context.Context, name string, fn func(context.Context) error, attrs ...attribute.KeyValue) error {
	ctx, span := t.StartSpan(ctx, name)
	defer span.End()

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	if err := fn(ctx); err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// 輔助方法: WithSpanCatch 包裝一個函數調用，自動捕獲 panic
func (t *Tracer) WithSpanCatch(ctx context.Context, name string, fn func(context.Context) error, attrs ...attribute.KeyValue) (err error) {
	ctx, span := t.StartSpan(ctx, name)
	defer span.End()

	if len(attrs) > 0 {
		span.SetAttributes(attrs...)
	}

	defer func() {
		if r := recover(); r != nil {
			err = fmt.Errorf("panic: %v", r)
			span.RecordError(err)
			span.SetStatus(codes.Error, err.Error())
			// 重新拋出 panic
			panic(r)
		}
	}()

	err = fn(ctx)
	if err != nil {
		span.RecordError(err)
		span.SetStatus(codes.Error, err.Error())
		return err
	}

	span.SetStatus(codes.Ok, "")
	return nil
}

// AddEvent 向 span 添加一個事件
func AddEvent(span trace.Span, name string, attrs ...attribute.KeyValue) {
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// RecordError 記錄錯誤到 span
func RecordError(span trace.Span, err error, opts ...trace.EventOption) {
	span.RecordError(err, opts...)
}

// SetAttributes 設置 span 的屬性
func SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	span.SetAttributes(attrs...)
}

// GetTracer 獲取底層的 trace.Tracer
func (t *Tracer) GetTracer() trace.Tracer {
	return t.tracer
}

// 輔助函數: 從 context 中獲取當前活動的 span
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

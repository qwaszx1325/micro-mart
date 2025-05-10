package mmotel

import (
	"context"
	"fmt"
	"log"
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

// 全局 tracer 變數
var globalTracer *Tracer

// Tracer 是我們自定義的追蹤器
type Tracer struct {
	provider *sdktrace.TracerProvider
	tracer   trace.Tracer
	options  *Options
}

// InitTracer 初始化全局追蹤器並返回關閉函數
func InitTracer(serviceName string, opts ...Option) func() {
	if serviceName == "" {
		log.Printf("服務名稱不能為空")
		return func() {}
	}

	tracer, err := NewTracer(serviceName, opts...)
	if err != nil {
		log.Printf("初始化追蹤器失敗: %v", err)
		return func() {}
	}

	globalTracer = tracer

	// 返回關閉函數
	return func() {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		if err := tracer.Shutdown(ctx); err != nil {
			log.Printf("關閉追蹤器失敗: %v", err)
		}
	}
}

// GetTracer 獲取全局追蹤器
func GetTracer() *Tracer {
	return globalTracer
}

// 創建新的追蹤器
func NewTracer(serviceName string, opts ...Option) (*Tracer, error) {
	if serviceName == "" {
		return nil, fmt.Errorf("服務名稱不能為空")
	}

	options := defaultOptions()
	options.ServiceName = serviceName

	// 應用自定義選項
	for _, opt := range opts {
		if opt != nil {
			opt(options)
		}
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
	if options == nil {
		return nil, fmt.Errorf("選項不能為空")
	}

	if options.ServiceName == "" {
		return nil, fmt.Errorf("服務名稱不能為空")
	}

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
		// 使用默認端點
		log.Println("警告: Jaeger 端點為空，使用默認端點 localhost:4317")
		endpoint = "localhost:4317"
	}

	// 使用最新的 API
	exporter, err := otlptracegrpc.New(
		ctx,
		otlptracegrpc.WithEndpoint(endpoint),
		otlptracegrpc.WithInsecure(),
		otlptracegrpc.WithTimeout(5*time.Second),
	)
	if err != nil {
		return nil, fmt.Errorf("創建 OTLP 導出器失敗: %w", err)
	}

	return exporter, nil
}

// 關閉追蹤器並釋放資源
func (t *Tracer) Shutdown(ctx context.Context) error {
	if t == nil || t.provider == nil {
		return nil
	}
	return t.provider.Shutdown(ctx)
}

// StartSpan 開始一個新的 span
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if ctx == nil {
		ctx = context.Background()
	}

	if globalTracer == nil {
		// 如果沒有初始化全局追蹤器，使用 NoopTracer
		return ctx, trace.SpanFromContext(ctx)
	}
	return globalTracer.tracer.Start(ctx, name, opts...)
}

// WithSpan 包裝一個函數調用，創建 span 並自動結束
func WithSpan(ctx context.Context, name string, fn func(context.Context) error, attrs ...attribute.KeyValue) error {
	if ctx == nil {
		ctx = context.Background()
	}

	if globalTracer == nil {
		// 如果沒有初始化追蹤器，只運行函數
		return fn(ctx)
	}

	ctx, span := StartSpan(ctx, name)
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

// WithSpanCatch 包裝一個函數調用，自動捕獲 panic
func WithSpanCatch(ctx context.Context, name string, fn func(context.Context) error, attrs ...attribute.KeyValue) (err error) {
	if ctx == nil {
		ctx = context.Background()
	}

	if globalTracer == nil {
		// 如果沒有初始化追蹤器，只運行函數但添加 panic 恢復
		defer func() {
			if r := recover(); r != nil {
				err = fmt.Errorf("panic: %v", r)
				panic(r) // 重新拋出 panic
			}
		}()
		return fn(ctx)
	}

	ctx, span := StartSpan(ctx, name)
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
	if span == nil {
		return
	}
	span.AddEvent(name, trace.WithAttributes(attrs...))
}

// RecordError 記錄錯誤到 span
func RecordError(span trace.Span, err error, opts ...trace.EventOption) {
	if span == nil || err == nil {
		return
	}
	span.RecordError(err, opts...)
}

// SetAttributes 設置 span 的屬性
func SetAttributes(span trace.Span, attrs ...attribute.KeyValue) {
	if span == nil {
		return
	}
	span.SetAttributes(attrs...)
}

// 從上下文中獲取當前 span
func SpanFromContext(ctx context.Context) trace.Span {
	if ctx == nil {
		return trace.SpanFromContext(context.Background())
	}
	return trace.SpanFromContext(ctx)
}

// 日誌相關函數

// Info 記錄一個信息級別的日誌
func Info(ctx context.Context, message string, fields ...interface{}) {
	span := SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		// 如果 span 不是有效的或不在記錄，只打印到標準日誌
		log.Printf("[INFO] %s %v", message, fields)
		return
	}

	// 添加事件到 span
	attrs := convertFieldsToAttributes(fields)
	AddEvent(span, "INFO: "+message, attrs...)

	// 同時打印到標準日誌
	log.Printf("[INFO] %s", message)
}

// Error 記錄一個錯誤級別的日誌
func Error(ctx context.Context, message string, fields ...Field) {
	span := SpanFromContext(ctx)
	if span == nil || !span.IsRecording() {
		// 如果 span 不是有效的或不在記錄，只打印到標準日誌
		log.Printf("[ERROR] %s %v", message, fields)
		return
	}

	// 添加事件到 span
	attrs := convertFieldArrayToAttributes(fields)
	AddEvent(span, "ERROR: "+message, attrs...)

	// 尋找錯誤對象
	for _, field := range fields {
		if errVal, ok := field.Value.(error); ok {
			RecordError(span, errVal)
			break
		}
	}

	// 同時打印到標準日誌
	log.Printf("[ERROR] %s", message)
}

// 將字段轉換為屬性
func convertFieldsToAttributes(fields []interface{}) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	for i := 0; i < len(fields); i++ {
		// 處理 Field 類型
		if field, ok := fields[i].(Field); ok {
			attrs = append(attrs, convertValueToAttribute(field.Key, field.Value))
			continue
		}

		// 處理鍵值對
		if i+1 < len(fields) {
			if key, ok := fields[i].(string); ok {
				attrs = append(attrs, convertValueToAttribute(key, fields[i+1]))
				i++ // 跳過值
			}
		}
	}

	return attrs
}

// 將 Field 數組轉換為屬性
func convertFieldArrayToAttributes(fields []Field) []attribute.KeyValue {
	var attrs []attribute.KeyValue

	for _, field := range fields {
		attrs = append(attrs, convertValueToAttribute(field.Key, field.Value))
	}

	return attrs
}

// 根據值的類型轉換為屬性
func convertValueToAttribute(key string, value interface{}) attribute.KeyValue {
	switch v := value.(type) {
	case string:
		return attribute.String(key, v)
	case int:
		return attribute.Int(key, v)
	case int64:
		return attribute.Int64(key, v)
	case float64:
		return attribute.Float64(key, v)
	case bool:
		return attribute.Bool(key, v)
	case error:
		return attribute.String(key, v.Error())
	default:
		return attribute.String(key, fmt.Sprintf("%v", v))
	}
}

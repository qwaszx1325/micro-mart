package mmotel

import (
	"fmt"
	"net/http"
	"time"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/trace"
)

// responseWriter 是一個包裝 http.ResponseWriter 的結構體，用於捕獲響應狀態碼
type responseWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func newResponseWriter(w http.ResponseWriter) *responseWriter {
	return &responseWriter{
		ResponseWriter: w,
		statusCode:     http.StatusOK,
	}
}

func (rw *responseWriter) WriteHeader(code int) {
	rw.statusCode = code
	rw.ResponseWriter.WriteHeader(code)
}

func (rw *responseWriter) Write(b []byte) (int, error) {
	size, err := rw.ResponseWriter.Write(b)
	rw.size += size
	return size, err
}

// HTTPMiddleware 返回一個 HTTP 中間件，自動為每個請求創建 span
// 使用最新推薦的 otelhttp 包
func (t *Tracer) HTTPMiddleware(next http.Handler) http.Handler {
	// 使用 otelhttp 包裝處理器
	return otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 獲取當前請求的 span
			ctx := r.Context()
			span := trace.SpanFromContext(ctx)

			// 添加自定義屬性
			span.SetAttributes(
				attribute.String("http.user_agent", r.UserAgent()),
			)

			// 創建包裝的響應寫入器來獲取更多信息
			rw := newResponseWriter(w)

			// 記錄開始時間
			startTime := time.Now()

			// 使用 defer 和 recover 來捕獲 panic
			defer func() {
				if err := recover(); err != nil {
					// 記錄 panic 為錯誤
					errorMsg := fmt.Sprintf("panic: %v", err)
					span.SetStatus(codes.Error, errorMsg)
					span.RecordError(fmt.Errorf(errorMsg))

					// 記錄響應時間
					duration := time.Since(startTime)
					span.SetAttributes(
						attribute.Int64("http.response_time_ms", duration.Milliseconds()),
					)

					// 重新拋出 panic
					panic(err)
				}

				// 記錄響應信息
				duration := time.Since(startTime)
				span.SetAttributes(
					attribute.Int64("http.response_time_ms", duration.Milliseconds()),
					attribute.Int64("http.response_size", int64(rw.size)),
				)
			}()

			// 調用下一個處理器
			next.ServeHTTP(rw, r)
		}),
		"http_server", // 操作名稱
		otelhttp.WithTracerProvider(t.provider),
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}

// WrapHandler 包裝一個 HTTP 處理器
func (t *Tracer) WrapHandler(handler http.Handler) http.Handler {
	return t.HTTPMiddleware(handler)
}

// WrapHandlerFunc 包裝一個 HTTP 處理器函數
func (t *Tracer) WrapHandlerFunc(handlerFunc http.HandlerFunc) http.Handler {
	return t.HTTPMiddleware(handlerFunc)
}

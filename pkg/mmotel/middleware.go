// pkg/mmotel/middleware.go
package mmotel

import (
	"fmt"
	"log"
	"net/http"

	"go.opentelemetry.io/contrib/instrumentation/net/http/otelhttp"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/propagation"
)

// HTTPMiddleware 返回一個 HTTP 中間件，自動為每個請求創建 span
func HTTPMiddleware(next http.Handler) http.Handler {
	if globalTracer == nil {
		log.Println("警告: 追蹤器未初始化，使用無追蹤中間件")
		return next
	}

	// 使用 OpenTelemetry 官方的 HTTP 處理器
	return otelhttp.NewHandler(
		http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			// 獲取當前請求的 span
			ctx := r.Context()
			span := SpanFromContext(ctx)

			// 添加自定義屬性
			span.SetAttributes(
				attribute.String("http.user_agent", r.UserAgent()),
			)

			// 調用下一個處理器
			next.ServeHTTP(w, r)
		}),
		"http_server", // 操作名稱
		otelhttp.WithTracerProvider(globalTracer.provider),
		otelhttp.WithPropagators(propagation.NewCompositeTextMapPropagator(
			propagation.TraceContext{},
			propagation.Baggage{},
		)),
		otelhttp.WithSpanNameFormatter(func(operation string, r *http.Request) string {
			return fmt.Sprintf("%s %s", r.Method, r.URL.Path)
		}),
	)
}

// WrapHTTPHandler 包裝一個 HTTP 處理器
func WrapHTTPHandler(handler http.Handler) http.Handler {
	return HTTPMiddleware(handler)
}

// WrapHTTPHandlerFunc 包裝一個 HTTP 處理器函數
func WrapHTTPHandlerFunc(handlerFunc http.HandlerFunc) http.Handler {
	return HTTPMiddleware(handlerFunc)
}

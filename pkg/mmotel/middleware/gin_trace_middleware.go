package middleware

import (
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

// GinTraceMiddleware creates a middleware for tracing HTTP requests in Gin
func GinTraceMiddleware() gin.HandlerFunc {
	tracer := otel.Tracer("gin-tracer")

	return func(c *gin.Context) {
		// 從 HTTP request 提取 context
		ctx := c.Request.Context()

		// 建立 span，span name 可用 method+path
		spanName := c.Request.Method + " " + c.FullPath()
		ctx, span := tracer.Start(ctx, spanName)
		defer span.End()

		// 設置回 c.Request，讓後續邏輯也能取得 trace context
		c.Request = c.Request.WithContext(ctx)

		// 處理請求
		c.Next()

	}
}

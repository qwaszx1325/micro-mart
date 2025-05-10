package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gin-gonic/gin"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/frontend_api/internal/api"
	frontendConfig "micro-mart/services/frontend_api/internal/config"
	"micro-mart/services/frontend_api/internal/infrastructure/grpc_client"
	userConfig "micro-mart/services/user/config"
)

func main() {
	// 加載配置
	cfg := frontendConfig.GetConfig("frontend_api")

	// 初始化 OpenTelemetry
	shutdown := mmotel.InitTracer(cfg.Host.ServiceName,
		mmotel.WithEnvironment("development"),
		mmotel.WithJaegerExporter(cfg.Otel.OtelUrl),
	)
	defer shutdown()

	// 初始化 user client
	userCfg := &userConfig.Config{}
	userCfg.UserUrl = cfg.User.UserUrl
	userClient, err := grpc_client.NewUserClient(userCfg)
	if err != nil {
		log.Fatalf("Failed to create user client: %v", err)
	}

	// 初始化 user handler
	userHandler := api.NewUserHandler(userClient)

	// 設置 Gin 路由
	router := gin.Default()

	// 註冊路由
	v1 := router.Group("/api/v1")
	{
		// 用戶相關路由
		userRoutes := v1.Group("/users")
		{
			userRoutes.POST("/register", userHandler.Register)
		}
	}

	// 啟動 HTTP 服務器
	srv := &http.Server{
		Addr:    ":8080",
		Handler: router,
	}

	// 優雅關閉
	go func() {
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("listen: %s\n", err)
		}
	}()

	// 等待中斷信號以優雅地關閉服務器
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// 設置 5 秒的超時時間
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := srv.Shutdown(ctx); err != nil {
		log.Fatal("Server forced to shutdown:", err)
	}

	log.Println("Server exiting")
}

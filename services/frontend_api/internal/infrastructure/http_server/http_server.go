package httpserver

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"github.com/gin-gonic/gin/binding"
	"github.com/go-playground/validator/v10"
	"go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin"
	"go.uber.org/fx"
	"micro-mart/pkg/helper"
	mmerror "micro-mart/pkg/mm_error"
	"micro-mart/pkg/mmotel"
	"micro-mart/pkg/responder"
	"micro-mart/services/frontend_api/internal/config"
	route "micro-mart/services/frontend_api/internal/route"
	"net/http"
	"os"
	"path/filepath"
)

// 修改NewHttpServer函數，使其返回一個HttpServer實例
func NewHttpServer(
	lc fx.Lifecycle,
	cfg *config.Config,
	route route.Route,
) *http.Server {

	// New gin server
	httpServer := http.Server{
		Addr: cfg.ServiceUrl,
	}

	// 不再需要這個變數
	// var shutdown func(context.Context) error

	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			if v, ok := binding.Validator.Engine().(*validator.Validate); ok {
				v.RegisterValidation("one_alpha", helper.ContainsAtLeastOneAlpha)
				v.RegisterValidation("one_num", helper.ContainsAtLeastOneNum)
			}

			r := gin.New()

			// jaeger trace用的中間件
			r.Use(otelgin.Middleware("frontend-api"))

			//response 的中間件
			r.Use(responder.GinResponser())
			
			// Register the routes
			route.RegisterRoutes(r)

			// Replace the handler
			httpServer.Handler = r

			// 啟動HTTP服務器
			go func() {
				if err := httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
					mmotel.Error(ctx, "Failed to start server", mmotel.NewField("error", err))
				}
			}()

			mmotel.Info(ctx, fmt.Sprintf("http server started at %s", cfg.ServiceUrl))
			return nil
		},
		OnStop: func(ctx context.Context) error {
			// 直接使用Server的Shutdown方法，而不是使用未初始化的shutdown變數
			err := httpServer.Shutdown(ctx)
			if err != nil {
				mmotel.Error(ctx, "Error shutting down http server", mmotel.NewField("error", err))
				return err
			}
			mmotel.Info(ctx, "http server shut down gracefully")
			return nil
		},
	})

	// 返回HttpServer實例
	return &httpServer
}

// startService starts the HTTP service using the provided Gin engine and configuration.
// It supports both TLS and non-TLS modes based on the configuration.
// todo 有空再做

// absPath returns the absolute path of the given path
func absPath(path string) (string, *mmerror.MmError) {
	if filepath.IsAbs(path) {
		return path, nil
	}

	workDir, err := os.Getwd()
	if err != nil {
		mmErr := mmerror.New(mmerror.InternalServerError, "failed to get current working directory", err)
		return "", mmErr
	}

	absPath := filepath.Join(workDir, path)
	return absPath, nil
}

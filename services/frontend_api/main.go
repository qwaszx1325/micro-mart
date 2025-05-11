package main

import (
	"go.uber.org/fx"
	"micro-mart/services/frontend_api/internal/config"
	"micro-mart/services/frontend_api/internal/infrastructure/grpc_client"
	"micro-mart/services/frontend_api/internal/infrastructure/http_server"
	route "micro-mart/services/frontend_api/internal/rote"
	"net/http"
)

func main() {
	fx.New(
		grpc_client.NewGrpcClientSet(),
		route.NewRouteV1Set(),
		fx.Provide(
			config.NewConfig,
			httpserver.NewHttpServer),
		fx.Invoke(func(*http.Server) {}),
	).Run()
}

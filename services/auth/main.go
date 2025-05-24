package main

import (
	"fmt"
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"micro-mart/pkg/db"
	redis_cache "micro-mart/pkg/db/redis"
	"micro-mart/services/auth/application"
	"micro-mart/services/auth/domain/service/service_impl"
	"micro-mart/services/auth/infrastructure/grpc_impl"
	"micro-mart/services/auth/infrastructure/redis_impl"
	"os"
)

func main() {

	fmt.Println("REDIS_MIN_IDLE_CONNS:", os.Getenv("REDIS_MIN_IDLE_CONNS"))
	fx.New(
		fx.Provide(
			grpc_impl.NewGrpcServer,
			application.NewAuthService,
			service_impl.NewAuthService,
			redis_impl.NewTokenRepository,
			redis_impl.NewRedisClient,
			fx.Annotate(
				redis_cache.NewRedisCache,
				fx.As(new(db.Cache)),
			),
		),
		fx.Invoke(
			func(server *grpc.Server) {},
		),
	).Run()

}

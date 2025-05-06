package main

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"micro-mart/services/user/application"
	"micro-mart/services/user/domain/service"
	"micro-mart/services/user/infrastructure/db_impl"
	"micro-mart/services/user/infrastructure/ent_impl"
	"micro-mart/services/user/infrastructure/grpc_impl"
)

func main() {

	fx.New(
		db_impl.NewEntDbFx(),
		fx.Provide(
			grpc_impl.NewGrpcServer,
			application.NewUserService,
			service.NewUserService,
			ent_impl.NewUserRepository,
		),
		fx.Invoke(
			func(server *grpc.Server) {},
		),
	).Run()

}

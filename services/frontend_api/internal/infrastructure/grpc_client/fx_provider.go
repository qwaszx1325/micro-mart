package grpc_client

import "go.uber.org/fx"

func NewGrpcClientSet() fx.Option {
	return fx.Module("grpc_client",
		// 有增加其他服務要加在裡面
		fx.Provide(NewUserClient))
}

package main

import (
	"go.uber.org/fx"
	"google.golang.org/grpc"
	"micro-mart/services/transaction-orchestrator/application"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/domain/service"
	"micro-mart/services/transaction-orchestrator/infrastructure/db_impl"
	"micro-mart/services/transaction-orchestrator/infrastructure/ent_impl"
	"micro-mart/services/transaction-orchestrator/infrastructure/grpc_client"
	"micro-mart/services/transaction-orchestrator/infrastructure/grpc_impl"
)

// ExtractTransactionService extracts the domain TransactionService from the application TransactionServiceImpl
func ExtractTransactionService(appService *application.TransactionServiceImpl) *service.TransactionService {
	return appService.GetTransactionService()
}

func main() {
	fx.New(
		db_impl.NewEntDbFx(),
		fx.Provide(
			config.GetConfig,
			grpc_impl.NewGrpcServer,
			application.NewTransactionService,
			ExtractTransactionService, // Provide the domain TransactionService
			application.NewRecoveryService,
			ent_impl.NewTransactionRepositoryImpl,
		),
		grpc_client.Module,
		fx.Invoke(
			func(server *grpc.Server) {},
		),
	).Run()
}

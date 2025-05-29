package main

import (
	"context"
	"log"
	"net"
	"os"
	"os/signal"
	"syscall"

	"go.uber.org/fx"
	"google.golang.org/grpc"
	"micro-mart/pkg/mmotel"
	"micro-mart/services/transaction-orchestrator/application"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/infrastructure/db_impl"
	"micro-mart/services/transaction-orchestrator/infrastructure/grpc_impl"
)

func main() {
	// Initialize tracer
	shutdown := mmotel.InitTracer("transaction-orchestrator", mmotel.WithStdoutExporter())
	defer shutdown()

	app := fx.New(
		fx.Provide(
			config.NewConfig,
			db_impl.NewTransactionRepository,
			application.NewTransactionService,
			application.NewRecoveryService,
			func(cfg *config.Config) *grpc.Server {
				server := grpc.NewServer(
					grpc.ChainUnaryInterceptor(
						mmotel.UnaryTraceInterceptor(),
						mmotel.ErrorLoggingInterceptor(),
					),
				)
				return server
			},
		),
		fx.Invoke(registerHooks),
	)

	startCtx, cancel := context.WithTimeout(context.Background(), app.StartTimeout())
	defer cancel()

	if err := app.Start(startCtx); err != nil {
		log.Fatal(err)
	}

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	stopCtx, cancel := context.WithTimeout(context.Background(), app.StopTimeout())
	defer cancel()

	if err := app.Stop(stopCtx); err != nil {
		log.Fatal(err)
	}
}

func registerHooks(
	lc fx.Lifecycle,
	cfg *config.Config,
	server *grpc.Server,
	recoveryService *application.RecoveryService,
	transactionService *application.TransactionServiceImpl,
) {
	lc.Append(fx.Hook{
		OnStart: func(ctx context.Context) error {
			// Register gRPC services
			grpc_impl.RegisterServer(server, transactionService)

			// Start gRPC server
			lis, err := net.Listen("tcp", cfg.GrpcServerAddress)
			if err != nil {
				return err
			}

			go func() {
				log.Printf("Starting gRPC server on %s", cfg.GrpcServerAddress)
				if err := server.Serve(lis); err != nil {
					log.Fatalf("Failed to serve: %v", err)
				}
			}()

			// Start recovery service
			go recoveryService.StartRecoveryProcess(ctx)

			return nil
		},
		OnStop: func(ctx context.Context) error {
			server.GracefulStop()
			return nil
		},
	})
}

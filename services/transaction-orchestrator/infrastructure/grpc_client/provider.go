package grpc_client

import (
	"go.uber.org/fx"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/domain/service"
)

// Module provides the gRPC clients for the transaction-orchestrator service
var Module = fx.Options(
	fx.Provide(
		NewUserClient,
		NewAuthClient,
		AsUserServiceClient,
		AsAuthServiceClient,
	),
)

// AsUserServiceClient is a provider that converts a *UserClient to a service.UserServiceClient
func AsUserServiceClient(client *UserClient) service.UserServiceClient {
	return client
}

// AsAuthServiceClient is a provider that converts a *AuthClient to a service.AuthServiceClient
func AsAuthServiceClient(client *AuthClient) service.AuthServiceClient {
	return client
}

// NewUserClientWithConfig creates a new UserClient with the given config
func NewUserClientWithConfig(cfg *config.Config) (service.UserServiceClient, error) {
	return NewUserClient(cfg)
}

// NewAuthClientWithConfig creates a new AuthClient with the given config
func NewAuthClientWithConfig(cfg *config.Config) (service.AuthServiceClient, error) {
	return NewAuthClient(cfg)
}
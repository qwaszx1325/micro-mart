package db_impl

import (
	"context"
	"log"
	"micro-mart/pkg/db"
	"micro-mart/services/transaction-orchestrator/config"
	"micro-mart/services/transaction-orchestrator/infrastructure/ent_impl/ent"
	"time"

	"go.uber.org/fx"
)

func NewDriver() ent.Option {
	cfg := config.GetConfig()
	driver := db.NewDriver(
		cfg.DB.User,
		cfg.DB.Pass,
		cfg.DB.Host,
		cfg.DB.Port,
		cfg.DB.Name,
	)

	// extra configurations
	db := driver.DB()
	db.SetMaxIdleConns(cfg.DB.MaxIdle)
	db.SetMaxOpenConns(cfg.DB.MaxConn)
	db.SetConnMaxLifetime(time.Duration(cfg.DB.ConnLife) * time.Second)

	return ent.Driver(driver)
}

func NewClient(driver ent.Option) *ent.Client {
	return ent.NewClient(driver)
}

func AutoMigrate(client *ent.Client) {
	cfg := config.GetConfig()

	// Auto migrate
	if cfg.DB.AutoMigrate {
		if err := client.Schema.Create(context.Background()); err != nil {
			log.Fatalf("failed to create schema resources: %v", err)
		}
	}
}

func NewEntDbFx() fx.Option {
	return fx.Module("ent",
		fx.Provide(NewDriver, NewClient, NewEntDb),
		fx.Invoke(AutoMigrate),
	)
}

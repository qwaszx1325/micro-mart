package config

import (
	"log"
	"micro-mart/pkg/cfgloader"
)

type (
	// 服務配置
	Host struct {
		ServiceName string `env:"SERVICE_NAME"`
		ServiceUrl  string `env:"SERVICE_URL"`
	}

	// OTEL 配置 (Jaeger)
	Otel struct {
		OtelUrl string `env:"OTEL_URL"`
	}

	// Redis 配置
	Redis struct {
		RedisUrl    string `env:"REDIS_URL"`
		Password    string `env:"REDIS_PASSWORD"`
		DB          int    `env:"REDIS_DB"`
		MaxActive   int    `env:"REDIS_MAX_ACTIVE_CONNS"`
		MinIdle     int    `env:"REDIS_MIN_IDLE_CONNS"`
		MaxIdle     int    `env:"REDIS_MAX_IDLE_CONNS"`
		ConnTimeout int    `env:"REDIS_CONN_TIMEOUT_SECS"`
	}

	// 資料庫配置 (PostgreSQL)
	DB struct {
		Host        string `env:"DB_HOST"`
		Port        int    `env:"DB_PORT"`
		User        string `env:"DB_USER"`
		Pass        string `env:"DB_PASS"`
		Name        string `env:"DB_NAME"`
		MaxConn     int    `env:"DB_MAX_CONN"`
		MaxIdle     int    `env:"DB_MAX_IDLE"`
		ConnLife    int    `env:"DB_MAX_CONN_LIFE_SECS"`
		AutoMigrate bool   `env:"AUTO_MIGRATE"`
	}

	// 外部服務配置
	Services struct {
		UserUrl string `env:"USER_URL"`
		AuthUrl string `env:"AUTH_URL"`
	}

	Config struct {
		Host
		Otel
		Redis
		DB
		Services
		RecoveryInterval int `env:"RECOVERY_INTERVAL_SECS" default:"60"`
	}
)

func GetConfig() *Config {
	config, err := cfgloader.LoadConfigFromEnv[Config]("transaction-orchestrator")
	if err != nil {
		log.Fatalf("load config from env failed: %v", err)
	}
	return config
}

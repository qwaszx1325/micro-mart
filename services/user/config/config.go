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
		MinIdle     int    `env:"REDIS_MIX_IDLE_CONNS"`
		MaxIdle     int    `env:"REDIS_MAX_IDLE_CONNS"`
		ConnTimeout int    `env:"REDIS_CONN_TIMEOUT_SECS"`
	}
	UserUrl struct {
		UserUrl string `env:"USER_URL"`
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

	// 驗證配置
	Verification struct {
		TokenExpiry           int `env:"VERIFICATION_TOKEN_EXPIRY"`
		TokenLockPeriod       int `env:"VERIFICATION_TOKEN_LOCK_PREIOD"`        // 注意拼寫錯誤保留
		TokenCountPeriod      int `env:"VERIFICATION_TOKEN_COUNT_PREIOD"`       // 注意拼寫錯誤保留
		TokenNotifyLockPeriod int `env:"VERIFICATION_TOKEN_NOTIFY_LOCK_PREIOD"` // 注意拼寫錯誤保留
		TokenTotalAttempts    int `env:"VERIFICATION_TOKEN_TOTAL_ATTEMPTS"`
	}

	Config struct {
		Host
		Otel
		Redis
		DB
		Verification
		UserUrl
	}
)

func GetConfig() *Config {
	config, err := cfgloader.LoadConfigFromEnv[Config]("user")
	if err != nil {
		log.Fatalf("load config from env failed: %v", err)
	}
	return config
}

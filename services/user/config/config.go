package config

import (
	"log"
	"micro-mart/pkg/cfgloader"
	"sync"
)

type (
	Host struct {
		ServiceName string `env:"SERVICE_NAME"`
		ServiceUrl  string `env:"SERVICE_URL"`
	}

	Otel struct {
		OtelUrl string `env:"OTEL_URL"`
	}

	Redis struct {
		RedisUrl    string `env:"REDIS_URL"`
		Password    string `env:"REDIS_PASSWORD"`
		DB          int    `env:"REDIS_DB"`
		MaxActive   int    `env:"REDIS_MAX_ACTIVE_CONNS"`
		MinIdle     int    `env:"REDIS_MIX_IDLE_CONNS"`
		MaxIdle     int    `env:"REDIS_MAX_IDLE_CONNS"`
		ConnTimeout int    `env:"REDIS_CONN_TIMEOUT_SECS"`
	}

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

	VERIFICATION struct {
		VerificationTokenExpiry           int `env:"VERIFICATION_TOKEN_EXPIRY"`
		VerificationTokenLockPeriod       int `env:"VERIFICATION_TOKEN_LOCK_PREIOD"`
		VerificationTokenCountPeriod      int `env:"VERIFICATION_TOKEN_COUNT_PREIOD"`
		VerificationTokenNotifyLockPeriod int `env:"VERIFICATION_TOKEN_NOTIFY_LOCK_PREIOD"`
		VerificationTokenTotalAttempts    int `env:"VERIFICATION_TOKEN_TOTAL_ATTEMPTS"`
	}

	Config struct {
		Host
		Otel
		Redis
		DB
		VERIFICATION
	}
)

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		config, err := cfgloader.LoadConfigFromEnv[Config]()
		if err != nil {
			log.Fatalf("load config from env failed: %v", err)
		}
		instance = config
	})
	return instance
}

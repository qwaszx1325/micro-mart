package config

import (
	"log"
	"micro-mart/pkg/cfgloader"
)

type (
	Host struct {
		ServiceName           string   `env:"SERVICE_NAME"`
		ServiceUrl            string   `env:"SERVICE_URL"`
		ServiceDomains        []string `env:"SERVICE_DOMAINS"`
		RateLimitIntervalSecs int      `env:"RATE_LIMIT_INTERVAL_SECS"`
		RateLimitMaxRequests  int      `env:"RATE_LIMIT_MAX_REQUESTS"`
		EnableTLS             bool     `env:"ENABLE_TLS"`
		CertFilePath          string   `env:"CERT_FILE_PATH"`
		KeyFilePath           string   `env:"KEY_FILE_PATH"`
	}

	Auth struct {
		AuthUrl string `env:"AUTH_URL"`
	}

	Merchant struct {
		MerchantUrl string `env:"MERCHANT_URL"`
	}

	User struct {
		UserUrl string `env:"USER_URL"`
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

	DBs struct {
		MerchantDB
		// AuthDB
	}

	MerchantDB struct {
		Host        string `env:"MERCHANT_DB_HOST"`
		Port        int    `env:"MERCHANT_DB_PORT"`
		User        string `env:"MERCHANT_DB_USER"`
		Pass        string `env:"MERCHANT_DB_PASS"`
		Name        string `env:"MERCHANT_DB_NAME"`
		MaxConn     int    `env:"MERCHANT_DB_MAX_CONN"`
		MaxIdle     int    `env:"MERCHANT_DB_MAX_IDLE"`
		ConnLife    int    `env:"MERCHANT_DB_MAX_CONN_LIFE_SECS"`
		AutoMigrate bool   `env:"MERCHANT_AUTO_MIGRATE"`
	}

	Config struct {
		Host
		Otel
		Auth
		Merchant
		User
		Redis
		DBs
	}
)

func NewConfig() *Config {
	config, err := cfgloader.LoadConfigFromEnv[Config]("frontend_api")
	if err != nil {
		log.Fatalf("load config from env failed: %v", err)
	}
	return config
}

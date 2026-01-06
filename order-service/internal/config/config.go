package config

import (
	"os"
	"strconv"
	"time"

	"github.com/gofor-little/env"
)

type Config struct {
	HTTPServer *HTTPServer
	Database   *Database
	AMQP       *AMQP
	Outbox     *Outbox
}

type HTTPServer struct {
	URL string
}

type Database struct {
	DSN string
}

type AMQP struct {
	Host                    string
	Port                    int
	Username                string
	Password                string
	PublishTimeout          time.Duration
	ConnectionRetryInterval time.Duration
	ConnectionRetryAttempts int
	Exchange                string
}

type Outbox struct {
	Interval       time.Duration
	MaxConcurrency int
	BatchSize      int
}

func NewConfig(envPath string) (*Config, error) {
	if err := env.Load(envPath); err != nil {
		return nil, err
	}

	cfg := &Config{
		HTTPServer: &HTTPServer{
			URL: getEnv("HTTP_SERVER_URL", ":4000"),
		},
		Database: &Database{
			DSN: getEnv("DATABASE_DSN", "postgres://postgres:password@order-db:5432/order-service?sslmode=disable"),
		},
		AMQP: &AMQP{
			Host:                    getEnv("AMQP_HOST", "rabbitmq"),
			Port:                    getEnvInt("AMQP_PORT", 5672),
			Username:                getEnv("AMQP_USERNAME", "default"),
			Password:                getEnv("AMQP_PASSWORD", "default"),
			PublishTimeout:          getEnvDuration("AMQP_PUBLISH_TIMEOUT_SECONDS", time.Second*2),
			ConnectionRetryInterval: getEnvDuration("AMQP_CONNECTION_RETRY_INTERVAL_SECONDS", time.Second*3),
			ConnectionRetryAttempts: getEnvInt("AMQP_CONNECTION_RETRY_ATTEMPTS", 5),
			Exchange:                getEnv("AMQP_EXCHANGE", "outbox.events"),
		},
		Outbox: &Outbox{
			MaxConcurrency: getEnvInt("AMQP_OUTBOX_MAX_CONCURRENCY", 10),
			BatchSize:      getEnvInt("AMQP_OUTBOX_BATCH_SIZE", 100),
			Interval:       getEnvDuration("OUTBOX_POLLING_INTERVAL", 2*time.Second),
		},
	}

	return cfg, nil
}

func getEnv(key string, defaultVal string) string {
	if val := os.Getenv(key); val != "" {
		return val
	}

	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if val, err := strconv.Atoi(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}

func getEnvDuration(key string, defaultVal time.Duration) time.Duration {
	if val, err := time.ParseDuration(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}

package config

import (
	"os"
	"strconv"
	"time"

	"github.com/gofor-little/env"
)

type Config struct {
	Database *Database
	AMQP     *AMQP
}

type Database struct {
	DSN string
}

type AMQP struct {
	Host           string
	Port           int
	Username       string
	Password       string
	PublishTimeout time.Duration
	Exchange       string
	Queue          string
}

func NewConfig(envPath string) (*Config, error) {
	if err := env.Load(envPath); err != nil {
		return nil, err
	}

	cfg := &Config{
		Database: &Database{
			DSN: getEnv("DATABASE_DSN", "postgres://postgres:password@notification-db:5432/notification-service?sslmode=disable"),
		},
		AMQP: &AMQP{
			Host:           getEnv("AMQP_HOST", "rabbitmq"),
			Port:           getEnvInt("AMQP_PORT", 5672),
			Username:       getEnv("AMQP_USERNAME", "default"),
			Password:       getEnv("AMQP_PASSWORD", "default"),
			PublishTimeout: getEnvDuration("AMQP_PUBLISH_TIMEOUT_SECONDS", time.Second*2),
			Exchange:       getEnv("AMQP_EXCHANGE", "outbox.events"),
			Queue:          getEnv("AMQP_QUEUE", "notification-service"),
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

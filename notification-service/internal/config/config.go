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
	Metrics    *Metrics
	Tracing    *Tracing
}

type HTTPServer struct {
	URL string
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
	DLX            string
	Queue          string
	DLQ            string
}

type Metrics struct {
	EnableDefaultMetrics bool
}

type Tracing struct {
	ServiceName  string `validate:"required"`
	CollectorURL string `validate:"required,hostname_port"`
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
			DSN: getEnv("DATABASE_DSN", "postgres://postgres:password@notification-db:5432/notification-service?sslmode=disable"),
		},
		AMQP: &AMQP{
			Host:           getEnv("AMQP_HOST", "rabbitmq"),
			Port:           getEnvInt("AMQP_PORT", 5672),
			Username:       getEnv("AMQP_USERNAME", "default"),
			Password:       getEnv("AMQP_PASSWORD", "default"),
			PublishTimeout: getEnvDuration("AMQP_PUBLISH_TIMEOUT", time.Second*2),
			Exchange:       getEnv("AMQP_EXCHANGE", "outbox.events"),
			DLX:            getEnv("AMQP_DLX", "outbox.dlx"),
			Queue:          getEnv("AMQP_QUEUE", "notification-service"),
			DLQ:            getEnv("AMQP_DLQ", "notification-service.dlq"),
		},
		Metrics: &Metrics{
			EnableDefaultMetrics: getEnvBool("METRICS_ENABLE_DEFAULT_METRICS", false),
		},
		Tracing: &Tracing{
			ServiceName:  getEnv("TRACING_SERVICE_NAME", "notification-service"),
			CollectorURL: getEnv("TRACING_COLLECTOR_URL", "jaeger:4318"),
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

func getEnvBool(key string, defaultVal bool) bool {
	if val, err := strconv.ParseBool(os.Getenv(key)); err == nil {
		return val
	}

	return defaultVal
}

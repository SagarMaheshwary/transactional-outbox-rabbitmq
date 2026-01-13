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
	DLQ            string
}

type Outbox struct {
	Interval              time.Duration
	MaxConcurrency        int
	BatchSize             int
	BacklogReportInterval time.Duration
	MaxRetryCount         int
	RetryDelay            time.Duration
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
			DSN: getEnv("DATABASE_DSN", "postgres://postgres:password@order-db:5432/order-service?sslmode=disable"),
		},
		AMQP: &AMQP{
			Host:           getEnv("AMQP_HOST", "rabbitmq"),
			Port:           getEnvInt("AMQP_PORT", 5672),
			Username:       getEnv("AMQP_USERNAME", "default"),
			Password:       getEnv("AMQP_PASSWORD", "default"),
			PublishTimeout: getEnvDuration("AMQP_PUBLISH_TIMEOUT", time.Second*2),
			Exchange:       getEnv("AMQP_EXCHANGE", "outbox.events"),
			DLX:            getEnv("AMQP_DLX", "outbox.dlx"),
			DLQ:            getEnv("AMQP_DLQ", "outbox.dlq.order-service"),
		},
		Outbox: &Outbox{
			MaxConcurrency:        getEnvInt("AMQP_OUTBOX_MAX_CONCURRENCY", 10),
			BatchSize:             getEnvInt("AMQP_OUTBOX_BATCH_SIZE", 100),
			Interval:              getEnvDuration("OUTBOX_POLLING_INTERVAL", 2*time.Second),
			BacklogReportInterval: getEnvDuration("OUTBOX_BACKLOG_REPORT_INTERVAL", 10*time.Second),
			MaxRetryCount:         getEnvInt("OUTBOX_MAX_RETRY_COUNT", 3),
			RetryDelay:            getEnvDuration("OUTBOX_RETRY_DELAY", 3*time.Second),
		},
		Metrics: &Metrics{
			EnableDefaultMetrics: getEnvBool("METRICS_ENABLE_DEFAULT_METRICS", false),
		},
		Tracing: &Tracing{
			ServiceName:  getEnv("TRACING_SERVICE_NAME", "order-service"),
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

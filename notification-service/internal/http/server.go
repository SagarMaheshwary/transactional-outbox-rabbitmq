package http

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/config"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/http/handler"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/observability/metrics"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/service"
)

type HTTPServer struct {
	URL    string
	Server *http.Server
}

type Opts struct {
	Log            logger.Logger
	Config         *config.Config
	MetricsService metrics.MetricsService
	HealthService  service.HealthService
}

func NewServer(url string, opts *Opts) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	// r.Use(
	// 	otelgin.Middleware(opts.Config.Tracing.ServiceName),
	// )

	healthHandler := handler.NewHealthHandler(&handler.HealthHandlerOpts{
		HealthService: opts.HealthService,
		Logger:        opts.Log,
	})

	r.GET("/metrics", gin.WrapH(opts.MetricsService.Handler()))
	r.GET("/livez", healthHandler.Livez)
	r.GET("/readyz", healthHandler.Readyz)

	return &HTTPServer{
		URL: url,
		Server: &http.Server{
			Addr:    url,
			Handler: r,
		},
	}
}

func (h *HTTPServer) ServeListener(listener net.Listener) error {
	if err := h.Server.Serve(listener); err != nil {
		return err
	}
	return nil
}

func (h *HTTPServer) Serve() error {
	listener, err := net.Listen("tcp", h.URL)
	if err != nil {
		return err
	}

	return h.ServeListener(listener)
}

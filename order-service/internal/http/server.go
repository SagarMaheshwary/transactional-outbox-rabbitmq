package http

import (
	"net"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/http/handler"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/service"
)

type HTTPServer struct {
	URL    string
	Server *http.Server
}

type Opts struct {
	OrderService service.OrderService
}

func NewServer(url string, opts *Opts) *HTTPServer {
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()

	orderHandler := handler.OrderHandler{
		OrderService: opts.OrderService,
	}

	r.POST("/orders", orderHandler.Create)

	return &HTTPServer{
		URL: url,
		Server: &http.Server{
			Addr:    url,
			Handler: r,
		},
	}
}

func (h *HTTPServer) ServeListener(listener net.Listener) error {
	if err := h.Server.Serve(listener); err != nil && err != http.ErrServerClosed {
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

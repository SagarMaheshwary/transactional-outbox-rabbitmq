package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/notification-service/internal/service"
)

type HealthHandlerOpts struct {
	HealthService service.HealthService
	Logger        logger.Logger
}

type HealthHandler struct {
	healthService service.HealthService
	logger        logger.Logger
}

func NewHealthHandler(opts *HealthHandlerOpts) *HealthHandler {
	return &HealthHandler{healthService: opts.HealthService, logger: opts.Logger}
}

func (h *HealthHandler) Livez(c *gin.Context) {
	c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (h *HealthHandler) Readyz(c *gin.Context) {
	status := h.healthService.Check(c)

	if status.Status == "ready" {
		c.JSON(http.StatusOK, status)
	} else {
		c.JSON(http.StatusServiceUnavailable, status)
	}
}

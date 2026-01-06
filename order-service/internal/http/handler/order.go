package handler

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/logger"
	"github.com/sagarmaheshwary/transactional-outbox-rabbitmq/order-service/internal/service"
)

type OrderHandler struct {
	OrderService service.OrderService
	Log          logger.Logger
}

type CreateOrderRequest struct {
	ProductID string `json:"product_id"`
	Quantity  int    `json:"quantity"`
}

func (o *OrderHandler) Create(c *gin.Context) {
	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	order, err := o.OrderService.Create(c.Request.Context(), &service.CreateOrder{
		ProductID: req.ProductID,
		Quantity:  req.Quantity,
	})
	if err != nil {
		o.Log.Error("Created order failed", logger.Field{Key: "error", Value: err.Error()})

		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"order": order,
	})
}

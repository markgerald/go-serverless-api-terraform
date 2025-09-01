package server

import (
	"go-serverless-api-terraform/internal/http/handlers"

	"github.com/gin-gonic/gin"

	// swagger
	swaggerFiles "github.com/swaggo/files"
	ginSwagger "github.com/swaggo/gin-swagger"
	_ "go-serverless-api-terraform/docs"
)

// NewRouter builds the Gin engine and registers routes
func NewRouter(h *handlers.Handler) *gin.Engine {
	r := gin.Default()

	// Swagger UI
	r.GET("/swagger/*any", ginSwagger.WrapHandler(swaggerFiles.Handler))

	// Orders routes
	r.GET("/orders", h.ListOrders)
	r.POST("/orders", h.CreateOrder)
	r.GET("/orders/:orderId", h.GetOrder)
	r.PUT("/orders/:orderId", h.UpdateOrder)
	r.DELETE("/orders/:orderId", h.DeleteOrder)

	// Order items routes
	r.GET("/orders/:orderId/items", h.ListItems)
	r.POST("/orders/:orderId/items", h.CreateItem)
	r.GET("/orders/:orderId/items/:itemId", h.GetItem)
	r.PUT("/orders/:orderId/items/:itemId", h.UpdateItem)
	r.DELETE("/orders/:orderId/items/:itemId", h.DeleteItem)

	return r
}

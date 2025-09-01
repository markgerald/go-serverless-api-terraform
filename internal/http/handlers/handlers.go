package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"go-serverless-api-terraform/internal/models"
	"go-serverless-api-terraform/internal/repository"
)

// Handler holds dependencies for HTTP endpoints
type Handler struct {
	repo repository.Repository
}

func New(repo repository.Repository) *Handler {
	return &Handler{repo: repo}
}

// DTOs

type createOrderReq struct {
	CustomerName string `json:"customer_name" binding:"required"`
	Status       string `json:"status"`
}

type updateOrderReq struct {
	CustomerName *string `json:"customer_name"`
	Status       *string `json:"status"`
}

type createItemReq struct {
	ProductName string  `json:"product_name" binding:"required"`
	Quantity    int     `json:"quantity" binding:"required"`
	Price       float64 `json:"price" binding:"required"`
}

type updateItemReq struct {
	ProductName *string  `json:"product_name"`
	Quantity    *int     `json:"quantity"`
	Price       *float64 `json:"price"`
}

// Orders
// ListOrders godoc
// @Summary List orders
// @Description Returns all orders
// @Tags orders
// @Produce json
// @Success 200 {array} models.Order
// @Failure 500 {object} map[string]string
// @Router /orders [get]
func (h *Handler) ListOrders(c *gin.Context) {
	orders, err := h.repo.ListOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

// CreateOrder godoc
// @Summary Create order
// @Description Creates a new order
// @Tags orders
// @Accept json
// @Produce json
// @Param order body createOrderReq true "Create order payload"
// @Success 201 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders [post]
func (h *Handler) CreateOrder(c *gin.Context) {
	var req createOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	order := &models.Order{
		ID:           uuid.NewString(),
		CustomerName: req.CustomerName,
		Status:       defaultIfEmpty(req.Status, "new"),
		CreatedAt:    now,
		UpdatedAt:    now,
	}
	if err := h.repo.CreateOrder(c.Request.Context(), order); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, order)
}

// GetOrder godoc
// @Summary Get order
// @Description Returns an order by ID
// @Tags orders
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {object} models.Order
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId} [get]
func (h *Handler) GetOrder(c *gin.Context) {
	id := c.Param("orderId")
	order, err := h.repo.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if order == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	c.JSON(http.StatusOK, order)
}

// UpdateOrder godoc
// @Summary Update order
// @Description Updates an existing order by ID
// @Tags orders
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Param order body updateOrderReq true "Update order payload"
// @Success 200 {object} models.Order
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId} [put]
func (h *Handler) UpdateOrder(c *gin.Context) {
	id := c.Param("orderId")
	existing, err := h.repo.GetOrder(c.Request.Context(), id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "order not found"})
		return
	}
	var req updateOrderReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.CustomerName != nil {
		existing.CustomerName = *req.CustomerName
	}
	if req.Status != nil {
		existing.Status = *req.Status
	}
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := h.repo.UpdateOrder(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteOrder godoc
// @Summary Delete order
// @Description Deletes an order by ID
// @Tags orders
// @Param orderId path string true "Order ID"
// @Success 204 {string} string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId} [delete]
func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Param("orderId")
	if err := h.repo.DeleteOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Items
// ListItems godoc
// @Summary List items of an order
// @Description Returns all items for a given order
// @Tags items
// @Produce json
// @Param orderId path string true "Order ID"
// @Success 200 {array} models.OrderItem
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId}/items [get]
func (h *Handler) ListItems(c *gin.Context) {
	orderID := c.Param("orderId")
	items, err := h.repo.ListOrderItems(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

// CreateItem godoc
// @Summary Create item
// @Description Creates a new item for a given order
// @Tags items
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Param item body createItemReq true "Create item payload"
// @Success 201 {object} models.OrderItem
// @Failure 400 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId}/items [post]
func (h *Handler) CreateItem(c *gin.Context) {
	orderID := c.Param("orderId")
	// Validate order exists
	ord, err := h.repo.GetOrder(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if ord == nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "order does not exist"})
		return
	}
	var req createItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.Quantity < 1 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be >= 1"})
		return
	}
	if req.Price < 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "price must be >= 0"})
		return
	}
	now := time.Now().UTC().Format(time.RFC3339)
	it := &models.OrderItem{
		OrderID:     orderID,
		ID:          uuid.NewString(),
		ProductName: req.ProductName,
		Quantity:    req.Quantity,
		Price:       req.Price,
		CreatedAt:   now,
		UpdatedAt:   now,
	}
	if err := h.repo.CreateOrderItem(c.Request.Context(), it); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusCreated, it)
}

// GetItem godoc
// @Summary Get item
// @Description Returns an item by ID for a given order
// @Tags items
// @Produce json
// @Param orderId path string true "Order ID"
// @Param itemId path string true "Item ID"
// @Success 200 {object} models.OrderItem
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId}/items/{itemId} [get]
func (h *Handler) GetItem(c *gin.Context) {
	orderID := c.Param("orderId")
	id := c.Param("itemId")
	it, err := h.repo.GetOrderItem(c.Request.Context(), orderID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if it == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	c.JSON(http.StatusOK, it)
}

// UpdateItem godoc
// @Summary Update item
// @Description Updates an item by ID for a given order
// @Tags items
// @Accept json
// @Produce json
// @Param orderId path string true "Order ID"
// @Param itemId path string true "Item ID"
// @Param item body updateItemReq true "Update item payload"
// @Success 200 {object} models.OrderItem
// @Failure 400 {object} map[string]string
// @Failure 404 {object} map[string]string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId}/items/{itemId} [put]
func (h *Handler) UpdateItem(c *gin.Context) {
	orderID := c.Param("orderId")
	id := c.Param("itemId")
	existing, err := h.repo.GetOrderItem(c.Request.Context(), orderID, id)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if existing == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "item not found"})
		return
	}
	var req updateItemReq
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	if req.ProductName != nil {
		existing.ProductName = *req.ProductName
	}
	if req.Quantity != nil {
		if *req.Quantity < 1 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "quantity must be >= 1"})
			return
		}
		existing.Quantity = *req.Quantity
	}
	if req.Price != nil {
		if *req.Price < 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "price must be >= 0"})
			return
		}
		existing.Price = *req.Price
	}
	existing.UpdatedAt = time.Now().UTC().Format(time.RFC3339)
	if err := h.repo.UpdateOrderItem(c.Request.Context(), existing); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, existing)
}

// DeleteItem godoc
// @Summary Delete item
// @Description Deletes an item by ID for a given order
// @Tags items
// @Param orderId path string true "Order ID"
// @Param itemId path string true "Item ID"
// @Success 204 {string} string
// @Failure 500 {object} map[string]string
// @Router /orders/{orderId}/items/{itemId} [delete]
func (h *Handler) DeleteItem(c *gin.Context) {
	orderID := c.Param("orderId")
	id := c.Param("itemId")
	if err := h.repo.DeleteOrderItem(c.Request.Context(), orderID, id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

func defaultIfEmpty(s, d string) string {
	if s == "" {
		return d
	}
	return s
}

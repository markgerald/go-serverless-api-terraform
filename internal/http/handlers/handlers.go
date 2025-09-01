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
func (h *Handler) ListOrders(c *gin.Context) {
	orders, err := h.repo.ListOrders(c.Request.Context())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, orders)
}

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

func (h *Handler) DeleteOrder(c *gin.Context) {
	id := c.Param("orderId")
	if err := h.repo.DeleteOrder(c.Request.Context(), id); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.Status(http.StatusNoContent)
}

// Items
func (h *Handler) ListItems(c *gin.Context) {
	orderID := c.Param("orderId")
	items, err := h.repo.ListOrderItems(c.Request.Context(), orderID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, items)
}

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

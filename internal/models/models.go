package models

// Order represents a customer order
// Stored in DynamoDB table configured by TABLE_ORDERS (PK: id)
type Order struct {
	ID           string `json:"id"`
	CustomerName string `json:"customer_name"`
	Status       string `json:"status"`
	CreatedAt    string `json:"created_at"`
	UpdatedAt    string `json:"updated_at"`
}

// OrderItem represents an item within an Order
// Stored in DynamoDB table configured by TABLE_ORDER_ITEMS (PK: order_id, SK: id)
type OrderItem struct {
	OrderID     string  `json:"order_id"`
	ID          string  `json:"id"`
	ProductName string  `json:"product_name"`
	Quantity    int     `json:"quantity"`
	Price       float64 `json:"price"`
	CreatedAt   string  `json:"created_at"`
	UpdatedAt   string  `json:"updated_at"`
}

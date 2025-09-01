package config

import (
	"os"

	"github.com/joho/godotenv"
)

// Config holds environment-driven configuration
// Keep fields aligned with previous main.go behavior
// so no behavior change for users or deployment.
type Config struct {
	Port            string
	AWSRegion       string
	DynamoEndpoint  string // optional for local DynamoDB
	OrdersTable     string
	OrderItemsTable string
	Env             string // e.g., "local" or "lambda"
}

// Load loads env vars and .env (if present)
func Load() (*Config, error) {
	_ = godotenv.Load() // ignore error; only for local convenience
	cfg := &Config{
		Port:            getenvDefault("APP_PORT", "8080"),
		AWSRegion:       getenvDefault("AWS_REGION", "us-east-1"),
		DynamoEndpoint:  os.Getenv("DYNAMODB_ENDPOINT"),
		OrdersTable:     getenvDefault("TABLE_ORDERS", "orders"),
		OrderItemsTable: getenvDefault("TABLE_ORDER_ITEMS", "order_items"),
		Env:             getenvDefault("APP_ENV", "local"),
	}
	return cfg, nil
}

func getenvDefault(k, d string) string {
	v := os.Getenv(k)
	if v == "" {
		return d
	}
	return v
}

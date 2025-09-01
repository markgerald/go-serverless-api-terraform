package main

import (
	"context"
	"log"
	"os"

	"github.com/aws/aws-lambda-go/lambda"
	ginadapter "github.com/awslabs/aws-lambda-go-api-proxy/gin"

	"go-serverless-api-terraform/internal/config"
	"go-serverless-api-terraform/internal/db"
	"go-serverless-api-terraform/internal/http/handlers"
	"go-serverless-api-terraform/internal/repository"
	"go-serverless-api-terraform/internal/server"
)

// @title Orders API
// @version 1.0
// @description API for managing orders and order items
// @BasePath /
// @schemes http https
// @accept json
// @produce json
func main() {
	cfg, err := config.Load()
	if err != nil {
		log.Fatalf("failed to load config: %v", err)
	}

	ctx := context.Background()
	dynamo, err := db.NewDynamoClient(ctx, cfg)
	if err != nil {
		log.Fatalf("failed to create dynamodb client: %v", err)
	}

	repo := repository.NewDynamoRepository(dynamo, cfg.OrdersTable, cfg.OrderItemsTable)
	h := handlers.New(repo)
	r := server.NewRouter(h)

	// Determine run mode: default local, otherwise Lambda
	env := cfg.Env
	if env == "" {
		env = os.Getenv("APP_ENV")
	}

	if env == "local" {
		addr := ":" + cfg.Port
		log.Printf("listening on %s", addr)
		if err := r.Run(addr); err != nil {
			log.Fatalf("server stopped: %v", err)
		}
		return
	}

	// Lambda mode (API Gateway / ALB)
	log.Printf("starting in Lambda mode (env=%s)", env)
	adapter := ginadapter.New(r)
	lambda.Start(adapter.ProxyWithContext)
}

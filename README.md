# go-serverless-api-terraform

Go project with an Orders API.

For the complete version of this project including Terraform/CDK for Terraform infrastructure, see:
https://github.com/markgerald/go-serverless-api-terraform/tree/terraform-cdk-go


## Table of Contents
- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [API](#api)
  - [Environment variables (API)](#environment-variables-api)
  - [Running locally](#running-locally)
  - [Endpoints](#endpoints)
  - [Swagger (documentation)](#swagger-documentation)
  - [Request examples](#request-examples)
- [Project Structure](#project-structure)


## Overview
- Go (Gin) API to manage orders and order items.
- Can run locally (APP_ENV=local) or in AWS Lambda with API Gateway (using aws-lambda-go and aws-lambda-go-api-proxy/gin).
- For infrastructure (Terraform/CDKTF), see the full project: https://github.com/markgerald/go-serverless-api-terraform/tree/terraform-cdk-go


## Prerequisites
- Go (as in go.mod; Go 1.22+ recommended)
- AWS CLI configured (profile or environment credentials)
- For DynamoDB Local (optional): Docker


## API

### Environment variables (API)
Example file: `.env.example`

- APP_PORT: local API port (default: 8080)
- AWS_REGION: AWS region used by the SDK (default: us-east-1)
- DYNAMODB_ENDPOINT: DynamoDB endpoint (use for DynamoDB Local, e.g.: http://localhost:8000)
- TABLE_ORDERS: orders table name (default: orders)
- TABLE_ORDER_ITEMS: order items table name (default: order_items)
- APP_ENV: runtime environment ("local" to run as a local HTTP server; any other value runs in Lambda mode)

Note (DynamoDB Local): besides the endpoint, set dummy credentials in your shell/.env when running locally:
- AWS_ACCESS_KEY_ID=dummy
- AWS_SECRET_ACCESS_KEY=dummy


### Running locally
1) Copy the example file and adjust:
   cp .env.example .env
2) (Optional) Start DynamoDB Local via Docker:
   docker run --rm -p 8000:8000 amazon/dynamodb-local
3) Start the API (APP_ENV=local is the default via config):
   go run ./main.go
4) Open: http://localhost:8080/swagger/index.html (Swagger) and the endpoints listed below.


### Endpoints
As defined by the router (internal/server/router.go) and Swagger (docs/docs.go):

- GET    /orders
- POST   /orders
- GET    /orders/:orderId
- PUT    /orders/:orderId
- DELETE /orders/:orderId
- GET    /orders/:orderId/items
- POST   /orders/:orderId/items
- GET    /orders/:orderId/items/:itemId
- PUT    /orders/:orderId/items/:itemId
- DELETE /orders/:orderId/items/:itemId


### Swagger (documentation)
- Local: http://localhost:8080/swagger/index.html
- Production (API Gateway): the /swagger route is not mapped in API Gateway in this project; the Swagger UI is intended for local use. See docs/docs.go if you need the reference for models and routes.


### Request examples
- Create order:
  curl -X POST http://localhost:8080/orders \
    -H 'Content-Type: application/json' \
    -d '{"customer_name":"Alice","status":"new"}'

- List orders:
  curl http://localhost:8080/orders

- Create item:
  curl -X POST http://localhost:8080/orders/<orderId>/items \
    -H 'Content-Type: application/json' \
    -d '{"product_name":"Keyboard","quantity":2,"price":99.99}'




## Project Structure
- `internal/` — application code (config, db, handlers, server, models, repository)
- `docs/` — minimal Swagger docs (loaded without code generation)
- `main.go` — API entrypoint (local/Lambda)
- `README.md` — this file

---

Questions or suggestions are welcome!
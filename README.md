# mark-go-lambda

Go project with an Orders API and AWS infrastructure managed with Terraform (CDK for Terraform in Go).

This README documents separately:
- API (how to run locally, environment variables, endpoints, and Swagger)
- Terraform (provisioned resources, variables, and deployment steps)


## Table of Contents
- [Overview](#overview)
- [Prerequisites](#prerequisites)
- [API](#api)
  - [Environment variables (API)](#environment-variables-api)
  - [Running locally](#running-locally)
  - [Endpoints](#endpoints)
  - [Swagger (documentation)](#swagger-documentation)
  - [Request examples](#request-examples)
- [Terraform (Infra)](#terraform-infra)
  - [Provisioned resources](#provisioned-resources)
  - [Environment variables (Terraform)](#environment-variables-terraform)
  - [Build the Lambda artifact](#build-the-lambda-artifact)
  - [Deploy (synthesize and apply)](#deploy-synthesize-and-apply)
  - [Destroy](#destroy)
  - [Troubleshooting](#troubleshooting)
- [Project Structure](#project-structure)


## Overview
- Go (Gin) API to manage orders and order items.
- Can run locally (APP_ENV=local) or in AWS Lambda with API Gateway (using aws-lambda-go and aws-lambda-go-api-proxy/gin).
- Infrastructure defined with CDK for Terraform (CDKTF) in Go, generating Terraform configurations for:
  - DynamoDB (orders and order items tables)
  - S3 (artifact bucket for the Lambda)
  - AWS Lambda (function running the API)
  - API Gateway HTTP API (routing to the Lambda)


## Prerequisites
- Go (as in go.mod; Go 1.22+ recommended)
- AWS CLI configured (profile or environment credentials)
- For DynamoDB Local (optional): Docker
- For infrastructure:
  - Terraform CLI
  - (Optional) CDKTF CLI — not strictly required; the project synthesizes and you can use Terraform directly in the generated directory


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


## Terraform (Infra)

### Provisioned resources
From `terraform/infra/infra.go` and modules in `terraform/modules`:
- DynamoDB
  - Orders table (hash key: id; PAY_PER_REQUEST billing; PITR enabled)
  - Order items table (hash: order_id; range: id; PAY_PER_REQUEST; PITR enabled)
- S3
  - Artifact bucket for the Lambda: name is `<S3_BUCKET_PREFIX>-<ENV>` (e.g., `go-lambda-artifacts-dev`)
  - Public access block enabled
  - Versioning/encryption options according to variables
- Lambda
  - Function created from S3 object (bucket + key)
  - Injected environment variables: TABLE_ORDERS, TABLE_ORDER_ITEMS, ENV
- API Gateway HTTP API
  - Routes for /orders and /orders/{orderId}/items endpoints mapped to the Lambda
  - Stage named after the environment (ENV)


### Environment variables (Terraform)
You can use `terraform/.env.sample` as a base and/or set variables in the environment when running `go run terraform/main.go`.

Main variables read by `terraform/main.go` (with defaults):
- AWS_REGION: AWS region (default: us-east-1)
- AWS_PROFILE: AWS CLI profile (optional)
- TABLE_ORDERS: orders table name (default: orders)
- TABLE_ORDER_ITEMS: order items table name (default: order_items)
- S3_BUCKET_PREFIX: artifact bucket prefix (default: go-lambda-artifacts)
- ENV: environment name (default: dev)
- S3_ENCRYPTION: NONE | AES256 | KMS (default: NONE if empty)
- S3_KMS_KEY_ID: KMS key ARN/ID (used if S3_ENCRYPTION=KMS)
- S3_VERSIONING: "true"/"1" to enable versioning (default: disabled)
- S3_FORCE_DESTROY: "true"/"1" to allow destroying a non-empty bucket (caution!)
- LAMBDA_NAME: function name (default: go-lambda-api)
- LAMBDA_RUNTIME: runtime (default: provided.al2023)
- LAMBDA_HANDLER: handler (default: bootstrap)
- LAMBDA_S3_KEY: object key in the artifact bucket (default: api.zip)
- LAMBDA_MEMORY: memory in MB (default: 128)
- LAMBDA_TIMEOUT: timeout in seconds (default: 10)


### Build the Lambda artifact
The function uses the custom runtime `provided.al2023` — build a `bootstrap` binary and zip it as `api.zip` (or another name if you set LAMBDA_S3_KEY).

Example (Linux/macOS):
```
# At the project root
GOOS=linux GOARCH=amd64 go build -o bootstrap ./
zip api.zip bootstrap
```

After the bucket exists, upload to S3:
```
aws s3 cp api.zip s3://<S3_BUCKET_PREFIX>-<ENV>/<LAMBDA_S3_KEY>
# e.g.: aws s3 cp api.zip s3://go-lambda-artifacts-dev/api.zip
```


### Deploy (synthesize and apply)
This project uses CDKTF in Go to generate Terraform JSON files, then you can apply them with the Terraform CLI.

1) Set variables (e.g., create `terraform/.env` based on `terraform/.env.sample`).
2) Synthesize the stack (generates `cdktf.out/`):
```
go run ./terraform/main.go
```
3) Go to the stack directory and initialize Terraform:
```
cd cdktf.out/stacks/infra
terraform init
```
4) First create only the artifact bucket (so you can upload the Lambda zip):
```
terraform apply -target=aws_s3_bucket.ArtifactsBucket -auto-approve
```
5) Upload the artifact to the created bucket:
```
# In another terminal (or after step 4):
aws s3 cp <your-zip> s3://<S3_BUCKET_PREFIX>-<ENV>/<LAMBDA_S3_KEY>
```
6) Apply the rest of the infrastructure:
```
terraform apply
```

At the end, check API Gateway in the AWS console to get the URL. The stage will have the name defined in `ENV` (e.g., `dev`).


### Destroy
To destroy the resources:
```
cd cdktf.out/stacks/infra
terraform destroy
```
If the bucket has versioning/objects, enable `S3_FORCE_DESTROY=true` and apply before destroying if needed.


### Troubleshooting
- Error creating the Lambda due to a missing S3 object:
  - Ensure the bucket was created (step 4) and you uploaded the correct `LAMBDA_S3_KEY` (step 5), then run `terraform apply` again.
- DynamoDB Local:
  - Set `DYNAMODB_ENDPOINT=http://localhost:8000` and dummy credentials in your environment when running locally.
- Swagger in production:
  - The `/swagger` route is not mapped in API Gateway; use it locally at `http://localhost:8080/swagger/index.html`.


## Project Structure
- `internal/` — application code (config, db, handlers, server, models, repository)
- `docs/` — minimal Swagger docs (loaded without code generation)
- `terraform/` — CDKTF code in Go and infrastructure modules
- `main.go` — API entrypoint (local/Lambda)
- `README.md` — this file

---

Questions or suggestions are welcome!
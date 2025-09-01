package main

import (
	"os"
	"strconv"
	"strings"

	"github.com/hashicorp/terraform-cdk-go/cdktf"
	"github.com/joho/godotenv"

	"go-serverless-api-terraform/terraform/infra"
)

func main() {
	_ = godotenv.Load()

	app := cdktf.NewApp(nil)

	region := os.Getenv("AWS_REGION")
	if region == "" {
		region = "us-east-1"
	}
	profile := os.Getenv("AWS_PROFILE")

	stack := infra.NewStack(app, "infra", &infra.StackConfig{Region: region, Profile: profile})

	ordersTable := os.Getenv("TABLE_ORDERS")
	if ordersTable == "" {
		ordersTable = "orders"
	}
	orderItemsTable := os.Getenv("TABLE_ORDER_ITEMS")
	if orderItemsTable == "" {
		orderItemsTable = "order_items"
	}

	bucketPrefix := os.Getenv("S3_BUCKET_PREFIX")
	if bucketPrefix == "" {
		bucketPrefix = "go-lambda-artifacts"
	}
	env := os.Getenv("ENV")
	if env == "" {
		env = "dev"
	}
	encryptionMode := strings.ToUpper(os.Getenv("S3_ENCRYPTION"))
	kmsKeyId := os.Getenv("S3_KMS_KEY_ID")
	versioning := strings.ToLower(os.Getenv("S3_VERSIONING"))
	forceDestroy := strings.ToLower(os.Getenv("S3_FORCE_DESTROY"))

	lambdaName := os.Getenv("LAMBDA_NAME")
	if lambdaName == "" {
		lambdaName = "go-lambda-api"
	}
	lambdaRuntime := os.Getenv("LAMBDA_RUNTIME")
	if lambdaRuntime == "" {
		lambdaRuntime = "provided.al2023"
	}
	lambdaHandler := os.Getenv("LAMBDA_HANDLER")
	if lambdaHandler == "" {
		lambdaHandler = "bootstrap"
	}
	lambdaKey := os.Getenv("LAMBDA_S3_KEY")
	if lambdaKey == "" {
		lambdaKey = "api.zip"
	}
	mem := 128
	if m, err := strconv.Atoi(os.Getenv("LAMBDA_MEMORY")); err == nil && m > 0 {
		mem = m
	}
	tout := 10
	if t, err := strconv.Atoi(os.Getenv("LAMBDA_TIMEOUT")); err == nil && t > 0 {
		tout = t
	}

	infra.SetupResources(stack, &infra.ResourcesConfig{
		OrdersTableName:     ordersTable,
		OrderItemsTableName: orderItemsTable,
		BucketPrefix:        bucketPrefix,
		Env:                 env,
		EncryptionMode:      encryptionMode,
		KmsKeyId:            kmsKeyId,
		Versioning:          versioning,
		ForceDestroy:        forceDestroy,
		LambdaName:          lambdaName,
		LambdaRuntime:       lambdaRuntime,
		LambdaHandler:       lambdaHandler,
		LambdaS3Key:         lambdaKey,
		MemoryMb:            mem,
		TimeoutSec:          tout,
	})

	app.Synth()
}

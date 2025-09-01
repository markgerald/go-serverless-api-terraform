package infra

import (
	apigwmod "go-serverless-api-terraform/terraform/modules/apigateway"
	dynmod "go-serverless-api-terraform/terraform/modules/dynamodb"
	lambdamod "go-serverless-api-terraform/terraform/modules/lambda"
	s3mod "go-serverless-api-terraform/terraform/modules/s3"

	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
	"github.com/cdktf/cdktf-provider-aws-go/aws/v21/provider"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

// StackConfig defines provider-level configuration for the stack.
type StackConfig struct {
	Region  string
	Profile string
}

// NewStack creates a new Terraform stack and configures the AWS provider.
func NewStack(scope constructs.Construct, id string, cfg *StackConfig) cdktf.TerraformStack {
	stack := cdktf.NewTerraformStack(scope, &id)

	// AWS Provider
	provider.NewAwsProvider(stack, jsii.String("aws"), &provider.AwsProviderConfig{
		Region:  jsii.String(cfg.Region),
		Profile: jsii.String(cfg.Profile),
	})

	return stack
}

// ResourcesConfig carries all inputs required to create application resources.
type ResourcesConfig struct {
	// Dynamo
	OrdersTableName     string
	OrderItemsTableName string

	// S3
	BucketPrefix   string
	Env            string
	EncryptionMode string // NONE|AES256|KMS
	KmsKeyId       string
	Versioning     string // "true"/"1" to enable
	ForceDestroy   string // "true"/"1" to enable

	// Lambda
	LambdaName    string
	LambdaRuntime string
	LambdaHandler string
	LambdaS3Key   string
	MemoryMb      int
	TimeoutSec    int
}

// SetupResources creates all application resources using the provided configuration.
func SetupResources(stack cdktf.TerraformStack, rc *ResourcesConfig) {
	// DynamoDB tables
	dynmod.NewDynamoTable(stack, "OrdersTable", &dynmod.DynamoTableProps{
		Name:        rc.OrdersTableName,
		HashKey:     dynmod.AttributeDef{Name: "id", Type: "S"},
		BillingMode: "PAY_PER_REQUEST",
		PitrEnabled: true,
		Tags:        map[string]string{"Project": "mark-go-lambda", "Table": rc.OrdersTableName},
	})

	dynmod.NewDynamoTable(stack, "OrderItemsTable", &dynmod.DynamoTableProps{
		Name:        rc.OrderItemsTableName,
		HashKey:     dynmod.AttributeDef{Name: "order_id", Type: "S"},
		RangeKey:    &dynmod.AttributeDef{Name: "id", Type: "S"},
		BillingMode: "PAY_PER_REQUEST",
		PitrEnabled: true,
		Tags:        map[string]string{"Project": "mark-go-lambda", "Table": rc.OrderItemsTableName},
	})

	// S3 bucket for artifacts
	var enc *s3mod.ServerSideEncryption
	if rc.EncryptionMode == "AES256" {
		enc = &s3mod.ServerSideEncryption{Enabled: true}
	} else if rc.EncryptionMode == "KMS" {
		if rc.KmsKeyId != "" {
			enc = &s3mod.ServerSideEncryption{Enabled: true, KmsKeyId: &rc.KmsKeyId}
		} else {
			enc = &s3mod.ServerSideEncryption{Enabled: true}
		}
	}

	pa := &s3mod.PublicAccessBlock{
		BlockPublicAcls: true, BlockPublicPolicy: true, IgnorePublicAcls: true, RestrictPublicBuckets: true,
	}

	artifactsBucket := s3mod.NewBucket(stack, "ArtifactsBucket", &s3mod.BucketProps{
		Prefix:       jsii.String(rc.BucketPrefix),
		Environment:  jsii.String(rc.Env),
		ForceDestroy: rc.ForceDestroy == "true" || rc.ForceDestroy == "1",
		Versioning:   &s3mod.Versioning{Enabled: rc.Versioning == "true" || rc.Versioning == "1"},
		Encryption:   enc,
		PublicAccess: pa,
		Tags:         map[string]string{"Project": "mark-go-lambda", "Purpose": "artifacts"},
	})

	// Lambda function
	apiLambda := lambdamod.NewLambdaFunction(stack, "ApiLambda", &lambdamod.LambdaProps{
		FunctionName: rc.LambdaName,
		S3Bucket:     artifactsBucket.Bucket(),
		S3Key:        rc.LambdaS3Key,
		Runtime:      rc.LambdaRuntime,
		Handler:      rc.LambdaHandler,
		MemoryMb:     rc.MemoryMb,
		TimeoutS:     rc.TimeoutSec,
		Environment: map[string]string{
			"TABLE_ORDERS":      rc.OrdersTableName,
			"TABLE_ORDER_ITEMS": rc.OrderItemsTableName,
			"ENV":               rc.Env,
		},
		Tags: map[string]string{"Project": "mark-go-lambda", "Service": "api"},
	})

	// API Gateway routing to Lambda
	routes := []apigwmod.RouteSpec{
		{Method: "GET", Path: "/orders"},
		{Method: "POST", Path: "/orders"},
		{Method: "GET", Path: "/orders/{orderId}"},
		{Method: "PUT", Path: "/orders/{orderId}"},
		{Method: "DELETE", Path: "/orders/{orderId}"},
		{Method: "GET", Path: "/orders/{orderId}/items"},
		{Method: "POST", Path: "/orders/{orderId}/items"},
		{Method: "GET", Path: "/orders/{orderId}/items/{itemId}"},
		{Method: "PUT", Path: "/orders/{orderId}/items/{itemId}"},
		{Method: "DELETE", Path: "/orders/{orderId}/items/{itemId}"},
	}

	apigwmod.NewHttpApiWithLambda(stack, "ApiGateway", &apigwmod.HttpApiProps{
		Name:               "mark-go-http-api",
		Description:        jsii.String("HTTP API for mark-go-lambda"),
		TargetLambdaArn:    apiLambda.Arn(),
		Routes:             routes,
		StageName:          rc.Env,
		AutoDeploy:         true,
		AllowInvokeFromApi: true,
		Tags:               map[string]string{"Project": "mark-go-lambda", "Service": "api"},
	})
}

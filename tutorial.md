# Terraform CDK + Go: Complete Serverless API Project (Part 1)
🚀 Infrastructure as Code in action!

If you already use Terraform but want the power of programming languages like Go, you’ll love this tutorial I just published.

I show you step by step how to build a complete Serverless API on AWS with Lambda + API Gateway + DynamoDB + S3, using Terraform CDK (CDKTF).
> File‑by‑file, hands‑on guide to build a **serverless Orders API** on AWS with **Terraform CDK (CDKTF) + Go**. We’ll create **every folder and file** under `terraform/`, explain what each piece does, and deploy. **Part 2 (coming soon)** will add **CI/CD with GitHub Actions** for both the API and the CDKTF code using **GitOps**.

---

## Why CDKTF (CDK for Terraform) is great

- **Type‑safe** infra in Go (structs, IDE autocomplete, refactors).
- **Reusable modules/constructs** instead of huge `.tf` files.
- Still **Terraform under the hood** (plan/apply, state backends, providers).
- Easy to grow into **multi‑stack** projects.

---

## What we’ll build

A small **Orders API** backed by **AWS Lambda** (Go), exposed by **API Gateway HTTP API**, storing data in **DynamoDB**, plus an **S3 bucket** to host the Lambda artifact (`api.zip`).

Endpoints in the app (outside scope here) include: `/orders`, `/orders/:orderId`, `/orders/:orderId/items`, etc. We’ll focus on the **infra**.

---

## Requirements

- **Go** ≥ 1.21/1.22
- **Node.js + npm** (for the CDKTF CLI)
- **Terraform CLI** ≥ 1.5
- **AWS CLI** configured (profile or env vars)

**Install CDKTF CLI**:
```bash
npm i -g cdktf-cli@latest
cdktf --version
```

---

## Repo context

Base project: `go-serverless-api-terraform` (contains the Go API code). We’ll create **all infra files** under `terraform/` (matching the branch `terraform-cdk-go`).

Clone and enter:
```bash
git clone https://github.com/markgerald/go-serverless-api-terraform.git
cd go-serverless-api-terraform
mkdir -p terraform && cd terraform
```

---

## Final folder layout (everything we’ll create now)

```
terraform/
├─ cdktf.json
├─ go.mod
├─ go.sum             # (auto‑generated)
├─ .env.sample
├─ main.go            # App entrypoint: reads env, builds stack, synths
├─ infra/
│  └─ infra.go        # Stack wiring (provider, modules, outputs)
└─ modules/
   ├─ dynamodb/
   │  └─ dynamodb.go  # Orders + Order Items tables
   ├─ s3/
   │  └─ bucket.go    # Artifact bucket + PAB + versioning/encryption
   ├─ iam/
   │  └─ role.go      # Lambda execution role + policies
   ├─ lambda/
   │  └─ lambda.go    # Lambda from S3 object + env vars
   └─ apigateway/
      └─ api.go       # HTTP API + integration + routes + stage + permission
```

We’ll build this tree step by step ⬇️

---

## 1) `cdktf.json` — project config

Create `terraform/cdktf.json`:
```json
{
  "language": "go",
  "app": "go run main.go",
  "terraformProviders": [
    "hashicorp/aws@~> 6.0"
  ],
  "context": {
    "awsRegion": "us-east-1"
  }
}
```
> Adjust `awsRegion` to your default. We’ll still allow overriding via `AWS_REGION`.

---

## 2) `go.mod` — module + dependencies

Create `terraform/go.mod`:
```go
module terraform

go 1.22

require (
    github.com/aws/constructs-go/constructs/v10 v10.3.0
    github.com/cdktf/cdktf-provider-aws-go/aws/v21 v21.0.0
    github.com/hashicorp/terraform-cdk-go/cdktf v0.21.0
)
```
Fetch deps:
```bash
go mod tidy
```

---

## 3) `.env.sample` — variables used by the stack

Create `terraform/.env.sample`:
```dotenv
# Common
AWS_REGION=us-east-1
AWS_PROFILE=
ENV=dev

# DynamoDB
TABLE_ORDERS=orders
TABLE_ORDER_ITEMS=order_items

# S3 artifact bucket
S3_BUCKET_PREFIX=go-lambda-artifacts
S3_VERSIONING=false      # true/false
S3_FORCE_DESTROY=false   # true/false (careful!)
S3_ENCRYPTION=NONE       # NONE|AES256|KMS
S3_KMS_KEY_ID=

# Lambda
LAMBDA_NAME=go-lambda-api
LAMBDA_RUNTIME=provided.al2023
LAMBDA_HANDLER=bootstrap
LAMBDA_MEMORY=128
LAMBDA_TIMEOUT=10
LAMBDA_S3_KEY=api.zip
```
> Copy to `.env` and export, or set directly in your shell before running.

---

## 4) `main.go` — entrypoint (reads env, builds stack, synths)

Create `terraform/main.go`:
```go
package main

import (
    "log"
    "os"

    "github.com/aws/jsii-runtime-go"
    "github.com/hashicorp/terraform-cdk-go/cdktf"

    "terraform/infra"
)

func getenv(k, def string) string {
    if v := os.Getenv(k); v != "" {
        return v
    }
    return def
}

func main() {
    // Read env (with sane defaults)
    cfg := infra.Config{
        AwsRegion:        getenv("AWS_REGION", "us-east-1"),
        AwsProfile:       os.Getenv("AWS_PROFILE"),
        Env:              getenv("ENV", "dev"),
        TableOrders:      getenv("TABLE_ORDERS", "orders"),
        TableOrderItems:  getenv("TABLE_ORDER_ITEMS", "order_items"),
        S3BucketPrefix:   getenv("S3_BUCKET_PREFIX", "go-lambda-artifacts"),
        S3Versioning:     getenv("S3_VERSIONING", "false"),
        S3ForceDestroy:   getenv("S3_FORCE_DESTROY", "false"),
        S3Encryption:     getenv("S3_ENCRYPTION", "NONE"),
        S3KmsKeyID:       os.Getenv("S3_KMS_KEY_ID"),
        LambdaName:       getenv("LAMBDA_NAME", "go-lambda-api"),
        LambdaRuntime:    getenv("LAMBDA_RUNTIME", "provided.al2023"),
        LambdaHandler:    getenv("LAMBDA_HANDLER", "bootstrap"),
        LambdaMemory:     getenv("LAMBDA_MEMORY", "128"),
        LambdaTimeout:    getenv("LAMBDA_TIMEOUT", "10"),
        LambdaS3Key:      getenv("LAMBDA_S3_KEY", "api.zip"),
    }

    app := cdktf.NewApp(nil)

    if _, err := infra.NewStack(app, "infra", &cfg); err != nil {
        log.Fatalf("failed to build stack: %v", err)
    }

    app.Synth() // generates cdktf.out/stacks/infra
    log.Printf("Synth complete. Now: cd cdktf.out/stacks/infra && terraform init")
}
```

---

## 5) `infra/infra.go` — wire provider + modules + outputs

Create folder and file:
```bash
mkdir -p infra && touch infra/infra.go
```

Add:
```go
package infra

import (
    "fmt"

    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
    "github.com/hashicorp/terraform-cdk-go/cdktf"

    awsprovider "github.com/cdktf/cdktf-provider-aws-go/aws/v21/provider"

    mddb "terraform/modules/dynamodb"
    mds3 "terraform/modules/s3"
    mdrole "terraform/modules/iam"
    mdlambda "terraform/modules/lambda"
    mdapi "terraform/modules/apigateway"
)

type Config struct {
    AwsRegion, AwsProfile string
    Env string

    TableOrders, TableOrderItems string

    S3BucketPrefix  string
    S3Versioning    string // "true"/"false"
    S3ForceDestroy  string // "true"/"false"
    S3Encryption    string // NONE|AES256|KMS
    S3KmsKeyID      string

    LambdaName, LambdaRuntime, LambdaHandler string
    LambdaMemory, LambdaTimeout, LambdaS3Key string
}

func NewStack(scope constructs.Construct, id string, cfg *Config) (cdktf.TerraformStack, error) {
    stack := cdktf.NewTerraformStack(scope, jsii.String(id))

    // Provider
    awsprovider.NewAwsProvider(stack, jsii.String("aws"), &awsprovider.AwsProviderConfig{
        Region:  jsii.String(cfg.AwsRegion),
        Profile: nilIfEmpty(cfg.AwsProfile),
    })

    // S3 bucket for artifacts
    bucket := mds3.NewArtifactBucket(stack, "Artifacts", &mds3.BucketConfig{
        Prefix:       cfg.S3BucketPrefix,
        Env:          cfg.Env,
        Versioning:   cfg.S3Versioning == "true" || cfg.S3Versioning == "1",
        ForceDestroy: cfg.S3ForceDestroy == "true" || cfg.S3ForceDestroy == "1",
        Encryption:   cfg.S3Encryption,
        KmsKeyID:     cfg.S3KmsKeyID,
    })

    // DynamoDB tables
    tables := mddb.NewTables(stack, "Tables", &mddb.TablesConfig{
        OrdersName:     cfg.TableOrders,
        OrderItemsName: cfg.TableOrderItems,
    })

    // IAM role for Lambda
    role := mdrole.NewLambdaRole(stack, "LambdaRole")

    // Lambda function from S3 object
    fn := mdlambda.NewApiLambda(stack, "ApiLambda", &mdlambda.LambdaConfig{
        Name:       cfg.LambdaName,
        Runtime:    cfg.LambdaRuntime,
        Handler:    cfg.LambdaHandler,
        Memory:     cfg.LambdaMemory,
        Timeout:    cfg.LambdaTimeout,
        S3Bucket:   bucket.Name(),
        S3Key:      cfg.LambdaS3Key,
        Env:        cfg.Env,
        TableOrders:     tables.OrdersName(),
        TableOrderItems: tables.OrderItemsName(),
        RoleArn:    role.Arn(),
    })

    // API Gateway HTTP API + routes + integration + permission
    api := mdapi.NewHttpApi(stack, "HttpApi", &mdapi.ApiConfig{
        Env:          cfg.Env,
        LambdaArn:    fn.Arn(),
        LambdaName:   fn.FunctionName(),
    })

    // Outputs
    cdktf.NewTerraformOutput(stack, jsii.String("api_url"), &cdktf.TerraformOutputConfig{
        Value: api.Endpoint(),
    })

    cdktf.NewTerraformOutput(stack, jsii.String("artifact_bucket"), &cdktf.TerraformOutputConfig{
        Value: bucket.Name(),
    })

    return stack, nil
}

func nilIfEmpty(s string) *string {
    if s == "" {
        return nil
    }
    return jsii.String(s)
}
```

---

## 6) Modules — one file at a time

Create the modules tree:
```bash
mkdir -p modules/dynamodb modules/s3 modules/iam modules/lambda modules/apigateway
```

### 6.1) `modules/dynamodb/dynamodb.go`
```go
package dynamodb

import (
    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
    "github.com/hashicorp/terraform-cdk-go/cdktf"
    dynamodbtable "github.com/cdktf/cdktf-provider-aws-go/aws/v21/dynamodbtable"
    ddbpitr "github.com/cdktf/cdktf-provider-aws-go/aws/v21/dynamodbtablepointintimerecovery"
)

type TablesConfig struct {
    OrdersName     string
    OrderItemsName string
}

type Tables struct {
    orders     dynamodbtable.DynamodbTable
    orderItems dynamodbtable.DynamodbTable
}

func NewTables(scope constructs.Construct, id string, cfg *TablesConfig) *Tables {
    s := constructs.NewConstruct(scope, jsii.String(id))

    orders := dynamodbtable.NewDynamodbTable(s, jsii.String("Orders"), &dynamodbtable.DynamodbTableConfig{
        Name:        jsii.String(cfg.OrdersName),
        BillingMode: jsii.String("PAY_PER_REQUEST"),
        HashKey:     jsii.String("id"),
        Attribute: &[]*dynamodbtable.DynamodbTableAttribute{{
            Name: jsii.String("id"), Type: jsii.String("S"),
        }},
    })
    ddbpitr.NewDynamodbTablePointInTimeRecovery(s, jsii.String("OrdersPITR"), &ddbpitr.DynamodbTablePointInTimeRecoveryConfig{
        TableName: orders.Name(),
        Enabled:   jsii.Bool(true),
    })

    items := dynamodbtable.NewDynamodbTable(s, jsii.String("OrderItems"), &dynamodbtable.DynamodbTableConfig{
        Name:        jsii.String(cfg.OrderItemsName),
        BillingMode: jsii.String("PAY_PER_REQUEST"),
        HashKey:     jsii.String("order_id"),
        RangeKey:    jsii.String("id"),
        Attribute: &[]*dynamodbtable.DynamodbTableAttribute{
            { Name: jsii.String("order_id"), Type: jsii.String("S") },
            { Name: jsii.String("id"),       Type: jsii.String("S") },
        },
    })
    ddbpitr.NewDynamodbTablePointInTimeRecovery(s, jsii.String("OrderItemsPITR"), &ddbpitr.DynamodbTablePointInTimeRecoveryConfig{
        TableName: items.Name(),
        Enabled:   jsii.Bool(true),
    })

    return &Tables{orders: orders, orderItems: items}
}

func (t *Tables) OrdersName() *string     { return t.orders.Name() }
func (t *Tables) OrderItemsName() *string { return t.orderItems.Name() }
```

### 6.2) `modules/s3/bucket.go`
```go
package s3

import (
    "fmt"

    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
    s3bucket "github.com/cdktf/cdktf-provider-aws-go/aws/v21/s3bucket"
    s3pab "github.com/cdktf/cdktf-provider-aws-go/aws/v21/s3bucketpublicaccessblock"
    s3ver "github.com/cdktf/cdktf-provider-aws-go/aws/v21/s3bucketversioning"
    s3serverenc "github.com/cdktf/cdktf-provider-aws-go/aws/v21/s3bucketserverSideencryptionconfiguration"
)

type BucketConfig struct {
    Prefix       string
    Env          string
    Versioning   bool
    ForceDestroy bool
    Encryption   string // NONE|AES256|KMS
    KmsKeyID     string
}

type Bucket struct {
    bucket s3bucket.S3Bucket
}

func NewArtifactBucket(scope constructs.Construct, id string, cfg *BucketConfig) *Bucket {
    s := constructs.NewConstruct(scope, jsii.String(id))
    name := fmt.Sprintf("%s-%s", cfg.Prefix, cfg.Env)

    b := s3bucket.NewS3Bucket(s, jsii.String("ArtifactsBucket"), &s3bucket.S3BucketConfig{
        Bucket:       jsii.String(name),
        ForceDestroy: jsii.Bool(cfg.ForceDestroy),
    })

    s3pab.NewS3BucketPublicAccessBlock(s, jsii.String("PAB"), &s3pab.S3BucketPublicAccessBlockConfig{
        Bucket:                b.Bucket(),
        BlockPublicAcls:       jsii.Bool(true),
        BlockPublicPolicy:     jsii.Bool(true),
        IgnorePublicAcls:      jsii.Bool(true),
        RestrictPublicBuckets: jsii.Bool(true),
    })

    s3ver.NewS3BucketVersioning(s, jsii.String("Versioning"), &s3ver.S3BucketVersioningConfig{
        Bucket: b.Bucket(),
        VersioningConfiguration: &s3ver.S3BucketVersioningVersioningConfiguration{
            Status: jsii.String(ifThen(cfg.Versioning, "Enabled", "Suspended")),
        },
    })

    if cfg.Encryption == "AES256" || cfg.Encryption == "KMS" {
        var rule *s3serverenc.S3BucketServerSideEncryptionConfigurationRule
        if cfg.Encryption == "AES256" {
            rule = &s3serverenc.S3BucketServerSideEncryptionConfigurationRule{
                ApplyServerSideEncryptionByDefault: &s3serverenc.S3BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefault{
                    SseAlgorithm: jsii.String("AES256"),
                },
            }
        } else {
            rule = &s3serverenc.S3BucketServerSideEncryptionConfigurationRule{
                ApplyServerSideEncryptionByDefault: &s3serverenc.S3BucketServerSideEncryptionConfigurationRuleApplyServerSideEncryptionByDefault{
                    SseAlgorithm: jsii.String("aws:kms"),
                    KmsMasterKeyId: nilIfEmpty(cfg.KmsKeyID),
                },
            }
        }
        s3serverenc.NewS3BucketServerSideEncryptionConfiguration(s, jsii.String("Enc"), &s3serverenc.S3BucketServerSideEncryptionConfigurationConfig{
            Bucket: b.Bucket(),
            Rule:   &[]*s3serverenc.S3BucketServerSideEncryptionConfigurationRule{rule},
        })
    }

    return &Bucket{bucket: b}
}

func (b *Bucket) Name() *string { return b.bucket.Bucket() }

func ifThen[T any](cond bool, a, b T) T { if cond { return a }; return b }
func nilIfEmpty(s string) *string { if s == "" { return nil }; return jsii.String(s) }
```

### 6.3) `modules/iam/role.go`
```go
package iam

import (
    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
    iamrole "github.com/cdktf/cdktf-provider-aws-go/aws/v21/iamrole"
    iampa "github.com/cdktf/cdktf-provider-aws-go/aws/v21/iamrolepolicyattachment"
)

type Role struct{ role iamrole.IamRole }

func NewLambdaRole(scope constructs.Construct, id string) *Role {
    s := constructs.NewConstruct(scope, jsii.String(id))

    assume := `{"Version":"2012-10-17","Statement":[{"Effect":"Allow","Principal":{"Service":"lambda.amazonaws.com"},"Action":"sts:AssumeRole"}]}`

    r := iamrole.NewIamRole(s, jsii.String("LambdaExecRole"), &iamrole.IamRoleConfig{
        Name:             jsii.String("orders-api-lambda-role"),
        AssumeRolePolicy: jsii.String(assume),
    })

    // Basic logging
    iampa.NewIamRolePolicyAttachment(s, jsii.String("BasicLogs"), &iampa.IamRolePolicyAttachmentConfig{
        Role:      r.Name(),
        PolicyArn: jsii.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
    })

    return &Role{role: r}
}

func (r *Role) Arn() *string { return r.role.Arn() }
```

> Keep policies minimal. If your handler needs DynamoDB access directly, attach the specific permissions here later.

### 6.4) `modules/lambda/lambda.go`
```go
package lambda

import (
    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
    lambdafn "github.com/cdktf/cdktf-provider-aws-go/aws/v21/lambdafunction"
)

type LambdaConfig struct {
    Name, Runtime, Handler string
    Memory, Timeout        string // keep as string for simplicity

    S3Bucket *string
    S3Key    string

    Env               string
    TableOrders       *string
    TableOrderItems   *string
    RoleArn           *string
}

type Fn struct{ fn lambdafn.LambdaFunction }

func NewApiLambda(scope constructs.Construct, id string, c *LambdaConfig) *Fn {
    s := constructs.NewConstruct(scope, jsii.String(id))

    mem := toNum(c.Memory, 128)
    tout := toNum(c.Timeout, 10)

    fn := lambdafn.NewLambdaFunction(s, jsii.String("Fn"), &lambdafn.LambdaFunctionConfig{
        FunctionName: jsii.String(c.Name),
        Role:         c.RoleArn,
        Runtime:      jsii.String(c.Runtime),
        Handler:      jsii.String(c.Handler),
        MemorySize:   jsii.Number(float64(mem)),
        Timeout:      jsii.Number(float64(tout)),
        S3Bucket:     c.S3Bucket,
        S3Key:        jsii.String(c.S3Key),
        Environment: &lambdafn.LambdaFunctionEnvironment{
            Variables: &map[string]*string{
                "ENV":               jsii.String(c.Env),
                "TABLE_ORDERS":      c.TableOrders,
                "TABLE_ORDER_ITEMS": c.TableOrderItems,
            },
        },
        Architectures: &[]*string{jsii.String("x86_64")},
    })

    return &Fn{fn: fn}
}

func (f *Fn) Arn() *string          { return f.fn.Arn() }
func (f *Fn) FunctionName() *string { return f.fn.FunctionName() }

func toNum(s string, def int) int { if s == "" { return def }; var n int; _, _ = fmt.Sscanf(s, "%d", &n); if n==0 {return def}; return n }
```

### 6.5) `modules/apigateway/api.go`
```go
package apigateway

import (
    "github.com/aws/constructs-go/constructs/v10"
    "github.com/aws/jsii-runtime-go"
    api "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2api"
    integ "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2integration"
    route "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2route"
    stage "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2stage"
    lperm "github.com/cdktf/cdktf-provider-aws-go/aws/v21/lambdapermission"
)

type ApiConfig struct {
    Env string
    LambdaArn  *string
    LambdaName *string
}

type Http struct {
    api api.Apigatewayv2Api
}

func NewHttpApi(scope constructs.Construct, id string, c *ApiConfig) *Http {
    s := constructs.NewConstruct(scope, jsii.String(id))

    a := api.NewApigatewayv2Api(s, jsii.String("Api"), &api.Apigatewayv2ApiConfig{
        Name:         jsii.String("orders-http"),
        ProtocolType: jsii.String("HTTP"),
    })

    ig := integ.NewApigatewayv2Integration(s, jsii.String("Integration"), &integ.Apigatewayv2IntegrationConfig{
        ApiId:                a.Id(),
        IntegrationType:      jsii.String("AWS_PROXY"),
        IntegrationUri:       c.LambdaArn,
        PayloadFormatVersion: jsii.String("2.0"),
    })

    // Minimal routes mapped to the Lambda (expand as needed)
    route.NewApigatewayv2Route(s, jsii.String("Orders"), &route.Apigatewayv2RouteConfig{
        ApiId:    a.Id(),
        RouteKey: jsii.String("ANY /orders"),
        Target:   jsii.String("integrations/" + *ig.Id()),
    })
    route.NewApigatewayv2Route(s, jsii.String("OrderItems"), &route.Apigatewayv2RouteConfig{
        ApiId:    a.Id(),
        RouteKey: jsii.String("ANY /orders/{orderId}/items"),
        Target:   jsii.String("integrations/" + *ig.Id()),
    })

    stage.NewApigatewayv2Stage(s, jsii.String("Stage"), &stage.Apigatewayv2StageConfig{
        ApiId:      a.Id(),
        Name:       jsii.String(c.Env),
        AutoDeploy: jsii.Bool(true),
    })

    lperm.NewLambdaPermission(s, jsii.String("AllowApiGw"), &lperm.LambdaPermissionConfig{
        Action:       jsii.String("lambda:InvokeFunction"),
        FunctionName: c.LambdaName,
        Principal:    jsii.String("apigateway.amazonaws.com"),
        SourceArn:    a.Arn(),
    })

    return &Http{api: a}
}

func (h *Http) Endpoint() *string { return h.api.ApiEndpoint() }
```

---

## 7) Build & upload the Lambda artifact (`api.zip`)

From the **repo root** (not inside `terraform/`), build the custom runtime binary and zip it:
```bash
GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o bootstrap ./
zip api.zip bootstrap
```
After the artifact bucket exists (next section step 4), upload:
```bash
aws s3 cp api.zip s3://<S3_BUCKET_PREFIX>-<ENV>/<LAMBDA_S3_KEY>
# ex.: aws s3 cp api.zip s3://go-lambda-artifacts-dev/api.zip
```

---

## 8) Synthesize & apply with Terraform

From `terraform/`:
```bash
# 1) Synthesize CDKTF -> Terraform JSON
GOFLAGS=-mod=mod go run ./main.go

# 2) Initialize Terraform in the synthesized stack
cd cdktf.out/stacks/infra
terraform init

# 3) First create only the artifact bucket (so you can upload api.zip)
terraform apply -target=aws_s3_bucket.ArtifactsBucket -auto-approve

# 4) Upload the artifact (in another shell, see step 7)
aws s3 cp api.zip s3://<S3_BUCKET_PREFIX>-<ENV>/<LAMBDA_S3_KEY>

# 5) Apply the rest
terraform apply
```
Grab the API URL from the Terraform outputs or the AWS console (API Gateway → Stages → your `ENV`).

Destroy when done:
```bash
terraform destroy
```
> If the bucket has objects, set `S3_FORCE_DESTROY=true` and re‑apply before destroying.

---

## Notes & pitfalls

- Runtime: **`provided.al2023`** + handler `bootstrap` (custom Go runtime).
- Keep IAM policies tight; start basic and iterate with least privilege.
- S3 bucket names are global; `prefix-env` helps avoid collisions.
- Consider configuring an **S3 backend + DynamoDB locks** later for team state.

Example backend (add in `infra.NewStack` when ready):
```go
cdktf.NewS3Backend(stack, &cdktf.S3BackendConfig{
  Bucket: jsii.String("tf-state-bucket"),
  Key:    jsii.String("serverless-api/terraform.tfstate"),
  Region: jsii.String("us-east-1"),
  DynamoDbTable: jsii.String("terraform-locks"),
})
```

---

## Full project link

The complete, ready‑to‑run code lives here (same structure as above):

➡️ https://github.com/markgerald/go-serverless-api-terraform/tree/terraform-cdk-go

---

## Coming in Part 2

- **GitHub Actions** pipelines for:
  - Building, testing and releasing the **Go API** artifact
  - Synthesizing and deploying **CDKTF** changes
- **GitOps**: PR‑driven infra changes with plans as checks, environments, and approvals.


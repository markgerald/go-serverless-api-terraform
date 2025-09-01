package db

import (
	"context"
	"errors"

	"github.com/aws/aws-sdk-go-v2/aws"
	awsconfig "github.com/aws/aws-sdk-go-v2/config"
	"github.com/aws/aws-sdk-go-v2/service/dynamodb"

	"go-serverless-api-terraform/internal/config"
)

// NewDynamoClient creates a DynamoDB client, optionally targeting a custom endpoint (e.g., DynamoDB Local)
func NewDynamoClient(ctx context.Context, cfg *config.Config) (*dynamodb.Client, error) {
	if cfg == nil {
		return nil, errors.New("config is nil")
	}
	if cfg.DynamoEndpoint != "" {
		awsCfg, err := awsconfig.LoadDefaultConfig(ctx,
			awsconfig.WithRegion(cfg.AWSRegion),
			awsconfig.WithEndpointResolverWithOptions(
				aws.EndpointResolverWithOptionsFunc(func(service, region string, options ...any) (aws.Endpoint, error) {
					return aws.Endpoint{URL: cfg.DynamoEndpoint}, nil
				}),
			),
		)
		if err != nil {
			return nil, err
		}
		return dynamodb.NewFromConfig(awsCfg), nil
	}
	awsCfg, err := awsconfig.LoadDefaultConfig(ctx, awsconfig.WithRegion(cfg.AWSRegion))
	if err != nil {
		return nil, err
	}
	return dynamodb.NewFromConfig(awsCfg), nil
}

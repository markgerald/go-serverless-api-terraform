package dynamodb

import (
	"github.com/aws/jsii-runtime-go"
	awsDyn "github.com/cdktf/cdktf-provider-aws-go/aws/v21/dynamodbtable"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type AttributeDef struct {
	Name string
	Type string // "S"|"N"|"B"
}

type DynamoTableProps struct {
	Name        string
	HashKey     AttributeDef
	RangeKey    *AttributeDef
	BillingMode string
	PitrEnabled bool
	Tags        map[string]string
}

func NewDynamoTable(stack cdktf.TerraformStack, id string, p *DynamoTableProps) {
	// attributes
	attrs := &[]*awsDyn.DynamodbTableAttribute{
		{Name: &p.HashKey.Name, Type: &p.HashKey.Type},
	}
	var rangeKey *string
	if p.RangeKey != nil {
		*attrs = append(*attrs, &awsDyn.DynamodbTableAttribute{Name: &p.RangeKey.Name, Type: &p.RangeKey.Type})
		rangeKey = &p.RangeKey.Name
	}

	// convert tags to map[string]*string
	tagMap := map[string]*string{}
	for k, v := range p.Tags {
		vv := v
		tagMap[k] = &vv
	}

	cfg := &awsDyn.DynamodbTableConfig{
		Name:        &p.Name,
		BillingMode: &p.BillingMode,
		HashKey:     &p.HashKey.Name,
		Attribute:   attrs,
		Tags:        &tagMap,
	}
	if rangeKey != nil {
		cfg.RangeKey = rangeKey
	}
	if p.PitrEnabled {
		cfg.PointInTimeRecovery = &awsDyn.DynamodbTablePointInTimeRecovery{Enabled: jsii.Bool(true)}
	}

	awsDyn.NewDynamodbTable(stack, &id, cfg)
}

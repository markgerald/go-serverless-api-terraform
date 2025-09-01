package lambda

import (
	"github.com/aws/jsii-runtime-go"
	iamrole "github.com/cdktf/cdktf-provider-aws-go/aws/v21/iamrole"
	iampolicyattach "github.com/cdktf/cdktf-provider-aws-go/aws/v21/iamrolepolicyattachment"
	awsLambda "github.com/cdktf/cdktf-provider-aws-go/aws/v21/lambdafunction"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type LambdaProps struct {
	FunctionName string
	S3Bucket     string
	S3Key        string
	Runtime      string
	Handler      string
	MemoryMb     int
	TimeoutS     int
	Environment  map[string]string
	Tags         map[string]string
	// Optional: if provided, use this role ARN for the Lambda. If empty, a basic execution role will be created.
	RoleArn *string
}

type Function interface {
	Arn() string
}

type functionImpl struct{ arn string }

func (f *functionImpl) Arn() string { return f.arn }

func NewLambdaFunction(stack cdktf.TerraformStack, id string, p *LambdaProps) Function {
	// Convert maps to *map[string]*string
	env := map[string]*string{}
	for k, v := range p.Environment {
		vv := v
		env[k] = &vv
	}
	tags := map[string]*string{}
	for k, v := range p.Tags {
		vv := v
		tags[k] = &vv
	}

	// Determine execution role ARN
	var roleArn *string
	if p.RoleArn != nil && *p.RoleArn != "" {
		roleArn = p.RoleArn
	} else {
		// Create a minimal IAM role for Lambda with basic logging permissions
		roleName := p.FunctionName + "-exec"
		assumeRolePolicy := `{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": { "Service": "lambda.amazonaws.com" },
      "Action": "sts:AssumeRole"
    }
  ]
}`
		role := iamrole.NewIamRole(stack, jsii.String(id+"Role"), &iamrole.IamRoleConfig{
			Name:             jsii.String(roleName),
			AssumeRolePolicy: jsii.String(assumeRolePolicy),
		})
		// Attach AWS managed basic execution policy for CloudWatch Logs
		iampolicyattach.NewIamRolePolicyAttachment(stack, jsii.String(id+"RoleBasicLogs"), &iampolicyattach.IamRolePolicyAttachmentConfig{
			Role:      role.Name(),
			PolicyArn: jsii.String("arn:aws:iam::aws:policy/service-role/AWSLambdaBasicExecutionRole"),
		})
		roleArn = role.Arn()
	}

	cfg := &awsLambda.LambdaFunctionConfig{
		FunctionName: &p.FunctionName,
		S3Bucket:     &p.S3Bucket,
		S3Key:        &p.S3Key,
		Runtime:      &p.Runtime,
		Handler:      &p.Handler,
		Role:         roleArn,
		MemorySize:   jsii.Number(float64(p.MemoryMb)),
		Timeout:      jsii.Number(float64(p.TimeoutS)),
		Environment: &awsLambda.LambdaFunctionEnvironment{
			Variables: &env,
		},
		Tags: &tags,
	}

	fn := awsLambda.NewLambdaFunction(stack, &id, cfg)
	// In CDKTF, getters return *string; dereference safely if present
	arn := ""
	if fn.Arn() != nil {
		arn = *fn.Arn()
	}
	return &functionImpl{arn: arn}
}

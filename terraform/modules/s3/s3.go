package s3

import (
	"fmt"
	"github.com/aws/jsii-runtime-go"
	awsS3 "github.com/cdktf/cdktf-provider-aws-go/aws/v21/s3bucket"
	s3pab "github.com/cdktf/cdktf-provider-aws-go/aws/v21/s3bucketpublicaccessblock"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type Versioning struct{ Enabled bool }

type ServerSideEncryption struct {
	Enabled  bool
	KmsKeyId *string
}

type PublicAccessBlock struct {
	BlockPublicAcls       bool
	IgnorePublicAcls      bool
	BlockPublicPolicy     bool
	RestrictPublicBuckets bool
}

type BucketProps struct {
	Prefix       *string
	Environment  *string
	ForceDestroy bool
	Versioning   *Versioning
	Encryption   *ServerSideEncryption
	PublicAccess *PublicAccessBlock
	Tags         map[string]string
}

type Bucket interface{ Bucket() string }

type bucketImpl struct{ name string }

func (b *bucketImpl) Bucket() string { return b.name }

func NewBucket(stack cdktf.TerraformStack, id string, p *BucketProps) Bucket {
	name := fmt.Sprintf("%s-%s", *p.Prefix, *p.Environment)
	tags := map[string]*string{}
	for k, v := range p.Tags {
		vv := v
		tags[k] = &vv
	}

	b := awsS3.NewS3Bucket(stack, &id, &awsS3.S3BucketConfig{
		Bucket:       &name,
		ForceDestroy: jsii.Bool(p.ForceDestroy),
		Tags:         &tags,
	})

	// Public access block
	if p.PublicAccess != nil {
		s3pab.NewS3BucketPublicAccessBlock(stack, jsii.String(id+"PublicAccessBlock"), &s3pab.S3BucketPublicAccessBlockConfig{
			Bucket:                b.Bucket(),
			BlockPublicAcls:       jsii.Bool(p.PublicAccess.BlockPublicAcls),
			BlockPublicPolicy:     jsii.Bool(p.PublicAccess.BlockPublicPolicy),
			IgnorePublicAcls:      jsii.Bool(p.PublicAccess.IgnorePublicAcls),
			RestrictPublicBuckets: jsii.Bool(p.PublicAccess.RestrictPublicBuckets),
		})
	}

	// Versioning and Encryption can be added with auxiliary resources; omitted for brevity.

	return &bucketImpl{name: *b.Bucket()}
}

package apigateway

import (
	"strconv"

	"github.com/aws/jsii-runtime-go"
	apigwv2api "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2api"
	apigwv2integration "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2integration"
	apigwv2route "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2route"
	apigwv2stage "github.com/cdktf/cdktf-provider-aws-go/aws/v21/apigatewayv2stage"
	lambdaPerm "github.com/cdktf/cdktf-provider-aws-go/aws/v21/lambdapermission"
	"github.com/hashicorp/terraform-cdk-go/cdktf"
)

type RouteSpec struct{ Method, Path string }

type HttpApiProps struct {
	Name               string
	Description        *string
	TargetLambdaArn    string
	Routes             []RouteSpec
	StageName          string
	AutoDeploy         bool
	AllowInvokeFromApi bool
	Tags               map[string]string
}

func NewHttpApiWithLambda(stack cdktf.TerraformStack, id string, p *HttpApiProps) {
	tags := map[string]*string{}
	for k, v := range p.Tags {
		vv := v
		tags[k] = &vv
	}

	api := apigwv2api.NewApigatewayv2Api(stack, &id, &apigwv2api.Apigatewayv2ApiConfig{
		Name:         &p.Name,
		ProtocolType: jsii.String("HTTP"),
		Description:  p.Description,
		Tags:         &tags,
	})

	integ := apigwv2integration.NewApigatewayv2Integration(stack, jsii.String(id+"Integration"), &apigwv2integration.Apigatewayv2IntegrationConfig{
		ApiId:                api.Id(),
		IntegrationType:      jsii.String("AWS_PROXY"),
		IntegrationUri:       &p.TargetLambdaArn,
		PayloadFormatVersion: jsii.String("2.0"),
	})

	for i, r := range p.Routes {
		routeKey := r.Method + " " + r.Path
		routeId := id + "Route" + strconv.Itoa(i)
		apigwv2route.NewApigatewayv2Route(stack, jsii.String(routeId), &apigwv2route.Apigatewayv2RouteConfig{
			ApiId:    api.Id(),
			RouteKey: &routeKey,
			Target:   jsii.String("integrations/" + *integ.Id()),
		})
	}

	apigwv2stage.NewApigatewayv2Stage(stack, jsii.String(id+"Stage"), &apigwv2stage.Apigatewayv2StageConfig{
		ApiId:      api.Id(),
		Name:       &p.StageName,
		AutoDeploy: jsii.Bool(p.AutoDeploy),
	})

	if p.AllowInvokeFromApi {
		lambdaPerm.NewLambdaPermission(stack, jsii.String(id+"LambdaPerm"), &lambdaPerm.LambdaPermissionConfig{
			StatementId:  jsii.String("AllowInvokeFromHttpApi"),
			Action:       jsii.String("lambda:InvokeFunction"),
			FunctionName: &p.TargetLambdaArn,
			Principal:    jsii.String("apigateway.amazonaws.com"),
		})
	}
}

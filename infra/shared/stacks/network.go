package stacks

import (
	"infra/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// NetworkStackProps defines the properties for the network stack
type NetworkStackProps struct {
	awscdk.StackProps
	Environment string
	Config      config.EnvironmentConfig
}

// NetworkStackOutputs contains the outputs from the network stack
type NetworkStackOutputs struct {
	Stack awscdk.Stack
	Vpc   awsec2.Vpc
}

// NewNetworkStack creates a new network stack with VPC, subnets, and security groups
func NewNetworkStack(scope constructs.Construct, id string, props *NetworkStackProps) *NetworkStackOutputs {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create VPC with public and private subnets across 2 AZs
	vpc := awsec2.NewVpc(stack, jsii.String("Vpc"), &awsec2.VpcProps{
		MaxAzs:      jsii.Number(2),
		NatGateways: jsii.Number(props.Config.NatGateways),
		SubnetConfiguration: &[]*awsec2.SubnetConfiguration{
			{
				Name:       jsii.String("Public"),
				SubnetType: awsec2.SubnetType_PUBLIC,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Private"),
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
				CidrMask:   jsii.Number(24),
			},
			{
				Name:       jsii.String("Isolated"),
				SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
				CidrMask:   jsii.Number(24),
			},
		},
	})

	// Add VPC Flow Logs for security monitoring
	vpc.AddFlowLog(jsii.String("FlowLog"), &awsec2.FlowLogOptions{
		Destination: awsec2.FlowLogDestination_ToCloudWatchLogs(nil, nil),
		TrafficType: awsec2.FlowLogTrafficType_ALL,
	})

	// Output the VPC ID
	awscdk.NewCfnOutput(stack, jsii.String("VpcId"), &awscdk.CfnOutputProps{
		Value:       vpc.VpcId(),
		Description: jsii.String("VPC ID"),
		ExportName:  jsii.String(id + "-VpcId"),
	})

	return &NetworkStackOutputs{
		Stack: stack,
		Vpc:   vpc,
	}
}

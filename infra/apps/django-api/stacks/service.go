package stacks

import (
	"infra/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsapplicationautoscaling"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecr"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecs"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsecspatterns"
	"github.com/aws/aws-cdk-go/awscdk/v2/awselasticloadbalancingv2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// ServiceStackProps defines the properties for the Django service stack
type ServiceStackProps struct {
	awscdk.StackProps
	Environment string
	Vpc         awsec2.Vpc
	Cluster     awsrds.DatabaseCluster
	Bucket      awss3.Bucket
	DbSecret    awssecretsmanager.ISecret
	Config      config.EnvironmentConfig
}

// ServiceStackOutputs contains the outputs from the service stack
type ServiceStackOutputs struct {
	Stack      awscdk.Stack
	EcsCluster awsecs.Cluster
	Service    awsecspatterns.ApplicationLoadBalancedFargateService
}

// NewServiceStack creates a new Django service stack with ECS Fargate and ALB
func NewServiceStack(scope constructs.Construct, id string, props *ServiceStackProps) *ServiceStackOutputs {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Create ECS Cluster
	ecsCluster := awsecs.NewCluster(stack, jsii.String("EcsCluster"), &awsecs.ClusterProps{
		Vpc:                 props.Vpc,
		ContainerInsightsV2: awsecs.ContainerInsights_ENHANCED,
	})

	// Create or reference ECR repository
	repo := awsecr.NewRepository(stack, jsii.String("DjangoRepo"), &awsecr.RepositoryProps{
		RepositoryName:     jsii.String("django-" + props.Environment),
		ImageScanOnPush:    jsii.Bool(true),
		ImageTagMutability: awsecr.TagMutability_MUTABLE,
		LifecycleRules: &[]*awsecr.LifecycleRule{
			{
				MaxImageCount: jsii.Number(10),
				Description:   jsii.String("Keep only 10 images"),
			},
		},
	})

	// Create the Fargate service with ALB
	service := awsecspatterns.NewApplicationLoadBalancedFargateService(stack, jsii.String("DjangoService"),
		&awsecspatterns.ApplicationLoadBalancedFargateServiceProps{
			Cluster:        ecsCluster,
			Cpu:            jsii.Number(props.Config.CPU),
			MemoryLimitMiB: jsii.Number(props.Config.Memory),
			DesiredCount:   jsii.Number(props.Config.DesiredCount),
			TaskImageOptions: &awsecspatterns.ApplicationLoadBalancedTaskImageOptions{
				Image:         awsecs.ContainerImage_FromEcrRepository(repo, jsii.String("latest")),
				ContainerPort: jsii.Number(8000),
				Environment: &map[string]*string{
					"DJANGO_SETTINGS_MODULE": jsii.String("config.settings.production"),
					"ALLOWED_HOSTS":          jsii.String("*"),
				},
				Secrets: &map[string]awsecs.Secret{
					"DATABASE_URL": awsecs.Secret_FromSecretsManager(props.DbSecret, nil),
				},
			},
			PublicLoadBalancer: jsii.Bool(true),
			AssignPublicIp:     jsii.Bool(false), // Tasks in private subnets
			TaskSubnets: &awsec2.SubnetSelection{
				SubnetType: awsec2.SubnetType_PRIVATE_WITH_EGRESS,
			},
			CircuitBreaker: &awsecs.DeploymentCircuitBreaker{
				Enable:   jsii.Bool(true),
				Rollback: jsii.Bool(true),
			},
			MinHealthyPercent:      jsii.Number(100),
			MaxHealthyPercent:      jsii.Number(200),
			HealthCheckGracePeriod: awscdk.Duration_Seconds(jsii.Number(60)),
		},
	)

	// Configure health check
	service.TargetGroup().ConfigureHealthCheck(&awselasticloadbalancingv2.HealthCheck{
		Path:                    jsii.String("/health/"),
		Interval:                awscdk.Duration_Seconds(jsii.Number(30)),
		Timeout:                 awscdk.Duration_Seconds(jsii.Number(5)),
		HealthyThresholdCount:   jsii.Number(2),
		UnhealthyThresholdCount: jsii.Number(3),
	})

	// Grant S3 access to the task
	props.Bucket.GrantReadWrite(service.TaskDefinition().TaskRole(), nil)

	// Auto-scaling
	scaling := service.Service().AutoScaleTaskCount(&awsapplicationautoscaling.EnableScalingProps{
		MinCapacity: jsii.Number(props.Config.MinCount),
		MaxCapacity: jsii.Number(props.Config.MaxCount),
	})

	scaling.ScaleOnCpuUtilization(jsii.String("CpuScaling"), &awsecs.CpuUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(70),
		ScaleInCooldown:          awscdk.Duration_Seconds(jsii.Number(60)),
		ScaleOutCooldown:         awscdk.Duration_Seconds(jsii.Number(60)),
	})

	scaling.ScaleOnMemoryUtilization(jsii.String("MemoryScaling"), &awsecs.MemoryUtilizationScalingProps{
		TargetUtilizationPercent: jsii.Number(70),
		ScaleInCooldown:          awscdk.Duration_Seconds(jsii.Number(60)),
		ScaleOutCooldown:         awscdk.Duration_Seconds(jsii.Number(60)),
	})

	// Outputs
	awscdk.NewCfnOutput(stack, jsii.String("LoadBalancerDNS"), &awscdk.CfnOutputProps{
		Value:       service.LoadBalancer().LoadBalancerDnsName(),
		Description: jsii.String("Application Load Balancer DNS"),
		ExportName:  jsii.String(id + "-AlbDns"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("EcrRepositoryUri"), &awscdk.CfnOutputProps{
		Value:       repo.RepositoryUri(),
		Description: jsii.String("ECR Repository URI"),
		ExportName:  jsii.String(id + "-EcrUri"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ClusterName"), &awscdk.CfnOutputProps{
		Value:       ecsCluster.ClusterName(),
		Description: jsii.String("ECS Cluster Name"),
		ExportName:  jsii.String(id + "-ClusterName"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ServiceName"), &awscdk.CfnOutputProps{
		Value:       service.Service().ServiceName(),
		Description: jsii.String("ECS Service Name"),
		ExportName:  jsii.String(id + "-ServiceName"),
	})

	return &ServiceStackOutputs{
		Stack:      stack,
		EcsCluster: ecsCluster,
		Service:    service,
	}
}

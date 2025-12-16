package stacks

import (
	"infra/config"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsec2"
	"github.com/aws/aws-cdk-go/awscdk/v2/awsrds"
	"github.com/aws/aws-cdk-go/awscdk/v2/awss3"
	"github.com/aws/aws-cdk-go/awscdk/v2/awssecretsmanager"
	"github.com/aws/constructs-go/constructs/v10"
	"github.com/aws/jsii-runtime-go"
)

// DataStackProps defines the properties for the data stack
type DataStackProps struct {
	awscdk.StackProps
	Environment string
	Vpc         awsec2.Vpc
	Config      config.EnvironmentConfig
}

// DataStackOutputs contains the outputs from the data stack
type DataStackOutputs struct {
	Stack    awscdk.Stack
	Cluster  awsrds.DatabaseCluster
	Bucket   awss3.Bucket
	DbSecret awssecretsmanager.ISecret
}

// NewDataStack creates a new data stack with Aurora PostgreSQL and S3
func NewDataStack(scope constructs.Construct, id string, props *DataStackProps) *DataStackOutputs {
	var sprops awscdk.StackProps
	if props != nil {
		sprops = props.StackProps
	}
	stack := awscdk.NewStack(scope, &id, &sprops)

	// Security group for database
	dbSecurityGroup := awsec2.NewSecurityGroup(stack, jsii.String("DbSecurityGroup"), &awsec2.SecurityGroupProps{
		Vpc:              props.Vpc,
		Description:      jsii.String("Security group for Aurora PostgreSQL"),
		AllowAllOutbound: jsii.Bool(false),
	})

	// Allow inbound PostgreSQL from private subnets
	dbSecurityGroup.AddIngressRule(
		awsec2.Peer_Ipv4(props.Vpc.VpcCidrBlock()),
		awsec2.Port_Tcp(jsii.Number(5432)),
		jsii.String("Allow PostgreSQL from VPC"),
		jsii.Bool(false),
	)

	// Create Aurora PostgreSQL cluster
	cluster := awsrds.NewDatabaseCluster(stack, jsii.String("AuroraCluster"), &awsrds.DatabaseClusterProps{
		Engine: awsrds.DatabaseClusterEngine_AuroraPostgres(&awsrds.AuroraPostgresClusterEngineProps{
			Version: awsrds.AuroraPostgresEngineVersion_VER_16_4(),
		}),
		Credentials:         awsrds.Credentials_FromGeneratedSecret(jsii.String("postgres"), nil),
		DefaultDatabaseName: jsii.String("django"),
		Writer: awsrds.ClusterInstance_Provisioned(jsii.String("Writer"), &awsrds.ProvisionedClusterInstanceProps{
			InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_R6G, awsec2.InstanceSize_LARGE),
		}),
		Readers: &[]awsrds.IClusterInstance{
			awsrds.ClusterInstance_Provisioned(jsii.String("Reader"), &awsrds.ProvisionedClusterInstanceProps{
				InstanceType: awsec2.InstanceType_Of(awsec2.InstanceClass_R6G, awsec2.InstanceSize_LARGE),
			}),
		},
		Vpc: props.Vpc,
		VpcSubnets: &awsec2.SubnetSelection{
			SubnetType: awsec2.SubnetType_PRIVATE_ISOLATED,
		},
		SecurityGroups:     &[]awsec2.ISecurityGroup{dbSecurityGroup},
		StorageEncrypted:   jsii.Bool(true),
		DeletionProtection: jsii.Bool(props.Environment == "production" || props.Environment == "sandbox"),
		Backup: &awsrds.BackupProps{
			Retention: awscdk.Duration_Days(jsii.Number(7)),
		},
		RemovalPolicy: awscdk.RemovalPolicy_SNAPSHOT,
	})

	// S3 bucket for static/media files
	// Let CloudFormation generate a unique bucket name to avoid conflicts
	bucket := awss3.NewBucket(stack, jsii.String("MediaBucket"), &awss3.BucketProps{
		Versioned:         jsii.Bool(true),
		Encryption:        awss3.BucketEncryption_S3_MANAGED,
		BlockPublicAccess: awss3.BlockPublicAccess_BLOCK_ALL(),
		EnforceSSL:        jsii.Bool(true),
		RemovalPolicy:     awscdk.RemovalPolicy_RETAIN,
	})

	// Outputs
	awscdk.NewCfnOutput(stack, jsii.String("ClusterEndpoint"), &awscdk.CfnOutputProps{
		Value:       cluster.ClusterEndpoint().Hostname(),
		Description: jsii.String("Aurora Cluster Endpoint"),
		ExportName:  jsii.String(id + "-ClusterEndpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("ClusterReaderEndpoint"), &awscdk.CfnOutputProps{
		Value:       cluster.ClusterReadEndpoint().Hostname(),
		Description: jsii.String("Aurora Cluster Reader Endpoint"),
		ExportName:  jsii.String(id + "-ClusterReaderEndpoint"),
	})

	awscdk.NewCfnOutput(stack, jsii.String("BucketName"), &awscdk.CfnOutputProps{
		Value:       bucket.BucketName(),
		Description: jsii.String("S3 Bucket Name"),
		ExportName:  jsii.String(id + "-BucketName"),
	})

	return &DataStackOutputs{
		Stack:    stack,
		Cluster:  cluster,
		Bucket:   bucket,
		DbSecret: cluster.Secret(),
	}
}

package main

import (
	"os"

	api "infra/apps/api/stacks"
	"infra/config"
	shared "infra/shared/stacks"

	"github.com/aws/aws-cdk-go/awscdk/v2"
	"github.com/aws/jsii-runtime-go"
)

func main() {
	defer jsii.Close()

	app := awscdk.NewApp(nil)

	// Get account from environment or use default
	account := os.Getenv("CDK_DEFAULT_ACCOUNT")
	if account == "" {
		account = os.Getenv("AWS_ACCOUNT_ID")
	}

	// Create stacks for each environment
	for envName, cfg := range config.Environments {
		primaryEnv := &awscdk.Environment{
			Account: jsii.String(account),
			Region:  jsii.String("us-east-1"),
		}

		// Shared Network stack
		network := shared.NewNetworkStack(app, envName+"-shared-network", &shared.NetworkStackProps{
			StackProps: awscdk.StackProps{
				Env:         primaryEnv,
				Description: jsii.String("Shared network infrastructure for " + envName),
			},
			Environment: envName,
			Config:      cfg,
		})

		// Shared Data stack (Aurora, S3) - depends on network
		data := shared.NewDataStack(app, envName+"-shared-data", &shared.DataStackProps{
			StackProps: awscdk.StackProps{
				Env:         primaryEnv,
				Description: jsii.String("Shared data infrastructure for " + envName),
			},
			Environment: envName,
			Vpc:         network.Vpc,
			Config:      cfg,
		})
		data.Stack.AddDependency(network.Stack, jsii.String("Data stack requires VPC from network stack"))

		// API App stack (ECS, ALB) - depends on network and data
		apiStack := api.NewServiceStack(app, envName+"-app-api", &api.ServiceStackProps{
			StackProps: awscdk.StackProps{
				Env:         primaryEnv,
				Description: jsii.String("API service for " + envName),
			},
			Environment: envName,
			Vpc:         network.Vpc,
			Cluster:     data.Cluster,
			Bucket:      data.Bucket,
			DbSecret:    data.DbSecret,
			Config:      cfg,
		})
		apiStack.Stack.AddDependency(data.Stack, jsii.String("API stack requires database and bucket from data stack"))
	}

	app.Synth(nil)
}

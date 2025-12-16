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
		})

		// Shared Data stack (Aurora, S3)
		data := shared.NewDataStack(app, envName+"-shared-data", &shared.DataStackProps{
			StackProps: awscdk.StackProps{
				Env:         primaryEnv,
				Description: jsii.String("Shared data infrastructure for " + envName),
			},
			Environment: envName,
			Vpc:         network.Vpc,
			Config:      cfg,
		})

		// API App stack (ECS, ALB)
		api.NewServiceStack(app, envName+"-app-api", &api.ServiceStackProps{
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
	}

	app.Synth(nil)
}

package main

import (
	"os"

	"infra/config"
	djangoapi "infra/apps/django-api/stacks"
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

		// Django API App stack (ECS, ALB)
		djangoapi.NewServiceStack(app, envName+"-app-django-api", &djangoapi.ServiceStackProps{
			StackProps: awscdk.StackProps{
				Env:         primaryEnv,
				Description: jsii.String("Django API service for " + envName),
			},
			Environment: envName,
			Vpc:         network.Vpc,
			Cluster:     data.Cluster,
			Bucket:      data.Bucket,
			DbSecret:    data.DbSecret,
			Config:      cfg,
		})

		// DR stacks for environments with DR enabled
		if cfg.EnableDR {
			drEnv := &awscdk.Environment{
				Account: jsii.String(account),
				Region:  jsii.String(cfg.DRRegion),
			}

			// DR Shared Network stack
			drNetwork := shared.NewNetworkStack(app, envName+"-dr-shared-network", &shared.NetworkStackProps{
				StackProps: awscdk.StackProps{
					Env:         drEnv,
					Description: jsii.String("DR shared network infrastructure for " + envName),
				},
				Environment: envName + "-dr",
			})

			// DR Shared Data stack
			drData := shared.NewDataStack(app, envName+"-dr-shared-data", &shared.DataStackProps{
				StackProps: awscdk.StackProps{
					Env:         drEnv,
					Description: jsii.String("DR shared data infrastructure for " + envName),
				},
				Environment: envName + "-dr",
				Vpc:         drNetwork.Vpc,
				Config:      cfg,
			})

			// DR Django API App stack (warm standby - 1 minimal task)
			drConfig := cfg
			drConfig.DesiredCount = 1 // Warm standby
			drConfig.MinCount = 1

			djangoapi.NewServiceStack(app, envName+"-dr-app-django-api", &djangoapi.ServiceStackProps{
				StackProps: awscdk.StackProps{
					Env:         drEnv,
					Description: jsii.String("DR Django API service for " + envName),
				},
				Environment: envName + "-dr",
				Vpc:         drNetwork.Vpc,
				Cluster:     drData.Cluster,
				Bucket:      drData.Bucket,
				DbSecret:    drData.DbSecret,
				Config:      drConfig,
			})
		}
	}

	app.Synth(nil)
}

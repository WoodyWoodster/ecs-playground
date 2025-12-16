# ECS Infrastructure (CDK Go)

Infrastructure as Code for deploying applications on AWS ECS Fargate with Aurora PostgreSQL.

## Architecture

```
infra/
├── shared/stacks/          # Shared infrastructure
│   ├── network.go          # VPC, subnets, NAT gateways
│   └── data.go             # Aurora PostgreSQL, S3
├── apps/api/stacks/        # App-specific infrastructure
│   └── service.go          # ECS Fargate, ALB, ECR
├── config/
│   └── environments.go     # Environment configurations
└── infra.go                # Entry point
```

## Environments & Branching Strategy

| Environment | Branch  | Region    | Purpose              |
| ----------- | ------- | --------- | -------------------- |
| production  | main    | us-east-1 | Production workloads |
| sandbox     | main    | us-east-1 | Customer API testing |
| demo        | main    | us-east-1 | Sales demos          |
| uat         | develop | us-east-1 | Internal testing     |

```
main branch ──────► production, sandbox, demo
develop branch ───► uat
```

## Stack Naming Convention

```
{environment}-shared-network     # VPC, subnets
{environment}-shared-data        # Aurora PostgreSQL, S3
{environment}-app-api            # ECS service, ALB
```

## Prerequisites

- AWS CLI configured with appropriate credentials
- Go 1.21+
- Bun (for CDK CLI): `curl -fsSL https://bun.sh/install | bash`

## Deployment Commands

| Command                   | Description                                 |
| ------------------------- | ------------------------------------------- |
| `make bootstrap`          | Bootstrap CDK (run once per account/region) |
| `make build`              | Build Go CDK                                |
| `make list`               | List all stacks                             |
| `make synth`              | Synthesize CloudFormation templates         |
| `make deploy-uat`         | Deploy all UAT stacks                       |
| `make deploy-production`  | Deploy all production stacks                |
| `make deploy-sandbox`     | Deploy all sandbox stacks                   |
| `make deploy-demo`        | Deploy all demo stacks                      |
| `make deploy-all`         | Deploy all environments                     |
| `make destroy-uat`        | Destroy all UAT stacks                      |
| `make destroy-production` | Destroy all production stacks               |
| `make destroy-sandbox`    | Destroy all sandbox stacks                  |
| `make destroy-demo`       | Destroy all demo stacks                     |
| `make destroy-all`        | Destroy all environments                    |

## Outputs

After deployment, stack outputs include:

- **Network**: VPC ID, subnet IDs
- **Data**: Aurora endpoint, S3 bucket name, secrets ARN
- **App**: ALB DNS, ECR repository URI, ECS cluster/service names

View outputs:

```bash
aws cloudformation describe-stacks --stack-name uat-shared-data \
  --query 'Stacks[0].Outputs' --output table
```

## CI/CD

Workflows are configured for both GitHub Actions and GitLab CI:

- `.github/workflows/deploy.yml` - GitHub Actions
- `.gitlab-ci.yml` - GitLab CI

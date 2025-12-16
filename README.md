# ECS Infrastructure (CDK Go)

Infrastructure as Code for deploying Django applications on AWS ECS Fargate with Aurora PostgreSQL.

## Architecture

```
infra/
├── shared/stacks/          # Shared infrastructure
│   ├── network.go          # VPC, subnets, NAT gateways
│   └── data.go             # Aurora PostgreSQL, S3
├── apps/django-api/stacks/ # App-specific infrastructure
│   └── service.go          # ECS Fargate, ALB, ECR
├── config/
│   └── environments.go     # Environment configurations
└── infra.go                # Entry point
```

## Environments & Branching Strategy

| Environment | Branch | Region | Purpose |
|-------------|--------|--------|---------|
| production  | main   | us-east-1 | Production workloads |
| sandbox     | main   | us-east-1 | Customer API testing |
| demo        | main   | us-east-1 | Sales demos |
| uat         | develop| us-east-1 | Internal testing |

```
main branch ──────► production, sandbox, demo
develop branch ───► uat
```

## Stack Naming Convention

```
{environment}-shared-network     # VPC, subnets
{environment}-shared-data        # Aurora PostgreSQL, S3
{environment}-app-django-api     # ECS service, ALB
```

## Prerequisites

- AWS CLI configured with appropriate credentials
- Go 1.21+
- Bun (for CDK CLI): `curl -fsSL https://bun.sh/install | bash`
- CDK bootstrapped in target region:
  ```bash
  bunx cdk bootstrap aws://ACCOUNT_ID/us-east-1
  ```

## Deployment Commands

### List All Stacks
```bash
bunx cdk list
```

### Deploy by Environment

```bash
# Deploy all UAT stacks
bunx cdk deploy uat-* --require-approval never

# Deploy all production stacks
bunx cdk deploy production-* --require-approval never
```

### Deploy by Layer (Recommended)

```bash
# 1. Deploy shared infrastructure first
bunx cdk deploy uat-shared-* --require-approval never

# 2. Deploy app stacks
bunx cdk deploy uat-app-* --require-approval never
```

### Deploy Individual Stacks

```bash
# Network only
bunx cdk deploy uat-shared-network --require-approval never

# Data only (depends on network)
bunx cdk deploy uat-shared-data --require-approval never

# Django API only (depends on network + data)
bunx cdk deploy uat-app-django-api --require-approval never
```

## Useful Commands

| Command | Description |
|---------|-------------|
| `bunx cdk list` | List all stacks |
| `bunx cdk diff uat-*` | Compare deployed vs local |
| `bunx cdk synth uat-shared-network` | Generate CloudFormation template |
| `bunx cdk destroy uat-*` | Delete all UAT stacks |
| `go build .` | Build Go code |
| `go test` | Run unit tests |

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

### Branching Strategy

```
main branch ──────► production, sandbox, demo
develop branch ───► uat
```

| Trigger | Environments | Stacks |
|---------|--------------|--------|
| Push to `main` | production, sandbox, demo | All app stacks |
| Push to `develop` | uat | All app stacks |
| Manual | Any | Any (selectable) |

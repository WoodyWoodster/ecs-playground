.PHONY: build synth list deploy-% destroy-% deploy-all destroy-all

# Build Go CDK
build:
	cd infra && go build .

# Synthesize CloudFormation templates
synth:
	cd infra && bunx cdk synth

# List all stacks
list:
	cd infra && bunx cdk list

# Deploy specific environment (e.g., make deploy-uat)
deploy-%:
	cd infra && bunx cdk deploy $*-* --require-approval never

# Deploy all environments
deploy-all:
	cd infra && bunx cdk deploy --all --require-approval never

# Destroy specific environment (e.g., make destroy-uat)
destroy-%:
	cd infra && bunx cdk destroy $*-* --force

# Destroy all environments
destroy-all:
	cd infra && bunx cdk destroy --all --force

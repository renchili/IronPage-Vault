# AWS Serverless Container Deployment

This directory contains an AWS serverless-container deployment plan for IronPage Vault.

IronPage Vault is a stateful backend. The recommended serverless target is ECS Fargate rather than a Lambda-only rewrite, because the service needs a long-running HTTP API process, PostgreSQL metadata, local PDF processing tools, and durable file storage.

## Files

```text
template.yaml    CloudFormation compatible ECS Fargate skeleton
deploy.sh        validates and deploys the template with AWS CLI
local-test.sh    no-AWS validation path
```

## Required AWS Inputs

The template expects existing AWS infrastructure values:

- VPC ID
- private subnet IDs
- application security group ID
- ECS cluster name
- container image URI
- EFS file system ID
- PostgreSQL host, user, database, and credential reference

This keeps the template reviewable and avoids hiding networking or database assumptions.

## Deploy

```bash
bash deploy/aws/serverless/deploy.sh \
  --stack-name ironpage-vault-serverless \
  --image-uri '<account>.dkr.ecr.<region>.amazonaws.com/ironpage-vault:latest' \
  --vpc-id '<vpc-id>' \
  --subnet-ids '<subnet-a>,<subnet-b>' \
  --security-group-id '<security-group-id>' \
  --efs-file-system-id '<efs-file-system-id>' \
  --db-host '<postgres-host>' \
  --db-credential-arn '<credential-reference-arn>'
```

## Test Without AWS

```bash
bash deploy/aws/serverless/local-test.sh
```

The local test validates YAML syntax when tools are available, builds the Docker image, and runs the local Docker acceptance path. It does not create ECS, ALB, EFS, or RDS resources.

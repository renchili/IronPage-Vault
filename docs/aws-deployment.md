# AWS Deployment Guide

This document describes two AWS deployment targets for IronPage Vault and how to test their delivery artifacts without an AWS account.

IronPage Vault is a stateful Go/Echo backend that stores PostgreSQL metadata and local PDF binaries. The recommended AWS production shape is therefore container-based, not a Lambda-only rewrite.

## Option A: AWS Serverless Container Deployment

Use AWS Fargate as the serverless compute layer.

### Target Architecture

- Amazon ECS Fargate service runs the IronPage Vault container.
- Amazon RDS for PostgreSQL stores metadata.
- Amazon EFS stores PDF binaries and backup artifacts when POSIX filesystem semantics are required.
- AWS Secrets Manager or SSM Parameter Store stores database credentials and application secrets.
- Application Load Balancer exposes HTTPS traffic to the service.
- Private subnets host the service and database. Public subnets host the ALB.
- CloudWatch Logs stores application logs.

This is the preferred serverless option because IronPage Vault requires a long-running HTTP server, local PDF processing tools, and durable file storage. Fargate keeps the operational model serverless while preserving the existing backend container.

### Deployment Artifacts

See:

```text
deploy/aws/serverless/
```

The provided CloudFormation template is intentionally parameterized. It can be validated without AWS credentials and adapted to an existing VPC, RDS instance, and EFS file system.

### No-AWS Test Strategy

Without an AWS account, test the serverless package at three levels:

1. Build the Docker image locally.
2. Validate CloudFormation/SAM syntax locally.
3. Run the same container through Docker Compose acceptance.

Commands:

```bash
bash deploy/aws/serverless/local-test.sh
```

The local test cannot create real ALB, ECS, RDS, or EFS resources. It verifies that the deployment artifacts are syntactically valid and that the container can still pass local acceptance.

## Option B: AWS EKS Deployment

Use Amazon EKS when the target environment already standardizes on Kubernetes or needs Kubernetes-native rollout, ingress, and policy controls.

### Target Architecture

- EKS managed node group or Fargate profile runs the application Pod.
- RDS PostgreSQL is preferred for production metadata storage.
- EFS CSI driver mounts durable PDF and backup storage.
- AWS Load Balancer Controller maps Ingress to ALB.
- External Secrets Operator or Kubernetes Secret stores credentials.
- CloudWatch Container Insights or Prometheus handles observability.

The repository includes plain Kubernetes manifests so they can run unchanged on EKS, kind, or minikube.

### Deployment Artifacts

See:

```text
deploy/aws/eks/
```

### No-AWS Test Strategy

Without AWS, run the EKS manifests on kind:

```bash
bash deploy/aws/eks/local-kind-test.sh
```

This validates the Kubernetes resources, starts the Pod locally, checks rollout status, and port-forwards the service for health/API checks.

## Acceptance Expectations

A reviewer without AWS can still verify:

- Docker image builds successfully.
- CloudFormation/SAM template validates locally.
- Kubernetes manifests apply to a local cluster.
- The container exposes `/healthz` on port 8080.
- The same `run_tests.sh` acceptance path remains the functional source of truth.

A reviewer with AWS can additionally verify:

- ECS/Fargate service reaches steady state.
- EKS deployment reaches available status.
- RDS, EFS, ALB, and Secrets integrations are wired with real AWS resources.
- CloudWatch receives container logs.

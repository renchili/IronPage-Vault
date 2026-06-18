#!/usr/bin/env bash
set -euo pipefail

STACK_NAME=""
IMAGE_URI=""
SUBNET_IDS=""
SECURITY_GROUP_ID=""
ECS_CLUSTER_NAME="ironpage-vault"
EFS_FILE_SYSTEM_ID=""
DB_HOST=""
DB_PORT="5432"
DB_NAME="ironpage"
DB_USER="ironpage"
DB_PASSWORD=""
REGION="${AWS_REGION:-ap-southeast-1}"

while [ $# -gt 0 ]; do
  case "$1" in
    --stack-name) STACK_NAME="$2"; shift 2 ;;
    --image-uri) IMAGE_URI="$2"; shift 2 ;;
    --subnet-ids) SUBNET_IDS="$2"; shift 2 ;;
    --security-group-id) SECURITY_GROUP_ID="$2"; shift 2 ;;
    --ecs-cluster-name) ECS_CLUSTER_NAME="$2"; shift 2 ;;
    --efs-file-system-id) EFS_FILE_SYSTEM_ID="$2"; shift 2 ;;
    --db-host) DB_HOST="$2"; shift 2 ;;
    --db-port) DB_PORT="$2"; shift 2 ;;
    --db-name) DB_NAME="$2"; shift 2 ;;
    --db-user) DB_USER="$2"; shift 2 ;;
    --db-password) DB_PASSWORD="$2"; shift 2 ;;
    --region) REGION="$2"; shift 2 ;;
    *) echo "unknown argument: $1" >&2; exit 2 ;;
  esac
done

require() {
  local name="$1" value="$2"
  if [ -z "$value" ]; then
    echo "missing required argument: $name" >&2
    exit 2
  fi
}

require --stack-name "$STACK_NAME"
require --image-uri "$IMAGE_URI"
require --subnet-ids "$SUBNET_IDS"
require --security-group-id "$SECURITY_GROUP_ID"
require --efs-file-system-id "$EFS_FILE_SYSTEM_ID"
require --db-host "$DB_HOST"
require --db-password "$DB_PASSWORD"

aws cloudformation validate-template \
  --region "$REGION" \
  --template-body file://deploy/aws/serverless/template.yaml >/dev/null

aws cloudformation deploy \
  --region "$REGION" \
  --stack-name "$STACK_NAME" \
  --template-file deploy/aws/serverless/template.yaml \
  --parameter-overrides \
    ImageUri="$IMAGE_URI" \
    SubnetIds="$SUBNET_IDS" \
    SecurityGroupId="$SECURITY_GROUP_ID" \
    EcsClusterName="$ECS_CLUSTER_NAME" \
    EfsFileSystemId="$EFS_FILE_SYSTEM_ID" \
    DbHost="$DB_HOST" \
    DbPort="$DB_PORT" \
    DbName="$DB_NAME" \
    DbUser="$DB_USER" \
    DbPassword="$DB_PASSWORD"

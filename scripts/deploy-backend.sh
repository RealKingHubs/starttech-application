#!/bin/bash
set -euo pipefail

IMAGE="${1:-}"
REGION="${AWS_REGION:-us-east-1}"
ENVIRONMENT="${ENVIRONMENT:-dev}"
BACKEND_TAG_NAME="${BACKEND_TAG_NAME:-${ENVIRONMENT}-backend}"

if [ -z "$IMAGE" ]; then
  echo "Usage: $0 <image-uri>"
  exit 1
fi

EC2_INSTANCE_IDS=$(aws ec2 describe-instances \
  --region "$REGION" \
  --filters \
    "Name=tag:Name,Values=$BACKEND_TAG_NAME" \
    "Name=instance-state-name,Values=running" \
  --query "Reservations[*].Instances[*].InstanceId" \
  --output text)

if [ -z "$EC2_INSTANCE_IDS" ]; then
  echo "No running backend EC2 instances found for tag Name=$BACKEND_TAG_NAME in region $REGION."
  exit 1
fi

SSM_MANAGED_INSTANCE_IDS=$(aws ssm describe-instance-information \
  --region "$REGION" \
  --query "InstanceInformationList[*].InstanceId" \
  --output text)

INSTANCE_IDS=""

for EC2_INSTANCE_ID in $EC2_INSTANCE_IDS; do
  for SSM_INSTANCE_ID in $SSM_MANAGED_INSTANCE_IDS; do
    if [ "$EC2_INSTANCE_ID" = "$SSM_INSTANCE_ID" ]; then
      INSTANCE_IDS="$INSTANCE_IDS $EC2_INSTANCE_ID"
      break
    fi
  done
done

INSTANCE_IDS="$(echo "$INSTANCE_IDS" | xargs)"

if [ -z "$INSTANCE_IDS" ]; then
  echo "Backend EC2 instances exist, but none are registered as SSM managed instances in region $REGION."
  echo "Expected backend tag: Name=$BACKEND_TAG_NAME"
  echo "EC2 instances found: $EC2_INSTANCE_IDS"
  echo "This usually means the SSM agent is not registered yet or the instances need to be refreshed after a launch template/userdata change."
  exit 1
fi

aws ssm send-command \
  --instance-ids $INSTANCE_IDS \
  --document-name "AWS-RunShellScript" \
  --region "$REGION" \
  --parameters 'commands=[
    "MONGO_URI=$(aws ssm get-parameter --name /starttech/dev/mongo_uri --with-decryption --query Parameter.Value --output text --region '"$REGION"')",
    "JWT_SECRET=$(aws ssm get-parameter --name /starttech/dev/jwt_secret --with-decryption --query Parameter.Value --output text --region '"$REGION"')",
    "DB_NAME=$(aws ssm get-parameter --name /starttech/dev/db_name --query Parameter.Value --output text --region '"$REGION"')",
    "REDIS_HOST=$(aws ssm get-parameter --name /starttech/dev/redis_host --query Parameter.Value --output text --region '"$REGION"')",

    "aws ecr get-login-password --region '"$REGION"' | docker login --username AWS --password-stdin 093796422475.dkr.ecr.us-east-1.amazonaws.com",

    "docker system prune -af || true",
    "docker pull '"$IMAGE"'",

    "docker stop backend || true",
    "docker rm backend || true",

    "docker run -d --name backend -p 8080:8080 \
      -e PORT=8080 \
      -e MONGO_URI=$MONGO_URI \
      -e DB_NAME=$DB_NAME \
      -e JWT_SECRET_KEY=$JWT_SECRET \
      -e ENABLE_CACHE=true \
      -e REDIS_ADDR=$REDIS_HOST:6379 \
      '"$IMAGE"'"
  ]'

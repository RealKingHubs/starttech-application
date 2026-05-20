#!/bin/bash

IMAGE=$1

INSTANCE_IDS=$(aws ssm describe-instance-information \
  --region us-east-1 \
  --query "InstanceInformationList[*].InstanceId" \
  --output text)

aws ssm send-command \
  --instance-ids $INSTANCE_IDS \
  --document-name "AWS-RunShellScript" \
  --region us-east-1 \
  --parameters 'commands=[
    "MONGO_URI=$(aws ssm get-parameter --name /starttech/dev/mongo_uri --with-decryption --query Parameter.Value --output text --region us-east-1)",
    "JWT_SECRET=$(aws ssm get-parameter --name /starttech/dev/jwt_secret --with-decryption --query Parameter.Value --output text --region us-east-1)",
    "DB_NAME=$(aws ssm get-parameter --name /starttech/dev/db_name --query Parameter.Value --output text --region us-east-1)",

    "aws ecr get-login-password --region us-east-1 | docker login --username AWS --password-stdin 093796422475.dkr.ecr.us-east-1.amazonaws.com",

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
      -e REDIS_ADDR='dev-redis.yd4jsv.0001.use1.cache.amazonaws.com' \
      '"$IMAGE"'"
  ]'
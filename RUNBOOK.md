````markdown id="o5a4oj"
# RUNBOOK.md

```md
# StartTech Operations Runbook

This runbook documents the operational procedures, deployment checks, troubleshooting steps, and recovery actions used during the StartTech deployment.

This document reflects actual deployment issues encountered and how they were resolved.

---

# Infrastructure Operations

## Terraform Deployment

### Initialize

```bash
terraform init
```

### Plan infrastructure

```bash
terraform plan
```

### Apply infrastructure

```bash
terraform apply
```

### Destroy infrastructure

```bash
terraform destroy
```

# Backend Operations

## Check ALB Health

```bash
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn>
```

Healthy targets should show:

```text
healthy
```

## Check Docker Containers

```bash
docker ps -a
```

## Check Backend Logs

```bash
docker logs backend
```

## Check SSM Parameters

```bash
aws ssm get-parameters-by-path \
  --path "/starttech/dev/"
```

# CI/CD Operations

Deployment pipeline:

1. Run Go tests
2. Run security scans
3. Build Docker image
4. Push image to ECR
5. Deploy backend
6. Build frontend
7. Deploy frontend to S3
8. Run smoke tests

Workflow file:

```text
.github/workflows/backend-ci-cd.yml
```

# Troubleshooting Guide

# 1. Terraform Module Errors

## Symptoms

```text
Reference to undeclared resource
```

or:

```text
Reference to undeclared module
```

## Cause

Resources were referenced directly across Terraform modules.

## Resolution

Pass required values using:

- outputs
- variables

instead of direct resource references.

# 2. S3 Bucket Policy Fails

## Symptoms

```text
AccessDenied: PutBucketPolicy
```

## Cause

S3 Block Public Access settings blocked website policies.

## Resolution

Adjusted public access configuration to support static website hosting.

# 3. ALB Returns 502 Bad Gateway

## Symptoms

- frontend cannot reach backend
- health checks fail
- target marked unhealthy

## Cause

Backend container failed to start.

## Verification

### Check target health

```bash
aws elbv2 describe-target-health
```

### Check containers

```bash
docker ps -a
```

### Check logs

```bash
docker logs backend
```

## Resolution

- redeploy backend
- verify Docker image exists in ECR
- verify userdata execution
- verify backend listens on port 8080

# 4. Registration Fails From Frontend

## Symptoms

Browser console errors:

```text
No 'Access-Control-Allow-Origin' header
```

or:

```text
CORS policy blocked request
```

## Cause

Terraform generated a new S3 bucket URL but backend still trusted an old hardcoded origin.

## Resolution

Updated backend middleware to dynamically allow StartTech S3 website origins.

# 5. Login Appears Successful But Session Fails

## Symptoms

- login succeeds
- `/users/me` fails
- frontend appears logged out

## Cause

Frontend relied on cookies across origins.

S3 website frontend and ALB backend are separate origins.

## Resolution

Authentication was redesigned around JWT Bearer tokens stored in localStorage.

# 6. Backend Health Check Fails

## Symptoms

ALB target shows:

```text
Target.FailedHealthChecks
```

## Cause

Backend container missing or crashed.

## Verification

```bash
docker ps -a
docker logs backend
```

## Resolution

Fix deployment script and redeploy container.

# 7. SSM Command Failures

## Symptoms

```text
InvalidInstanceId
```

or failed SSM commands.

## Cause

- wrong instance ID
- unhealthy instance
- terminated instance

## Resolution

Verify active EC2 instance ID before sending commands.

# Frontend Deployment Operations

## Build frontend

```bash
npm run build
```

## Deploy frontend

```bash
aws s3 sync Client/dist/ s3://<bucket-name> --delete
```

# Backend Health Validation

Health endpoint:

```bash
curl http://<alb-dns>/health
```

Expected response:

```json
{
  "status": "ok"
}
```

# Monitoring

## CloudWatch Logs

Primary backend logs:

```text
/starttech/backend
```

# Recovery Procedure

If deployment becomes unstable:

1. Verify EC2 instance health
2. Verify ALB target health
3. Verify backend container status
4. Verify Docker logs
5. Verify ECR image exists
6. Verify SSM secrets
7. Redeploy workflow

# Important Lessons From Deployment

The deployment evolved significantly during implementation.

## Key lessons learned

- CloudFront limitations can reshape architecture decisions
- S3 static hosting requires careful CORS handling
- Cross-origin cookie auth is unreliable for this setup
- JWT Bearer authentication is more stable for static hosting
- Secrets should never remain hardcoded in userdata
- ALB health checks are critical for debugging deployments
- Dynamic infrastructure requires flexible backend configuration

These operational fixes transformed the deployment from an unstable prototype into a repeatable cloud deployment workflow.
```
````

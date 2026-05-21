# RUNBOOK.md

````md id="qj4v0m"
# StartTech Operations Runbook

This runbook documents the actual operational procedures used while deploying, debugging, stabilizing, and maintaining the StartTech platform.

The guide focuses only on the real deployment issues, fixes, and workflows encountered during implementation.

---

# Environment Overview

The platform consists of:

| Component | Technology |
|---|---|
| Frontend | React + Vite |
| Backend | Go + Gin |
| Container Runtime | Docker |
| Infrastructure | Terraform |
| CI/CD | GitHub Actions |
| Backend Compute | EC2 Auto Scaling Group |
| Load Balancing | Application Load Balancer |
| Frontend Hosting | S3 Static Website Hosting |
| Database | MongoDB Atlas |
| Cache | Redis ElastiCache |
| Secret Management | AWS SSM Parameter Store |
| Monitoring | CloudWatch + SSM |

---

# Operational Architecture

```text
Browser
   ↓
S3 Static Website
   ↓
Application Load Balancer
   ↓
EC2 Backend Containers
   ↓
MongoDB Atlas + Redis
````

---

# CI/CD Operations

Two deployment pipelines exist.

---

# 1. Infrastructure Pipeline

Location:

```text id="4htq5z"
starttech-infra/.github/workflows/terraform.yml
```

Purpose:

* Provision AWS infrastructure
* Update networking
* Create AWS services
* Apply Terraform changes

---

# Common Infrastructure Commands

## Initialize Terraform

```bash id="a7vtdh"
terraform init
```

---

## Review Infrastructure Changes

```bash id="lcbhql"
terraform plan
```

---

## Apply Infrastructure

```bash id="cf7jlwm"
terraform apply
```

---

# 2. Application Pipeline

Location:

```text id="h8my0g"
starttech-application/.github/workflows/backend-ci-cd.yml
```

Purpose:

* Run backend tests
* Build Docker images
* Push images to ECR
* Deploy backend containers
* Build frontend
* Upload frontend assets to S3
* Run smoke tests

---

# Deployment Workflow

## Backend Deployment

```text
GitHub Push
   ↓
Go Tests
   ↓
Docker Build
   ↓
Push To ECR
   ↓
Deploy Script Executes
   ↓
EC2 Pulls Container
   ↓
Container Starts
   ↓
ALB Health Check Passes
```

---

## Frontend Deployment

```text
GitHub Push
   ↓
Vite Build
   ↓
Generate dist/
   ↓
Upload dist/ To S3
```

---

# SSM Parameter Management

Sensitive credentials are stored in AWS Systems Manager Parameter Store.

---

# Parameters Used

| Parameter Path              | Purpose                  |
| --------------------------- | ------------------------ |
| `/starttech/dev/mongo_uri`  | MongoDB Atlas connection |
| `/starttech/dev/jwt_secret` | JWT signing key          |
| `/starttech/dev/redis_host` | Redis endpoint           |
| `/starttech/dev/db_name`    | Database name            |

---

# Creating Parameters

Script used:

```text id="xfwk7g"
starttech-infra/scripts/create-ssm-parameters.sh
```

---

# Important Security Change

Early deployment versions used hardcoded secrets inside userdata scripts.

This was later removed and replaced with:

```text id="o6u5ff"
AWS SSM Parameter Store
```

This reduced secret exposure significantly.

---

# Frontend Hosting Operations

The frontend is hosted using:

```text id="m4czt5"
S3 Static Website Hosting
```

---

# Why S3 Static Hosting Was Used

The original design intended to use:

```text id="j3s8s5"
CloudFront + Private S3
```

However, CloudFront access was unavailable in the AWS account being used.

The deployment strategy was redesigned around:

```text id="6h6mgu"
Public S3 Website Hosting
```

This directly affected:

* authentication design
* CORS handling
* bucket policies
* frontend deployment

---

# Frontend Deployment Command

```bash id="0xxzv8"
aws s3 sync frontend/dist/ s3://<frontend-bucket> --delete
```

---

# Backend Health Monitoring

The backend is monitored through:

* ALB health checks
* GitHub smoke tests
* Docker container status
* CloudWatch logs

---

# Common Operational Issue:

# ALB Returning 502 Bad Gateway

## Symptoms

* Frontend fails to connect
* Smoke tests fail
* ALB target group becomes unhealthy
* Browser shows 502 Bad Gateway

---

# Root Cause Encountered

During deployment:

* backend containers were not running correctly
* EC2 instances registered as unhealthy
* ALB health checks failed

This caused:

```text id="r4rq84"
Target.FailedHealthChecks
```

---

# How It Was Investigated

The backend instances were private and inaccessible through SSH.

AWS Systems Manager was used instead.

---

# SSM Debugging Commands

## List Docker Containers

```bash id="7yjlwm"
aws ssm send-command \
  --instance-ids "<instance-id>" \
  --document-name "AWS-RunShellScript" \
  --parameters commands="docker ps -a"
```

---

## Inspect Backend Logs

```bash id="wxpkic"
aws ssm send-command \
  --instance-ids "<instance-id>" \
  --document-name "AWS-RunShellScript" \
  --parameters commands="docker logs backend --tail 200"
```

---

## Check Docker Runtime

```bash id="0fhtj4"
aws ssm send-command \
  --instance-ids "<instance-id>" \
  --document-name "AWS-RunShellScript" \
  --parameters commands="journalctl -u docker --no-pager -n 50"
```

---

# Common Issue:

# No Such Container: backend

## Symptoms

```text id="9jtnwe"
Error response from daemon: No such container: backend
```

---

# Meaning

The deployment script completed incorrectly or the backend container failed before startup completed.

This was one of the main indicators that backend deployment failed before ALB health checks.

---

# How To Verify Target Health

## Check Target Group State

```bash id="wz6fkv"
aws elbv2 describe-target-health \
  --target-group-arn <target-group-arn>
```

---

# Healthy Output

```text
"State": "healthy"
```

---

# Unhealthy Output

```text
"State": "unhealthy"
"Reason": "Target.FailedHealthChecks"
```

---

# CORS Troubleshooting

CORS became one of the largest runtime issues after migrating to S3 website hosting.

---

# Symptoms

Browser errors:

```text id="w72h4p"
No 'Access-Control-Allow-Origin' header
```

and:

```text id="dnt3c4"
Response to preflight request doesn't pass access control check
```

---

# Root Cause

Terraform dynamically generated frontend bucket names:

```text id="7s0h9d"
dev-starttech-frontend-ee4128bc
```

The backend originally trusted only one hardcoded frontend origin.

Whenever the bucket changed, the browser blocked requests.

---

# Final CORS Fix

The backend middleware was redesigned to:

* support localhost development
* allow dynamic S3 website origins
* correctly handle OPTIONS requests

---

# Authentication Troubleshooting

Authentication changed significantly during deployment.

---

# Original Problem

The frontend originally depended on cookies for authentication.

This failed because:

* frontend and backend were cross-origin
* S3 website hosting complicated cookie handling
* browser restrictions blocked session persistence

---

# Symptoms

* login appeared successful
* authenticated requests failed afterward
* `/users/me` requests returned errors
* frontend displayed "Registration Failed"

---

# Final Authentication Fix

The frontend was redesigned to:

* store JWT tokens in localStorage
* send bearer tokens in Authorization headers
* stop depending on browser cookies

---

# Frontend Environment Configuration

The frontend requires:

```env id="z7j6lz"
VITE_API_BASE_URL=http://<alb-dns>
```

This value is injected during GitHub Actions deployment using repository secrets.

---

# Common Issue:

# Frontend Calls Wrong Backend URL

## Symptoms

Frontend requests fail with:

```text id="9f4jlwm"
404 Not Found
```

or:

```text id="pb5m5j"
ERR_NETWORK
```

---

# Fix

Verify:

1. GitHub secret exists
2. `.env` file is generated during CI/CD
3. Frontend rebuild completed successfully
4. Browser is loading latest deployed assets

---

# Smoke Test Operations

The deployment pipeline includes automated smoke tests.

Script:

```text id="91klm7"
scripts/health-check.sh
```

---

# Smoke Test Failure Symptoms

```text id="q1qj5r"
Health check failed
```

---

# Meaning

Usually indicates:

* backend container not running
* ALB unhealthy target
* failed deployment
* application crash

---

# CloudWatch Operations

CloudWatch is used for:

* backend logs
* deployment visibility
* runtime debugging

---

# Log Group

```text id="v84q4m"
/starttech/backend
```

---

# Recommended Runtime Checks

After deployment:

1. Verify target health
2. Verify backend container exists
3. Verify frontend loads
4. Test registration flow
5. Test login flow
6. Test authenticated endpoints
7. Inspect CloudWatch for runtime errors

---

# Operational Lessons Learned

Several major deployment decisions changed during implementation.

---

# 1. CloudFront Was Removed

Reason:

* AWS account limitation

Impact:

* frontend architecture redesign
* authentication redesign
* CORS redesign

---

# 2. Cookie Authentication Failed

Reason:

* cross-origin browser restrictions

Impact:

* migration to JWT token auth

---

# 3. Hardcoded Secrets Were Removed

Reason:

* security risk

Impact:

* migration to AWS SSM Parameter Store

---

# 4. SSM Became Essential

Reason:

* backend instances were private

Impact:

* operational debugging depended heavily on SSM Run Command

---

# Final Operational Notes

This deployment was stabilized through iterative troubleshooting rather than a purely theoretical setup.

The final operational workflow now supports:

* repeatable deployments
* automated infrastructure provisioning
* automated backend deployment
* automated frontend deployment
* secure secret management
* runtime debugging
* production-style operational monitoring

```
```

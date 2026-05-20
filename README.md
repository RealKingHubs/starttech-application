# README.md

````md
# StartTech Infrastructure And Application Platform

StartTech is a full-stack cloud-native task management platform deployed on AWS using Terraform, Docker, GitHub Actions, and a production-style DevOps workflow.

The project combines:

- Infrastructure as Code with Terraform
- Containerized backend deployment
- Automated CI/CD pipelines
- AWS networking and load balancing
- Static frontend hosting on S3
- Secure secret management with AWS SSM
- Redis caching
- MongoDB Atlas integration
- Operational monitoring and troubleshooting workflows

This repository reflects the actual engineering decisions, deployment issues, fixes, and architecture evolution during the build process.

---

# Project Structure

```text
Starttech/
│
├── starttech-application/
│   │
│   ├── Client/
│   │   ├── src/
│   │   ├── public/
│   │   ├── package.json
│   │   └── vite.config.ts
│   │
│   ├── Server/
│   │   └── MuchToDo/
│   │       ├── cmd/api/
│   │       ├── internal/
│   │       ├── docs/
│   │       ├── Dockerfile
│   │       └── go.mod
│   │
│   ├── scripts/
│   │   ├── deploy-backend.sh
│   │   └── health-check.sh
│   │
│   └── .github/
│       └── workflows/
│           └── backend-ci-cd.yml
│
├── starttech-infra/
│   │
│   ├── terraform/
│   │   ├── modules/
│   │   ├── environments/
│   │   └── main.tf
│   │
│   ├── scripts/
│   │   └── create-ssm-parameters.sh
│   │
│   └── .github/
│       └── workflows/
│           └── terraform.yml
│
├── README.md
├── ARCHITECTURE.md
└── RUNBOOK.md
````

---

# What The Platform Does

The application allows users to:

* Create accounts
* Authenticate securely
* Create and manage tasks
* Persist task data in MongoDB
* Cache frequently accessed data with Redis

The platform was designed to simulate a production-ready AWS deployment pipeline rather than a simple local development setup.

---

# Infrastructure Overview

The infrastructure is provisioned entirely with Terraform.

## AWS Services Used

| Service                   | Purpose                         |
| ------------------------- | ------------------------------- |
| VPC                       | Network isolation               |
| Public Subnets            | ALB and public-facing resources |
| Private Subnets           | Backend and Redis               |
| Security Groups           | Network access control          |
| Application Load Balancer | Traffic routing                 |
| EC2 Auto Scaling Group    | Backend compute layer           |
| ECR                       | Docker image registry           |
| S3                        | Frontend static hosting         |
| ElastiCache Redis         | Caching layer                   |
| SSM Parameter Store       | Secret management               |
| IAM + OIDC                | GitHub Actions authentication   |
| CloudWatch                | Logging and monitoring          |

---

# Why The Frontend Uses S3 Static Hosting

The original architecture used:

```text
CloudFront → Private S3 Bucket
```

However, CloudFront distribution access was unavailable in the AWS account being used.

Because of this limitation, the frontend deployment strategy was redesigned to use:

```text
Public S3 Static Website Hosting
```

This changed several parts of the system:

* Bucket access policies
* Frontend deployment process
* Browser authentication flow
* CORS configuration
* Token handling strategy

The final deployment architecture became:

```text
Browser
   ↓
S3 Static Website Hosting
   ↓
Application Load Balancer
   ↓
EC2 Backend Containers
```

---

# Backend Deployment Flow

The backend application is written in Go using Gin.

The backend is:

1. Built into a Docker image
2. Pushed to Amazon ECR
3. Pulled onto EC2 instances
4. Started through deployment automation scripts

Deployment happens automatically through GitHub Actions.

---

# Frontend Deployment Flow

The frontend is built with React and Vite.

The deployment pipeline:

1. Builds the frontend application
2. Injects the backend API URL dynamically
3. Generates optimized static assets
4. Uploads the build output to S3

Deployment command:

```bash
aws s3 sync Client/dist/ s3://<frontend-bucket> --delete
```

---

# Authentication Design Evolution

## Original Design

The frontend originally depended on browser cookies for authentication.

This worked locally but failed in the deployed AWS environment because:

* Frontend and backend were hosted on different origins
* S3 static website hosting introduced cross-origin browser restrictions
* Cookies became unreliable across origins

This caused:

* Registration failures
* Login session failures
* Browser CORS blocking

---

## Final Design

The authentication system was redesigned to use JWT bearer tokens.

The backend already returned JWT tokens during login, so the frontend was updated to:

* Store JWT tokens in localStorage
* Automatically attach Authorization headers
* Use bearer-token authentication for protected routes

This completely removed the dependency on cross-site cookies.

---

# CORS Redesign

Terraform generates frontend buckets dynamically using random suffixes:

```tf
bucket = "${var.environment}-starttech-frontend-${random_id.suffix.hex}"
```

This meant frontend URLs constantly changed.

Originally, the backend used hardcoded CORS origins, which caused browser requests to fail whenever the bucket name changed.

The backend CORS middleware was redesigned to:

* Support localhost development
* Support dynamically generated S3 website URLs
* Handle OPTIONS preflight requests correctly

This fixed frontend registration and authentication failures.

---

# Secret Management

Sensitive credentials were initially hardcoded during early infrastructure setup.

This was replaced with AWS Systems Manager Parameter Store.

Secrets now include:

* MongoDB connection URI
* JWT signing secret
* Redis endpoint
* Database name

Parameters are stored using:

```bash
aws ssm put-parameter
```

Secure values use:

```text
SecureString
```

---

# CI/CD Pipelines

Two separate GitHub Actions pipelines exist.

---

## Application Pipeline

Location:

```text
starttech-application/.github/workflows/backend-ci-cd.yml
```

Responsibilities:

* Go testing
* Docker image build
* Security scanning
* Push image to ECR
* Deploy backend
* Build frontend
* Deploy frontend to S3
* Smoke testing

---

## Infrastructure Pipeline

Location:

```text
starttech-infra/.github/workflows/terraform.yml
```

Responsibilities:

* Terraform validation
* Terraform formatting checks
* Terraform planning
* Terraform apply

---

# Security Improvements Made During The Project

Several security improvements were introduced during development:

* Removal of hardcoded secrets
* Migration to SSM Parameter Store
* Restricted security groups
* Private subnet backend deployment
* Redis isolation
* JWT-based authentication
* GitHub OIDC authentication
* Reduced IAM credential exposure

---

# Operational Challenges Solved

The project involved several real deployment issues:

## 1. CloudFront Access Restriction

CloudFront could not be used in the AWS account.

Solution:

* Migrated frontend to S3 static website hosting

---

## 2. Backend Health Check Failures

The backend container failed to start correctly, causing:

* ALB 502 errors
* Unhealthy target groups
* Smoke test failures

Solution:

* Debugged EC2 instances using AWS SSM
* Investigated Docker container state
* Corrected deployment issues

---

## 3. CORS Failures

Frontend requests were blocked by browser CORS protection.

Solution:

* Redesigned backend middleware for dynamic origins

---

## 4. Authentication Failures

Cross-origin cookies failed in the browser.

Solution:

* Migrated frontend to JWT token authentication

---

# Local Development

## Backend

```bash
cd starttech-application/Server/MuchToDo

go mod download
go run cmd/api/main.go
```

---

## Frontend

```bash
cd starttech-application/Client

npm install
npm run dev
```

---

# Terraform Deployment

```bash
cd starttech-infra/terraform

terraform init
terraform plan
terraform apply
```

---

# Environment Variables

## Frontend

```env
VITE_API_BASE_URL=http://<alb-dns>
```

Stored in GitHub Secrets for CI/CD deployment.

---

## Backend

Loaded from AWS SSM Parameter Store.

---

# Monitoring And Troubleshooting

Operational debugging was performed using:

* AWS SSM Run Command
* CloudWatch Logs
* ALB Target Health
* Docker container logs
* GitHub Actions logs

Detailed operational procedures are documented in:

```text
RUNBOOK.md
```

---

# Documentation

| File            | Purpose                          |
| --------------- | -------------------------------- |
| README.md       | Project overview and setup       |
| ARCHITECTURE.md | Infrastructure and system design |
| RUNBOOK.md      | Operations and troubleshooting   |

---

# Final Notes

This project evolved significantly during deployment.

Several production-style problems were encountered and solved in real time, including:

* browser CORS restrictions
* container deployment failures
* ALB health check failures
* static hosting limitations
* authentication redesign
* secret management migration

The final platform reflects the operational decisions made to stabilize and successfully deploy the system on AWS.

```
```

# ARCHITECTURE.md

````md
# StartTech System Architecture

This document explains the actual architecture implemented for the StartTech platform, including the infrastructure decisions, deployment flow, networking design, authentication redesign, and operational fixes introduced during deployment.

The architecture evolved during implementation as real deployment constraints and runtime issues were discovered.

---

# High-Level Architecture

```text
                        ┌─────────────────────┐
                        │     GitHub          │
                        │   Source Control    │
                        └─────────┬───────────┘
                                  │
                                  │ Push
                                  ▼
                    ┌─────────────────────────┐
                    │   GitHub Actions CI/CD  │
                    └─────────┬───────────────┘
                              │
              ┌───────────────┴────────────────┐
              │                                │
              ▼                                ▼
   ┌────────────────────┐         ┌────────────────────┐
   │ Terraform Pipeline │         │ Application Pipeline│
   └─────────┬──────────┘         └─────────┬──────────┘
             │                              │
             ▼                              ▼
 ┌────────────────────────┐      ┌────────────────────────┐
 │ AWS Infrastructure     │      │ Docker Image Build     │
 │ Provisioning           │      │ + Frontend Build       │
 └─────────┬──────────────┘      └─────────┬──────────────┘
           │                               │
           │                               ▼
           │                   ┌────────────────────────┐
           │                   │ Amazon ECR             │
           │                   └─────────┬──────────────┘
           │                             │
           ▼                             ▼
 ┌────────────────────────┐   ┌──────────────────────────┐
 │ VPC                    │   │ EC2 Auto Scaling Group   │
 │ Subnets                │   │ Dockerized Go Backend    │
 │ Security Groups        │   └─────────┬────────────────┘
 │ ALB                    │             │
 │ Redis                  │             ▼
 └─────────┬──────────────┘   ┌──────────────────────────┐
           │                  │ Application Load Balancer│
           │                  └─────────┬────────────────┘
           │                            │
           ▼                            ▼
 ┌────────────────────────┐   ┌──────────────────────────┐
 │ S3 Static Hosting      │   │ MongoDB Atlas            │
 │ React Frontend         │   │ External Database        │
 └────────────────────────┘   └──────────────────────────┘
````

---

# Infrastructure Architecture

The infrastructure was provisioned entirely using Terraform.

The environment was separated into reusable modules to simplify management and improve maintainability.

---

# Terraform Module Design

```text
terraform/
│
├── modules/
│   ├── networking/
│   ├── security/
│   ├── storage/
│   ├── compute/
│   ├── loadbalancer/
│   └── monitoring/
│
├── environments/
│   └── dev/
│
└── main.tf
```

---

# Networking Architecture

The infrastructure uses a custom VPC with both public and private subnets.

---

## Public Subnets

Public subnets contain internet-facing resources:

* Application Load Balancer
* NAT access
* Public routing

These subnets allow inbound traffic from users.

---

## Private Subnets

Private subnets contain internal resources:

* Backend EC2 instances
* Redis cluster

These services are intentionally isolated from direct internet access.

---

# Why The Backend Was Kept Private

The backend API is not directly exposed publicly.

Instead:

```text
User → ALB → Backend EC2
```

Benefits:

* Reduced attack surface
* Controlled traffic entry point
* Centralized health checking
* Easier scaling

---

# Application Load Balancer

The Application Load Balancer acts as the single public entry point for backend traffic.

Responsibilities:

* Route HTTP requests
* Perform health checks
* Forward traffic to healthy EC2 targets
* Detect failed backend containers

---

# ALB Health Checks

The ALB continuously checks backend health.

During deployment, backend containers initially failed to start correctly, causing:

```text
502 Bad Gateway
```

and:

```text
Target.FailedHealthChecks
```

This issue became one of the main operational debugging tasks during deployment.

---

# Compute Layer

The backend application runs on EC2 instances managed by an Auto Scaling Group.

Each instance:

1. Pulls backend Docker images from ECR
2. Starts the API container
3. Registers with the ALB target group

---

# Why Docker Was Used

The backend was containerized to ensure:

* consistent runtime environments
* reproducible deployments
* simpler rollback capability
* cleaner CI/CD automation

---

# Container Deployment Flow

```text
GitHub Actions
    ↓
Build Docker Image
    ↓
Push To ECR
    ↓
EC2 Pulls Image
    ↓
Container Starts
    ↓
ALB Health Check Passes
```

---

# Frontend Architecture

The frontend was built using:

* React
* TypeScript
* Vite

The frontend is deployed as static assets to Amazon S3.

---

# Original Frontend Design

The original deployment design was:

```text
CloudFront → Private S3 Bucket
```

This architecture would have provided:

* CDN caching
* HTTPS edge delivery
* private bucket access through OAC

---

# Why CloudFront Was Removed

CloudFront distribution access was unavailable in the AWS account being used.

Because of this limitation, the frontend architecture was redesigned to:

```text
Public S3 Static Website Hosting
```

This changed multiple parts of the system:

* frontend hosting strategy
* bucket policies
* authentication flow
* browser behavior
* CORS handling

---

# Final Frontend Architecture

```text
Browser
   ↓
S3 Static Website Endpoint
   ↓
Application Load Balancer
   ↓
Backend API
```

---

# S3 Static Hosting Design

The frontend bucket uses:

* static website hosting
* public read access
* frontend asset deployment through CI/CD

Terraform dynamically creates bucket names using:

```tf
bucket = "${var.environment}-starttech-frontend-${random_id.suffix.hex}"
```

This prevents bucket name collisions globally.

---

# Authentication Architecture Evolution

Authentication changed significantly during deployment.

---

# Original Authentication Design

The frontend originally relied on browser cookies for session persistence.

This worked locally but failed in production because:

* frontend and backend existed on different origins
* S3 static website hosting introduced cross-origin limitations
* browser cookie handling became unreliable

This caused:

* login failures
* registration failures
* failed authenticated requests

---

# Final Authentication Design

The backend already returned JWT tokens during login.

The frontend was redesigned to:

* store JWT tokens in localStorage
* send Authorization headers automatically
* use bearer-token authentication for all protected requests

---

# Why Token Authentication Was Chosen

Token-based authentication removed dependency on:

* cross-origin cookies
* browser cookie policies
* credential forwarding problems

This made authentication stable across:

```text
S3 Static Website → ALB Backend
```

---

# CORS Architecture

CORS became a major deployment issue.

---

# Root Cause

Terraform dynamically generates frontend bucket names.

Example:

```text
dev-starttech-frontend-ee4128bc
```

The backend originally used hardcoded frontend origins.

Whenever the bucket changed, browser requests failed.

---

# CORS Redesign

The backend middleware was redesigned to:

* allow localhost development
* support dynamic S3 website origins
* correctly process OPTIONS preflight requests

This fixed browser request blocking.

---

# Redis Architecture

Redis was introduced using Amazon ElastiCache.

Purpose:

* caching
* username lookup optimization
* reducing repeated database queries

Redis runs inside private subnets and is inaccessible publicly.

---

# MongoDB Architecture

MongoDB Atlas was used instead of self-hosted MongoDB.

Reasons:

* managed database operations
* reduced infrastructure overhead
* easier cloud integration

The backend connects securely using a connection URI stored in SSM Parameter Store.

---

# Secret Management Architecture

Secrets were originally hardcoded during early setup.

This was later redesigned using AWS Systems Manager Parameter Store.

---

# Parameters Stored

| Parameter     | Type         |
| ------------- | ------------ |
| MongoDB URI   | SecureString |
| JWT Secret    | SecureString |
| Redis Host    | String       |
| Database Name | String       |

---

# Why SSM Was Introduced

This removed secrets from:

* Terraform
* userdata scripts
* deployment pipelines
* repository files

It also improved operational security significantly.

---

# CI/CD Architecture

Two separate GitHub Actions pipelines were implemented.

---

# Infrastructure Pipeline

Location:

```text
starttech-infra/.github/workflows/terraform.yml
```

Responsibilities:

* Terraform validation
* Terraform planning
* Terraform apply

---

# Application Pipeline

Location:

```text
starttech-application/.github/workflows/backend-ci-cd.yml
```

Responsibilities:

* Go testing
* Docker build
* ECR push
* Frontend build
* S3 deployment
* Backend deployment
* Smoke testing

---

# Why Frontend Deployment Exists In Application CI/CD

Frontend assets are application artifacts.

Terraform provisions infrastructure only.

Application CI/CD handles:

```text
React Build → S3 Upload
```

This separation keeps infrastructure and application deployment independent.

---

# Runtime Debugging Architecture

AWS Systems Manager became critical during deployment debugging.

---

# Why SSM Was Important

Backend instances existed inside private infrastructure and could not be accessed directly.

SSM allowed operational debugging without SSH access.

Commands used included:

```bash
aws ssm send-command
```

and:

```bash
aws ssm get-command-invocation
```

---

# Problems Diagnosed Using SSM

SSM helped identify:

* backend container failures
* missing containers
* unhealthy targets
* Docker runtime issues
* deployment script failures

---

# Monitoring And Logging

Monitoring used:

* CloudWatch Logs
* ALB Target Health
* GitHub Actions logs
* Docker logs
* SSM command output

---

# Final Architecture Summary

The final deployed architecture reflects several real deployment-driven design decisions.

The most significant changes included:

* replacing CloudFront with S3 static hosting
* redesigning authentication from cookies to JWT tokens
* implementing dynamic CORS handling
* migrating secrets into SSM Parameter Store
* using SSM for operational debugging
* separating infrastructure and application pipelines

The resulting system is a fully automated AWS deployment platform built from real operational troubleshooting and iterative infrastructure refinement.

```
```

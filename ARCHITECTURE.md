````markdown
# ARCHITECTURE.md

```md
# StartTech System Architecture

# Overview

StartTech is a distributed cloud-native application deployed on AWS.

The system consists of:

- React frontend hosted on Amazon S3
- Go backend running inside Docker containers on EC2
- Application Load Balancer for routing
- Redis cache using ElastiCache
- MongoDB Atlas database
- GitHub Actions CI/CD pipeline
- Terraform-managed infrastructure

---

# High-Level Architecture

```text
Users
   │
   ▼
S3 Static Website Hosting
(React Frontend)
   │
   ▼
Application Load Balancer
   │
   ▼
EC2 Auto Scaling Group
(Go API Containers)
   │
   ├── Redis ElastiCache
   │
   └── MongoDB Atlas
```

# Frontend Architecture

## Technology Stack

- React + Vite

## Hosting Design

Originally the frontend was intended to use:

- CloudFront + private S3 bucket

CloudFront permissions were unavailable in the deployment environment.

The architecture was therefore redesigned around:

- Amazon S3 Static Website Hosting

This changed several parts of the system:

| Change | Architectural Impact |
|---|---|
| Public S3 hosting | Frontend publicly accessible |
| Cross-origin backend access | Required CORS handling |
| Dynamic bucket URLs | Required flexible origin validation |
| Cookie auth instability | Required JWT auth redesign |

# Backend Architecture

## Technology Stack

- Go + Gin

## Runtime

The backend runs inside Docker containers on EC2 instances.

Containers are pulled from:

- Amazon Elastic Container Registry (ECR)

## Load Balancing

Traffic enters through:

- Application Load Balancer

Responsibilities:

- distribute traffic
- perform health checks
- route requests to healthy instances

Health endpoint:

```text
/health
```

Backend application port:

```text
8080
```

## Compute Layer

Backend compute is managed using:

- EC2 Auto Scaling Group

Launch Templates bootstrap instances automatically.

### Startup process

1. Install Docker
2. Authenticate to ECR
3. Retrieve secrets from SSM
4. Create environment configuration
5. Pull Docker image
6. Start backend container

# Networking

Infrastructure is deployed inside a custom VPC.

## Components

| Component | Purpose |
|---|---|
| Public Subnets | ALB + EC2 |
| Private Subnets | Redis |
| Security Groups | Access control |
| Internet Gateway | External access |

# Database Architecture

Primary database:

- MongoDB Atlas

MongoDB Atlas was used instead of self-hosted MongoDB to simplify infrastructure management.

# Caching Layer

Caching uses:

- Amazon ElastiCache Redis

Redis is used for:

- username caching
- performance optimization
- reducing repeated database reads

# Secret Management

Secrets are stored in:

- AWS Systems Manager Parameter Store

Secrets were intentionally removed from:

- Terraform variables
- userdata hardcoding
- GitHub repository code

This improved security and deployment flexibility.

# Authentication Architecture

## Original Design

Initial authentication relied on cookies.

This failed reliably because:

- frontend and backend were cross-origin
- S3 website hosting does not work well with cross-site cookies

## Final Design

Authentication was redesigned around:

- JWT Bearer Tokens

### Flow

1. User logs in
2. Backend returns JWT token
3. Frontend stores token in localStorage
4. Axios interceptor attaches token
5. Backend validates Authorization header

This became the stable production authentication model.

# CORS Architecture

Because frontend bucket names change dynamically:

```text
dev-starttech-frontend-<random-id>
```

hardcoded origins caused deployment instability.

The backend middleware was redesigned to:

- support localhost development
- support dynamic StartTech S3 website origins
- handle browser preflight requests correctly

# CI/CD Architecture

GitHub Actions pipeline stages:

```text
Push To GitHub
   │
   ▼
Run Go Tests
   │
   ▼
Run Security Scans
   │
   ▼
Build Docker Image
   │
   ▼
Push Image To ECR
   │
   ▼
Deploy Backend To EC2
   │
   ▼
Build Frontend
   │
   ▼
Deploy Frontend To S3
```

# Monitoring & Operations

## Monitoring tools used

| Tool | Purpose |
|---|---|
| CloudWatch Logs | Backend logs |
| ALB Health Checks | Service health |
| SSM Run Command | Remote diagnostics |

# Major Operational Challenges

| Problem | Resolution |
|---|---|
| Terraform module references | Fixed variable/output structure |
| CloudFront unavailable | Switched to S3 website hosting |
| Bucket policy failures | Updated S3 public access settings |
| CORS failures | Dynamic origin validation |
| Cookie auth failures | JWT token authentication |
| Backend unhealthy targets | Fixed deployment and userdata |
| Hardcoded secrets | Migrated to SSM |

# Final Architecture Characteristics

The final system provides:

- Infrastructure as Code
- Automated deployments
- Centralized secret management
- Containerized backend services
- Scalable compute layer
- Static frontend hosting
- Token-based authentication
- Operational monitoring
- Production-style cloud deployment
```
````

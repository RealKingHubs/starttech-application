# StartTech Application

StartTech is a full-stack task management application with a React/Vite frontend, a Go/Gin backend, deployment scripts, and separate GitHub Actions pipelines for each app surface.

## Project Structure

```text
starttech-application/
├── .github/
│   └── workflows/
│       ├── frontend-ci-cd.yml
│       └── backend-ci-cd.yml
├── frontend/
├── backend/
├── scripts/
│   ├── deploy-frontend.sh
│   ├── deploy-backend.sh
│   ├── health-check.sh
│   └── rollback.sh
└── README.md
```

## Local Development

### Backend

```bash
cd starttech-application/backend
go mod download
go run ./cmd/api/main.go
```

### Frontend

```bash
cd starttech-application/frontend
npm install
npm run dev
```

## CI/CD

- `backend-ci-cd.yml` tests, builds, scans, pushes, and deploys the backend.
- `frontend-ci-cd.yml` installs dependencies, builds the frontend, and deploys the static assets to S3.

## Deployment Scripts

- `scripts/deploy-backend.sh`: deploys the backend container to EC2 through SSM.
- `scripts/deploy-frontend.sh`: syncs the built frontend assets to the target S3 bucket.
- `scripts/health-check.sh`: polls the backend health endpoint after deployment.
- `scripts/rollback.sh`: redeploys a previous backend image by tag or full ECR image URI.

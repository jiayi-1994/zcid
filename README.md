# zcid

Cloud-native CI/CD platform built on **Tekton** (CI engine) and **ArgoCD** (CD engine).

zcid acts as a "translation layer + status dashboard" — it translates user intent from a visual UI into Kubernetes CRD resources, monitors execution status, and presents real-time results. It does not reinvent task scheduling or deployment orchestration.

## Architecture

```
┌──────────────────────────────────────────────────────────────┐
│                     zcid Platform                            │
│                                                              │
│  ┌──────────┐   ┌──────────┐   ┌───────────┐               │
│  │  React    │   │  Go/Gin  │   │ WebSocket │               │
│  │ Frontend  │──▶│ Backend  │──▶│   Hub     │               │
│  │ Arco UI   │   │  API     │   │ Real-time │               │
│  └──────────┘   └────┬─────┘   └───────────┘               │
│                      │                                       │
│         ┌────────────┼────────────┐                         │
│         ▼            ▼            ▼                          │
│  ┌────────────┐ ┌─────────┐ ┌──────────┐                   │
│  │ PostgreSQL │ │  Redis   │ │  MinIO   │                   │
│  │ (Data)     │ │ (Cache)  │ │ (Logs)   │                   │
│  └────────────┘ └─────────┘ └──────────┘                   │
│                      │                                       │
│         ┌────────────┼────────────┐                         │
│         ▼            ▼            ▼                          │
│  ┌────────────┐ ┌─────────┐ ┌──────────┐                   │
│  │  Tekton    │ │ ArgoCD  │ │  Harbor  │                   │
│  │  (CI Run)  │ │ (CD)    │ │(Registry)│                   │
│  └────────────┘ └─────────┘ └──────────┘                   │
└──────────────────────────────────────────────────────────────┘
```

### Tech Stack

| Layer | Technology |
|-------|-----------|
| **Backend** | Go 1.25, Gin, GORM, golang-migrate |
| **Frontend** | React 19, TypeScript, Arco Design, @xyflow/react |
| **Database** | PostgreSQL 16 |
| **Cache** | Redis 7 |
| **Object Storage** | MinIO |
| **CI Engine** | Tekton Pipelines v0.62 |
| **CD Engine** | ArgoCD v2.13 |
| **Container Registry** | Harbor (or any OCI registry) |
| **Container Runtime** | Kubernetes 1.25+ |

## Features

- **Visual Pipeline Editor** — Drag-and-drop DAG editor with stage/step nodes, or switch to YAML mode
- **Pipeline Templates** — Built-in templates (Go, Java, Node.js) for one-click pipeline creation
- **Tekton CRD Translation** — Converts visual pipeline config to Tekton PipelineRun resources
- **Dual Build Chains** — Containerized (Kaniko → Harbor → ArgoCD) and traditional (compile → MinIO archive)
- **Real-time Logs** — WebSocket-based live log streaming with secret masking
- **ArgoCD Deployment** — GitOps-style deployment with sync, status monitoring, and rollback
- **Git Integration** — GitHub/GitLab repository connection, branch selection, webhook auto-trigger
- **Variable Management** — Global / project / pipeline scoped variables with AES-256-GCM encryption for secrets
- **RBAC** — Role-based access control (admin / owner / maintainer / member)
- **Notification Rules** — Configurable webhook notifications for pipeline and deployment events
- **Audit Logging** — Comprehensive audit trail for all sensitive operations
- **Dashboard** — Project overview with build statistics and quick actions
- **Onboarding** — Guided first-use experience for new users

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 22+
- PostgreSQL 16+
- Redis 7+

### Local Development

```bash
# 1. Clone the repo
git clone https://github.com/jiayi-1994/zcid.git
cd zcid

# 2. Copy and configure environment
cp config/config.yaml.example config/config.yaml
# Edit config/config.yaml with your database credentials
# Or use environment variables (see Configuration section)

# 3. Run database migrations
export DB_URL="postgres://zcid:password@localhost:5432/zcid?sslmode=disable"
go run cmd/migrate/main.go up

# 4. Start the backend
go run cmd/server/main.go

# 5. Start the frontend (in another terminal)
cd web
npm install
npm run dev
```

The frontend will be available at `http://localhost:5173` and the API at `http://localhost:8080`.

**Default admin account**: `admin` / `admin123` (created by migration seed)

### Configuration

zcid uses a layered configuration: YAML file defaults → environment variable overrides. Environment variables always take precedence.

| Environment Variable | Description | Default |
|---------------------|-------------|---------|
| `SERVER_PORT` | HTTP server port | `8080` |
| `GIN_MODE` | Gin mode (debug/release) | `debug` |
| `DB_HOST` | PostgreSQL host | `localhost` |
| `DB_PORT` | PostgreSQL port | `5432` |
| `DB_NAME` | Database name | `zcid` |
| `DB_USER` | Database user | `zcid` |
| `DB_PASSWORD` | Database password | — |
| `DB_SSL_MODE` | SSL mode | `disable` |
| `DB_URL` | Full database URL (for migrate CLI) | — |
| `REDIS_HOST` | Redis host | `localhost` |
| `REDIS_PORT` | Redis port | `6379` |
| `REDIS_PASSWORD` | Redis password | — |
| `REDIS_DB` | Redis database number | `0` |
| `MINIO_ENDPOINT` | MinIO endpoint | `localhost:9000` |
| `MINIO_ACCESS_KEY` | MinIO access key | — |
| `MINIO_SECRET_KEY` | MinIO secret key | — |
| `MINIO_USE_SSL` | Use SSL for MinIO | `false` |
| `JWT_SECRET` | JWT signing key | — |
| `ZCID_ENCRYPTION_KEY` | AES-256 encryption key (hex, 32 bytes) | — |

## Deployment

### Docker

The project builds a multi-stage Docker image (frontend + backend in a single image):

```bash
docker build -t zcid:latest .
docker run -p 8080:8080 \
  -e DB_HOST=your-db-host \
  -e DB_PASSWORD=your-password \
  -e JWT_SECRET=your-jwt-secret \
  -e ZCID_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef \
  zcid:latest
```

### Kubernetes (Helm)

```bash
# Install middleware first (see deploy/middleware/README.md)
# Then install zcid:
helm install zcid deploy/helm/zcid/ \
  --namespace zcid \
  --set image.tag=main \
  --set secrets.dbPassword=your-db-password \
  --set secrets.jwtSecret=your-jwt-secret \
  --set config.encryptionKey=0123456789abcdef0123456789abcdef
```

Enable ingress:

```bash
helm install zcid deploy/helm/zcid/ \
  --namespace zcid \
  --set ingress.enabled=true \
  --set ingress.hosts[0].host=zcid.example.com \
  --set ingress.hosts[0].paths[0].path=/ \
  --set ingress.hosts[0].paths[0].pathType=Prefix
```

### Full Kubernetes Deployment

For a complete deployment including all dependencies:

1. **Middleware** (PostgreSQL, Redis, MinIO): [`deploy/middleware/README.md`](deploy/middleware/README.md)
2. **Tekton** (CI engine): [`deploy/tekton/README.md`](deploy/tekton/README.md)
3. **ArgoCD** (CD engine): [`deploy/argocd/README.md`](deploy/argocd/README.md)
4. **zcid** (application): Helm chart at `deploy/helm/zcid/`

#### Deployment Order

```bash
# Step 1: Create namespace
kubectl create namespace zcid

# Step 2: Deploy middleware
# Follow deploy/middleware/README.md
# This installs PostgreSQL, Redis, MinIO

# Step 3: Deploy Tekton
# Follow deploy/tekton/README.md
# This installs Tekton Pipelines + Dashboard + RBAC

# Step 4: Deploy ArgoCD
# Follow deploy/argocd/README.md
# This installs ArgoCD + zcid account + AppProject

# Step 5: Deploy zcid
helm install zcid deploy/helm/zcid/ \
  --namespace zcid \
  --set secrets.dbPassword=zcid-password \
  --set secrets.jwtSecret=change-me-in-production \
  --set config.encryptionKey=0123456789abcdef0123456789abcdef

# Step 6: Verify
kubectl get pods -n zcid
kubectl port-forward svc/zcid -n zcid 8080:8080
# Visit http://localhost:8080
```

#### Resource Requirements (Total)

| Component | CPU Request | Memory Request | Disk |
|-----------|------------|----------------|------|
| PostgreSQL | 250m | 256Mi | 20Gi |
| Redis | 100m | 128Mi | 2Gi |
| MinIO | 250m | 256Mi | 50Gi |
| Tekton | 500m | 512Mi | — |
| ArgoCD | 550m | 640Mi | — |
| zcid | 100m | 128Mi | — |
| **Total** | **1750m** | **1920Mi** | **72Gi** |

## Usage Guide

### 1. First Login

After deployment, visit the zcid URL and log in with:
- **Username**: `admin`
- **Password**: `admin123`

> Change the admin password immediately after first login via the user profile page.

### 2. Create a Project

1. Click "**新建项目**" on the dashboard
2. Fill in the project name and description
3. The project is created with default environments (dev, staging, production)

### 3. Connect Git Repository

1. Navigate to **Project → Git 仓库**
2. Click "**添加连接**"
3. Select provider (GitHub or GitLab)
4. Enter the repository URL and access token
5. Test the connection and save

### 4. Create a Pipeline

**From template (recommended for beginners):**
1. Navigate to **Project → 流水线**
2. Click "**从模板创建**"
3. Select a template (e.g., `go-docker-build`, `java-maven-docker`)
4. Fill in template parameters (image name, registry, etc.)
5. Save — the pipeline is ready to run

**Visual editor:**
1. Click "**新建流水线**"
2. Add stages (e.g., "Build", "Test", "Deploy")
3. Add steps to each stage (Git Clone, Build, Push, etc.)
4. Configure step parameters in the side panel
5. Toggle to YAML mode to review the generated configuration
6. Save

### 5. Run a Pipeline

- **Manual**: Click the "**运行**" button on any pipeline
- **Webhook**: Configure a Git webhook to auto-trigger on push events
  1. Navigate to **Pipeline → 设置 → 触发方式**
  2. Select "Webhook"
  3. Copy the webhook URL and configure it in your Git provider

### 6. Monitor Pipeline Runs

1. Navigate to **Project → 流水线运行**
2. View the real-time status of each stage/step
3. Click a run to see live logs (WebSocket streaming)
4. Failed steps show error details inline

### 7. Deploy with ArgoCD

1. Navigate to **Project → 部署**
2. Click "**发起部署**"
3. Select the target environment and service
4. Specify the image tag or Git revision
5. Click "**部署**" — zcid creates/updates an ArgoCD Application
6. Monitor sync status and health in real-time
7. Rollback to a previous version if needed

### 8. Manage Variables & Secrets

1. Navigate to **Project → 变量**
2. Add variables at global, project, or pipeline scope
3. Mark sensitive values as "secret" — they are encrypted with AES-256-GCM
4. Variables are injected into pipeline runs as environment variables
5. Secrets in logs are automatically masked with `***`

### 9. Notification Rules

1. Navigate to **Project → 通知**
2. Create rules for pipeline success/failure events
3. Configure webhook URLs for notifications (DingTalk, Feishu, Slack, etc.)
4. Toggle rules on/off as needed

### 10. Admin Operations

Accessible to `admin` role only:

- **用户管理**: Create/edit/disable users, assign roles
- **系统设置**: Configure K8s API URL, default registry, ArgoCD endpoint
- **审计日志**: View all platform operations with filters (user, action, time range)
- **全局变量**: Manage platform-wide variables and secrets
- **集成管理**: View integration health status (Tekton, ArgoCD, Registry)

## Project Structure

```
zcid/
├── cmd/
│   ├── server/          # Application entry point
│   └── migrate/         # Database migration CLI
├── config/              # Configuration package
├── internal/            # Business logic (not importable externally)
│   ├── admin/           # System admin & health endpoints
│   ├── audit/           # Audit logging
│   ├── auth/            # Authentication (JWT) & user management
│   ├── crdclean/        # CRD cleanup scheduler
│   ├── deployment/      # ArgoCD deployment management
│   ├── environment/     # Environment CRUD
│   ├── git/             # Git connection & webhook handling
│   ├── logarchive/      # Log archival (MinIO)
│   ├── notification/    # Notification rules & dispatch
│   ├── pipeline/        # Pipeline CRUD, templates, config
│   ├── pipelinerun/     # Pipeline run orchestration
│   ├── project/         # Project management
│   ├── rbac/            # Role-based access control
│   ├── registry/        # Container registry management
│   ├── svcdef/          # Service definition CRUD
│   ├── variable/        # Variable management (scoped, encrypted)
│   └── ws/              # WebSocket hub, log streaming, watcher
├── pkg/                 # Shared packages (importable)
│   ├── argocd/          # ArgoCD gRPC client
│   ├── cache/           # Redis cache wrapper
│   ├── crypto/          # AES-256-GCM encryption
│   ├── database/        # PostgreSQL connection & migrations
│   ├── gitprovider/     # GitHub/GitLab API abstraction
│   ├── k8s/             # Kubernetes secret helpers
│   ├── logging/         # Structured logging (slog)
│   ├── middleware/       # Gin middleware (auth, RBAC, error handling)
│   ├── response/        # Standardized API response & error codes
│   ├── storage/         # MinIO storage client
│   └── tekton/          # Tekton CRD types, translator, build chains
├── migrations/          # SQL migration files (000001 - 000018)
├── web/                 # React frontend
│   ├── src/
│   │   ├── components/  # Reusable UI components
│   │   ├── pages/       # Page components (routing)
│   │   ├── services/    # API client layer
│   │   ├── stores/      # Zustand state management
│   │   ├── styles/      # Global CSS & design tokens
│   │   ├── hooks/       # Custom React hooks
│   │   └── lib/         # Utility libraries
│   └── ...
├── deploy/              # Deployment manifests
│   ├── helm/zcid/       # Helm chart
│   ├── middleware/      # PostgreSQL, Redis, MinIO deployment guide
│   ├── tekton/          # Tekton deployment guide + RBAC
│   └── argocd/          # ArgoCD deployment guide + AppProject
├── files/               # BMAD development artifacts
│   ├── planning-artifacts/   # PRD, architecture, epics
│   └── implementation-artifacts/  # Story specs, retrospectives
├── Dockerfile           # Multi-stage build
├── Makefile             # Build commands
└── .github/workflows/   # GitHub Actions CI
```

## Development

### Running Tests

```bash
# Backend tests
go test ./pkg/... ./internal/... -count=1

# Backend tests with race detection
go test ./pkg/... ./internal/... -count=1 -race

# Frontend type check
cd web && npx tsc --noEmit

# Frontend unit tests
cd web && npx vitest run
```

### Database Migrations

```bash
# Apply all pending migrations
go run cmd/migrate/main.go up

# Rollback the latest migration
go run cmd/migrate/main.go down

# Create a new migration
go run cmd/migrate/main.go new --name add_some_table
```

### Building

```bash
# Build backend
go build -o zcid-server ./cmd/server

# Build frontend
cd web && npm run build

# Build Docker image
docker build -t zcid:latest .
```

## CI/CD

GitHub Actions automatically runs on push to `main` or version tags:

1. **Test** — Go build/test (with race detection), frontend TypeScript check, Vitest
2. **Build & Push** — Multi-stage Docker build, push to `ghcr.io/jiayi-1994/zcid`

Image tags:
- `main` branch → `ghcr.io/jiayi-1994/zcid:main`
- `v1.0.0` tag → `ghcr.io/jiayi-1994/zcid:1.0.0`
- Each commit → `ghcr.io/jiayi-1994/zcid:<sha>`

## License

Private project. All rights reserved.

# zcid

Cloud-native CI/CD platform built on **Tekton** (CI) and **ArgoCD** (CD).

zcid translates visual pipeline configurations into Kubernetes CRD resources, monitors execution status in real time, and manages the full CI/CD lifecycle from code checkout to production deployment.

## Screenshots

| Dashboard | Pipeline Editor | Pipeline Run |
|-----------|----------------|--------------|
| Two-column layout with metrics, project list, quick actions | Fullscreen horizontal DAG editor with stage/step nodes | Status lifecycle: queued → running → succeeded |

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                    zcid Platform                     │
│                                                     │
│  React 19 + Arco Design  ←→  Go/Gin REST API       │
│  Visual Pipeline Editor       WebSocket Hub         │
│                                                     │
│         PostgreSQL    Redis    MinIO                 │
│         (data)        (cache)  (logs/artifacts)      │
│                                                     │
│         Tekton v0.62  ArgoCD v2.13  Harbor           │
│         (CI runs)     (CD deploy)   (registry)       │
└─────────────────────────────────────────────────────┘
```

## Tech Stack

| Layer | Technology |
|-------|-----------|
| Backend | Go 1.25, Gin, GORM, golang-migrate |
| Frontend | React 19, TypeScript, Arco Design, @xyflow/react |
| Database | PostgreSQL 16 |
| Cache | Redis 7 |
| Object Storage | MinIO |
| CI Engine | Tekton Pipelines v0.56–v0.62 |
| CD Engine | ArgoCD v2.11–v2.13 |
| Container Registry | Harbor (or any OCI registry) |

## Features

### CI/CD Core
- **Visual Pipeline Editor** — Fullscreen horizontal DAG editor, or JSON mode
- **Pipeline Templates** — Built-in templates: Go, Java Maven, Node.js, Docker, Java JAR
- **Step Types** — Git Clone, Shell script (dark-themed editor), Kaniko build, BuildKit build
- **Tekton Translation** — Converts visual config → Tekton PipelineRun CRD
- **Run Lifecycle** — Mock: queued→running→succeeded/failed with timing. Real: Tekton CRD status sync
- **ArgoCD Deployment** — GitOps deploy with sync, status monitoring, rollback
- **Dual Build Chains** — Containerized (Kaniko→Harbor→ArgoCD) and traditional (compile→MinIO)

### Platform
- **Git Integration** — GitHub/GitLab connection, branch selection, webhook auto-trigger
- **Variable Management** — Global/project/pipeline scoped, AES-256-GCM encrypted secrets
- **RBAC** — Role-based access control (admin / owner / maintainer / member)
- **Notification Rules** — Webhook notifications for build/deploy events
- **Audit Logging** — Full audit trail for all operations
- **System Settings** — K8s, ArgoCD, registry configuration + health monitoring

### UI/UX (shadcn/ui-inspired)
- **Neutral Design System** — Zinc palette, Inter font, minimal shadows
- **Clean Sidebar** — Section-grouped navigation with user menu
- **Project Cards** — Card grid with avatars, status badges, search
- **Dashboard** — Time-based greeting, metrics, two-column layout
- **Fullscreen Editor** — No sidebar interference, horizontal stage flow
- **Step Template Picker** — One-click stage+step creation
- **Branch Dropdown** — Select from common branches or type custom
- **Inline Stage Rename** — Double-click to edit stage names

## Quick Start

### Prerequisites

- Go 1.25+
- Node.js 22+
- PostgreSQL 16+
- Redis 7+
- MinIO (for log archival)

### Local Development

```bash
# 1. Clone
git clone https://github.com/jiayi-1994/zcid.git && cd zcid

# 2. Configure
cp config/config.yaml.example config/config.yaml
# Edit config/config.yaml with DB password, MinIO keys, JWT secret

# 3. Migrate
export DB_URL="postgres://zcid:password@localhost:5432/zcid?sslmode=disable"
go run cmd/migrate/main.go up

# 4. Start backend
export ZCID_ENCRYPTION_KEY="0123456789abcdef0123456789abcdef"
go run cmd/server/main.go

# 5. Start frontend
cd web && npm install && npm run dev
```

Frontend: `http://localhost:5173` · API: `http://localhost:8080` · Login: `admin` / `admin123`

> Without K8s/Tekton/ArgoCD, zcid runs in **mock mode** — pipeline runs simulate a realistic lifecycle (queued→running→succeeded in ~10s) and deployments simulate ArgoCD sync.

### Configuration

Environment variables override `config/config.yaml`:

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` / `DB_PORT` / `DB_NAME` / `DB_USER` / `DB_PASSWORD` | PostgreSQL | `localhost:5432/zcid` |
| `REDIS_HOST` / `REDIS_PORT` / `REDIS_PASSWORD` | Redis | `localhost:6379` |
| `MINIO_ENDPOINT` / `MINIO_ACCESS_KEY` / `MINIO_SECRET_KEY` | MinIO | `localhost:9000` |
| `JWT_SECRET` | JWT signing key | — |
| `ZCID_ENCRYPTION_KEY` | AES-256 key (32-byte hex) | — |
| `ZCID_K8S_ENABLED` | Enable real K8s integration | `false` (auto-detected) |
| `ARGOCD_SERVER` / `ARGOCD_AUTH_TOKEN` | ArgoCD connection | — (mock if unset) |

## Deployment

### Docker

```bash
docker build -t zcid:latest .
docker run -p 8080:8080 \
  -e DB_HOST=db -e DB_PASSWORD=secret \
  -e JWT_SECRET=your-key \
  -e ZCID_ENCRYPTION_KEY=0123456789abcdef0123456789abcdef \
  zcid:latest
```

### Kubernetes (Helm)

Helm chart automatically handles all K8s/ArgoCD/Tekton environment variables. See [`deploy/README.md`](deploy/README.md) for the complete deployment guide.

```bash
helm install zcid deploy/helm/zcid/ --namespace zcid \
  --set secrets.dbPassword=your-password \
  --set secrets.jwtSecret=your-jwt-key \
  --set config.encryptionKey=0123456789abcdef0123456789abcdef
```

## Development

```bash
# Backend
go test ./pkg/... ./internal/... -count=1       # unit tests
go test ./pkg/... ./internal/... -count=1 -race  # with race detection
go build -o zcid-server ./cmd/server             # build binary

# Frontend
cd web
npx tsc --noEmit     # type check
npx vitest run       # unit tests
npm run build        # production build

# Database
go run cmd/migrate/main.go up                     # apply migrations
go run cmd/migrate/main.go down                   # rollback
go run cmd/migrate/main.go new --name add_table   # create migration
```

## Project Structure

```
zcid/
├── cmd/server/              # Application entry point
├── cmd/migrate/             # Database migration CLI
├── config/                  # Configuration loading
├── internal/                # Business logic
│   ├── auth/                # JWT authentication + user management
│   ├── pipeline/            # Pipeline CRUD + templates
│   ├── pipelinerun/         # Run orchestration + K8s executor
│   ├── deployment/          # ArgoCD deployment management
│   ├── git/                 # Git connections + webhooks
│   ├── variable/            # Scoped variable management
│   ├── notification/        # Webhook notifications
│   ├── rbac/                # Casbin RBAC
│   ├── audit/               # Audit logging
│   ├── ws/                  # WebSocket hub + log streaming
│   └── ...                  # environment, project, svcdef, etc.
├── pkg/                     # Shared packages
│   ├── tekton/              # Tekton CRD types + translator
│   ├── argocd/              # ArgoCD client (real + mock)
│   └── ...                  # database, cache, crypto, etc.
├── migrations/              # SQL migrations (000001–000018)
├── web/                     # React frontend
│   └── src/
│       ├── components/pipeline/  # Visual editor, step config
│       ├── pages/               # All page components
│       ├── services/            # API client layer
│       └── styles/              # Global CSS (shadcn-inspired)
├── deploy/                  # Deployment manifests
│   ├── helm/zcid/           # Helm chart
│   ├── middleware/           # PostgreSQL, Redis, MinIO guides
│   ├── tekton/              # Tekton deployment + RBAC
│   └── argocd/              # ArgoCD deployment + config
├── Dockerfile               # Multi-stage build
├── Makefile                 # Build commands
└── .github/workflows/       # GitHub Actions CI
```

## CI/CD (Self-hosted)

GitHub Actions on push to `main` or version tags:

1. **Test** — Go build/test + frontend TypeScript check + Vitest
2. **Build & Push** — Docker multi-stage → `ghcr.io/jiayi-1994/zcid`

## License

Private project. All rights reserved.

# AGENTS.md

## Cursor Cloud specific instructions

### Overview

zcid is a cloud-native CI/CD platform with a Go/Gin backend and React/TypeScript frontend. It requires PostgreSQL 16+, Redis 7+, and MinIO as infrastructure dependencies. Kubernetes, Tekton, and ArgoCD are optional — the app gracefully falls back to mock mode without them.

### Services

| Service | Port | Purpose |
|---------|------|---------|
| Go backend | 8080 | API server (Gin) |
| Vite dev server | 5173 | React frontend with HMR |
| PostgreSQL | 5432 | Primary data store |
| Redis | 6379 | Cache, sessions, RBAC watcher |
| MinIO | 9000 | Object storage for logs/artifacts |

### Starting infrastructure services

```bash
sudo pg_ctlcluster 16 main start
sudo redis-server --daemonize yes
sudo chmod -R 777 /data/minio && MINIO_ROOT_USER=minioadmin MINIO_ROOT_PASSWORD=minioadmin minio server /data/minio --console-address ":9001" &
```

### Configuration

Copy `config/config.yaml.example` to `config/config.yaml` and set credentials. Key env vars:
- `ZCID_ENCRYPTION_KEY` — 32-byte hex key for AES-256-GCM (e.g. `0123456789abcdef0123456789abcdef`)
- `DB_URL` — only needed for the migrate CLI (`postgres://zcid:password@localhost:5432/zcid?sslmode=disable`)

Default admin login: `admin` / `admin123` (seeded by migration 000002).

### Running the application

See `README.md` "Local Development" section. Key commands:
- Backend: `go run cmd/server/main.go` (or `make dev`)
- Frontend: `cd web && npm run dev`
- Migrations: `DB_URL="postgres://zcid:password@localhost:5432/zcid?sslmode=disable" go run cmd/migrate/main.go up`

### Lint, test, build

- **Backend lint**: `go build ./...` (or `make lint` if golangci-lint is installed)
- **Backend tests**: `go test ./pkg/... ./internal/... -count=1`
- **Frontend lint**: `cd web && npx tsc --noEmit`
- **Frontend tests**: `cd web && npx vitest run`
- **Frontend build**: `cd web && npm run build`

### Gotchas

- MinIO needs write permissions on its data directory (`/data/minio`). If you see "file access denied", run `sudo chmod -R 777 /data/minio` before starting MinIO.
- The backend requires `config/config.yaml` to exist (copied from `config/config.yaml.example`) with DB password and MinIO secret key filled in.
- The `ZCID_ENCRYPTION_KEY` env var must be set when running the backend for variable encryption features to work.
- Redis is started without a password in local dev. The config.yaml `redis.password` field should be left empty.
- Without K8s, Tekton, and ArgoCD, the app automatically uses mock clients — no configuration needed.

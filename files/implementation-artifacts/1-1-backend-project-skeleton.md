# Story 1.1: 后端项目骨架与开发环境

Status: done

## Story

As a 开发者,
I want 一个可运行的 Go+Gin 后端项目骨架和本地开发环境,
so that 我可以立即开始编写业务代码。

## Acceptance Criteria (BDD)

1. **Given** 开发者克隆了代码仓库 **When** 执行 `docker-compose up -d` 和 `make dev` **Then** PostgreSQL、Redis、MinIO 容器启动，后端服务在 localhost:8080 运行 **And** `GET /healthz` 返回 200

2. **Given** 后端服务启动 **When** 访问 `GET /readyz` **Then** 返回 DB 和 Redis 连接状态 **And** 连接正常时返回 200，异常时返回 503

3. **Given** config.yaml 和环境变量均配置了同一字段 **When** 服务启动加载配置 **Then** 环境变量值覆盖 config.yaml 值 **And** 敏感配置（密码、密钥）仅通过环境变量注入

## Tasks / Subtasks

- [x] Task 1: 初始化 Go 模块与核心依赖 (AC: #1)
  - [x] 1.1 `go mod init github.com/xjy/zcid`
  - [x] 1.2 安装核心依赖：gin@v1.12.0, gorm@v1.30.1, gorm/driver/postgres, go-redis/v9, minio-go/v7
  - [x] 1.3 安装工具依赖：swaggo/swag/v2, golang-jwt/jwt/v5, casbin/v2（预留，不在本 Story 实现业务逻辑）
  - [x] 1.4 创建 `cmd/server/main.go` 入口文件

- [x] Task 2: 创建后端目录结构 (AC: #1)
  - [x] 2.1 创建 `internal/` 业务模块目录（暂为空目录或占位 .gitkeep）
  - [x] 2.2 创建 `pkg/middleware/` 目录，添加空的中间件文件占位
  - [x] 2.3 创建 `pkg/response/` 目录（统一响应格式 — 本 Story 只建骨架，Story 1.3 详细实现）
  - [x] 2.4 创建 `migrations/` 目录
  - [x] 2.5 创建 `docs/` 目录（swag 生成输出目标）

- [x] Task 3: 应用配置管理 (AC: #3)
  - [x] 3.1 创建 `config/config.go`：定义 Config struct（Server/DB/Redis/MinIO 配置块）
  - [x] 3.2 实现 YAML 加载 + 环境变量覆盖逻辑（使用 `os.Getenv` 或 viper 轻量方案）
  - [x] 3.3 创建 `config.yaml` 模板（开发环境默认值）
  - [x] 3.4 敏感字段（DB_PASSWORD、REDIS_PASSWORD、MINIO_SECRET_KEY、ENCRYPT_KEY）仅从环境变量读取

- [x] Task 4: 数据库连接 (AC: #1, #2)
  - [x] 4.1 创建 `pkg/database/postgres.go`：GORM 初始化，连接池配置
  - [x] 4.2 配置连接池参数：MaxOpenConns=25, MaxIdleConns=10, ConnMaxLifetime=5min
  - [x] 4.3 Ping 检测用于健康检查

- [x] Task 5: Redis 连接 (AC: #1, #2)
  - [x] 5.1 创建 `pkg/database/redis.go`：go-redis/v9 客户端初始化
  - [x] 5.2 Ping 检测用于健康检查

- [x] Task 6: MinIO 连接 (AC: #1)
  - [x] 6.1 创建 `pkg/storage/minio.go`：minio-go/v7 客户端初始化
  - [x] 6.2 启动时检查/创建默认 bucket（如 `zcid-logs`, `zcid-artifacts`）

- [x] Task 7: 健康检查端点 (AC: #1, #2)
  - [x] 7.1 `GET /healthz` — 无依赖检查，直接返回 200（存活探针）
  - [x] 7.2 `GET /readyz` — 检查 DB 连接 + Redis 连接，全部正常返回 200，任一失败返回 503 + 详细状态 JSON
  - [x] 7.3 响应格式：`{"status": "ok/degraded", "checks": {"db": "ok", "redis": "ok"}}`

- [x] Task 8: docker-compose.yml (AC: #1)
  - [x] 8.1 PostgreSQL 16 容器（端口 5432，DB=zcid，用户=zcid）
  - [x] 8.2 Redis 7-alpine 容器（端口 6379）
  - [x] 8.3 MinIO 容器（端口 9000 API + 9001 Console）
  - [x] 8.4 Volume 持久化配置

- [x] Task 9: Makefile (AC: #1)
  - [x] 9.1 `make dev` — 启动后端服务（`go run cmd/server/main.go`）
  - [x] 9.2 `make build` — 编译二进制
  - [x] 9.3 `make test` — 运行测试
  - [x] 9.4 `make swag` — 生成 OpenAPI v3 文档（swag v2）
  - [x] 9.5 `make lint` — golangci-lint（如果安装）

- [x] Task 10: 基础测试 (AC: #1, #2)
  - [x] 10.1 健康检查端点测试（httptest）
  - [x] 10.2 配置加载测试（环境变量覆盖验证）

## Dev Notes

### 架构约束（必须遵守）

- **代码组织**：handler→service→repo 三层结构，按业务模块组织。本 Story 只建目录结构，不创建业务模块代码
  - [Source: architecture.md#Backend Code Organization Convention]
- **命名约定**：模块目录名单数形式（`pipeline` 非 `pipelines`），文件名 `handler.go`/`service.go`/`repo.go`/`model.go`/`dto.go`
  - [Source: architecture.md#Backend Code Organization Convention]
- **API 路径前缀**：`/api/v1`
  - [Source: architecture.md#Core Architectural Decisions]
- **无 Starter Template**：架构评估结论为自定义初始化，不使用任何现成脚手架
  - [Source: architecture.md#Starter Template Evaluation]

### 关键技术版本（已验证）

| 依赖 | 版本 | 说明 |
|------|------|------|
| Go | 1.24+ | 当前稳定版 |
| Gin | v1.12.0 | 2026-02-28 发布 |
| GORM | v1.30.1 | Go Generics 支持 |
| go-redis | v9 | 最新大版本 |
| minio-go | v7 | 最新大版本 |
| swaggo/swag | **v2.0.0** | OpenAPI v3（非 v1 的 Swagger 2.0）|
| PostgreSQL | 16 | docker-compose |
| Redis | 7-alpine | docker-compose |

[Source: architecture.md#Verified Dependency Versions]

### docker-compose 精确配置

```yaml
services:
  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: zcid
      POSTGRES_USER: zcid
      POSTGRES_PASSWORD: zcid_dev
    volumes:
      - pgdata:/var/lib/postgresql/data
  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]
  minio:
    image: minio/minio
    ports: ["9000:9000", "9001:9001"]
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: zcid
      MINIO_ROOT_PASSWORD: zcid_dev_key
    volumes:
      - miniodata:/data
volumes:
  pgdata:
  miniodata:
```

[Source: architecture.md#Development Environment docker-compose.yml]

### 后端目录结构（本 Story 需创建）

```
zcid/
├── cmd/server/main.go
├── config/
│   ├── config.go
│   └── config.yaml
├── internal/               # 暂为空目录，后续 Story 按需填充
│   ├── auth/
│   ├── project/
│   ├── pipeline/
│   └── ...
├── pkg/
│   ├── database/
│   │   ├── postgres.go
│   │   └── redis.go
│   ├── storage/
│   │   └── minio.go
│   ├── middleware/          # 占位，Story 1.3 详细实现
│   ├── response/            # 占位，Story 1.3 详细实现
│   └── masking/             # 占位，Story 1.4 详细实现
├── migrations/
├── docs/
├── Makefile
├── Dockerfile
├── docker-compose.yml
├── go.mod
└── go.sum
```

[Source: architecture.md#Backend Code Organization Convention]

### 配置管理模式

```go
type Config struct {
    Server   ServerConfig
    Database DatabaseConfig
    Redis    RedisConfig
    MinIO    MinIOConfig
}

type ServerConfig struct {
    Port string `yaml:"port" env:"SERVER_PORT" default:"8080"`
}

type DatabaseConfig struct {
    Host     string `yaml:"host" env:"DB_HOST" default:"localhost"`
    Port     int    `yaml:"port" env:"DB_PORT" default:"5432"`
    Name     string `yaml:"name" env:"DB_NAME" default:"zcid"`
    User     string `yaml:"user" env:"DB_USER" default:"zcid"`
    Password string `env:"DB_PASSWORD"` // 仅环境变量，不写入 YAML
    SSLMode  string `yaml:"ssl_mode" env:"DB_SSL_MODE" default:"disable"`
}
```

[Source: architecture.md#Core Architectural Decisions, prd.md#Implementation Constraints]

### 健康检查端点规范

- `/healthz` — 存活探针（Liveness），无依赖检查，直接 200
- `/readyz` — 就绪探针（Readiness），检查 DB + Redis
- `/api/v1/health` — 完整健康状态（后续 Story 10.3 实现，含 K8s/Tekton/ArgoCD）

[Source: architecture.md#ARCH-12]

### 不做的事（明确边界）

- **不创建业务模块代码**：internal/ 下只建目录结构
- **不创建数据库表**：migrations/ 目录就绪但不添加 SQL 文件（Story 1.2 实现迁移框架）
- **不实现统一错误处理**：pkg/response/ 只建占位（Story 1.3 实现）
- **不实现日志脱敏**：pkg/masking/ 只建占位（Story 1.4 实现）
- **不实现 JWT/Casbin**：只安装依赖，不编写认证逻辑（Story 2.1/2.3 实现）
- **不创建前端项目**：前端骨架是 Story 1.6

### Project Structure Notes

- 本 Story 建立的目录结构是所有后续 Story 的基础，必须严格遵循架构文档定义
- `internal/` 下的业务模块目录用 `.gitkeep` 占位，确保目录结构进入版本控制
- `cmd/server/main.go` 是唯一入口，负责初始化所有基础组件并注册路由

### References

- [Source: architecture.md#Backend Code Organization Convention] — 目录结构和命名约定
- [Source: architecture.md#Selected Approach: Custom Initialization] — 依赖安装命令
- [Source: architecture.md#Verified Dependency Versions] — 精确版本号
- [Source: architecture.md#Development Environment docker-compose.yml] — 容器配置
- [Source: architecture.md#Core Architectural Decisions] — JSONB 策略、配置管理、API 版本
- [Source: prd.md#Implementation Constraints] — 错误码段、API 设计约束
- [Source: epics.md#Epic 1] — Epic 上下文和 Story 依赖关系

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- Gin v1.12.0 requires Go 1.25+; go.mod auto-upgraded from 1.24 to 1.25.0
- swaggo/swag v2 resolved to v2.0.0-rc5 (latest available)
- Windows bash mkdir with backslashes concatenated path segments; fixed with forward slashes

### Completion Notes List

- All 10 tasks and 32 subtasks completed
- Go module initialized with github.com/xjy/zcid, Go 1.25.0
- Core deps: gin@v1.12.0, gorm@v1.30.1, go-redis/v9@v9.18.0, minio-go/v7@v7.0.98
- Tool deps: swag/v2@v2.0.0-rc5, jwt/v5@v5.3.1, casbin/v2@v2.135.0
- Config: YAML load + env override, sensitive fields (DB_PASSWORD, REDIS_PASSWORD, MINIO_SECRET_KEY) env-only via yaml:"-" tag
- DB: GORM + postgres driver, pool config (25/10/5min), Ping health check
- Redis: go-redis/v9, Ping health check
- MinIO: minio-go/v7, auto-create buckets (zcid-logs, zcid-artifacts)
- Health: GET /healthz (200 always), GET /readyz (200 ok / 503 degraded with checks JSON)
- docker-compose: PostgreSQL 16, Redis 7-alpine, MinIO with volumes
- Makefile: dev, build, test, swag, lint targets
- Tests: 8 tests (3 health endpoint + 5 config) all passing
- AC #1: docker-compose + make dev + /healthz 200 — satisfied
- AC #2: /readyz with DB+Redis status, 200/503 — satisfied
- AC #3: env overrides YAML, sensitive fields env-only — satisfied (verified by tests)

### Change Log

- 2026-03-02: Initial implementation of backend project skeleton (Story 1.1)
- 2026-03-02: Code review (AI) — H3 fix: admin routes wrapped in route group with TODO for Epic 2 auth middleware

### File List

- cmd/server/main.go (new) — application entry point, health routes
- cmd/server/main_test.go (new) — health endpoint tests
- config/config.go (new) — config structs, YAML + env loading
- config/config.yaml (new) — default dev config template
- config/config_test.go (new) — config loading and env override tests
- pkg/database/postgres.go (new) — GORM PostgreSQL init, pool config, ping
- pkg/database/redis.go (new) — go-redis client init, ping
- pkg/storage/minio.go (new) — MinIO client init, bucket creation
- docker-compose.yml (new) — PostgreSQL 16, Redis 7, MinIO containers
- Makefile (new) — dev, build, test, swag, lint targets
- go.mod (new) — Go module definition
- go.sum (new) — dependency checksums
- internal/auth/.gitkeep (new) — placeholder
- internal/project/.gitkeep (new) — placeholder
- internal/pipeline/.gitkeep (new) — placeholder
- pkg/middleware/.gitkeep (new) — placeholder
- pkg/response/.gitkeep (new) — placeholder
- pkg/masking/.gitkeep (new) — placeholder
- migrations/.gitkeep (new) — placeholder
- docs/.gitkeep (new) — placeholder

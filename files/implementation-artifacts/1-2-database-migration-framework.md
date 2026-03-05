# Story 1.2: 数据库迁移框架

Status: done

## Story

As a 开发者,
I want golang-migrate 迁移框架就绪,
so that 后续每个 Story 可以按需创建数据库表。

## Acceptance Criteria (BDD)

1. **Given** 迁移框架已集成 **When** 执行 `make migrate-up` **Then** migrations/ 目录下的 SQL 文件按序号执行 **And** 数据库 schema_migrations 表记录当前版本

2. **Given** 需要回滚 **When** 执行 `make migrate-down` **Then** 最近一次迁移被回滚

3. **Given** 开发者需要新建迁移 **When** 执行 `make migrate-new name=create_xxx` **Then** 生成带序号的 up/down SQL 文件对

## Tasks / Subtasks

- [x] Task 1: 安装 golang-migrate 依赖 (AC: #1)
  - [x] 1.1 `go get github.com/golang-migrate/migrate/v4`
  - [x] 1.2 安装 postgres 驱动：`go get github.com/golang-migrate/migrate/v4/database/postgres`
  - [x] 1.3 安装文件源驱动：`go get github.com/golang-migrate/migrate/v4/source/file`

- [x] Task 2: 创建初始迁移文件 (AC: #1, #3)
  - [x] 2.1 创建 `migrations/000001_init_schema.up.sql`：创建扩展（如 uuid-ossp）和基础设置
  - [x] 2.2 创建 `migrations/000001_init_schema.down.sql`：回滚初始 schema
  - [x] 2.3 确认 SQL 文件使用顺序编号格式（000001、000002...）

- [x] Task 3: 实现迁移执行逻辑 (AC: #1)
  - [x] 3.1 创建 `pkg/database/migrate.go`：封装 golang-migrate 调用逻辑
  - [x] 3.2 实现 `RunMigrations(dsn, migrationsPath string) error` 函数
  - [x] 3.3 实现 `RollbackMigration(dsn, migrationsPath string) error` 函数
  - [x] 3.4 迁移执行时打印当前版本和执行状态日志

- [x] Task 4: 集成到应用启动流程 (AC: #1)
  - [x] 4.1 在 `cmd/server/main.go` 中添加启动时自动执行迁移逻辑
  - [x] 4.2 通过配置或环境变量控制是否启动时自动迁移（`AUTO_MIGRATE=true`）
  - [x] 4.3 迁移失败时服务启动中止并输出错误日志

- [x] Task 5: Makefile 迁移命令 (AC: #1, #2, #3)
  - [x] 5.1 `make migrate-up` — 执行所有待运行的迁移
  - [x] 5.2 `make migrate-down` — 回滚最近一次迁移
  - [x] 5.3 `make migrate-new name=xxx` — 生成新的 up/down SQL 文件对
  - [x] 5.4 迁移命令使用环境变量 `DB_URL` 或从 config 构建连接字符串

- [x] Task 6: 基础测试 (AC: #1, #2)
  - [x] 6.1 迁移文件 SQL 语法验证测试
  - [x] 6.2 迁移 up/down 幂等性验证测试（需要真实数据库或测试容器）
  - [x] 6.3 迁移执行函数单元测试

## Dev Notes

### 架构约束（必须遵守）

- **迁移工具**：使用 golang-migrate v4，SQL 文件驱动（非 GORM AutoMigrate）
  - [Source: architecture.md#Core Architectural Decisions]
- **迁移文件目录**：`migrations/`，文件命名格式 `{序号}_{描述}.up.sql` / `{序号}_{描述}.down.sql`
  - [Source: architecture.md#Backend Code Organization Convention]
- **启动时自动执行**：服务启动时自动运行 migrate up
  - [Source: architecture.md#Implementation Constraints: "golang-migrate SQL 驱动，启动时自动执行"]
- **数据库命名约定**：
  - 表名：snake_case 复数 — `users`、`projects`、`pipelines`
  - 列名：snake_case — `project_id`、`created_at`
  - 外键：`{关联表单数}_id` — `project_id`、`user_id`
  - 索引：`idx_{表名}_{列名}` — `idx_pipelines_project_id`
  - 唯一约束：`uk_{表名}_{列名}` — `uk_users_username`
  - [Source: architecture.md#Naming Patterns]
- **GORM 模型映射**：Go struct `PipelineRun` → 表 `pipeline_runs`（GORM 默认行为）
  - [Source: architecture.md#Naming Patterns]
- **时间字段**：PostgreSQL `timestamptz`（UTC 存储）
  - [Source: architecture.md#Format Patterns]
- **Null 处理**：数据库 NULL 使用 Go 指针类型（`*string`、`*int64`）
  - [Source: architecture.md#Format Patterns]

### 关键技术版本

| 依赖 | 版本 | 说明 |
|------|------|------|
| golang-migrate | v4.19.1 | SQL 文件驱动 |
| PostgreSQL | 16 | docker-compose |
| GORM | v1.30.1 | 已安装（Story 1.1） |

[Source: architecture.md#Verified Dependency Versions]

### Makefile 命令规范（来自架构文档）

```
migrate-up:  # migrate -path migrations -database $DB_URL up
migrate-down: # migrate -path migrations -database $DB_URL down 1
migrate-new: # migrate create -ext sql -dir migrations -seq $(name)
```

[Source: architecture.md#Makefile Targets]

### 初始迁移文件说明

- 本 Story 只创建迁移框架和一个初始 schema 迁移（如启用 uuid-ossp 扩展）
- 不创建业务表（users、projects 等），业务表在各自 Story 中创建
- 迁移文件示例路径：
  - `migrations/000001_init_schema.up.sql`
  - `migrations/000001_init_schema.down.sql`

[Source: architecture.md#Backend Code Organization Convention]

### 不做的事（明确边界）

- **不创建业务表**：users 表在 Story 2.1，projects 表在 Story 3.1，pipelines 表在 Story 6.1
- **不实现 GORM AutoMigrate**：架构明确使用 golang-migrate SQL 文件方式
- **不修改已有的 GORM 连接逻辑**：`pkg/database/postgres.go` 保持不变
- **不修改健康检查端点**：保持 Story 1.1 实现不变

### Previous Story (1.1) 经验

- Go module 已是 1.25.0（Gin v1.12.0 要求）
- GORM + postgres driver 已安装并配置
- Config 系统支持 YAML + env 覆盖，敏感字段 env-only
- DB 连接池已配置（25/10/5min）
- Makefile 已有 dev、build、test、swag、lint 目标
- Windows bash 使用正斜杠路径避免问题
- docker-compose 已配置 PostgreSQL 16（端口 5432，用户=zcid，密码=zcid_dev，数据库=zcid）

### DB_URL 构建

从现有配置系统构建 golang-migrate 所需的连接字符串格式：
```
postgres://zcid:zcid_dev@localhost:5432/zcid?sslmode=disable
```

环境变量 `DB_URL` 可以覆盖，Makefile 命令优先使用 `DB_URL`。

### Project Structure Notes

- `pkg/database/migrate.go` 是新增文件，与已有的 `postgres.go` 和 `redis.go` 同目录
- `migrations/` 目录已在 Story 1.1 创建（含 .gitkeep）
- 删除 `migrations/.gitkeep`（有实际 SQL 文件后不再需要）
- Makefile 添加 migrate-up、migrate-down、migrate-new 三个新目标

### References

- [Source: architecture.md#Core Architectural Decisions] — golang-migrate 选型
- [Source: architecture.md#Verified Dependency Versions] — golang-migrate v4.19.1
- [Source: architecture.md#Backend Code Organization Convention] — migrations/ 目录结构
- [Source: architecture.md#Naming Patterns] — 数据库命名约定
- [Source: architecture.md#Format Patterns] — 时间格式、Null 处理
- [Source: architecture.md#Makefile Targets] — 迁移命令规范
- [Source: architecture.md#Implementation Constraints] — 启动时自动执行迁移
- [Source: epics.md#Epic 1 Story 1.2] — Story 定义和验收条件

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `go test ./... -v`
- `go build ./...`

### Completion Notes List

- Integrated golang-migrate up/down execution into `pkg/database/migrate.go` with migration version logging and no-change handling.
- Added sequential migration file generator support via `CreateMigrationFiles(migrationsPath, name)`.
- Added migration CLI at `cmd/migrate/main.go` with `up`, `down`, and `new --name` commands.
- Added `DatabaseConfig.MigrationURL()` for migrate-compatible PostgreSQL URL generation.
- Integrated startup auto-migration in server boot path guarded by `AUTO_MIGRATE=true`.
- Added Makefile targets `migrate-up`, `migrate-down`, and `migrate-new name=...`.
- Removed obsolete `migrations/.gitkeep` after real SQL migration files were committed.
- Verified full test/build pass (`go test ./... -v`, `go build ./...`), with integration migration test skipped when `MIGRATION_TEST_DB_URL` is not set.

### Change Log

- 2026-03-02: Implemented Story 1.2 database migration framework end-to-end; moved status to `review`.
- 2026-03-02: Code review (AI) — no issues found, moved to `done`.

### File List

- `cmd/server/main.go`
- `cmd/migrate/main.go`
- `config/config.go`
- `pkg/database/migrate.go`
- `pkg/database/migrate_test.go`
- `migrations/000001_init_schema.up.sql`
- `migrations/000001_init_schema.down.sql`
- `Makefile`
- `files/implementation-artifacts/1-2-database-migration-framework.md`
- `files/implementation-artifacts/sprint-status.yaml`

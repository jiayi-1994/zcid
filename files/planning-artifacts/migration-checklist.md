# Migration Checklist for Epic 3-11

## Epic 3: 项目与资源管理

### ✅ Story 3.1: 项目 CRUD
**需要 Migration**: 是
**表结构**:
- `projects` 表：id, name, description, creator_id, status, created_at, updated_at
- 索引：name (unique), creator_id, status

### ✅ Story 3.2: 环境管理与 Namespace 映射
**需要 Migration**: 是
**表结构**:
- `environments` 表：id, project_id, name, k8s_namespace, created_at, updated_at
- 索引：project_id, k8s_namespace (unique)

### ✅ Story 3.3: 服务管理
**需要 Migration**: 是
**表结构**:
- `services` 表：id, project_id, name, description, created_at, updated_at
- 索引：project_id, name

### ✅ Story 3.4: 项目成员与角色管理
**需要 Migration**: 是
**表结构**:
- `project_members` 表：id, project_id, user_id, role, created_at, updated_at
- 索引：(project_id, user_id) unique, user_id
- **注意**: Casbin 策略存储在 `casbin_rule` 表（已在 Epic 2 创建）

### ❌ Story 3.5: 前端项目管理页面
**需要 Migration**: 否（纯前端）

---

## Epic 4: 变量与凭证管理

### ✅ Story 4.1: 多层级变量 CRUD
**需要 Migration**: 是
**表结构**:
- `variables` 表：id, scope (global/project/pipeline), scope_id, key, value, type (plain/secret), encrypted, created_at, updated_at
- 索引：(scope, scope_id, key) unique

### ✅ Story 4.2: 密钥变量加密与安全
**需要 Migration**: 否（复用 Story 4.1 的 variables 表，增加 encrypted 字段逻辑）

### ❌ Story 4.3: 运行时密钥注入与清理
**需要 Migration**: 否（运行时逻辑，不涉及持久化表）

### ❌ Story 4.4: 前端变量管理页面
**需要 Migration**: 否（纯前端）

---

## Epic 5: Git 仓库集成

### ✅ Story 5.1: Git 仓库连接配置
**需要 Migration**: 是
**表结构**:
- `git_repos` 表：id, project_id, name, url, auth_type, credentials (encrypted), created_at, updated_at
- 索引：project_id

### ✅ Story 5.2: 仓库/分支选择
**需要 Migration**: 否（复用 pipelines 表，增加 repo_id, branch 字段）

### ✅ Story 5.3: Webhook 接收与签名验证
**需要 Migration**: 是
**表结构**:
- `webhook_events` 表：id, repo_id, event_type, payload, signature, processed, created_at
- 索引：repo_id, processed, created_at

### ❌ Story 5.4: 前端 Git 页面
**需要 Migration**: 否（纯前端）

---

## Epic 6: 流水线可视化编排

### ✅ Story 6.1: 流水线 CRUD (JSONB 存储)
**需要 Migration**: 是
**表结构**:
- `pipelines` 表：id, project_id, name, definition (JSONB), repo_id, branch, created_at, updated_at
- 索引：project_id, name

### ❌ Story 6.2-6.6: 其他流水线功能
**需要 Migration**: 否（复用 pipelines 表，增加字段或 JSONB 内嵌配置）

---

## Epic 7: 流水线执行与构建

### ✅ Story 7.1: CRD 翻译引擎
**需要 Migration**: 否（运行时逻辑）

### ✅ Story 7.2: 流水线运行编排
**需要 Migration**: 是
**表结构**:
- `pipeline_runs` 表：id, pipeline_id, trigger_type, status, started_at, finished_at, created_at
- 索引：pipeline_id, status, created_at

### ✅ Story 7.3: 取消运行与构建产物
**需要 Migration**: 是
**表结构**:
- `build_artifacts` 表：id, run_id, type, path, size, created_at
- 索引：run_id

### ❌ Story 7.4-7.5: 构建链路
**需要 Migration**: 否（配置存储在 pipelines JSONB 或系统配置表）

---

## Epic 8: 实时日志与状态监控

### ✅ Story 8.1: WebSocket 连接
**需要 Migration**: 否（运行时连接管理）

### ✅ Story 8.2-8.3: 实时日志与状态
**需要 Migration**: 否（流式数据，存储在 MinIO）

### ✅ Story 8.4: 日志归档
**需要 Migration**: 是
**表结构**:
- `log_archives` 表：id, run_id, storage_path, size, created_at
- 索引：run_id

### ✅ Story 8.5: 运行历史列表
**需要 Migration**: 否（复用 pipeline_runs 表）

---

## Epic 9: 部署与环境管理

### ✅ Story 9.1: ArgoCD 部署触发
**需要 Migration**: 是
**表结构**:
- `deployments` 表：id, run_id, environment_id, status, argocd_app_name, created_at, updated_at
- 索引：run_id, environment_id, status

### ✅ Story 9.2-9.3: 部署状态与历史
**需要 Migration**: 否（复用 deployments 表）

---

## Epic 10: 通知、审计与平台运维

### ✅ Story 10.1: 通知规则配置
**需要 Migration**: 是
**表结构**:
- `notification_rules` 表：id, project_id, event_type, webhook_url, enabled, created_at
- 索引：project_id, enabled

### ✅ Story 10.2: 审计日志
**需要 Migration**: 是
**表结构**:
- `audit_logs` 表：id, user_id, action, resource_type, resource_id, ip, created_at
- 索引：user_id, resource_type, created_at

### ✅ Story 10.3: 系统设置与健康检测
**需要 Migration**: 是
**表结构**:
- `system_settings` 表：key, value, updated_at
- 主键：key

### ❌ Story 10.4: CRD 清理与告警
**需要 Migration**: 否（定时任务逻辑）

---

## Epic 11: 全局概览与用户引导

### ❌ Story 11.1-11.3: 概览与引导
**需要 Migration**: 否（聚合查询或前端状态）

---

## 总结

**需要 Migration 的 Stories**: 18 个
**不需要 Migration 的 Stories**: 剩余所有

## 开发建议

1. **每个需要 Migration 的 Story**，在 Task 1 中必须包含：
   - 创建 migration 文件（`make migrate-new name=xxx`）
   - 在 Dev Notes 中记录完整表结构（字段、类型、索引、外键）
   - 考虑 seed data（如果需要初始数据）

2. **Migration 命名规范**：
   - `000005_create_projects.up.sql`
   - `000006_create_environments.up.sql`
   - 依此类推

3. **字段设计原则**：
   - 所有表必须有 `created_at`, `updated_at`
   - 软删除场景增加 `deleted_at` 或 `status` 字段
   - 外键字段统一命名：`{table}_id`（如 `project_id`, `user_id`）

4. **索引设计原则**：
   - 查询频繁的字段加索引
   - 唯一约束用 UNIQUE 索引
   - 复合索引按查询条件顺序排列

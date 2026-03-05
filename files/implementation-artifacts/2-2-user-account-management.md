# Story 2.2: 用户账号管理

Status: done

## Story

As a 管理员,
I want 创建、编辑、禁用用户账号,
so that 我可以管理平台的用户。

## Acceptance Criteria

1. **管理员创建用户**
   - Given 管理员已登录
   - When POST `/api/v1/admin/users` 提交用户信息
   - Then 创建新用户，密码使用 bcrypt 哈希存储

2. **管理员编辑用户**
   - Given 管理员编辑用户
   - When PUT `/api/v1/admin/users/:uid`
   - Then 用户信息更新成功

3. **管理员禁用用户后会话失效**
   - Given 管理员禁用某用户
   - When 设置用户状态为 disabled
   - Then 该用户所有 Refresh Token 从 Redis 删除
   - And 该用户后续登录返回“账号已禁用”错误

4. **非管理员禁止访问用户管理接口**
   - Given 非管理员用户
   - When 尝试访问 `/api/v1/admin/users`
   - Then 返回 403 权限不足

## Tasks / Subtasks

- [x] Task 1: 扩展用户领域模型与存储结构（AC: 1,2,3）
  - [x] 在 `internal/auth/model.go` 增加用户状态字段（如 `status`），定义可用/禁用枚举
  - [x] 在 `internal/auth/repo.go` 增加用户创建、按 ID 查询、更新能力
  - [x] 在迁移中确保 users 表具备状态字段与必要索引（若已存在则按迁移规范演进）

- [x] Task 2: 实现管理员用户管理服务（AC: 1,2,3）
  - [x] 在 `internal/auth/service.go` 增加 CreateUser/UpdateUser/DisableUser 服务方法
  - [x] 创建用户时统一复用 `HashPassword`，禁止明文存储
  - [x] 禁用用户时删除该用户 Refresh Token 会话（与 Story 2.1 Redis 约定一致）
  - [x] 登录流程增加 disabled 状态校验并返回“账号已禁用”业务错误

- [x] Task 3: 实现管理员用户管理接口（AC: 1,2,4）
  - [x] 在 `internal/auth/handler.go` 增加 `POST /api/v1/admin/users`、`PUT /api/v1/admin/users/:uid`
  - [x] 在 `cmd/server/main.go` 将 admin 用户路由挂到 `/api/v1/admin/users`
  - [x] 接入鉴权中间件，确保仅管理员可访问，非管理员返回 403
  - [x] 对接统一响应格式（成功 code=0，失败 400xx/500xx）

- [ ] Task 4: 权限与错误码对齐（AC: 3,4）
  - [ ] 复用 Casbin RBAC 规则，补充 admin 用户管理对象与动作策略
  - [x] 确认“账号已禁用”“权限不足”落在 400xx 认证鉴权码段
  - [x] 统一错误信息与 HTTP 状态码映射（403 对应权限不足）

- [x] Task 5: 测试与回归验证（AC: 1,2,3,4）
  - [x] 为 service 新增单元测试：创建用户哈希校验、编辑成功、禁用后清会话、禁用用户登录失败
  - [x] 为 handler 新增接口测试：管理员成功、非管理员 403、参数校验失败
  - [x] 覆盖错误路径：重复用户名、用户不存在、Redis 异常
  - [x] 跑通 `go test ./...`，确认 Story 2.1 认证能力无回归

## Dev Notes

### 技术上下文与强约束

- Story 2.2 是 Epic 2 的延续，必须在 Story 2.1 已实现 JWT 双 Token 基础上扩展，不可重写认证主流程。
- 用户管理能力放在 `internal/auth` 模块内，遵循 handler → service → repo 分层。
- 密码存储强制 bcrypt；任何路径禁止明文密码落库与日志输出。
- 用户禁用后必须立即失效会话（删除 Refresh Token），确保后续 refresh/login 受限。
- 权限控制依赖 Casbin：仅管理员可访问 `/api/v1/admin/users*`。

### 相关代码位置（按架构约定）

- `internal/auth/handler.go`
- `internal/auth/service.go`
- `internal/auth/repo.go`
- `internal/auth/model.go`
- `internal/auth/dto.go`
- `pkg/middleware/auth.go`
- `pkg/response/response.go`、`pkg/response/errors.go`
- `cmd/server/main.go`

### 安全与边界

- 登录错误信息需要区分“账号禁用”与“账号密码错误”时，仅在明确业务要求场景返回禁用提示；其他认证失败保持最小信息泄露。
- 禁用用户应清理 Refresh Token，会话撤销必须可验证。
- 禁止将密码哈希、Token、凭证写入日志。
- 非管理员访问用户管理接口必须严格 403，不可仅依赖前端隐藏入口。

### 测试要求

- Service 层：
  - 创建用户密码哈希正确
  - 编辑用户信息成功
  - 禁用用户触发会话删除
  - 禁用用户登录返回业务错误
- Handler 层：
  - 管理员访问成功
  - 非管理员访问返回 403
  - 参数不合法返回验证错误
- 回归：
  - Story 2.1 的 login/refresh/logout 全部通过

### Project Structure Notes

- 严格保持 `internal/auth` 领域边界，不将用户管理逻辑散落到其他模块。
- `cmd/server/main.go` 仅负责路由编排，业务逻辑不可进入 main。

### References

- Epic 2 与 Story 2.2 AC：[Source: files/planning-artifacts/epics.md#Story 2.2: 用户账号管理]
- 认证与权限 FR（FR1-FR5）：[Source: files/planning-artifacts/prd.md#Functional Requirements]
- 认证安全 NFR（NFR9 等）：[Source: files/planning-artifacts/prd.md#Non-Functional Requirements]
- JWT/Casbin/错误码架构：[Source: files/planning-artifacts/architecture.md#Authentication & Security]
- API 路由与统一响应：[Source: files/planning-artifacts/architecture.md#API & Communication Patterns]
- 后端分层规范：[Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- create-story workflow

### Completion Notes List

- 已基于 Epic 2 自动生成 Story 2.2 开发文档
- 已补全用户管理、权限约束、会话失效和测试门禁要求
- 已明确与 Story 2.1 的衔接边界，避免重复实现认证基础能力

### File List

- files/implementation-artifacts/2-2-user-account-management.md

## Change Log

- 2026-03-02: 创建 Story 2.2（ready-for-dev），补充实现约束、任务拆解与测试要求。

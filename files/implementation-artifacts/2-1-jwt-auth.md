# Story 2.1: 用户登录与 JWT 双 Token 认证

Status: done

## Story

As a 用户,
I want 通过账号密码登录平台并获取认证凭证,
so that 我可以安全地访问平台功能。

## Acceptance Criteria

1. **登录签发双 Token**
   - Given 用户提交正确的用户名和密码
   - When POST `/api/v1/auth/login`
   - Then 返回 Access Token（30min）和 Refresh Token（7天）
   - And Refresh Token 存储到 Redis

2. **刷新 Access Token**
   - Given Access Token 过期
   - When POST `/api/v1/auth/refresh` 携带有效 Refresh Token
   - Then 返回新的 Access Token
   - And Refresh Token 不变

3. **登出失效 Refresh Token**
   - Given 用户登出
   - When POST `/api/v1/auth/logout`
   - Then Redis 中该用户的 Refresh Token 被删除
   - And 后续使用该 Refresh Token 刷新失败

4. **密码安全存储**
   - Given 密码存储
   - When 用户注册或修改密码
   - Then 密码使用 bcrypt 哈希存储，不存明文

## Tasks / Subtasks

- [ ] Task 1: 实现登录接口与双 Token 签发（AC: 1）
  - [ ] 在 `internal/auth/handler.go` 新增 `POST /api/v1/auth/login` handler
  - [ ] 在 `internal/auth/service.go` 实现账号密码校验与 Token 签发逻辑（Access 30min，Refresh 7d）
  - [ ] 在 `internal/auth/repo.go` 增加按用户名查询用户能力
  - [ ] 将 Refresh Token 写入 Redis（含过期时间）
  - [ ] 对接统一响应格式（成功 code=0，失败返回 400xx）

- [ ] Task 2: 实现刷新接口（AC: 2）
  - [ ] 新增 `POST /api/v1/auth/refresh` handler
  - [ ] 校验 Refresh Token 有效性与 Redis 中会话存在性
  - [ ] 仅重签 Access Token，保持 Refresh Token 不变
  - [ ] 覆盖常见失败场景（过期、伪造、Redis 不存在）

- [ ] Task 3: 实现登出接口（AC: 3）
  - [ ] 新增 `POST /api/v1/auth/logout` handler
  - [ ] 删除 Redis 中对应 Refresh Token
  - [ ] 验证登出后刷新失败路径

- [ ] Task 4: 实现密码哈希策略（AC: 4）
  - [ ] 统一使用 bcrypt 进行密码哈希与比对（禁止明文存储）
  - [ ] 在用户创建/改密路径复用同一套哈希工具
  - [ ] 为哈希与比对逻辑补充单元测试

- [ ] Task 5: 测试与质量门禁（AC: 1,2,3,4）
  - [ ] 为 login/refresh/logout 编写单元测试（成功 + 失败）
  - [ ] 为 auth service 增加 Redis 交互相关测试
  - [ ] 跑通 `go test ./...` 并确保无回归
  - [ ] 验证错误码与统一响应结构符合约定

## Dev Notes

### 技术上下文与强约束

- 认证方案固定为 **JWT 双 Token**：Access Token 30 分钟，Refresh Token 7 天（Redis 持久化会话）。
- 密码必须使用 **bcrypt** 哈希存储，不允许任何明文落库。
- API 必须走统一响应格式：
  - 成功：`{"code":0,"message":"success","data":...,"requestId":"..."}`
  - 失败：`{"code":4xxxx,"message":"...","detail":"...","requestId":"..."}`
- 错误码需落在认证权限段（400xx）。
- 代码分层必须遵循 handler → service → repo，不在 handler 写业务逻辑。

### 相关代码位置（按架构约定）

- `cmd/server/main.go`（路由注册入口）
- `internal/auth/handler.go`
- `internal/auth/service.go`
- `internal/auth/repo.go`
- `internal/auth/model.go`
- `internal/auth/dto.go`
- `pkg/middleware/auth.go`（鉴权中间件集成点）
- `pkg/response/response.go`、`pkg/response/errors.go`（统一响应与业务错误）
- `migrations/000001_init_schema.up.sql`（users 表结构）
- `migrations/000004_seed_admin_user.up.sql`（初始 admin 用户）
- `config/config.yaml`（开发环境配置，包含数据库/Redis/JWT 密钥）

### Redis 与 Token 约定

- Refresh Token 必须可吊销（logout/禁用用户场景）。
- 建议将用户标识与 token 唯一标识绑定存储，设置 7 天 TTL。
- refresh 流程必须二次校验：
  1) JWT 本身合法且未过期；
  2) Redis 会话存在且匹配。

### 安全与边界

- 登录失败信息避免泄露用户存在性（用户名不存在/密码错误尽量同类提示）。
- 禁止在日志输出密码、Token、敏感凭证。
- 若项目已接入登录限流策略，确保 `/api/v1/auth/login` 保持兼容。

### 测试要求

- 单元测试：
  - 正常登录签发双 Token
  - 用户不存在/密码错误
  - refresh 成功与失败（过期、伪造、会话不存在）
  - logout 后 refresh 失败
  - bcrypt 哈希与比对正确
- 集成校验：
  - 接口响应结构、HTTP 状态码与业务错误码映射正确
  - Redis token TTL 生效

### Project Structure Notes

- 严格遵循既有单体模块化结构，不新增跨模块耦合。
- Auth 能力仅落在 `internal/auth` 与通用 `pkg/*`（中间件/响应）中，避免污染其他域模块。

### Database Migration Notes

- `migrations/000001_init_schema.up.sql` 创建 `users` 表：
  - 字段：id, username (unique), password_hash, role, status, created_at, updated_at
  - 索引：username, status
- `migrations/000004_seed_admin_user.up.sql` 插入初始 admin 用户：
  - 用户名：admin
  - 密码：admin123（bcrypt hashed）
  - 角色：admin
- 开发环境配置在 `config/config.yaml`，包含数据库密码、JWT 密钥等
- 生产环境建议通过环境变量覆盖敏感配置

### References

- Epic 2 / Story 2.1 需求与 AC：[Source: files/planning-artifacts/epics.md#Epic 2: 用户认证与权限管理]
- 认证与安全 NFR（JWT、bcrypt、传输安全）：[Source: files/planning-artifacts/prd.md#Non-Functional Requirements]
- JWT 双 Token、Casbin 与错误码架构：[Source: files/planning-artifacts/architecture.md#Authentication & Security]
- API 路由与统一响应规范：[Source: files/planning-artifacts/architecture.md#API & Communication Patterns]
- 后端分层与目录规范：[Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention]

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6

### Debug Log References

- create-story workflow

### Completion Notes List

- 已基于 Epic 2 自动生成 Story 2.1 开发文档
- 已补全可直接执行的任务拆解、实现边界、测试要求与代码落点

### File List

- files/implementation-artifacts/2-1-jwt-auth.md

## Change Log

- 2026-03-02: 创建 Story 2.1（ready-for-dev），并补充完整开发上下文与实现约束。

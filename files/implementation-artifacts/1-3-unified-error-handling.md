# Story 1.3: 统一错误处理与响应格式

Status: done

## Story

As a 开发者,
I want 统一的 API 响应格式和错误处理机制,
so that 所有 API 返回一致的结构，前端可以统一处理。

## Acceptance Criteria (BDD)

1. **Given** 任意 API 请求成功 **When** 返回响应 **Then** 格式为 `{"code": 0, "message": "success", "data": {...}, "requestId": "req-xxx"}`
2. **Given** 业务逻辑返回错误 **When** handler 调用 `response.HandleError(c, err)` **Then** 返回对应错误码和 HTTP 状态码 **And** 格式为 `{"code": 40201, "message": "...", "detail": "...", "requestId": "req-xxx"}`
3. **Given** handler 发生未捕获的 panic **When** 全局错误中间件拦截 **Then** 返回 500 + 错误码 50001 **And** 记录 ERROR 级别日志含 stack trace
4. **Given** 请求进入 Gin 路由 **When** RequestID 中间件执行 **Then** 生成唯一 requestId 并注入 context，响应 header 包含 `X-Request-ID`

## Tasks / Subtasks

- [x] Task 1: 建立统一响应与业务错误模型 (AC: #1, #2)
  - [x] 1.1 新建 `pkg/response/response.go`：实现 `Success`、`Error` 响应方法，输出统一 JSON 格式
  - [x] 1.2 新建 `pkg/response/errors.go`：定义 `BizError`（code/message/detail/httpStatus）
  - [x] 1.3 新建 `pkg/response/codes.go`：维护错误码到 HTTP 状态码映射（覆盖 400xx/500xx）
  - [x] 1.4 实现 `HandleError(c, err)`：兼容 `BizError` 与未知错误

- [x] Task 2: 实现 RequestID 中间件 (AC: #4)
  - [x] 2.1 新建 `pkg/middleware/requestid.go`：生成唯一 requestId（如 `req-<随机串>`）
  - [x] 2.2 将 requestId 注入 Gin context，并设置响应头 `X-Request-ID`
  - [x] 2.3 为后续日志和错误响应统一提供 requestId 获取函数

- [x] Task 3: 实现全局错误与 panic 恢复中间件 (AC: #2, #3)
  - [x] 3.1 新建 `pkg/middleware/error.go`：封装 panic recover
  - [x] 3.2 panic 时返回 `{code:50001,...}`，HTTP 500，带 requestId
  - [x] 3.3 记录包含 stack trace 的 ERROR 级别日志

- [x] Task 4: 集成到服务启动入口 (AC: #1, #2, #3, #4)
  - [x] 4.1 在 `cmd/server/main.go` 注册 RequestID 与 Error 中间件（靠前顺序）
  - [x] 4.2 新增/调整示例 handler，使用 `response.Success` 和 `response.HandleError`
  - [x] 4.3 验证现有 `/healthz`、`/readyz` 不受破坏

- [x] Task 5: 测试与回归验证 (AC: #1, #2, #3, #4)
  - [x] 5.1 单元测试：`pkg/response` 的成功/错误输出结构
  - [x] 5.2 中间件测试：requestId 生成与 header 注入
  - [x] 5.3 中间件测试：panic recover 返回 50001 + requestId
  - [x] 5.4 集成验证：`go test ./... -v` 与 `go build ./...`

## Dev Notes

### 架构约束（必须遵守）

- API 统一响应格式必须为：`code/message/data/requestId`，错误响应附带 `detail`（如有）
  - [Source: files/planning-artifacts/architecture.md#API & Communication Patterns]
- 错误传播链固定为：repo → service → handler，不在 handler 写业务判断
  - [Source: files/planning-artifacts/architecture.md#Communication Patterns]
- 统一错误码段规划已经约定，需复用并避免临时散落魔法数字
  - [Source: files/planning-artifacts/architecture.md#API & Communication Patterns]
- 中间件目录与命名约定：`pkg/middleware/*.go`，包名小写
  - [Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention]
- 日志规范：结构化、携带 requestId，错误级别区分 ERROR/WARN/INFO/DEBUG
  - [Source: files/planning-artifacts/architecture.md#Communication Patterns]

### Story 1.3 实现边界

- 本 Story 聚焦“错误处理 + 响应格式 + requestId 中间件”
- 不实现完整业务模块（auth/project/pipeline）
- 不改动数据库迁移逻辑与 Redis/MinIO 初始化
- 不引入新鉴权策略（JWT/Casbin 在 Epic 2 实现）

### 来自上一故事（1.2）的可复用经验

- `cmd/server/main.go` 已完成配置加载、PostgreSQL/Redis/MinIO 初始化与健康检查路由注册，可直接扩展中间件注册点
  - [Source: files/implementation-artifacts/1-2-database-migration-framework.md#Completion Notes List]
- 当前工程已通过 `go test ./... -v` 与 `go build ./...` 验证，Story 1.3 需保持该基线
  - [Source: files/implementation-artifacts/1-2-database-migration-framework.md#Debug Log References]
- 现有代码使用 Gin 默认中间件，新增中间件时需确认顺序避免覆盖默认行为
  - [Source: cmd/server/main.go:48]

### 技术实现建议（供 dev-story 执行）

- `requestId` 生成建议使用标准库随机源（避免全局可预测值）
- `HandleError` 应优先识别 `BizError`，未知错误统一落到 50001
- panic recover 需确保：
  - 响应结构统一
  - 不泄露敏感内部细节
  - 日志保留栈信息用于定位

### Project Structure Notes

- 建议新增文件：
  - `pkg/response/response.go`
  - `pkg/response/errors.go`
  - `pkg/response/codes.go`
  - `pkg/middleware/requestid.go`
  - `pkg/middleware/error.go`
  - 对应测试文件 `*_test.go`
- 重点修改文件：
  - `cmd/server/main.go`（中间件接入与示例路由响应统一）

### References

- [Source: files/planning-artifacts/epics.md#Story 1.3: 统一错误处理与响应格式]
- [Source: files/planning-artifacts/architecture.md#API & Communication Patterns]
- [Source: files/planning-artifacts/architecture.md#Communication Patterns]
- [Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention]
- [Source: files/planning-artifacts/architecture.md#Implementation Patterns & Consistency Rules]
- [Source: cmd/server/main.go:48]

## Project Context Reference

- `project-context.md` 未找到（`**/project-context.md` 无匹配），本 Story 依据 PRD / Architecture / Epics 制作。

## Story Completion Status

- Story implementation completed and verified via code review
- Unified response, requestId middleware, and panic recovery are integrated in server startup flow
- Validation commands executed: `go test ./... -v` and `go build ./...`

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `go test ./... -v` (PASS across packages; integration migration test skipped as expected when `MIGRATION_TEST_DB_URL` is unset)
- `go build ./...` (PASS)

### Completion Notes List

- Completed unified response utilities in `pkg/response` with business error mapping and fallback handling.
- Added request ID middleware and panic recovery middleware, then integrated both in server bootstrap.
- Added tests for response envelope, requestId propagation, and panic recovery response behavior.
- Verified health endpoints remain intact while adding example routes for success/error/panic flows.

### File List

- `cmd/server/main.go`
- `pkg/response/response.go`
- `pkg/response/errors.go`
- `pkg/response/codes.go`
- `pkg/response/response_test.go`
- `pkg/middleware/requestid.go`
- `pkg/middleware/error.go`
- `pkg/middleware/requestid_test.go`
- `pkg/middleware/error_test.go`
- `files/implementation-artifacts/1-3-unified-error-handling.md`
- `files/implementation-artifacts/sprint-status.yaml`

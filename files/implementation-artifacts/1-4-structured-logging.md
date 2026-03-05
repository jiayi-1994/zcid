# Story 1.4: 结构化日志与脱敏引擎

Status: done

## Story

As a 运维人员,
I want 结构化 JSON 日志输出和自动脱敏,
so that 日志可被日志系统采集且不泄露敏感信息。

## Acceptance Criteria (BDD)

1. **Given** 服务运行中 **When** 任意日志输出 **Then** 格式为 JSON，包含 level、msg、time、requestId 字段
2. **Given** 日志内容包含密钥变量值 **When** 脱敏 Handler 处理 **Then** 密钥值被替换为 `***`
3. **Given** 管理员调用 admin API 调整日志级别 **When** 设置为 DEBUG **Then** 运行时立即生效，无需重启 **And** `slog.LevelVar` 动态切换

## Tasks / Subtasks

- [x] Task 1: 实现结构化日志初始化与动态调级 (AC: #1, #3)
  - [x] 1.1 新建 `pkg/logging/logger.go`，实现 `Init/SetLevel/CurrentLevel`
  - [x] 1.2 使用 `slog.LevelVar` 实现运行时调级
  - [x] 1.3 在 `cmd/server/main.go` 接入 `logging.Init(cfg.Server.LogLevel)`

- [x] Task 2: 实现日志脱敏处理 (AC: #2)
  - [x] 2.1 新建 `pkg/logging/masking.go`，实现 `MaskingHandler`
  - [x] 2.2 对 message 与 attributes 进行敏感字段脱敏
  - [x] 2.3 递归处理 group attrs，避免嵌套字段泄露

- [x] Task 3: 集成请求访问日志 (AC: #1)
  - [x] 3.1 新建 `pkg/middleware/accesslog.go`
  - [x] 3.2 记录 method/path/status/latency/clientIp/requestId
  - [x] 3.3 在 `cmd/server/main.go` 注册 `middleware.AccessLogger()`

- [x] Task 4: 管理 API 动态调级 (AC: #3)
  - [x] 4.1 在 `cmd/server/main.go` 新增 `POST /admin/log-level`
  - [x] 4.2 参数校验失败和非法级别返回统一业务错误
  - [x] 4.3 更新后返回当前日志级别

- [x] Task 5: 配置与回归验证
  - [x] 5.1 `config/config.go` 修复 `SERVER_PORT` 环境变量覆盖回归
  - [x] 5.2 补充 `config/config_test.go` 的 `log_level` YAML / ENV 覆盖断言
  - [x] 5.3 补充 `pkg/logging/logger_test.go` 脱敏与动态调级测试
  - [x] 5.4 执行 `go test ./... -v` 与 `go build ./...`

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `go test ./... -v` (PASS)
- `go build ./...` (PASS)

### Completion Notes List

- 服务日志切换为 `slog` JSON 输出，支持配置启动级别与运行时动态切换。
- 新增脱敏 Handler，对 `password/secret/token/apikey/...` 等敏感键和值进行统一掩码。
- 新增访问日志中间件并接入服务启动链路，统一输出 requestId 与请求关键字段。
- 增加 `POST /admin/log-level` 管理接口，支持在线变更日志级别。
- 修复配置层 `SERVER_PORT` 环境覆盖回归，并补齐测试覆盖。

### Change Log

- 2026-03-02: 完成 Story 1.4，状态更新为 `review`。
- 2026-03-02: Code review (AI) — M1 fix: accesslog fallback to c.Request.URL.Path when FullPath() empty; M2 fix: added "passwd"/"pwd" to isSensitiveKey. Moved to `done`.

### File List

- `cmd/server/main.go`
- `config/config.go`
- `config/config_test.go`
- `pkg/logging/logger.go`
- `pkg/logging/masking.go`
- `pkg/logging/logger_test.go`
- `pkg/middleware/accesslog.go`
- `pkg/middleware/error.go`
- `files/implementation-artifacts/1-4-structured-logging.md`
- `files/implementation-artifacts/sprint-status.yaml`

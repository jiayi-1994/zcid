# Story 1.5: Redis 连接与缓存基础层

Status: done

## Story

As a 开发者,
I want Redis 连接池和基础缓存工具就绪,
so that 后续功能可以直接使用缓存能力。

## Acceptance Criteria (BDD)

1. **Given** 服务启动且 Redis 可达 **When** 健康检查执行 **Then** `/readyz` 包含 Redis 连接状态
2. **Given** Redis 连接断开 **When** 服务尝试缓存操作 **Then** 操作返回错误，不 panic **And** 业务逻辑可降级到直接查库

## Tasks / Subtasks

- [x] Task 1: 盘点并复用现有 Redis 连接初始化 (AC: #1)
  - [x] 1.1 复用 `pkg/database/redis.go` 的连接与 `PingRedis`
  - [x] 1.2 确认 `cmd/server/main.go` 的 `/readyz` 已纳入 Redis 检查

- [x] Task 2: 新增缓存基础层 (AC: #2)
  - [x] 2.1 新建 `pkg/cache/redis.go`，提供 `Get/Set/Delete`
  - [x] 2.2 统一 key prefix 处理与默认 TTL
  - [x] 2.3 未命中返回 `ErrCacheMiss`，连接异常返回可处理错误

- [x] Task 3: 覆盖基础错误路径测试 (AC: #2)
  - [x] 3.1 `nil client` 调用返回错误而非 panic
  - [x] 3.2 Redis 不可达时 `Get/Set/Delete` 返回错误

- [x] Task 4: 回归验证
  - [x] 4.1 执行 `go test ./... -v`
  - [x] 4.2 执行 `go build ./...`

## Dev Agent Record

### Agent Model Used

Claude Opus 4.6 (claude-opus-4-6)

### Debug Log References

- `go test ./... -v` (PASS)
- `go build ./...` (PASS)

### Completion Notes List

- 新增 `pkg/cache/redis.go` 缓存基础封装，支持 key 前缀与默认 TTL。
- 新增 `ErrCacheMiss` 语义错误，便于上层按“缓存未命中”分支进行降级处理。
- 在 Redis 不可用、client 为空等场景下，统一返回错误而不是 panic。
- 新增 `pkg/cache/redis_test.go` 覆盖不可用路径，确保基础层可安全落地到后续业务模块。

### Change Log

- 2026-03-02: 完成 Story 1.5，状态更新为 `review`。
- 2026-03-02: Code review (AI) — H4 fix: added ":" separator in cache buildKey to prevent namespace collision. Moved to `done`.

### File List

- `pkg/cache/redis.go`
- `pkg/cache/redis_test.go`
- `files/implementation-artifacts/1-5-redis-cache-layer.md`
- `files/implementation-artifacts/sprint-status.yaml`

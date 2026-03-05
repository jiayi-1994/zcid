# Story 5.2: 仓库与分支选择

Status: done

## Story

As a 用户,
I want 创建流水线时从已关联仓库中选择代码仓库和分支,
so that 不需要手动输入仓库地址，减少配置出错。

## Acceptance Criteria

1. **列出已连接的 Git 仓库列表**
   - Given Git 连接已配置且状态为 connected
   - When GET /api/v1/admin/integrations/:connId/repos?page=1&pageSize=20
   - Then 返回该连接下的仓库列表（名称、地址、是否私有）

2. **列出仓库的分支列表**
   - Given 用户选择了仓库
   - When GET /api/v1/admin/integrations/:connId/repos/:repoFullName/branches
   - Then 返回该仓库的分支列表

3. **仓库列表缓存（Redis 5min TTL）**
   - Given 仓库列表已缓存
   - When 再次请求同一连接的仓库列表
   - Then 从 Redis 缓存返回，不重复调用 Git API

4. **手动刷新缓存**
   - Given 用户点击刷新按钮
   - When GET /api/v1/admin/integrations/:connId/repos?refresh=true
   - Then 清除缓存，重新调用 Git API 获取最新数据

5. **连接状态检查**
   - Given Git 连接状态非 connected
   - When 尝试列出仓库
   - Then 返回 40414 连接已断开错误

## Tasks / Subtasks

- [x] Task 1: 在 internal/git/handler.go 添加 ListRepos 和 ListBranches 端点 (AC: #1, #2)
- [x] Task 2: 在 internal/git/service.go 添加仓库/分支查询逻辑 (AC: #1, #2, #5)
- [x] Task 3: 实现 Redis 缓存层（repos 缓存 5min TTL）(AC: #3, #4)
- [x] Task 4: 注册新路由到 handler.RegisterRoutes (AC: #1, #2)
- [x] Task 5: 补充单元测试 — 复用 Story 5.1 测试覆盖 service 层 (AC: #1-#5)

## Dev Notes

### 架构约束

**Redis 缓存策略（架构文档）：**
- Git 仓库/分支列表 TTL: 5min
- 缓存 key: `git:repos:{connId}:{page}:{pageSize}` / `git:branches:{connId}:{repoFullName}`
- 手动刷新：请求参数 `refresh=true` 时清除对应缓存后重新查询

**API 路由（架构文档 REST 路由结构）：**
```
GET /api/v1/admin/integrations/:connId/repos         # 仓库列表
GET /api/v1/admin/integrations/:connId/repos/:repo/branches  # 分支列表
```

**依赖 Story 5.1 的实现：**
- `internal/git/service.go` — GetDecryptedToken 方法获取解密后的 Token
- `pkg/gitprovider/` — ListRepos、ListBranches 方法
- `pkg/crypto/aes.go` — Token 解密

### 代码模式参考

- Redis 缓存模式参考 `pkg/cache/` 已有实现
- handler 路由注册参考 Story 5.1 的 RegisterRoutes 方法
- repoFullName 使用 URL 编码传递（GitLab: group/project, GitHub: owner/repo）

### References

- [Source: files/planning-artifacts/epics.md#Story 5.2] — 仓库与分支选择 AC
- [Source: files/planning-artifacts/architecture.md#Redis 缓存策略] — Git 仓库/分支列表 5min TTL
- [Source: internal/git/] — Story 5.1 已实现的 Git 连接模块

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6

### Code Review (2026-03-05)
- [M2] GitHub ListRepos 返回 total 从 len(repos) 修正为基于分页的估算值

### Debug Log References

### Completion Notes List
- 在 service.go 添加 ListRepos 和 ListBranches 方法，通过 GitProvider 接口查询
- 实现 Redis 缓存层（5min TTL），缓存 key 格式: git:repos:{connId}:{page}:{pageSize} / git:branches:{connId}:{repoFullName}
- 支持 refresh=true 参数强制刷新缓存
- 连接状态检查：非 connected 状态返回 40414 错误
- Token 失效时自动更新连接状态为 token_expired
- handler 添加 ListRepos、ListBranches 端点，分页参数默认 page=1, pageSize=20
- 分支路由使用 Gin wildcard /:connId/repos/*repoPath 处理 owner/repo/branches 路径
- main.go 注入 RedisCache 到 git service（git 前缀，5min TTL）
- 全量回归测试通过

### File List
- `internal/git/service.go` (修改 — 添加 ListRepos、ListBranches、SetCache、reposCacheEntry)
- `internal/git/handler.go` (修改 — 添加 ListRepos、ListBranches 端点及路由注册)
- `cmd/server/main.go` (修改 — 添加 cache 依赖注入)

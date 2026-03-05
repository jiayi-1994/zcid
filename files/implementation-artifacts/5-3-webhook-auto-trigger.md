# Story 5.3: Webhook 接收与自动触发

Status: done

## Story

As a 系统,
I want 接收 Git Webhook 推送事件并自动匹配触发流水线,
so that 代码提交后自动启动构建。

## Acceptance Criteria

1. **Webhook 端点接收事件**
   - Given GitLab/GitHub 推送 Webhook 事件
   - When POST /api/v1/webhooks/gitlab 或 /api/v1/webhooks/github
   - Then 解析事件 payload，提取仓库、分支、事件类型、commit SHA

2. **签名验证（NFR12）**
   - Given Webhook 携带签名头
   - When 验证签名
   - Then GitLab 使用 X-Gitlab-Token，GitHub 使用 X-Hub-Signature-256 HMAC-SHA256
   - And 签名无效返回 401 并记录审计日志

3. **幂等性去重（FR19）**
   - Given 相同 Webhook 事件重复到达
   - When 幂等键（event_type:repo:commit_sha:timestamp_minute）已存在于 Redis
   - Then 跳过触发，返回 200

4. **事件匹配**
   - Given Webhook 签名验证通过
   - When 系统匹配仓库 URL、分支、事件类型
   - Then 找到对应的流水线配置（预留接口，Epic 6 实现完整匹配）

5. **Webhook Secret 管理**
   - Given 管理员配置 Git 连接
   - When 创建连接时
   - Then 系统自动生成 webhook_secret 并加密存储
   - And 可通过 API 获取 webhook_secret 用于配置 Git 仓库

## Tasks / Subtasks

- [x] Task 1: 创建数据库迁移 — 添加 webhook_secret 字段到 git_connections 表 (AC: #5)
  - [x] 1.1 创建 000011_add_webhook_secret.up.sql
  - [x] 1.2 创建 000011_add_webhook_secret.down.sql
- [x] Task 2: 创建 Webhook 签名验证逻辑 (AC: #2)
  - [x] 2.1 创建 pkg/gitprovider/webhook.go — GitLab/GitHub 签名验证
- [x] Task 3: 创建 Webhook Handler (AC: #1, #2, #3)
  - [x] 3.1 创建 internal/git/webhook_handler.go — Webhook 接收端点
  - [x] 3.2 实现幂等性去重（Redis SETNX，TTL 5min）
- [x] Task 4: 在 git_connections 表中增加 webhook_secret (AC: #5)
  - [x] 4.1 更新 model.go 添加 WebhookSecret 字段
  - [x] 4.2 创建连接时自动生成 webhook_secret（32 字节随机 hex）
  - [x] 4.3 添加获取 webhook_secret 的 API 端点
- [x] Task 5: 注册 Webhook 路由（不需要 JWT 认证）(AC: #1)
- [x] Task 6: 单元测试 (AC: #1-#5)

## Dev Notes

### 架构约束

**Webhook 路由（架构文档）：**
```
POST /api/v1/webhooks/gitlab   # 不需要 JWT，使用签名验证
POST /api/v1/webhooks/github   # 不需要 JWT，使用签名验证
```

**签名验证规则（NFR12）：**
- GitLab: 请求头 `X-Gitlab-Token` 与存储的 webhook_secret 比较
- GitHub: 请求头 `X-Hub-Signature-256`，HMAC-SHA256(webhook_secret, body)

**幂等键设计（架构文档）：**
- Key: `webhook:{event_type}:{repo}:{commit_sha}:{timestamp_minute}`
- 存储: Redis SETNX，TTL 5 分钟
- 重复到达返回 200（静默跳过）

**MVP 说明：**
- Story 5.3 实现 Webhook 接收、签名验证、幂等去重的基础设施
- 流水线匹配和触发需要 Epic 6 的流水线 CRUD 完成后才能实现完整逻辑
- 本 Story 预留 triggerPipeline 接口，记录事件日志，暂不执行实际触发

### References

- [Source: files/planning-artifacts/epics.md#Story 5.3] — Webhook 接收与自动触发 AC
- [Source: files/planning-artifacts/architecture.md#幂等键设计模式] — Webhook 去重
- [Source: files/planning-artifacts/architecture.md#API 安全] — Rate Limiting /webhooks/* 100次/分钟

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6

### Code Review (2026-03-05)
- [H3+H4] Webhook Handler 改用 Service.VerifyGitLabWebhook / VerifyGitHubWebhook 公开方法，不再直接访问私有 decryptToken
- [M7] parseGitLabEvent/parseGitHubEvent 添加必要字段（repo name、commit SHA）校验
- [M8] 幂等键改为同时检查当前分钟和上一分钟，修复跨分钟边界去重失效

### Debug Log References

### Completion Notes List
- 创建 migration 000011 为 git_connections 添加 webhook_secret 字段
- 实现 GitLab 签名验证（X-Gitlab-Token 直比）和 GitHub 签名验证（X-Hub-Signature-256 HMAC-SHA256）
- webhook_handler.go 实现双端点 POST /webhooks/gitlab 和 /webhooks/github
- 解析 GitLab/GitHub push event payload，提取 repo/branch/commit/author 信息
- 幂等去重：Redis SETNX，key=webhook:{event_type}:{repo}:{commit_sha}:{timestamp_minute}，TTL 5min
- 创建连接时自动生成 32 字节随机 webhook_secret 并 AES 加密存储
- 添加 GET /:connId/webhook-secret 端点供管理员获取明文 secret 配置到 Git 仓库
- 流水线匹配触发留 TODO 给 Epic 6，当前仅记录事件日志
- 7 个 webhook 签名验证单元测试 + 1 个幂等键测试通过
- 全量回归测试通过

### File List
- `migrations/000011_add_webhook_secret.up.sql` (新建)
- `migrations/000011_add_webhook_secret.down.sql` (新建)
- `pkg/gitprovider/webhook.go` (新建)
- `pkg/gitprovider/webhook_test.go` (新建)
- `internal/git/webhook_handler.go` (新建)
- `internal/git/model.go` (修改 — 添加 WebhookSecret 字段)
- `internal/git/service.go` (修改 — 添加 GetWebhookSecret、FindConnectionByServerURL、generateWebhookSecret)
- `internal/git/handler.go` (修改 — 添加 GetWebhookSecret 端点)
- `cmd/server/main.go` (修改 — 注册 webhook 路由)

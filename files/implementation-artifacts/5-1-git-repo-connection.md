# Story 5.1: Git 仓库连接配置

Status: done

## Story

As a 管理员,
I want 配置 GitLab 和 GitHub 仓库连接,
so that 平台可以访问代码仓库，为流水线构建提供代码源。

## Acceptance Criteria

1. **创建 Git 连接（GitLab OAuth / GitHub OAuth）**
   - Given 管理员已登录
   - When POST /api/v1/admin/integrations 配置 GitLab/GitHub 连接
   - Then 连接信息（access_token, refresh_token）使用 AES-256-GCM 加密存储
   - And 创建成功返回连接 ID 和状态

2. **查询 Git 连接列表**
   - Given Git 连接已配置
   - When GET /api/v1/admin/integrations
   - Then 返回连接列表及状态（connected / disconnected / token_expired）
   - And Token 值不回显，仅显示掩码

3. **测试 Git 连接**
   - Given Git 连接已配置
   - When POST /api/v1/admin/integrations/:id/test
   - Then 系统使用存储的凭证调用 Git API 验证连接
   - And 返回测试结果（成功/失败原因）

4. **更新 Git 连接**
   - Given Git 连接已配置
   - When PUT /api/v1/admin/integrations/:id
   - Then 可更新连接名称、描述、凭证（Token 重新加密存储）

5. **删除 Git 连接**
   - Given Git 连接已配置
   - When DELETE /api/v1/admin/integrations/:id
   - Then 连接标记为删除，关联数据保留但不可用

6. **OAuth Token 自动刷新**
   - Given OAuth Token 过期
   - When 系统尝试访问 Git API
   - Then 自动使用 refresh_token 刷新 access_token
   - And 刷新失败时标记连接为 token_expired

7. **GitProvider 接口抽象**
   - Given 系统需要支持 GitLab 和 GitHub
   - When 实现 GitProvider 接口
   - Then 包含 TestConnection、ListRepos、ListBranches、RefreshToken 方法
   - And GitLab 和 GitHub 各自实现该接口

## Tasks / Subtasks

- [x] Task 1: 创建数据库迁移 — git_connections 表 (AC: #1, #2)
  - [x] 1.1 创建 000010_create_git_connections.up.sql
  - [x] 1.2 创建 000010_create_git_connections.down.sql
- [x] Task 2: 创建 GitProvider 接口抽象 (AC: #7)
  - [x] 2.1 创建 pkg/gitprovider/provider.go — 接口定义
  - [x] 2.2 创建 pkg/gitprovider/gitlab.go — GitLab 实现（REST API v4）
  - [x] 2.3 创建 pkg/gitprovider/github.go — GitHub 实现（REST API v3）
  - [x] 2.4 创建 pkg/gitprovider/types.go — 共享类型定义
- [x] Task 3: 创建 internal/git 模块 (AC: #1-#6)
  - [x] 3.1 创建 internal/git/model.go — GitConnection 数据模型
  - [x] 3.2 创建 internal/git/dto.go — 请求/响应 DTO
  - [x] 3.3 创建 internal/git/repo.go — 数据访问层
  - [x] 3.4 创建 internal/git/service.go — 业务逻辑（含加解密、Token 刷新）
  - [x] 3.5 创建 internal/git/handler.go — HTTP handler
  - [x] 3.6 创建 internal/git/service_test.go — 单元测试
- [x] Task 4: 注册路由到 main.go (AC: #1-#5)
  - [x] 4.1 在 cmd/server/main.go 添加 registerIntegrationRoutes
  - [x] 4.2 路由挂载到 /api/v1/admin/integrations
- [x] Task 5: 扩展 Config 支持 Git 集成配置 (AC: #6)
  - [x] 5.1 Config 无需修改，Git 集成使用已有的 AESCrypto 基础设施
- [x] Task 6: 错误码注册 (AC: #1-#6)
  - [x] 6.1 在 pkg/response/codes.go 注册 4041x 段错误码

## Dev Notes

### 架构决策

**GitProvider 接口抽象（架构文档 NFR22）：**

外部系统接口必须抽象化。Git 集成通过 `pkg/gitprovider/` 提供统一接口，后端业务模块（`internal/git/`）只依赖接口，不直接依赖 GitLab/GitHub SDK。

```go
// pkg/gitprovider/provider.go
type GitProvider interface {
    TestConnection(ctx context.Context) error
    ListRepos(ctx context.Context, page, pageSize int) ([]Repository, int, error)
    ListBranches(ctx context.Context, repoFullName string) ([]Branch, error)
    RefreshToken(ctx context.Context, refreshToken string) (*TokenPair, error)
    GetProviderType() ProviderType
}
```

**数据模型设计：**

```sql
CREATE TABLE git_connections (
    id            UUID PRIMARY KEY DEFAULT uuid_generate_v4(),
    name          VARCHAR(100) NOT NULL,
    provider_type VARCHAR(20) NOT NULL,  -- 'gitlab' | 'github'
    server_url    VARCHAR(500) NOT NULL, -- GitLab 实例 URL 或 https://github.com
    access_token  TEXT NOT NULL,         -- AES-256-GCM 加密
    refresh_token TEXT,                  -- AES-256-GCM 加密（可选）
    token_type    VARCHAR(20) NOT NULL DEFAULT 'pat', -- 'pat' | 'oauth'
    status        VARCHAR(20) NOT NULL DEFAULT 'connected',
    description   TEXT,
    created_by    UUID NOT NULL REFERENCES users(id),
    created_at    TIMESTAMPTZ NOT NULL DEFAULT NOW(),
    updated_at    TIMESTAMPTZ NOT NULL DEFAULT NOW()
);

CREATE INDEX idx_git_connections_provider_type ON git_connections(provider_type);
CREATE INDEX idx_git_connections_status ON git_connections(status);
```

**MVP 简化决策：**

MVP 阶段采用 Personal Access Token (PAT) 方式连接，OAuth 授权流程作为 P1 功能预留接口但暂不实现完整的 OAuth 回调流程。原因：
- PAT 方式实现简单，管理员在 GitLab/GitHub 生成 Token 后粘贴即可
- 核心功能（列出仓库、列出分支、验证连接）PAT 和 OAuth 方式完全相同
- 数据模型已预留 token_type 和 refresh_token 字段，后续支持 OAuth 零迁移

**Token 加密存储：**

复用已有的 `pkg/crypto/aes.go` AES-256-GCM 加密模块（Epic 4 已实现），与变量密钥加密使用同一套基础设施。

**错误码分配（架构文档错误码段 404xx）：**

| 错误码 | 说明 |
|--------|------|
| 40401 | OAuth 授权失败 / Token 无效 |
| 40402 | Webhook 签名验证失败（Story 5.3 使用） |
| 40403 | Git 连接不存在 |
| 40404 | Git 连接已断开 |
| 40405 | Git Provider 不支持 |
| 40406 | Git API 调用失败 |

### 代码模式参考

**handler→service→repo 三层结构 — 参考 internal/variable/ 模块：**

- handler.go: RegisterRoutes 注册路由，只做参数绑定和响应输出
- service.go: 编排业务逻辑，调用 repo 和 pkg/crypto
- repo.go: 数据访问层，单表 CRUD
- model.go: GORM 模型定义
- dto.go: 请求/响应 DTO

**路由注册模式 — 参考 main.go registerAdminRoutes：**

```go
// 管理员路由需要 RequireAdminRBAC 中间件
integrations := v1.Group("/admin/integrations")
integrations.Use(middleware.RequireAdminRBAC(jwtSecret))
gitHandler.RegisterRoutes(integrations)
```

**加密/解密模式 — 参考 internal/variable/service.go：**

```go
// 创建时加密
encrypted, err := s.crypto.Encrypt(plainToken)
// 读取时解密（内部使用，不返回给前端）
decrypted, err := s.crypto.Decrypt(encrypted)
```

### 技术约束

1. **Git API 库选择：**
   - GitLab: 使用 `github.com/xanzy/go-gitlab` — Go 社区最成熟的 GitLab API v4 客户端
   - GitHub: 使用 `github.com/google/go-github/v70` — Google 官方维护的 GitHub API v3 客户端
   - 两者均通过 PAT 认证，无需额外的 OAuth 库

2. **连接状态枚举：**
   - `connected` — 连接正常
   - `disconnected` — 连接已断开（管理员主动断开或 Token 被撤销）
   - `token_expired` — Token 过期（自动刷新失败）

3. **Git API 调用限制：**
   - GitLab: 默认 2000 次/分钟
   - GitHub: PAT 5000 次/小时
   - 后续 Story 5.2 实现仓库/分支缓存（Redis 5min TTL）

4. **安全考虑：**
   - Token 必须 AES-256-GCM 加密存储，密钥来自 ZCID_ENCRYPTION_KEY 环境变量
   - API 响应中 Token 显示为掩码 `****...xxxx`（最后 4 位）
   - 日志中禁止打印 Token 明文

### Project Structure Notes

新增文件遵循已有的 internal/ 模块化结构：

```
zcid/
├── internal/git/              # 新增模块
│   ├── handler.go
│   ├── service.go
│   ├── service_test.go
│   ├── repo.go
│   ├── model.go
│   └── dto.go
├── pkg/gitprovider/           # 新增公共包
│   ├── provider.go            # GitProvider 接口
│   ├── gitlab.go              # GitLab 实现
│   ├── github.go              # GitHub 实现
│   └── types.go               # 共享类型
├── migrations/
│   ├── 000010_create_git_connections.up.sql    # 新增
│   └── 000010_create_git_connections.down.sql  # 新增
└── cmd/server/main.go         # 修改：添加 registerIntegrationRoutes
```

### References

- [Source: files/planning-artifacts/architecture.md#Backend Code Organization Convention] — handler→service→repo 三层结构
- [Source: files/planning-artifacts/architecture.md#Authentication & Security] — AES-256-GCM 加密策略
- [Source: files/planning-artifacts/architecture.md#API & Communication Patterns] — REST 路由 /api/v1/admin/integrations
- [Source: files/planning-artifacts/architecture.md#错误码段分配] — 404xx 段 Git 集成
- [Source: files/planning-artifacts/epics.md#Story 5.1] — Git 仓库连接配置 AC
- [Source: files/planning-artifacts/epics.md#Epic 5] — FR15-FR19 覆盖
- [Source: files/planning-artifacts/ux-design-specification.md#集成状态 Dashboard] — 集成管理页面 wireframe
- [Source: internal/variable/] — 同结构模块实现参考
- [Source: pkg/crypto/aes.go] — AES-256-GCM 加密模块复用

### Previous Story Intelligence

**Epic 4 完成情况：**
- 4 个 Story 全部完成，变量 CRUD + AES 加密 + 运行时注入 + 前端页面
- `pkg/crypto/aes.go` 已实现并测试，可直接复用于 Git Token 加密
- `internal/variable/` 模块结构是本 Story 的最佳参考模板
- 前端 `services/variable.ts` API 层模式可用于 Story 5.4 参考

**关键学习：**
- 加密模块需要 ZCID_ENCRYPTION_KEY 环境变量，未设置时功能降级而非 panic
- handler 层使用 `middleware.RequireAdminRBAC` 保护管理员端点
- 路由注册在 main.go 中独立函数，接收 db、jwtSecret 等依赖
- 测试文件与源文件同目录（Go 标准实践）

## Dev Agent Record

### Agent Model Used
Claude Opus 4.6

### Debug Log References
- Build error: 错误码 40401 与 CodeNotFound 冲突，修改为 4041x 段

### Code Review (2026-03-05)
**Reviewer**: Adversarial AI Code Review
**Issues Found**: 4 HIGH, 10 MEDIUM, 3 LOW (across all Epic 5 stories)
**All HIGH and MEDIUM issues fixed**, verified by full regression tests.
**Key Fixes Applied**:
- [H1] Token 掩码改为基于明文而非密文（PlainTokenMask 字段 + 解密后计算）
- [H2] FindConnectionByServerURL 改用 DB 查询替代全表扫描
- [H3+H4] Webhook Handler 重构为使用 Service 公开方法（VerifyGitLabWebhook/VerifyGitHubWebhook）
- [M1] encryptToken 错误码从 CodeDecryptFailed 改为 CodeEncryptFailed（40503）

### Completion Notes List
- 创建 git_connections 数据库迁移（UUID 主键，唯一名称约束，provider_type/status 索引）
- 实现 GitProvider 接口抽象（pkg/gitprovider/），支持 GitLab REST API v4 和 GitHub REST API v3
- 使用原生 net/http 调用 Git API，无额外第三方 SDK 依赖（MVP 简洁策略）
- 实现 internal/git 模块完整的 handler→service→repo 三层结构
- Token 使用 AES-256-GCM 加密存储，复用 pkg/crypto 模块
- API 响应中 Token 显示掩码（****xxxx，保留最后 4 位）
- TestConnection 端点通过 GitProvider 接口验证连接，失败时自动更新 status
- 支持 PAT 和 OAuth 两种 token_type（MVP 以 PAT 为主）
- 错误码注册在 4041x 段（避开 40401 CodeNotFound 冲突）
- 9 个单元测试全部通过，全量回归测试零失败
- 路由挂载在 /api/v1/admin/integrations，使用 RequireAdminRBAC 中间件

### File List
- `migrations/000010_create_git_connections.up.sql` (新建)
- `migrations/000010_create_git_connections.down.sql` (新建)
- `pkg/gitprovider/provider.go` (新建)
- `pkg/gitprovider/types.go` (新建)
- `pkg/gitprovider/errors.go` (新建)
- `pkg/gitprovider/gitlab.go` (新建)
- `pkg/gitprovider/github.go` (新建)
- `internal/git/model.go` (新建)
- `internal/git/dto.go` (新建)
- `internal/git/repo.go` (新建)
- `internal/git/service.go` (新建)
- `internal/git/handler.go` (新建)
- `internal/git/service_test.go` (新建)
- `cmd/server/main.go` (修改)
- `pkg/response/codes.go` (修改)

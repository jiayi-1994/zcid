# Epic 5 Retrospective: Git 仓库集成

**日期**: 2026-03-05
**参与者**: xjy (Developer), Bob (Scrum Master), Alice (Tech Lead), Carol (QA Lead), Dave (Architect), Emma (Product Manager)
**Epic 状态**: Done (4/4 stories completed, code review passed)

---

## Epic 概览

### Stories 完成情况

| Story | 名称 | 状态 | 模型 | 关键成果 |
|-------|------|------|------|----------|
| 5.1 | Git 仓库连接配置 | Done | Opus 4.6 | GitProvider 接口抽象, CRUD + AES-256-GCM Token 加密, 9 个单元测试 |
| 5.2 | 仓库与分支选择 | Done | Opus 4.6 | Redis 缓存层 (5min TTL), 分页仓库/分支列表, refresh 参数支持 |
| 5.3 | Webhook 接收与自动触发 | Done | Opus 4.6 | GitLab/GitHub 双端点签名验证, Redis 幂等去重, 自动生成 webhook_secret |
| 5.4 | 前端 Git 集成管理页面 | Done | Opus 4.6 | 连接列表/创建/编辑/删除, 状态 Badge, 测试连接, Webhook Secret 复制 |

**完成率**: 100% (4/4)
**技术栈**: Go + Gin + PostgreSQL + Redis + net/http (后端), React + TypeScript + Arco Design + Zustand (前端)

---

## 亮点 (What Went Well)

### 1. GitProvider 接口抽象设计优秀
- 统一的 `GitProvider` 接口同时支持 GitLab 和 GitHub
- 使用原生 `net/http` 调用 Git API，无额外第三方 SDK 依赖（MVP 简洁策略）
- 新增 Git 平台（如 Gitee）只需实现接口，零侵入现有代码
- 符合架构文档 NFR22 外部系统接口抽象化要求

### 2. 安全设计全面
- Access Token 和 Webhook Secret 均使用 AES-256-GCM 加密存储
- API 响应中 Token 显示掩码，PlainTokenMask 基于明文计算（Code Review 修复后）
- GitLab `X-Gitlab-Token` 直比 + GitHub `X-Hub-Signature-256` HMAC-SHA256 验证
- Webhook 端点无需 JWT 认证，通过签名验证保障安全

### 3. 缓存与幂等策略合理
- 仓库/分支列表 Redis 缓存 5min TTL，减少 Git API 调用压力
- 支持 `refresh=true` 强制刷新缓存
- Webhook 事件 Redis SETNX 幂等去重，支持跨分钟边界检测（Code Review 修复后）
- Token 失效时自动更新连接状态为 `token_expired`

### 4. 开发流程顺畅
- 3 个后端 Story + 1 个前端 Story 有序推进，每个 Story 独立可测
- 复用已有基础设施（pkg/crypto, pkg/cache, pkg/response）减少重复代码
- handler->service->repo 三层架构模式一致，代码结构清晰
- 前端参考 Epic 4 变量管理页面模式，开发效率高

---

## 问题与挑战 (What Didn't Go Well)

### 1. Code Review 发现 11 个问题 (4 HIGH + 7 MEDIUM)

**这是本 Epic 最大的问题 — 开发阶段遗漏了多个设计和实现缺陷。**

#### HIGH 级别问题

| 编号 | 问题 | 根因 | 修复 |
|------|------|------|------|
| H1 | `ToConnectionResponse` 对 AES 密文做 `maskToken`，掩码无意义 | Token 加密后存储，DTO 层无法区分密文和明文 | 添加 `PlainTokenMask` 非 DB 字段，Service 层基于明文计算掩码 |
| H2 | `FindConnectionByServerURL` 全表扫描 O(n) | 缺少 DB 查询方法，直接遍历内存列表 | 添加 `repo.GetByServerURL` 使用 SQL WHERE 查询 |
| H3 | Webhook Handler 遍历所有连接逐个解密 Secret | 初始实现未考虑连接数增长的性能影响 | 添加 `repo.ListByProviderType` 按 provider 过滤 |
| H4 | Webhook Handler 直接访问 Service 私有方法 `decryptToken` | 同 package 内可访问小写方法，但违反层级封装 | Service 新增 `VerifyGitLabWebhook`/`VerifyGitHubWebhook` 公开方法 |

#### MEDIUM 级别问题

| 编号 | 问题 | 根因 | 修复 |
|------|------|------|------|
| M1 | `encryptToken` 使用 `CodeDecryptFailed` 错误码 | 复用已有错误码时未注意语义 | 新增 `CodeEncryptFailed = 40503` |
| M2 | GitHub `ListRepos` 返回 `len(repos)` 作为 total | GitHub API 不提供 X-Total header，开发时未处理 | 基于分页大小估算 total |
| M5 | 前端编辑模式 description 空字符串处理不一致 | name/accessToken 使用 `\|\| undefined`，description 未处理 | 统一使用 `?? undefined` |
| M7 | Webhook payload 缺少必要字段验证 | 解析后未检查 repo name 和 commit SHA 是否为空 | 添加必填字段验证 |
| M8 | 幂等键 `timestamp_minute` 跨分钟边界去重失效 | 分钟边界处同一事件生成不同 key | 同时检查当前分钟和上一分钟的 key |
| M9 | 前端集成管理页面无分页 | `pagination={false}` 硬编码 | 启用 Arco Table 客户端分页 |
| M10 | `ApiResponse<T>` 在 3 个 service 文件中重复定义 | 各 service 文件独立开发时未提取共享类型 | 提取到 `services/types.ts` |

### 2. 错误码冲突 (开发阶段)

- 架构文档规划 Git 错误码段 `404xx`，但 `40401` 已被 `CodeNotFound` 占用
- Go build 报 `duplicate key 40401 in map literal`
- **修复**: 改用 `4041x` 段（40411-40417）

### 3. 类型转换遗漏 (开发阶段)

- `service.go` 中 `updates["status"] = StatusConnected` 传入 `ConnectionStatus` 类型
- mock repo 期望 `string` 类型，导致 `panic: interface conversion`
- **修复**: 显式 `string(StatusConnected)` 类型转换

### 4. MVP 技术债

- OAuth 完整授权流程仅预留接口，未实现
- Webhook -> Pipeline 触发留 TODO 给 Epic 6
- `ListBranches` 硬编码 `per_page=100`，超过 100 分支的仓库会丢失数据

---

## 关键洞察 (Key Insights)

### 1. 加密存储引入了 DTO 层复杂度
Token 加密后，所有需要显示 Token 信息的地方都需要额外的解密步骤。H1 问题的根因是：DTO 层（`ToConnectionResponse`）直接操作了密文字段而未意识到数据已加密。**经验**: 加密字段应在模型层或文档中有明确标注，DTO 层不应直接处理加密字段。

### 2. 同 package 内的封装同样重要
Go 的 package-level 可见性使得同 package 内的所有类型可以互访。H4 问题表明，即使语言允许，`WebhookHandler` 直接调用 `Service.decryptToken` 也违反了分层原则。**经验**: handler 层只应调用 service 层的公开（大写）方法，即使在同一个 package 内。

### 3. API 差异需要适配层
M2 问题说明 GitLab 和 GitHub 的 API 行为差异（X-Total header vs Link pagination）需要在 provider 层统一处理。原始实现只对 GitLab 做了 total 处理，GitHub 返回了错误值。**经验**: 接口抽象不仅要统一调用方式，还要统一返回值语义。

### 4. 时间相关的幂等设计要考虑边界
M8 问题是经典的时间边界 bug。基于 `timestamp_minute` 的幂等键在分钟交界处失效。**经验**: 时间窗口去重应该检查相邻窗口，或使用固定 TTL 的 key（不含时间戳）。

### 5. Code Review 的价值
本 Epic 在开发阶段通过编译错误和单元测试发现了 2 个问题（错误码冲突、类型转换），但 Code Review 额外发现了 11 个问题（4 HIGH + 7 MEDIUM）。这些问题中有安全相关的（H1 密文掩码）、性能相关的（H2/H3 全表扫描）和正确性相关的（M2 total 错误、M8 幂等失效）。**结论**: Code Review 是不可或缺的质量关卡。

---

## 行动项 (Action Items)

### Action Item 1: 加密字段模型标注规范
- **问题**: DTO 层直接操作密文字段（H1）
- **行动**:
  1. 在 model 中为加密字段添加注释标注 `// encrypted: AES-256-GCM`
  2. 建立规范：加密字段的展示必须经过 Service 层解密
  3. 考虑在 GORM model 上添加 `PlainXxx` 虚拟字段模式作为标准做法
- **负责人**: Alice (Tech Lead)
- **截止时间**: Epic 6 开始前
- **优先级**: HIGH

### Action Item 2: 完善 Webhook -> Pipeline 触发链路
- **问题**: Story 5.3 的 Webhook 事件处理预留了 TODO
- **行动**:
  1. Epic 6 Story 6-5 实现 Webhook 事件到流水线的完整匹配
  2. 补充 Webhook -> Pipeline 匹配 -> 执行触发的集成测试
- **负责人**: 后端开发者
- **截止时间**: Epic 6 Story 6-5 完成时
- **优先级**: HIGH

### Action Item 3: 错误码全局唯一性校验
- **问题**: 40401 冲突直到 go build 才发现（开发阶段）
- **行动**:
  1. 在 `pkg/response/codes.go` 中添加 `init()` 检测重复码值
  2. 更新架构文档的错误码段规划表，标注已占用码位
- **负责人**: Dave (Architect)
- **截止时间**: Epic 6 开始前
- **优先级**: MEDIUM

### Action Item 4: ListBranches 分页支持
- **问题**: 硬编码 per_page=100，超过 100 分支的仓库丢失数据
- **行动**:
  1. 为 GitLab/GitHub provider 的 ListBranches 实现自动翻页
  2. 或提升 per_page 上限并添加文档说明
- **负责人**: 后端开发者
- **截止时间**: Epic 6 并行进行
- **优先级**: LOW

### Action Item 5: 前端集成页面 E2E 测试
- **问题**: 前端仅通过 TypeScript 类型检查和现有单元测试验证
- **行动**:
  1. 为集成管理页面编写 Playwright E2E 测试
  2. 覆盖: 创建连接 -> 测试 -> 复制 Webhook Secret -> 编辑 -> 删除
  3. 包含权限测试: 非管理员访问 /admin/integrations 应重定向
- **负责人**: Carol (QA Lead)
- **截止时间**: Epic 6 开始前
- **优先级**: MEDIUM

---

## 度量指标

### 完成情况
- **Stories 完成**: 4/4 (100%)
- **AC 达成率**: 100%
- **后端测试通过率**: 100% (全量回归零失败)
- **前端测试通过率**: 100% (16/16 通过)

### 质量指标
- **开发阶段 Build Errors**: 2 个（错误码冲突 + 类型转换，均即时修复）
- **Code Review Issues**: 4 HIGH + 7 MEDIUM + 3 LOW = 14 个
- **Code Review 修复率**: 11/11 HIGH+MEDIUM 全部修复 (100%)
- **技术债**: 3 个（OAuth 流程 TODO、Webhook->Pipeline 触发 TODO、ListBranches 分页限制）

### 文件统计

**新增文件 (21 个)**:
- `migrations/000010_create_git_connections.up.sql`
- `migrations/000010_create_git_connections.down.sql`
- `migrations/000011_add_webhook_secret.up.sql`
- `migrations/000011_add_webhook_secret.down.sql`
- `pkg/gitprovider/provider.go`
- `pkg/gitprovider/types.go`
- `pkg/gitprovider/errors.go`
- `pkg/gitprovider/gitlab.go`
- `pkg/gitprovider/github.go`
- `pkg/gitprovider/webhook.go`
- `pkg/gitprovider/webhook_test.go`
- `internal/git/model.go`
- `internal/git/dto.go`
- `internal/git/repo.go`
- `internal/git/service.go`
- `internal/git/handler.go`
- `internal/git/service_test.go`
- `internal/git/webhook_handler.go`
- `web/src/services/integration.ts`
- `web/src/services/types.ts`
- `web/src/pages/admin/integrations/IntegrationsPage.tsx`
- `web/src/pages/admin/integrations/ConnectionFormModal.tsx`

**修改文件 (8 个)**:
- `cmd/server/main.go`
- `pkg/response/codes.go`
- `web/src/App.tsx`
- `web/src/stores/auth.ts`
- `web/src/components/layout/AppLayout.tsx`
- `web/src/services/variable.ts`
- `web/src/services/project.ts`

### 时间指标
- **Epic 开始**: 2026-03-05
- **Epic 完成**: 2026-03-05
- **Code Review**: 2026-03-05 (11 issues fixed)
- **总耗时**: 1 天

---

## 经验教训总结

### 对下个 Epic 的建议

1. **接口抽象后验证返回值语义**: Epic 6 流水线编排可能涉及不同执行引擎，确保接口统一后返回值的一致性
2. **错误码预分配并自动检测**: 提前为 Epic 6 分配错误码段，添加 init() 唯一性检查
3. **加密字段显式标注**: 所有新增的加密字段必须在 model 注释中标明
4. **Webhook 触发集成**: Epic 6 Story 6-5 需要与 Story 5.3 的 Webhook 基础设施对接
5. **时间边界测试**: 涉及时间窗口的逻辑必须测试边界条件

### Code Review 改进

1. **Review 应在开发完成后立即执行**: 本次 4 个 Story 一起 Review 效果好，但单个 Story 即时 Review 可以更早发现问题
2. **封装规则应明确文档化**: handler 只调用 service 公开方法的规则应写入项目规范
3. **前端共享类型应提前规划**: ApiResponse 重复定义说明缺少前端 TypeScript 类型规范

### 团队协作改进

1. **错误码治理**: 引入 init() 自动检测，防止冲突
2. **测试左移**: 开发阶段即运行全量回归测试，本 Epic 表现良好
3. **代码复用**: handler->service->repo 三层架构和 AES 加密模块在多个 Epic 间成功复用
4. **前端模式统一**: 参考已有管理页面模式开发新页面，显著提升效率

---

## 回顾会结论

Epic 5 整体交付质量高，在一天内完成了 4 个 Story 的全部开发和 Code Review。Code Review 发现并修复了 11 个 HIGH+MEDIUM 级问题，显著提升了代码质量。Git 仓库集成模块的设计（GitProvider 接口抽象、缓存策略、Webhook 安全验证）为后续 Epic 6 流水线编排提供了坚实的基础。

**关键成功因素**:
- GitProvider 接口抽象使后端 Story 之间的依赖关系清晰
- 复用已有基础设施（加密、缓存、响应格式）减少了重复开发
- 严格的 Code Review 发现了开发阶段遗漏的安全、性能和正确性问题
- 全量回归测试确保修复不引入新问题

**最大教训**:
- 加密字段的 DTO 展示需要特别注意（H1）
- 同 package 内也要遵守分层封装原则（H4）
- API 接口抽象要统一返回值语义而非仅统一调用方式（M2）

**对 Epic 6 的信心**: HIGH（Git 集成基础设施已就绪，Code Review 流程已验证有效）

---

**下次回顾会**: Epic 6 完成后

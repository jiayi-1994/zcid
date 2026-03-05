---
stepsCompleted: [1, 2, 3, 4]
inputDocuments:
  - 'files/planning-artifacts/prd.md'
  - 'files/planning-artifacts/architecture.md'
  - 'files/planning-artifacts/ux-design-specification.md'
---

# zcid - Epic Breakdown

## Overview

This document provides the complete epic and story breakdown for zcid, decomposing the requirements from the PRD, UX Design, and Architecture into implementable stories.

## Requirements Inventory

### Functional Requirements

- FR1: 用户可以通过账号密码登录平台并获取认证凭证
- FR2: 管理员可以创建、编辑、禁用用户账号
- FR3: 管理员可以为用户分配系统级角色（管理员 / 项目管理员 / 普通成员）
- FR4: 系统根据用户角色和项目归属，控制其对资源和操作的访问权限
- FR5: 密钥类型变量对普通成员完全不可见
- FR6: 管理员可以创建和删除项目
- FR7: 项目管理员可以在项目内创建、编辑、删除环境，并将环境映射到 K8s Namespace
- FR8: 项目管理员可以在项目内创建、编辑、删除服务
- FR9: 项目管理员可以将用户添加到项目并分配项目内角色
- FR10: 不同项目之间的流水线、环境、服务、变量互不可见
- FR11: 项目管理员可以在全局、项目、流水线三个层级创建和管理变量，下级覆盖上级
- FR12: 项目管理员可以创建密钥类型变量，密钥变量加密存储且界面不可回显
- FR13: 流水线运行时，系统自动将密钥变量以临时方式注入执行环境，运行结束后自动清理
- FR14: 系统在所有日志输出中自动脱敏密钥变量值
- FR15: 管理员可以配置 GitLab 和 GitHub 仓库连接（OAuth 授权）
- FR16: 用户在创建流水线时可以从已关联的仓库中选择代码仓库和分支
- FR17: 系统可以接收 GitLab/GitHub 的 Webhook 推送事件并验证签名
- FR18: 系统根据 Webhook 事件的仓库、分支、事件类型自动匹配并触发对应的流水线
- FR19: 系统对 Webhook 事件进行幂等性去重，防止重复触发
- FR20: 项目管理员可以通过可视化界面编排流水线（Stage→Step 模型）
- FR21: 用户可以从预置模板一键创建流水线，只需填写少量参数
- FR22: 高级用户可以切换到 YAML 模式直接编辑流水线配置
- FR23: 系统将用户的流水线配置翻译为 Tekton PipelineRun CRD 并提交执行
- FR24: 用户可以手动触发流水线运行
- FR25: 用户可以为流水线配置 Webhook 自动触发规则
- FR26: 项目管理员可以为流水线配置并发控制策略
- FR27: 多个流水线可以并行运行，互不干扰
- FR28: 流水线运行时自动注入本次触发的 Git 信息
- FR29: 项目管理员可以复制已有流水线配置来创建新流水线
- FR30: 用户手动触发流水线时可以临时指定或覆盖运行时参数
- FR31: 用户可以取消正在运行的流水线
- FR32: 系统支持容器化构建链路：代码拉取→编译构建→镜像构建→镜像推送
- FR33: 系统支持传统构建链路：代码拉取→编译打包→产物上传到对象存储
- FR34: 管理员可以配置镜像仓库连接（Harbor）
- FR35: 用户可以查看构建产物信息
- FR36: 用户可以实时查看流水线每个 Step 的执行状态
- FR37: 用户可以实时查看正在运行的构建步骤的日志输出
- FR38: 构建失败时，系统醒目标识失败的 Step 并高亮显示错误日志
- FR39: 日志连接断开后可自动重连并从断点续传
- FR40: 构建完成后，系统将日志归档以供历史查看
- FR41: 用户可以查看历史构建运行的日志
- FR42: 系统持续监听已提交的流水线运行状态变更，并实时同步到平台
- FR43: 用户可以查看流水线的运行历史列表
- FR44: 系统可以将构建产物部署到指定 K8s 环境
- FR45: 用户可以查看每个环境中各服务的部署状态
- FR46: 用户可以查看部署的同步详情和错误信息
- FR47: 项目管理员可以手动触发重新同步
- FR48: 项目管理员可以查看部署历史并回滚到指定版本
- FR49: 普通成员只能在 dev 环境触发部署，staging/prod 需要更高权限
- FR50: 项目管理员可以为流水线配置通知规则
- FR51: 系统通过 Webhook（HTTP POST）发送通知
- FR52: 系统记录所有写操作的审计日志
- FR53: 用户登录后可以看到项目及最近构建和环境状态概览
- FR54: 技术管理者可以跨项目查看所有环境的健康状态汇总
- FR55: 首次登录的用户可以看到所属项目和快捷操作入口
- FR56: 用户可以在主要列表中进行筛选和搜索
- FR57: 管理员可以配置系统级设置
- FR58: 管理员可以查看外部集成的连接状态
- FR59: 系统启动时自动检测 Tekton 和 ArgoCD 的版本兼容性
- FR60: 系统可以检测关键依赖服务的健康状态并展示降级提示
- FR61: 系统自动清理过期的 PipelineRun CRD 资源
- FR62: 系统检测 ArgoCD Application 被外部修改时展示告警信息

### Non-Functional Requirements

- NFR1: API 响应时间 — 常规 CRUD < 500ms（P95），列表查询含分页 < 1s（P95）
- NFR2: CRD 翻译与提交延迟 — 触发到 PipelineRun CRD 提交 < 5 秒
- NFR3: 部署触发延迟 — 触发到 ArgoCD Application 开始同步 < 30 秒
- NFR4: WebSocket 日志推送延迟 < 2 秒，首次连接到日志流出现 < 5 秒
- NFR5: 首屏加载 < 3 秒，页面内导航切换 < 1 秒
- NFR6: 单实例支持 50 并发用户、20 条流水线并发运行
- NFR7: 密钥变量 AES-256-GCM 加密存储，密钥独立管理
- NFR8: 所有 API 支持 HTTPS/TLS，WebSocket 支持 WSS
- NFR9: JWT 合理过期 + Token 刷新，密码 bcrypt 哈希
- NFR10: 密钥变量值在所有日志输出中脱敏为 `***`，纳入自动化测试
- NFR11: 临时 K8s Secret 注入，PipelineRun 结束后 30 秒内清理
- NFR12: Webhook 签名验证，失败返回 401 并记录审计
- NFR13: ServiceAccount ClusterRole 最小权限，不使用通配符
- NFR14: 月可用率 > 99.5%
- NFR15: 流水线失败隔离，不影响其他运行中的流水线
- NFR16: 构建日志 Pod 回收后仍可归档查看，归档成功率 > 99%
- NFR17: K8s API 短暂不可达时展示缓存状态和降级提示
- NFR18: Harbor 推送失败自动重试 3 次
- NFR19: 支持 100+ 项目、1000+ 流水线、10 万+ 运行记录
- NFR20: PipelineRun/TaskRun TTL 自动清理
- NFR21: 单实例 200+ 并发 WebSocket 连接
- NFR22: 外部系统接口抽象（GitProvider/RegistryProvider/Notifier/ClusterManager）
- NFR23: Tekton Pipeline v1 API（v0.44+）兼容，启动检测
- NFR24: ArgoCD gRPC API 版本差异处理
- NFR25: RESTful 统一响应格式 + 错误码段 + OpenAPI 文档自动生成
- NFR26: 审计日志保留 90 天
- NFR27: 健康检查端点覆盖 DB/Redis/K8s
- NFR28: 应用日志 JSON 结构化输出，支持动态调级

### Additional Requirements

**架构文档额外需求：**

- ARCH-1: 项目基础骨架搭建 — Go+Gin 后端 + React+TS+Arco 前端 + docker-compose 开发环境（PostgreSQL/Redis/MinIO）
- ARCH-2: 应用配置管理 — config.yaml + 环境变量覆盖，敏感配置通过环境变量注入
- ARCH-3: 结构化日志 — Go slog JSON Handler + 日志脱敏 Handler 包装层 + 运行时动态调级
- ARCH-4: 数据库连接 + GORM + golang-migrate 迁移框架
- ARCH-5: Redis 连接 + 缓存层（Casbin 策略缓存、Session 缓存、健康检查缓存、Git 仓库缓存）
- ARCH-6: JWT 双 Token 认证（Access 30min + Refresh 7天 Redis 存储）
- ARCH-7: Casbin RBAC 四元组模型（sub, proj, obj, act）+ Redis Watcher 热更新
- ARCH-8: 统一错误处理 + 响应格式 + 错误码段分配
- ARCH-9: WebSocket 消息协议（type/payload/timestamp/seq）+ 连接管理器 + 心跳 + 断点续传
- ARCH-10: @xyflow/react v12 流水线可视化编排器 + dagre 布局
- ARCH-11: Helm Chart 部署（charts/zcid/）
- ARCH-12: 健康检查三级端点（/healthz, /readyz, /api/v1/health）
- ARCH-13: 前端 API 客户端自动生成（swag v2 → @hey-api/openapi-ts）
- ARCH-14: 前端状态管理分层（TanStack Query + Zustand + WebSocket Manager + React Router）
- ARCH-15: 前端路由 + 权限守卫（React Router v7 + Casbin 数据）

**UX 设计文档额外需求：**

- UX-1: 三层组件架构（Base Arco → Extension → Domain），依赖规则强制执行
- UX-2: PipelineRenderer 共享渲染基座，StagePreview 复用无交互开销
- UX-3: LogViewer xterm.js 架构（scrollback 50k、MinIO 归档分页、性能基线 10k<1s/100k<3s）
- UX-4: DynamicForm @rjsf/core + @zcid/rjsf-arco-theme 独立包
- UX-5: Monaco Editor 懒加载（React.lazy + hover 预加载）
- UX-6: STATUS_MAP 全局状态字典（状态→颜色→图标单一来源）
- UX-7: Hooks 数据解耦层（usePipeline/useLogStream/useDeployStatus）
- UX-8: Design Token 双轨消费（Arco Less 编译时 + CSS Variables 运行时）
- UX-9: 响应式断点实现（768/1024/1280/1440px + sidebar 自动折叠）
- UX-10: 可访问性测试流水线（axe-core CI blocking + eslint-plugin-jsx-a11y + Lighthouse CI score>90）
- UX-11: 按钮三级层次（Primary 每页最多1个 / Secondary / Text）
- UX-12: 反馈四级模式（按钮 loading / Message 3s / Notification 手动关闭 / 持久状态变更）
- UX-13: 键盘快捷键体系（Cmd+S 保存、Cmd+Enter 触发运行、Escape 关闭面板）
- UX-14: Apple+Linear 混合设计方向（中低密度、蓝白中性色 + 精确语义色）
- UX-15: StageNode/StepNode 三模式（edit/runtime/mini）统一流水线视觉语言

### FR Coverage Map

| FR | Epic | 描述 |
|----|------|------|
| FR1 | Epic 2 | 账号密码登录 |
| FR2 | Epic 2 | 用户账号管理 |
| FR3 | Epic 2 | 系统级角色分配 |
| FR4 | Epic 2 | 角色权限控制 |
| FR5 | Epic 2 | 密钥变量对普通成员不可见 |
| FR6 | Epic 3 | 项目创建删除 |
| FR7 | Epic 3 | 环境管理与 Namespace 映射 |
| FR8 | Epic 3 | 服务管理 |
| FR9 | Epic 3 | 项目成员与角色 |
| FR10 | Epic 3 | 项目间隔离 |
| FR11 | Epic 4 | 多层级变量管理 |
| FR12 | Epic 4 | 密钥变量加密存储 |
| FR13 | Epic 4 | 运行时密钥注入与清理 |
| FR14 | Epic 4 | 日志脱敏 |
| FR15 | Epic 5 | Git 仓库连接配置 |
| FR16 | Epic 5 | 仓库/分支选择 |
| FR17 | Epic 5 | Webhook 接收与签名验证 |
| FR18 | Epic 5 | Webhook 自动匹配触发 |
| FR19 | Epic 5 | Webhook 幂等去重 |
| FR20 | Epic 6 | 可视化流水线编排 |
| FR21 | Epic 6 | 模板一键创建 |
| FR22 | Epic 6 | YAML 模式编辑 |
| FR23 | Epic 7 | CRD 翻译与提交 |
| FR24 | Epic 7 | 手动触发运行 |
| FR25 | Epic 6 | Webhook 触发规则配置 |
| FR26 | Epic 6 | 并发控制策略 |
| FR27 | Epic 7 | 多流水线并行 |
| FR28 | Epic 7 | Git 信息自动注入 |
| FR29 | Epic 6 | 复制流水线 |
| FR30 | Epic 6 | 运行时参数覆盖 |
| FR31 | Epic 7 | 取消运行中流水线 |
| FR32 | Epic 7 | 容器化构建链路 |
| FR33 | Epic 7 | 传统构建链路 |
| FR34 | Epic 7 | 镜像仓库连接 |
| FR35 | Epic 7 | 构建产物查看 |
| FR36 | Epic 8 | Step 执行状态实时查看 |
| FR37 | Epic 8 | 实时构建日志 |
| FR38 | Epic 8 | 失败 Step 高亮 |
| FR39 | Epic 8 | 日志断线重连续传 |
| FR40 | Epic 8 | 日志归档 |
| FR41 | Epic 8 | 历史日志查看 |
| FR42 | Epic 8 | 状态变更实时同步 |
| FR43 | Epic 8 | 运行历史列表 |
| FR44 | Epic 9 | 部署到 K8s 环境 |
| FR45 | Epic 9 | 部署状态查看 |
| FR46 | Epic 9 | 同步详情与错误 |
| FR47 | Epic 9 | 手动重新同步 |
| FR48 | Epic 9 | 部署历史与回滚 |
| FR49 | Epic 9 | 部署权限分级 |
| FR50 | Epic 10 | 通知规则配置 |
| FR51 | Epic 10 | Webhook 通知发送 |
| FR52 | Epic 10 | 审计日志 |
| FR53 | Epic 11 | 项目概览 |
| FR54 | Epic 11 | 跨项目环境健康汇总 |
| FR55 | Epic 11 | 首次登录引导 |
| FR56 | Epic 11 | 列表筛选搜索 |
| FR57 | Epic 10 | 系统级设置 |
| FR58 | Epic 10 | 集成连接状态 |
| FR59 | Epic 10 | 版本兼容性检测 |
| FR60 | Epic 10 | 依赖健康检测与降级 |
| FR61 | Epic 10 | PipelineRun CRD 清理 |
| FR62 | Epic 10 | ArgoCD 外部修改告警 |

## Epic List

### Epic 1: 项目基础骨架与开发环境
搭建完整的项目骨架和开发环境，使开发团队可以启动本地开发、运行测试、构建部署，为所有后续功能提供技术基础。
**FRs covered:** 无直接 FR（基础设施 Epic）
**额外需求:** ARCH-1, ARCH-2, ARCH-3, ARCH-4, ARCH-5, ARCH-8, ARCH-11, ARCH-12, ARCH-13, UX-1, UX-6, UX-8, UX-9, UX-14

### Epic 2: 用户认证与权限管理
用户可以登录平台、管理账号、分配角色，系统根据角色和项目归属控制资源访问权限。
**FRs covered:** FR1, FR2, FR3, FR4, FR5
**额外需求:** ARCH-6, ARCH-7, ARCH-15

### Epic 3: 项目与资源管理
管理员可以创建项目，项目管理员可以管理环境、服务和成员，不同项目之间完全隔离。
**FRs covered:** FR6, FR7, FR8, FR9, FR10

### Epic 4: 变量与凭证管理
项目管理员可以在多层级管理变量和密钥，系统保证加密存储、运行时安全注入和日志脱敏。
**FRs covered:** FR11, FR12, FR13, FR14
**NFR 关联:** NFR7, NFR10, NFR11

### Epic 5: Git 仓库集成
管理员可以配置 Git 仓库连接，用户创建流水线时选择仓库和分支，系统接收 Webhook 并自动匹配触发流水线。
**FRs covered:** FR15, FR16, FR17, FR18, FR19
**NFR 关联:** NFR12

### Epic 6: 流水线可视化编排
用户可以通过可视化界面或模板创建流水线，高级用户可切换 YAML 模式，支持并发策略、触发规则、复制和参数覆盖。
**FRs covered:** FR20, FR21, FR22, FR25, FR26, FR29, FR30
**额外需求:** ARCH-10, UX-2, UX-4, UX-5, UX-7, UX-11, UX-12, UX-13, UX-15

### Epic 7: 流水线执行与构建
系统将流水线配置翻译为 Tekton CRD 并执行，支持容器化和传统两条构建链路，用户可以触发、取消运行，查看构建产物。
**FRs covered:** FR23, FR24, FR27, FR28, FR31, FR32, FR33, FR34, FR35
**NFR 关联:** NFR2, NFR6, NFR15

### Epic 8: 实时日志与状态监控
用户可以实时查看构建日志和 Step 状态，失败时高亮错误，支持断线重连、断点续传和历史日志归档查看。
**FRs covered:** FR36, FR37, FR38, FR39, FR40, FR41, FR42, FR43
**额外需求:** ARCH-9, UX-3, UX-7
**NFR 关联:** NFR4, NFR16, NFR21

### Epic 9: 部署与环境管理
用户可以将构建产物部署到 K8s 环境，查看部署状态和同步详情，手动同步和回滚，部署权限分级控制。
**FRs covered:** FR44, FR45, FR46, FR47, FR48, FR49
**额外需求:** UX-7
**NFR 关联:** NFR3

### Epic 10: 通知、审计与平台运维
管理员可以配置通知规则和系统设置，系统记录审计日志，自动检测依赖健康状态和版本兼容性，清理过期资源。
**FRs covered:** FR50, FR51, FR52, FR57, FR58, FR59, FR60, FR61, FR62
**NFR 关联:** NFR26, NFR27, NFR28

### Epic 11: 全局概览与用户引导
用户登录后看到项目概览和环境健康汇总，首次用户有引导入口，主要列表支持筛选搜索。
**FRs covered:** FR53, FR54, FR55, FR56
**额外需求:** UX-10

## Epic 1: 项目基础骨架与开发环境

搭建完整的项目骨架和开发环境，使开发团队可以启动本地开发、运行测试、构建部署，为所有后续功能提供技术基础。

### Story 1.1: 后端项目骨架与开发环境

As a 开发者,
I want 一个可运行的 Go+Gin 后端项目骨架和本地开发环境,
So that 我可以立即开始编写业务代码。

**Acceptance Criteria:**

**Given** 开发者克隆了代码仓库
**When** 执行 `docker-compose up -d` 和 `make dev`
**Then** PostgreSQL、Redis、MinIO 容器启动，后端服务在 localhost:8080 运行
**And** `GET /healthz` 返回 200

**Given** 后端服务启动
**When** 访问 `GET /readyz`
**Then** 返回 DB 和 Redis 连接状态
**And** 连接正常时返回 200，异常时返回 503

**Given** config.yaml 和环境变量均配置了同一字段
**When** 服务启动加载配置
**Then** 环境变量值覆盖 config.yaml 值
**And** 敏感配置（密码、密钥）仅通过环境变量注入

### Story 1.2: 数据库迁移框架

As a 开发者,
I want golang-migrate 迁移框架就绪,
So that 后续每个 Story 可以按需创建数据库表。

**Acceptance Criteria:**

**Given** 迁移框架已集成
**When** 执行 `make migrate-up`
**Then** migrations/ 目录下的 SQL 文件按序号执行
**And** 数据库 schema_migrations 表记录当前版本

**Given** 需要回滚
**When** 执行 `make migrate-down`
**Then** 最近一次迁移被回滚

**Given** 开发者需要新建迁移
**When** 执行 `make migrate-new name=create_xxx`
**Then** 生成带序号的 up/down SQL 文件对

### Story 1.3: 统一错误处理与响应格式

As a 开发者,
I want 统一的 API 响应格式和错误处理机制,
So that 所有 API 返回一致的结构，前端可以统一处理。

**Acceptance Criteria:**

**Given** 任意 API 请求成功
**When** 返回响应
**Then** 格式为 `{"code": 0, "message": "success", "data": {...}, "requestId": "req-xxx"}`

**Given** 业务逻辑返回错误
**When** handler 调用 `response.HandleError(c, err)`
**Then** 返回对应错误码和 HTTP 状态码
**And** 格式为 `{"code": 40201, "message": "...", "detail": "...", "requestId": "req-xxx"}`

**Given** handler 发生未捕获的 panic
**When** 全局错误中间件拦截
**Then** 返回 500 + 错误码 50001
**And** 记录 ERROR 级别日志含 stack trace

**Given** 请求进入 Gin 路由
**When** RequestID 中间件执行
**Then** 生成唯一 requestId 并注入 context，响应 header 包含 X-Request-ID

### Story 1.4: 结构化日志与脱敏引擎

As a 运维人员,
I want 结构化 JSON 日志输出和自动脱敏,
So that 日志可被日志系统采集且不泄露敏感信息。

**Acceptance Criteria:**

**Given** 服务运行中
**When** 任意日志输出
**Then** 格式为 JSON，包含 level、msg、time、requestId 字段

**Given** 日志内容包含密钥变量值
**When** 脱敏 Handler 处理
**Then** 密钥值被替换为 `***`

**Given** 管理员调用 admin API 调整日志级别
**When** 设置为 DEBUG
**Then** 运行时立即生效，无需重启
**And** `slog.LevelVar` 动态切换

### Story 1.5: Redis 连接与缓存基础层

As a 开发者,
I want Redis 连接池和基础缓存工具就绪,
So that 后续功能可以直接使用缓存能力。

**Acceptance Criteria:**

**Given** 服务启动且 Redis 可达
**When** 健康检查执行
**Then** `/readyz` 包含 Redis 连接状态

**Given** Redis 连接断开
**When** 服务尝试缓存操作
**Then** 操作返回错误，不 panic
**And** 业务逻辑降级到直接查库

### Story 1.6: 前端项目骨架

As a 前端开发者,
I want React+TS+Vite+Arco 前端项目骨架可运行,
So that 我可以开始开发页面。

**Acceptance Criteria:**

**Given** 开发者进入 web/ 目录
**When** 执行 `npm install && npm run dev`
**Then** Vite dev server 启动，浏览器访问 localhost:5173 显示空白布局页

**Given** 前端项目初始化
**When** 查看目录结构
**Then** 包含 pages/、components/common/、components/layout/、hooks/、stores/、lib/ws/、theme/、utils/、constants/、styles/ 目录
**And** 符合三层组件架构目录规范

**Given** Arco Design 主题配置
**When** 应用加载
**Then** 使用 Apple 风格蓝白色调 Token（primary #1677FF、bg #FFFFFF/#F7F8FA、大圆角）
**And** CSS Variables 响应式断点（768/1024/1280/1440px）已定义

**Given** 浏览器窗口宽度 < 1280px
**When** AppLayout 渲染
**Then** Sidebar 自动折叠为 icon-only 模式（64px）

### Story 1.7: 前端基础组件与设计系统集成

As a 前端开发者,
I want 基础公共组件和设计系统常量就绪,
So that 后续页面开发可以直接复用。

**Acceptance Criteria:**

**Given** 前端项目已初始化
**When** 导入 STATUS_MAP
**Then** 包含 success/running/failed/warning/pending/cancelled/timeout 七种状态的 color/bg/icon/label 定义

**Given** 页面组件渲染出错
**When** ErrorBoundary 捕获错误
**Then** 显示友好错误页面而非白屏
**And** 提供"重试"按钮

**Given** 页面首次加载
**When** 数据尚未返回
**Then** PageSkeleton 组件显示骨架屏占位

**Given** 用户无某资源权限
**When** PermissionGate 包裹的内容渲染
**Then** 该内容不显示

### Story 1.8: API 客户端自动生成流水线

As a 前端开发者,
I want 后端 API 变更后自动生成 TypeScript 客户端代码,
So that 前后端类型始终同步，不手写 API 调用。

**Acceptance Criteria:**

**Given** 后端执行 `make swag` 生成 OpenAPI v3 文档
**When** 前端执行 `npm run codegen`
**Then** @hey-api/openapi-ts 读取 OpenAPI spec 生成 services/generated/ 目录下的 TypeScript 客户端
**And** 包含请求函数、请求/响应类型定义

**Given** 后端新增或修改 API
**When** 重新执行 swag + codegen
**Then** 前端类型自动更新，旧类型不兼容处 TypeScript 编译报错

### Story 1.9: Helm Chart 基础骨架

As a 运维人员,
I want Helm Chart 骨架就绪,
So that 后续可以通过 `helm install` 部署到 K8s。

**Acceptance Criteria:**

**Given** charts/zcid/ 目录存在
**When** 执行 `helm template zcid ./charts/zcid`
**Then** 生成有效的 K8s Deployment、Service、ConfigMap 资源清单

**Given** values.yaml 参数化
**When** 修改 values.yaml 中的数据库连接、Redis 地址等
**Then** 渲染出的资源清单反映修改后的值

**Given** Helm Chart 就绪
**When** 执行 `helm lint ./charts/zcid`
**Then** 无 error 级别问题

## Epic 2: 用户认证与权限管理

用户可以登录平台、管理账号、分配角色，系统根据角色和项目归属控制资源访问权限。

### Story 2.1: 用户登录与 JWT 双 Token 认证

As a 用户,
I want 通过账号密码登录平台并获取认证凭证,
So that 我可以安全地访问平台功能。

**Acceptance Criteria:**

**Given** 用户提交正确的用户名和密码
**When** POST /api/v1/auth/login
**Then** 返回 Access Token（30min）和 Refresh Token（7天）
**And** Refresh Token 存储到 Redis

**Given** Access Token 过期
**When** POST /api/v1/auth/refresh 携带有效 Refresh Token
**Then** 返回新的 Access Token
**And** Refresh Token 不变

**Given** 用户登出
**When** POST /api/v1/auth/logout
**Then** Redis 中该用户的 Refresh Token 被删除
**And** 后续使用该 Refresh Token 刷新失败

**Given** 密码存储
**When** 用户注册或修改密码
**Then** 密码使用 bcrypt 哈希存储，不存明文

### Story 2.2: 用户账号管理

As a 管理员,
I want 创建、编辑、禁用用户账号,
So that 我可以管理平台的用户。

**Acceptance Criteria:**

**Given** 管理员已登录
**When** POST /api/v1/admin/users 提交用户信息
**Then** 创建新用户，密码 bcrypt 哈希存储

**Given** 管理员编辑用户
**When** PUT /api/v1/admin/users/:uid
**Then** 用户信息更新成功

**Given** 管理员禁用某用户
**When** 设置用户状态为 disabled
**Then** 该用户所有 Refresh Token 从 Redis 删除
**And** 该用户后续登录返回"账号已禁用"错误

**Given** 非管理员用户
**When** 尝试访问 /api/v1/admin/users
**Then** 返回 403 权限不足

### Story 2.3: 角色与权限管理

As a 管理员,
I want 为用户分配系统级角色并控制资源访问,
So that 不同角色的用户只能访问授权的功能。

**Acceptance Criteria:**

**Given** Casbin RBAC 模型已加载
**When** 管理员为用户分配角色（管理员/项目管理员/普通成员）
**Then** 策略写入 PostgreSQL 并通过 Redis Watcher 热更新

**Given** 用户请求受保护资源
**When** JWT 验证 + Casbin 鉴权中间件执行
**Then** 四元组 (sub, proj, obj, act) 匹配策略，通过则放行，否则返回 403

**Given** 密钥类型变量
**When** 普通成员查询变量列表
**Then** 密钥类型变量完全不可见（FR5）

### Story 2.4: 前端登录页与认证状态管理

As a 用户,
I want 一个简洁的登录页面和自动 Token 管理,
So that 我可以方便地登录和保持会话。

**Acceptance Criteria:**

**Given** 用户未登录
**When** 访问任意页面
**Then** 自动跳转到 /login 页面

**Given** 用户在登录页输入正确凭证
**When** 点击登录按钮
**Then** 按钮进入 loading 状态，登录成功后跳转到 /dashboard
**And** Token 存储到 authStore（Zustand）

**Given** Access Token 过期
**When** API 返回 401
**Then** Axios 拦截器自动使用 Refresh Token 刷新
**And** 刷新成功后重试原请求，刷新失败跳转登录页

### Story 2.5: 前端权限路由守卫

As a 用户,
I want 只看到我有权限的页面和操作,
So that 界面清晰且不会误操作。

**Acceptance Criteria:**

**Given** 用户已登录且角色为普通成员
**When** 路由渲染
**Then** 无权限的路由入口不显示（基于 Casbin 权限数据）

**Given** 用户直接访问无权限的 URL
**When** 路由守卫检查
**Then** 显示 403 无权限页面

**Given** PermissionGate 组件包裹操作按钮
**When** 用户无该操作权限
**Then** 按钮不渲染

### Story 2.6: 前端用户管理页面

As a 管理员,
I want 在界面上管理用户账号,
So that 不需要调用 API 即可创建、编辑、禁用用户。

**Acceptance Criteria:**

**Given** 管理员已登录
**When** 访问 /admin/users
**Then** 显示用户列表（用户名、角色、状态、创建时间）
**And** 非管理员访问返回 403

**Given** 管理员点击"新建用户"按钮
**When** 填写用户名、密码、角色并提交
**Then** 创建成功后刷新列表并显示成功提示

**Given** 管理员点击用户行的"编辑"按钮
**When** 修改用户信息并提交
**Then** 更新成功后刷新列表

**Given** 管理员点击用户行的"禁用"按钮
**When** 确认操作
**Then** 禁用后该用户无法登录

## Epic 3: 项目与资源管理

管理员可以创建项目，项目管理员可以管理环境、服务和成员，不同项目之间完全隔离。

### Story 3.1: 项目 CRUD

As a 管理员,
I want 创建和删除项目,
So that 团队可以按项目组织 CI/CD 资源。

**Acceptance Criteria:**

**Given** 管理员已登录
**When** POST /api/v1/projects 提交项目名称和描述
**Then** 项目创建成功，创建者自动成为项目管理员

**Given** 项目名称已存在
**When** 创建同名项目
**Then** 返回 40102 项目名重复错误

**Given** 管理员删除项目
**When** DELETE /api/v1/projects/:id
**Then** 项目及其关联的环境、服务、流水线、变量标记为删除

### Story 3.2: 环境管理与 Namespace 映射

As a 项目管理员,
I want 在项目内创建环境并映射到 K8s Namespace,
So that 不同环境（dev/staging/prod）隔离部署。

**Acceptance Criteria:**

**Given** 项目管理员已登录
**When** POST /api/v1/projects/:id/environments
**Then** 创建环境并关联 K8s Namespace

**Given** Namespace 已被其他环境占用
**When** 尝试映射
**Then** 返回 40302 Namespace 已占用错误

**Given** 编辑环境
**When** PUT /api/v1/projects/:id/environments/:eid
**Then** 环境信息更新成功

### Story 3.3: 服务管理

As a 项目管理员,
I want 在项目内管理服务,
So that 流水线和部署可以关联到具体服务。

**Acceptance Criteria:**

**Given** 项目管理员已登录
**When** POST /api/v1/projects/:id/services
**Then** 服务创建成功

**Given** 服务已关联部署
**When** 删除服务
**Then** 提示需先解除部署关联

### Story 3.4: 项目成员与角色管理

As a 项目管理员,
I want 将用户添加到项目并分配项目内角色,
So that 团队成员可以按角色协作。

**Acceptance Criteria:**

**Given** 项目管理员已登录
**When** POST /api/v1/projects/:id/members 添加用户
**Then** 用户加入项目并获得指定角色
**And** Casbin 策略通过 g2 三元组（user, role, project）写入

**Given** 修改成员角色
**When** PUT /api/v1/projects/:id/members/:uid
**Then** Casbin 策略热更新，权限立即生效

### Story 3.5: 前端项目管理页面

As a 用户,
I want 在界面上管理项目、环境、服务和成员,
So that 不需要调用 API 即可完成管理操作。

**Acceptance Criteria:**

**Given** 用户已登录
**When** 访问 /projects
**Then** 显示有权限的项目列表

**Given** 用户进入项目
**When** 访问 /projects/:id
**Then** 显示项目级布局（侧边栏 + 内容区 Outlet）
**And** 侧边栏包含流水线、环境、服务、变量、成员导航

**Given** 不同项目
**When** 用户在项目 A 中操作
**Then** 看不到项目 B 的任何资源（FR10 项目隔离）

## Epic 4: 变量与凭证管理

项目管理员可以在多层级管理变量和密钥，系统保证加密存储、运行时安全注入和日志脱敏。

### Story 4.1: 多层级变量 CRUD

As a 项目管理员,
I want 在全局、项目、流水线三个层级创建和管理变量,
So that 变量可以按层级覆盖，减少重复配置。

**Acceptance Criteria:**

**Given** 管理员已登录
**When** 创建全局变量 DB_HOST=global-db
**Then** 所有项目的流水线运行时可获取该变量

**Given** 项目管理员创建项目级变量 DB_HOST=project-db
**When** 该项目的流水线运行
**Then** DB_HOST 值为 project-db（项目级覆盖全局级）

**Given** 流水线级变量 DB_HOST=pipeline-db
**When** 该流水线运行
**Then** DB_HOST 值为 pipeline-db（流水线级覆盖项目级）

### Story 4.2: 密钥变量加密与安全

As a 项目管理员,
I want 创建加密存储的密钥变量,
So that 敏感信息安全存储且不可回显。

**Acceptance Criteria:**

**Given** 创建密钥类型变量
**When** 存储到数据库
**Then** 值使用 AES-256-GCM 加密，密钥来自环境变量 ZCID_ENCRYPTION_KEY

**Given** 查询变量列表
**When** 返回密钥类型变量
**Then** 值显示为 `******`，不可回显原文

**Given** 普通成员查询变量
**When** 列表返回
**Then** 密钥类型变量完全不可见（FR5）

### Story 4.3: 运行时密钥注入与清理

As a 系统,
I want 流水线运行时自动注入密钥并在结束后清理,
So that 密钥不持久暴露在集群中。

**Acceptance Criteria:**

**Given** 流水线运行触发
**When** executor 准备 Tekton PipelineRun
**Then** 密钥变量创建为临时 K8s Secret 注入 Pod

**Given** PipelineRun 结束（成功或失败）
**When** watcher 检测到终态
**Then** 临时 Secret 在 30 秒内自动清理（NFR11）

**Given** 构建日志输出包含密钥值
**When** 日志流经脱敏引擎
**Then** 密钥值替换为 `***`（FR14, NFR10）

### Story 4.4: 前端变量管理页面

As a 项目管理员,
I want 在界面上管理变量和密钥,
So that 不需要命令行即可配置变量。

**Acceptance Criteria:**

**Given** 项目管理员进入变量页面
**When** 访问 /projects/:id/variables
**Then** 显示变量列表，密钥类型值显示为 `******`

**Given** 创建新变量
**When** 选择类型为"密钥"并提交
**Then** 变量创建成功，值加密存储

**Given** 编辑普通变量
**When** 修改值并保存
**Then** 变量更新成功，页面显示 Message 成功提示（3s 自动消失）

## Epic 5: Git 仓库集成

管理员可以配置 Git 仓库连接，用户创建流水线时选择仓库和分支，系统接收 Webhook 并自动匹配触发流水线。

### Story 5.1: Git 仓库连接配置

As a 管理员,
I want 配置 GitLab 和 GitHub 仓库连接,
So that 平台可以访问代码仓库。

**Acceptance Criteria:**

**Given** 管理员已登录
**When** POST /api/v1/admin/integrations 配置 GitLab OAuth
**Then** 完成 OAuth 授权流程，连接信息加密存储

**Given** Git 连接已配置
**When** GET /api/v1/admin/integrations
**Then** 返回连接列表及状态（已连接/已断开）

**Given** OAuth Token 过期
**When** 系统尝试访问 Git API
**Then** 自动刷新 Token，刷新失败标记连接为断开

### Story 5.2: 仓库与分支选择

As a 用户,
I want 创建流水线时从已关联仓库中选择代码仓库和分支,
So that 不需要手动输入仓库地址。

**Acceptance Criteria:**

**Given** Git 连接已配置
**When** 用户在流水线创建页选择仓库
**Then** 下拉列表显示已关联的仓库（缓存 5min，支持手动刷新）

**Given** 用户选择了仓库
**When** 选择分支
**Then** 下拉列表显示该仓库的分支列表

### Story 5.3: Webhook 接收与自动触发

As a 系统,
I want 接收 Git Webhook 推送事件并自动匹配触发流水线,
So that 代码提交后自动启动构建。

**Acceptance Criteria:**

**Given** GitLab/GitHub 推送 Webhook 事件
**When** POST /api/v1/webhooks/gitlab 或 /github
**Then** 验证签名（NFR12），签名无效返回 401 并记录审计日志

**Given** Webhook 签名验证通过
**When** 系统匹配仓库、分支、事件类型
**Then** 自动触发对应流水线运行

**Given** 相同 Webhook 事件重复到达
**When** 幂等键（event_type:repo:commit_sha:timestamp_minute）已存在
**Then** 跳过触发，返回 200（FR19）

### Story 5.4: 前端 Git 集成管理页面

As a 管理员,
I want 在界面上配置和查看 Git 仓库连接,
So that 可以直观管理集成状态。

**Acceptance Criteria:**

**Given** 管理员访问集成页面
**When** 访问 /admin/integrations
**Then** 显示已配置的 Git 连接列表及状态

**Given** 管理员点击"添加 GitLab 连接"
**When** 完成 OAuth 授权流程
**Then** 连接添加成功，列表刷新显示新连接

**Given** 连接状态异常
**When** 页面渲染
**Then** StatusBadge 显示红色"已断开"状态

## Epic 6: 流水线可视化编排

用户可以通过可视化界面或模板创建流水线，高级用户可切换 YAML 模式，支持并发策略、触发规则、复制和参数覆盖。

### Story 6.1: 流水线 CRUD 与 JSONB 存储

As a 项目管理员,
I want 创建、编辑、删除流水线配置,
So that 我可以管理项目的 CI/CD 流程。

**Acceptance Criteria:**

**Given** 项目管理员已登录
**When** POST /api/v1/projects/:id/pipelines
**Then** 流水线配置以 JSONB 存储，包含 schemaVersion 字段

**Given** 流水线已存在
**When** PUT /api/v1/projects/:id/pipelines/:pid
**Then** 配置更新成功，高频字段（name/status）同步更新独立列

**Given** 项目管理员复制流水线
**When** POST /api/v1/projects/:id/pipelines/:pid/copy
**Then** 创建新流水线，配置复制自原流水线，名称加 "-copy" 后缀（FR29）

### Story 6.2: 模板一键创建流水线

As a 用户,
I want 从预置模板一键创建流水线,
So that 不需要从零配置即可快速开始。

**Acceptance Criteria:**

**Given** 用户进入创建流水线页面
**When** 选择模板（Go 微服务/Java Maven/前端 Node/通用 Docker）
**Then** 显示 TemplateSelector 组件，预览模板结构

**Given** 用户选择模板并填写少量参数（仓库、分支、镜像名）
**When** 点击创建
**Then** 基于模板 JSON 生成完整流水线配置并保存（FR21）

### Story 6.3: 可视化流水线编排器

As a 项目管理员,
I want 通过可视化界面拖拽编排流水线,
So that 直观地定义 Stage→Step 执行流程。

**Acceptance Criteria:**

**Given** 用户进入流水线编辑页
**When** PipelineEditor 加载
**Then** 使用 @xyflow/react 渲染 Stage（Group Node）和 Step（子 Node）
**And** dagre 自动布局，支持缩放、拖拽、选择、小地图

**Given** 用户添加新 Stage
**When** 拖拽到画布
**Then** StageNode 以 edit 模式渲染，可添加 Step

**Given** 用户点击 Step
**When** StepConfigPanel 打开
**Then** 显示 DynamicForm（@rjsf/core）根据 Step 类型的 JSON Schema 渲染配置表单

**Given** 用户编辑流水线
**When** 按 Cmd+S
**Then** 保存配置，按钮进入 loading 状态，成功后 Message 提示（UX-13）

### Story 6.4: YAML 模式编辑

As a 高级用户,
I want 切换到 YAML 模式直接编辑流水线配置,
So that 我可以精细控制配置细节。

**Acceptance Criteria:**

**Given** 用户在可视化编辑器中
**When** 点击 ModeSwitch 切换到 YAML 模式
**Then** Monaco Editor 通过 React.lazy 懒加载（UX-5）
**And** 加载期间显示 Spin + "加载编辑器..."

**Given** 用户 hover ModeSwitch 按钮
**When** 鼠标悬停
**Then** 触发 Monaco preload（hover 预加载策略）

**Given** YAML 编辑完成
**When** 切换回可视化模式
**Then** YAML 解析为配置对象，可视化编辑器同步更新

### Story 6.5: 并发控制与触发规则配置

As a 项目管理员,
I want 为流水线配置并发策略和 Webhook 触发规则,
So that 控制流水线的执行行为。

**Acceptance Criteria:**

**Given** 项目管理员编辑流水线设置
**When** 配置并发策略
**Then** 可选：排队等待 / 取消旧构建 / 拒绝并通知（FR26）

**Given** 配置 Webhook 触发规则
**When** 设置分支匹配模式和事件类型
**Then** 规则保存到流水线配置中（FR25）

### Story 6.6: 流水线列表与运行时参数

As a 用户,
I want 查看流水线列表并在触发时指定运行时参数,
So that 灵活控制每次运行。

**Acceptance Criteria:**

**Given** 用户访问流水线列表
**When** 访问 /projects/:id/pipelines
**Then** 显示流水线列表，每行包含 MiniStatusBar 显示最近运行状态色块

**Given** 用户手动触发流水线
**When** 点击"运行"按钮
**Then** 弹出参数确认面板，可临时覆盖运行时参数（FR30）
**And** 确认后按钮进入 loading 状态

## Epic 7: 流水线执行与构建

系统将流水线配置翻译为 Tekton CRD 并执行，支持容器化和传统两条构建链路，用户可以触发、取消运行，查看构建产物。

### Story 7.1: CRD 翻译引擎

As a 系统,
I want 将流水线 JSONB 配置翻译为 Tekton PipelineRun CRD,
So that 用户的可视化配置可以在 K8s 上执行。

**Acceptance Criteria:**

**Given** 流水线配置（JSONB）
**When** translator.Translate(config) 调用
**Then** 输出合法的 Tekton PipelineRun YAML
**And** 翻译延迟 < 5 秒（NFR2）

**Given** 翻译测试用例（testdata/）
**When** 运行 `go test ./pkg/tekton/...`
**Then** 4 种模板（Go/Java/Node/Docker）的输入→输出全部匹配期望

**Given** 配置中包含变量引用
**When** 翻译时
**Then** 四级变量合并后注入 CRD 的 params/env

### Story 7.2: 流水线运行编排与提交

As a 用户,
I want 触发流水线运行并提交到 K8s 执行,
So that 代码可以自动构建。

**Acceptance Criteria:**

**Given** 用户手动触发或 Webhook 触发
**When** executor 执行运行编排
**Then** 流程为：变量合并 → CRD 翻译 → 临时 Secret 创建 → K8s 提交 PipelineRun

**Given** 流水线运行时
**When** 自动注入 Git 信息
**Then** commit SHA、分支、提交者作为参数注入（FR28）

**Given** 并发策略为"排队等待"且已有运行中实例
**When** 新触发到达
**Then** 新运行进入 pending 状态排队

**Given** 并发策略为"取消旧构建"
**When** 新触发到达
**Then** 旧运行被取消，新运行立即开始

### Story 7.3: 取消运行与构建产物

As a 用户,
I want 取消运行中的流水线并查看构建产物,
So that 我可以中止错误的构建并确认产物。

**Acceptance Criteria:**

**Given** 流水线正在运行
**When** 用户点击取消
**Then** 系统删除对应 Tekton PipelineRun，状态更新为 cancelled（FR31）

**Given** 容器化构建成功
**When** 查看构建产物
**Then** 显示镜像地址 + Tag（FR35）

**Given** 传统构建成功
**When** 查看构建产物
**Then** 显示 MinIO 对象存储路径（FR35）

### Story 7.4: 容器化构建链路

As a 系统,
I want 支持完整的容器化构建链路,
So that 用户的代码可以构建为容器镜像并推送到仓库。

**Acceptance Criteria:**

**Given** 流水线配置为容器化构建
**When** PipelineRun 执行
**Then** 执行链路：代码拉取 → 编译构建 → Kaniko 镜像构建 → Harbor 推送（FR32）

**Given** Harbor 推送失败
**When** 第一次失败
**Then** 自动重试最多 3 次（NFR18），超时后标记失败并通知

### Story 7.5: 传统构建链路与镜像仓库配置

As a 系统,
I want 支持传统构建链路和镜像仓库管理,
So that 非容器化项目也可以使用平台。

**Acceptance Criteria:**

**Given** 流水线配置为传统构建
**When** PipelineRun 执行
**Then** 执行链路：代码拉取 → 编译打包 → 产物上传 MinIO（FR33）

**Given** 管理员配置镜像仓库
**When** POST /api/v1/admin/integrations/registry
**Then** Harbor 连接信息加密存储，连接测试通过（FR34）

## Epic 8: 实时日志与状态监控

用户可以实时查看构建日志和 Step 状态，失败时高亮错误，支持断线重连、断点续传和历史日志归档查看。

### Story 8.1: WebSocket 连接管理与消息协议

As a 系统,
I want WebSocket 连接管理器和统一消息协议,
So that 前端可以接收实时推送数据。

**Acceptance Criteria:**

**Given** 客户端连接 /ws/v1/logs/:runId
**When** WebSocket 握手成功
**Then** 连接注册到 Hub，开始推送日志消息

**Given** 服务端运行中
**When** 每 30 秒
**Then** 发送 heartbeat 消息，客户端 60 秒未收到视为断开

**Given** 客户端断线重连
**When** 携带 lastSeq 参数
**Then** 服务端从该序号后继续推送（断点续传）

**Given** 单用户 WebSocket 连接数达到 10
**When** 尝试新建连接
**Then** 拒绝并返回错误提示

### Story 8.2: 流水线状态实时监控

As a 用户,
I want 实时查看流水线每个 Step 的执行状态,
So that 我可以跟踪构建进度。

**Acceptance Criteria:**

**Given** PipelineRun 正在执行
**When** watcher 通过 K8s Informer 检测到状态变更
**Then** 通过 WebSocket /ws/v1/pipeline-status/:projectId 推送状态更新

**Given** 前端 usePipeline Hook 接收到状态更新
**When** 渲染运行详情页
**Then** StageNode 以 runtime 模式渲染，状态着色从 STATUS_MAP 读取（等待灰/运行蓝/成功绿/失败红）

**Given** 构建失败
**When** 某 Step 状态变为 failed
**Then** 该 StepNode 醒目标识失败状态（FR38）

### Story 8.3: 实时构建日志与 LogViewer

As a 用户,
I want 实时查看正在运行的构建日志,
So that 我可以监控构建过程和排查问题。

**Acceptance Criteria:**

**Given** 用户进入运行详情页
**When** LogViewer 组件加载
**Then** xterm.js 通过 React.lazy 懒加载，连接 WebSocket /ws/v1/logs/:runId

**Given** 构建正在运行
**When** 日志持续输出
**Then** xterm.js 实时渲染，scrollback buffer 50k 行（UX-3）

**Given** 构建失败
**When** 日志包含错误信息
**Then** 错误日志高亮显示（FR38）

**Given** WebSocket 连接断开
**When** ConnectionStatus 组件检测到
**Then** 显示黄色"重连中..."状态，自动重连并从 lastSeq 续传（FR39）

### Story 8.4: 日志归档与历史查看

As a 用户,
I want 查看已完成构建的历史日志,
So that 构建完成后仍可排查问题。

**Acceptance Criteria:**

**Given** PipelineRun 结束
**When** logger 检测到终态
**Then** 日志归档到 MinIO，按 1MB 分片存储（FR40）
**And** 归档成功率 > 99%（NFR16）

**Given** 用户查看历史运行日志
**When** useLogStream 切换到 archive 模式
**Then** 从 MinIO 分页加载归档日志，避免一次性拉取（FR41）

### Story 8.5: 运行历史列表

As a 用户,
I want 查看流水线的运行历史,
So that 我可以追踪每次运行的结果。

**Acceptance Criteria:**

**Given** 用户访问流水线详情
**When** 查看运行历史 Tab
**Then** 显示运行列表，包含状态、触发方式、触发人、时间（FR43）

**Given** 运行列表
**When** 渲染每行
**Then** StatusBadge 显示运行状态，颜色从 STATUS_MAP 读取

## Epic 9: 部署与环境管理

用户可以将构建产物部署到 K8s 环境，查看部署状态和同步详情，手动同步和回滚，部署权限分级控制。

### Story 9.1: ArgoCD 集成与部署触发

As a 系统,
I want 通过 ArgoCD gRPC API 管理部署,
So that 构建产物可以部署到 K8s 环境。

**Acceptance Criteria:**

**Given** 构建成功产出容器镜像
**When** 触发部署到指定环境
**Then** 系统创建或更新 ArgoCD Application，触发同步（FR44）
**And** 触发到开始同步 < 30 秒（NFR3）

**Given** ArgoCD Application 已存在
**When** 按 project_id:environment_id:service_id 查询
**Then** 更新镜像 Tag 并触发同步，不重复创建

### Story 9.2: 部署状态监控与同步操作

As a 用户,
I want 查看部署状态并手动触发同步,
So that 我可以了解环境健康状况并处理异常。

**Acceptance Criteria:**

**Given** 用户访问环境详情页
**When** useDeployStatus Hook 加载
**Then** 显示各服务部署状态（健康/同步中/异常），通过 WebSocket 实时更新（FR45）

**Given** 部署状态异常
**When** 用户查看同步详情
**Then** 显示 ArgoCD 同步错误信息（FR46）

**Given** 项目管理员点击"重新同步"
**When** 触发 ArgoCD Sync
**Then** Application 重新同步，状态实时更新（FR47）

### Story 9.3: 部署历史、回滚与权限控制

As a 项目管理员,
I want 查看部署历史并回滚到指定版本,
So that 部署出问题时可以快速恢复。

**Acceptance Criteria:**

**Given** 项目管理员查看部署历史
**When** 访问环境详情的历史 Tab
**Then** 显示部署版本列表，包含镜像 Tag、时间、操作人（FR48）

**Given** 项目管理员选择回滚版本
**When** 点击"回滚"并确认（ConfirmDialog）
**Then** ArgoCD 回滚到指定版本（FR48）

**Given** 普通成员尝试在 staging 环境触发部署
**When** 权限检查
**Then** 返回 403，只有 dev 环境允许普通成员操作（FR49）

## Epic 10: 通知、审计与平台运维

管理员可以配置通知规则和系统设置，系统记录审计日志，自动检测依赖健康状态和版本兼容性，清理过期资源。

### Story 10.1: 通知规则与 Webhook 发送

As a 项目管理员,
I want 为流水线配置通知规则,
So that 构建和部署结果自动通知相关人员。

**Acceptance Criteria:**

**Given** 项目管理员配置通知规则
**When** 设置触发条件（构建成功/失败、部署完成）和 Webhook URL
**Then** 规则保存成功（FR50）

**Given** 流水线运行结束且匹配通知规则
**When** Notifier 执行
**Then** 通过 HTTP POST 发送 Webhook 通知（FR51）

**Given** 相同事件重复触发通知
**When** 幂等键检查
**Then** 不重复发送

### Story 10.2: 审计日志记录与查询

As a 管理员,
I want 查看系统中所有写操作的审计日志,
So that 可以追踪谁在什么时间做了什么操作，满足合规要求。

**Acceptance Criteria:**

**Given** 用户执行写操作（创建/更新/删除）
**When** 请求通过 Gin 中间件
**Then** 系统异步记录审计日志，包含操作人、时间、操作类型、目标资源、结果（FR52）

**Given** 审计日志写入
**When** 使用结构化日志格式
**Then** 日志包含 user_id、action、resource_type、resource_id、result、ip、timestamp 字段，保留至少 90 天（NFR26）

**Given** 管理员查看审计日志
**When** 访问审计日志页面
**Then** 分页展示日志列表，支持按操作人、操作类型、时间范围筛选

**Given** 审计日志中涉及密钥操作
**When** 记录日志
**Then** 密钥值不出现在日志中，仅记录"密钥已更新"等脱敏描述（NFR10）

### Story 10.3: 系统设置与依赖健康检测

As a 管理员,
I want 配置系统级设置并查看外部依赖的健康状态,
So that 确保平台运行环境正确配置且关键依赖正常。

**Acceptance Criteria:**

**Given** 管理员访问系统设置
**When** 打开设置页面
**Then** 可配置 K8s 集群连接、全局变量、镜像仓库默认地址、通知渠道（FR57）

**Given** 管理员查看集成状态
**When** 访问集成面板
**Then** 显示已配置的 Git 仓库、镜像仓库、通知渠道的连接状态（在线/离线）（FR58）

**Given** 系统启动
**When** 初始化阶段
**Then** 自动检测 Tekton Pipeline API 版本（要求 v1 API，v0.44+）和 ArgoCD 版本兼容性，不兼容时记录 WARN 日志并在管理面板提示（FR59，NFR23，NFR24）

**Given** 关键依赖不可用
**When** K8s API Server / Tekton / ArgoCD 探针失败
**Then** 健康检查端点返回 degraded 状态，前端展示降级提示横幅（FR60，NFR17，NFR27）

**Given** 健康检查端点被访问
**When** GET /healthz
**Then** 返回数据库连接、Redis 连接、K8s API 可达性的逐项状态（NFR27）

**Given** 平台运行日志
**When** 任何日志输出
**Then** JSON 结构化格式，支持日志级别动态调整（NFR28）

### Story 10.4: CRD 资源清理与漂移告警

As a 管理员,
I want 系统自动清理过期的 CRD 资源并检测 ArgoCD 配置漂移,
So that 防止 etcd 数据膨胀，及时发现外部修改。

**Acceptance Criteria:**

**Given** PipelineRun 已完成且超过 TTL（默认 7 天）
**When** CRD Cleaner 定时任务执行
**Then** 自动删除过期的 PipelineRun/TaskRun CRD 资源（FR61，NFR20）

**Given** CRD 清理执行
**When** 删除资源
**Then** 记录清理日志（清理数量、耗时），失败时告警

**Given** ArgoCD Application 被外部直接修改（kubectl / ArgoCD UI）
**When** zcid 状态同步检测到配置与预期不一致
**Then** 在相关环境状态页面展示漂移告警信息（FR62）

**Given** 漂移告警展示
**When** 用户查看环境详情
**Then** 告警横幅显示"检测到外部修改，当前配置可能与 zcid 记录不一致"，提供"重新同步"操作

## Epic 11: 全局概览与用户引导

用户登录后看到项目总览与环境状态仪表盘，首次登录用户获得引导，主要列表支持筛选搜索。

### Story 11.1: 全局仪表盘

As a 用户,
I want 登录后看到我有权限的所有项目及其最近构建和环境状态概览,
So that 一目了然地了解整体情况，快速定位需要关注的问题。

**Acceptance Criteria:**

**Given** 用户登录成功
**When** 进入首页仪表盘
**Then** 展示用户有权限的项目卡片列表，每个卡片显示项目名称、最近一次构建状态、各环境健康摘要（FR53）

**Given** 技术管理者登录
**When** 查看全局视图
**Then** 可跨项目查看所有环境的健康状态汇总，包含 Healthy/Degraded/Progressing/Unknown 分类统计（FR54）

**Given** 仪表盘数据加载
**When** 首次渲染
**Then** 使用 TanStack Query 缓存策略，staleTime 30s，后台定时刷新（UX-10）

### Story 11.2: 首次登录引导与快捷入口

As a 首次登录用户,
I want 看到引导信息和快捷操作入口,
So that 快速了解平台并开始使用。

**Acceptance Criteria:**

**Given** 用户首次登录（无任何操作记录）
**When** 进入首页
**Then** 显示引导卡片：所属项目列表、"创建第一条流水线"快捷入口、文档链接（FR55）

**Given** 用户已有操作记录
**When** 进入首页
**Then** 不显示引导卡片，直接展示仪表盘

**Given** 引导卡片
**When** 用户点击"不再显示"
**Then** 记录偏好到 localStorage，后续不再展示

### Story 11.3: 全局列表筛选与搜索

As a 用户,
I want 在流水线列表、运行历史、服务列表等主要列表中进行筛选和搜索,
So that 快速找到目标资源。

**Acceptance Criteria:**

**Given** 用户在流水线列表页
**When** 使用筛选器
**Then** 支持按状态、分支、触发方式筛选（FR56）

**Given** 用户在运行历史页
**When** 输入搜索关键词
**Then** 支持按流水线名称、触发人、Commit SHA 搜索（FR56）

**Given** 用户在服务列表页
**When** 使用筛选器
**Then** 支持按环境、部署状态筛选（FR56）

**Given** 列表筛选条件
**When** 用户设置筛选参数
**Then** URL query params 同步更新，支持书签和分享（UX-10）

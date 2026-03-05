---
stepsCompleted: [1, 2, 3, 4, 5, 6, 7, 8]
inputDocuments:
  - 'files/planning-artifacts/prd.md'
  - 'files/planning-artifacts/product-brief-zcid-2026-03-01.md'
  - 'files/brainstorming/brainstorming-session-2026-03-01-001.md'
  - 'files/planning-artifacts/ux-design-specification.md'
workflowType: 'architecture'
project_name: 'zcid'
user_name: 'xjy'
date: '2026-03-01'
lastStep: 8
status: 'complete'
completedAt: '2026-03-01'
---

# Architecture Decision Document

_This document builds collaboratively through step-by-step discovery. Sections are appended as we work through each architectural decision together._

## Project Context Analysis

### Requirements Overview

**功能需求（62 条，11 个能力领域）：**

架构影响最大的 FR 集中在以下领域：

| 能力领域 | FR 数量 | 架构影响 |
|---------|---------|---------|
| 流水线编排与执行 | 12 | **最高** — Tekton CRD 翻译层是核心技术难点；三种模式（模板/可视化/YAML）需要统一内部数据模型作为基础；并发控制策略（FR26）需要流水线生命周期状态机 |
| 实时日志与状态监控 | 8 | **高** — WebSocket 连接管理、Pod 日志流、MinIO 归档、断点续传需专门的实时通信架构 |
| 部署与环境管理 | 6 | **高** — ArgoCD gRPC 集成层、Application 生命周期管理、状态同步 |
| 平台运维 | 6 | **高** — 健康检查、版本兼容检测、CRD 清理、外部修改检测需统一的后台任务调度框架 |
| 用户与权限管理 | 5 | **中高** — Casbin RBAC 贯穿所有 API，密钥变量可见性控制需前后端协同 |
| Git 仓库集成 | 5 | **中** — GitProvider 抽象、OAuth 流程、Webhook 接收/验证/去重/匹配 |
| 变量与凭证管理 | 4 | **中高** — 四级变量合并逻辑、AES-256-GCM 加解密、临时 Secret 生命周期 |
| 构建与产物管理 | 4 | **中** — 两条构建链路翻译为不同 Tekton TaskRun 结构 |
| 项目与资源管理 | 5 | **低** — 标准 CRUD + 项目隔离逻辑 |
| 通知与审计 | 3 | **低** — Notifier 接口 + Gin 中间件拦截 |
| 全局概览与引导 | 4 | **低** — 聚合查询 + 前端组件 |

**非功能需求（28 条，6 个质量维度）：**

驱动架构决策的关键 NFR：

- **性能：** API < 500ms、CRD 翻译 < 5s、WebSocket 延迟 < 2s — 要求低延迟的 K8s API 交互和高效的 WebSocket 管道
- **并发：** 50+ 用户、20+ 并发流水线、200+ WebSocket 连接 — 要求 goroutine 池管理和连接复用（fan-out 模式）
- **安全：** AES-256-GCM + 日志脱敏 + 临时 Secret 30s 清理 — 要求独立的安全层和凭证生命周期管理器
- **可靠性：** 99.5% 可用率 + 优雅降级 + 日志归档 99% — 要求健康检查、熔断、异步归档
- **集成：** 4 个接口抽象 + Tekton v1 兼容 + ArgoCD 版本兼容 — 要求版本检测和适配器模式
- **可扩展性：** 10 万+ 运行记录 + PipelineRun TTL 清理 — 要求分页优化和后台清理任务

### Core Data Flows

**CI 数据流（构建链路）：**
```
用户编排（可视化/模板/YAML）
  → 统一内部数据模型（流水线配置 JSONB）
    → 四级变量合并（全局→项目→流水线→运行时）
      → CRD 翻译引擎（输出 Tekton PipelineRun YAML）
        → K8s API 提交（Tekton Go Typed Client）
          → Tekton Pod 执行
            → Pod 日志流（client-go GetLogs+Follow）
              → 日志脱敏管道
                → WebSocket 推送至前端
          → PipelineRun 状态 Watch
            → 状态同步至 PostgreSQL + WebSocket 推送
```

**CD 数据流（部署链路）：**
```
构建产物（镜像 Tag / MinIO 路径）
  → ArgoCD Application 创建/更新（gRPC API）
    → ArgoCD 同步执行
      → 状态轮询（gRPC GetApplication）
        → 部署状态同步至 PostgreSQL + WebSocket 推送
```

### Scale & Complexity

- **主要领域：** 全栈 Web 平台 + Kubernetes 生态集成
- **复杂度等级：** 中高（技术集成驱动）
- **预估架构组件：** 约 15-18 个核心模块

**复杂度指标分析：**

| 维度 | 等级 | 说明 |
|------|------|------|
| 实时特性 | **高** | WebSocket 日志流 + 流水线状态推送 + 部署状态同步 |
| 隔离需求 | **中** | 项目逻辑隔离 + 环境 Namespace 物理隔离（非 SaaS 多租户） |
| 合规要求 | **低** | 无行业监管，安全需求集中在凭证和审计 |
| 集成复杂度 | **高** | 13 项外部集成，3 种 K8s 交互模式（Typed Client/gRPC/client-go） |
| 用户交互复杂度 | **高** | 可视化流水线编排器、实时日志查看器、三层用户体验 |
| 数据复杂度 | **中** | JSONB 存储流水线配置、运行记录增长快、日志归档到 MinIO |
| 前端架构复杂度 | **高** | 多 WebSocket 连接管理、可视化画布引擎、三模式双向同步 |

### Technical Constraints & Dependencies

**硬依赖：**
- Tekton Pipeline v0.44+（v1 API）— CRD 翻译层的基础
- ArgoCD — CD 引擎，gRPC API 交互
- Kubernetes 集群 — 运行时环境，in-cluster config（MVP）
- PostgreSQL — 主数据存储，Casbin 策略存储
- Redis — 缓存 + Streams 消息队列
- MinIO — 构建产物 + 日志归档

**已确定的技术选型（来自头脑风暴 42 项决策）：**
- 后端：Go + Gin + GORM + golang-migrate
- 前端：React + TypeScript + Arco Design + Zustand + TanStack Query
- API：RESTful + WebSocket + swaggo/swag
- 认证：JWT + Casbin RBAC
- 构建：Kaniko（P0）+ BuildKit（P1）
- 日志：Pod 直读（实时）+ MinIO 归档（历史）

**待决策的技术选型：**
- 流水线可视化编排器的前端实现方案（ReactFlow 等第三方库 vs 自研）

**实现约束：**
- 错误码段分配：400xx-406xx + 500xx（一旦发布即为 API 合同）
- JSONB schemaVersion 字段用于数据迁移
- golang-migrate SQL 驱动，启动时自动执行
- 单体架构，按业务模块组织代码

**架构级可测试性约束：**
- CRD 翻译层必须设计为可独立单元测试（输入流水线配置 → 输出 CRD YAML），不应只能通过集成测试（需要 Tekton 集群）验证。翻译正确率目标 > 99%，需要大量翻译用例的自动化回归

### Translation Layer Error Classification

CRD 翻译是系统核心，错误可能在三个层面发生，处理路径完全不同：

| 错误层面 | 原因 | 对用户的表现 | 处理策略 |
|---------|------|------------|---------|
| 翻译逻辑错误 | zcid 代码 bug，生成了无效 CRD | 流水线提交失败 | 系统错误，需开发修复；大量单元测试预防 |
| K8s API 提交失败 | 集群不可达、RBAC 权限不足、资源配额超限 | 流水线提交失败 | 平台错误，展示具体 K8s 错误信息 + 重试建议 |
| Tekton 执行失败 | 用户配置问题（Dockerfile 路径错误、构建命令失败等） | 流水线运行中某 Step 失败 | 用户错误，高亮失败 Step + 错误日志 |

### Background Task Architecture

以下后台任务需求需要统一的调度框架（建议基于 Redis Streams 或 goroutine 定时任务）：

| 后台任务 | 来源 | 触发方式 | 关键约束 |
|---------|------|---------|---------|
| PipelineRun 状态 Watch | FR42 | 持续运行 | 实时性要求高，Watch 断开需自动重连 |
| 日志归档到 MinIO | FR40 | 事件触发（PipelineRun 完成） | 归档成功率 > 99%，失败需重试 |
| 临时 Secret 清理 | NFR11 | 事件触发（PipelineRun 结束） | 30 秒内完成清理 |
| PipelineRun TTL 清理 | FR61 | 定时执行 | 防止 etcd 膨胀 |
| ArgoCD 外部修改检测 | FR62 | 定时轮询 | 检测到修改展示告警 |
| 健康检查 | FR60 | 定时执行 | 覆盖 K8s/Tekton/ArgoCD/DB/Redis |

### Cross-Cutting Concerns Identified

1. **认证与鉴权（Auth）：** JWT Token 验证 + Casbin RBAC 策略检查 — 贯穿所有 API 端点，Gin 中间件统一拦截
2. **审计日志（Audit）：** 所有写操作记录 — Gin 中间件拦截，独立审计表
3. **日志脱敏（Secret Masking）：** 三个独立日志来源需各自处理：
   - 构建日志（Pod 直出）→ WebSocket 转发管道中实施脱敏
   - 审计日志（Gin 中间件）→ 写入审计表前脱敏
   - 应用日志（zcid 自身）→ 结构化日志输出前脱敏
4. **错误处理（Error Handling）：** 分段错误码 + 统一 JSON 响应 — 全局错误中间件
5. **K8s 连接健康（K8s Health）：** API Server/Tekton/ArgoCD 可达性检测 — 健康检查端点 + 优雅降级逻辑
6. **项目隔离（Project Isolation）：** 数据查询默认加项目过滤 — 需要在数据层或中间件层统一处理
7. **变量合并（Variable Resolution）：** 全局→项目→流水线→运行时四级覆盖 — 运行时变量解析器
8. **幂等性（Idempotency）：** 不只限于 Webhook 去重（FR19），还包括 PipelineRun 提交、ArgoCD Application 创建/更新、通知发送 — 需要统一的幂等键设计模式
9. **前端实时数据管理（Frontend Real-time）：** React 应用需要统一的 WebSocket 管理层 — 连接池、自动重连、断点续传（sinceTime）、连接状态展示，覆盖日志流/流水线状态/部署状态三个通道
10. **流水线生命周期状态机（Pipeline Lifecycle）：** 流水线从配置→触发→排队→运行→完成/失败/取消的完整状态机，驱动并发控制策略（排队等待/取消旧构建/拒绝并通知）

## Starter Template Evaluation

### Evaluation Summary

评估了 6 个后端 Starter 和 4 个前端 Starter，结论：**无现成 Starter 适用，采用自定义初始化**。

**评估的后端 Starter：**

| Starter | 不适用原因 |
|---------|-----------|
| go-clean-template | DDD 过度分层，不适合管理平台单体 |
| go-blueprint | 脚手架工具，生成的结构缺少 K8s 集成层 |
| gin-boilerplate 类 | 标准 CRUD 脚手架，缺少 Tekton/ArgoCD/WebSocket 集成 |
| kratos | 微服务框架，与单体架构定位冲突 |
| go-zero | 微服务框架，引入不必要的复杂度 |
| fiber-boilerplate | 非 Gin 框架 |

**评估的前端 Starter：**

| Starter | 不适用原因 |
|---------|-----------|
| Arco Design Pro (React) | 最接近但内置 Redux，需替换为 Zustand；缺少 WebSocket 管理层和可视化画布 |
| Vite React TS 模板 | 太基础，缺少组件库和状态管理 |
| Create React App | 已过时，Vite 是当前标准 |
| Next.js 模板 | SSR 框架，zcid 是 SPA，不需要 SSR |

**不适用的核心原因：** zcid 的核心复杂度在 K8s 生态集成（Tekton Typed Client、ArgoCD gRPC、client-go Pod 日志）和实时通信（WebSocket fan-out），这些不是通用 Web Starter 覆盖的范围。使用通用 Starter 反而需要大量删改和重构。

### Selected Approach: Custom Initialization

**后端初始化：**
```bash
mkdir zcid && cd zcid
go mod init github.com/xjy/zcid
# 核心依赖
go get github.com/gin-gonic/gin@v1.12.0
go get gorm.io/gorm@v1.30.1
go get gorm.io/driver/postgres
go get github.com/casbin/casbin/v2
go get github.com/golang-jwt/jwt/v5
go get github.com/redis/go-redis/v9
go get github.com/minio/minio-go/v7
# Tekton + ArgoCD + K8s
go get github.com/tektoncd/pipeline/pkg/client
go get k8s.io/client-go
# API 文档
go get github.com/swaggo/swag/v2
go get github.com/swaggo/gin-swagger
# 数据库迁移
go get github.com/golang-migrate/migrate/v4
```

**前端初始化：**
```bash
npm create vite@latest web -- --template react-ts
cd web
npm install @arco-design/web-react
npm install zustand @tanstack/react-query
npm install react-router-dom
npm install @hey-api/openapi-ts    # OpenAPI v3 代码生成
npm install react-i18next i18next  # 国际化预留
```

### Verified Dependency Versions

| 依赖 | 版本 | 验证日期 | 备注 |
|------|------|---------|------|
| Go | 1.24+ | 2026-03-01 | 当前稳定版 |
| Gin | v1.12.0 | 2026-02-28 发布 | 最新稳定版 |
| GORM | v1.30.1 | 2026-03-01 | Go Generics 支持 |
| golang-migrate | v4.19.1 | 2026-03-01 | SQL 文件驱动 |
| swaggo/swag | **v2.0.0** | 2026-03-01 | **OpenAPI v3**（非 v1 的 Swagger 2.0） |
| React | 19 | 2026-03-01 | 当前稳定版 |
| Vite | 6 | 2026-03-01 | 当前稳定版 |
| Arco Design | v2.66.10 | 2026-03-01 | 最新稳定版 |
| TypeScript | 5.7+ | 2026-03-01 | 当前稳定版 |

### Key Decision: swaggo/swag v2.0.0 (OpenAPI v3)

头脑风暴中选定了 swaggo/swag 但未指定版本。经评估，选择 **v2.0.0**：

- v1.x 生成 Swagger 2.0 规范（已过时）
- v2.0.0 生成 OpenAPI v3 规范（行业标准）
- @hey-api/openapi-ts 前端代码生成器原生支持 OpenAPI v3
- 一次到位，避免后续从 v1 迁移到 v2 的成本

### Key Decision: @hey-api/openapi-ts

前端 API 代码生成工具选择 **@hey-api/openapi-ts**：

- 原生支持 OpenAPI v3（与 swag v2 配合）
- 生成 TypeScript 类型安全的 API 客户端
- 支持 TanStack Query 集成
- 活跃维护，社区认可度高

### Backend Code Organization Convention

采用 **handler→service→repo** 三层结构，按业务模块组织，公共基础层 `pkg/`：

```
zcid/
├── cmd/server/main.go              # 入口：初始化 DB/Redis/MinIO/K8s Client，注册路由
├── internal/                        # 业务模块（私有）
│   ├── pipeline/                    # 流水线模块
│   │   ├── handler.go               # HTTP handler（Gin handler func）
│   │   ├── service.go               # 业务逻辑
│   │   ├── repo.go                  # 数据访问（GORM）
│   │   ├── model.go                 # 数据库模型（GORM struct）
│   │   └── dto.go                   # 请求/响应 DTO
│   ├── project/                     # 项目模块（同结构）
│   ├── environment/                 # 环境模块
│   ├── service/                     # 服务模块
│   ├── registry/                    # 镜像仓库模块
│   ├── auth/                        # 认证权限模块
│   ├── variable/                    # 变量与凭证模块
│   ├── git/                         # Git 集成模块
│   ├── notification/                # 通知模块
│   ├── audit/                       # 审计模块
│   └── dashboard/                   # 概览模块
├── pkg/                             # 公共基础层（可被外部引用）
│   ├── tekton/                      # Tekton Go Typed Client 封装
│   │   ├── client.go                # 客户端初始化
│   │   └── translator.go           # CRD 翻译引擎（核心）
│   ├── argocd/                      # ArgoCD gRPC Client 封装
│   ├── k8s/                         # client-go 封装（Pod 日志、Secret 操作）
│   ├── crypto/                      # AES-256-GCM 加解密
│   ├── ws/                          # WebSocket fan-out 管理器
│   ├── middleware/                   # Gin 中间件
│   │   ├── auth.go                  # JWT 验证 + Casbin 鉴权
│   │   ├── audit.go                 # 审计日志拦截
│   │   ├── error.go                 # 全局错误处理
│   │   └── project_scope.go         # 项目隔离过滤
│   ├── response/                    # 统一响应格式 + 错误码定义
│   └── masking/                     # 日志脱敏引擎
├── migrations/                      # golang-migrate SQL 文件
│   ├── 000001_init_schema.up.sql
│   └── 000001_init_schema.down.sql
├── docs/                            # swag v2 生成的 OpenAPI v3 文档
├── Makefile
├── Dockerfile
└── docker-compose.yml
```

**命名约定：**
- 模块目录名：单数形式（`pipeline` 非 `pipelines`）
- 文件名：`handler.go`、`service.go`、`repo.go`、`model.go`、`dto.go`
- handler 层只做参数绑定和响应，不含业务逻辑
- service 层编排业务逻辑，调用 repo 和 pkg
- repo 层只做数据访问，不含业务判断

### Frontend Code Organization

```
web/
├── src/
│   ├── pages/                       # 页面组件（按路由组织）
│   ├── components/
│   │   └── pipeline-editor/         # 流水线可视化编排器（核心组件）
│   ├── hooks/                       # 自定义 React Hooks
│   ├── services/                    # @hey-api/openapi-ts 生成的 API 客户端
│   ├── stores/                      # Zustand 状态管理
│   ├── lib/
│   │   └── ws/                      # WebSocket 统一管理层（连接池/重连/断点续传）
│   ├── theme/
│   │   └── tokens.ts                # Arco Design 主题定制
│   └── utils/                       # 工具函数
├── vite.config.ts
├── vitest.config.ts
└── openapi-ts.config.ts             # API 代码生成配置
```

### Arco Design Theme Tokens (Apple 风格蓝白色调)

```typescript
export const zcidTheme = {
  // 主色 — Apple 蓝
  '--color-primary-6': '#1677FF',
  '--color-primary-5': '#4096FF',
  '--color-primary-7': '#0958D9',
  // 背景 — 干净白底
  '--color-bg-1': '#FFFFFF',
  '--color-bg-2': '#F7F8FA',
  // 圆角 — Apple 风格大圆角
  '--border-radius-small': '6px',
  '--border-radius-medium': '8px',
  '--border-radius-large': '12px',
};
```

### Development Environment: docker-compose.yml

```yaml
services:
  postgres:
    image: postgres:16
    ports: ["5432:5432"]
    environment:
      POSTGRES_DB: zcid
      POSTGRES_USER: zcid
      POSTGRES_PASSWORD: zcid_dev
    volumes:
      - pgdata:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    ports: ["6379:6379"]

  minio:
    image: minio/minio
    ports: ["9000:9000", "9001:9001"]
    command: server /data --console-address ":9001"
    environment:
      MINIO_ROOT_USER: zcid
      MINIO_ROOT_PASSWORD: zcid_dev_key
    volumes:
      - miniodata:/data

volumes:
  pgdata:
  miniodata:
```

### Testing Frameworks

**后端测试：**
- **testify** — 断言和 Mock 框架，Go 社区标准
- **testcontainers-go** — 集成测试使用真实 PostgreSQL/Redis/MinIO 容器，避免 Mock 与真实行为不一致
- CRD 翻译层使用纯单元测试（输入 JSONB 配置 → 输出 Tekton CRD YAML），不依赖 K8s 集群

**前端测试：**
- **vitest** — Vite 原生测试框架，与构建工具零配置集成
- **@testing-library/react** — 组件测试，以用户行为驱动
- **msw (Mock Service Worker)** — API Mock，拦截网络请求而非替换模块

## Core Architectural Decisions

### Decision Priority Analysis

**Critical Decisions (Block Implementation):**
- JSONB 存储策略（胖 JSONB + 独立列索引）
- JWT 双 Token 认证策略
- Casbin RBAC 四元组模型
- API 版本策略（URL 前缀 `/api/v1`）
- WebSocket 消息协议（统一格式 + seq 断点续传）
- 流水线可视化编排器（@xyflow/react）
- 错误码段分配

**Important Decisions (Shape Architecture):**
- 数据验证双层策略
- AES-256-GCM 密钥环境变量管理
- API 安全（敏感端点限流）
- 前端路由策略
- 结构化日志（slog）
- 应用配置管理（YAML + 环境变量覆盖）
- Helm Chart 部署
- 健康检查三级端点

**Deferred Decisions (Post-MVP):**
- 多集群配置管理
- OAuth 集成（PaaS 嵌入场景）
- ArgoCD 自动同步策略
- 多语言支持（i18n 结构已预留）

### Data Architecture

**JSONB 存储策略：胖 JSONB + 独立列索引**
- 流水线配置整体存为一个 JSONB 字段，包含 `schemaVersion` 用于数据迁移
- 高频查询字段（name、status、project_id、trigger_type、created_at）保持独立列 + B-tree 索引
- JSONB 只做存储，不做 GIN 索引查询
- 理由：管理平台 QPS 不高，独立列索引覆盖所有列表查询场景，避免 JSONB 查询的复杂性和性能不确定性

**数据验证：双层验证**
- Handler 层：Gin binding tag 做格式校验（required、类型、长度、正则）
- Service 层：业务规则校验（名称唯一、项目存在性、权限检查、状态合法性）
- 理由：关注点分离，handler 拦截格式错误返回 400，service 拦截业务错误返回具体业务错误码

**Redis 缓存策略：短 TTL 为主，减少外部调用**

| 缓存对象 | TTL | 失效方式 | 用途 |
|---------|-----|---------|------|
| Casbin 策略 | 内置缓存 | Redis Watcher 热更新 | 避免每次请求查库 |
| 用户 Session 信息 | 30min（随 Access Token） | Token 刷新时更新 | JWT payload 缓存 |
| 健康检查结果 | 30s | 自动过期 | 减少探测频率 |
| Git 仓库/分支列表 | 5min | 手动刷新按钮清除 | 减少 Git API 调用 |

**GORM 连接池：**
- MaxOpenConns: 25 / MaxIdleConns: 10 / ConnMaxLifetime: 30min / ConnMaxIdleTime: 5min

### Authentication & Security

**JWT 双 Token 认证：**
- Access Token TTL: 30 分钟（短过期，泄露影响窗口小）
- Refresh Token TTL: 7 天（存 Redis，支持主动吊销）
- 用户登出：删除 Redis 中的 Refresh Token
- 管理员禁用账号：删除该用户所有 Refresh Token，即时生效
- Token 签名算法：HS256（单体应用足够，无需 RS256 的公钥分发）

**Casbin RBAC 模型：**

```ini
[request_definition]
r = sub, proj, obj, act

[policy_definition]
p = sub, proj, obj, act

[role_definition]
g = _, _
g2 = _, _, _

[policy_effect]
e = some(where (p.eft == allow))

[matchers]
m = g(r.sub, p.sub) || \
    (g2(r.sub, p.sub, r.proj) && r.proj == p.proj && r.obj == p.obj && r.act == p.act)
```

- `g`：系统角色继承（admin 继承所有权限）
- `g2`：项目角色继承（user, role, project 三元组）
- 请求四元组：(用户ID, 项目ID, 资源类型, 操作)
- 策略存储：PostgreSQL（casbin GORM adapter）
- 策略热更新：Redis Watcher

**AES-256-GCM 密钥管理：**
- 密钥来源：环境变量 `ZCID_ENCRYPTION_KEY`（K8s Secret 注入）
- 密钥长度：32 bytes，启动时校验，不合法 panic
- 密钥轮换：需重启服务（MVP 可接受）
- 日志中永远不打印密钥值

**API 安全：**
- CORS：开发环境允许 `localhost:*`，生产环境只允许配置的前端域名
- Rate Limiting：仅敏感端点
  - `/api/v1/auth/login` — 10 次/分钟/IP
  - `/api/v1/webhooks/*` — 100 次/分钟/来源 IP
- 实现：Gin 中间件 + Redis 滑动窗口计数器

### API & Communication Patterns

**API 版本策略：URL 前缀**
- REST API：`/api/v1/...`
- WebSocket：`/ws/v1/...`
- 版本升级时并行存在（`/api/v2/...`）

**REST 路由结构：**

```
/api/v1/auth/login              POST
/api/v1/auth/refresh            POST
/api/v1/auth/logout             POST
/api/v1/projects                GET/POST
/api/v1/projects/:id            GET/PUT/DELETE
/api/v1/projects/:id/pipelines  GET/POST
/api/v1/projects/:id/pipelines/:pid          GET/PUT/DELETE
/api/v1/projects/:id/pipelines/:pid/runs     GET/POST
/api/v1/projects/:id/pipelines/:pid/runs/:rid GET
/api/v1/projects/:id/environments            GET/POST
/api/v1/projects/:id/environments/:eid       GET/PUT/DELETE
/api/v1/projects/:id/services                GET/POST
/api/v1/projects/:id/variables               GET/POST
/api/v1/projects/:id/members                 GET/POST
/api/v1/admin/users             GET/POST
/api/v1/admin/users/:uid        GET/PUT
/api/v1/admin/settings          GET/PUT
/api/v1/admin/integrations      GET
/api/v1/webhooks/gitlab         POST
/api/v1/webhooks/github         POST
```

**WebSocket 路由：**

```
/ws/v1/logs/:pipelineRunId         # 实时构建日志
/ws/v1/pipeline-status/:projectId  # 流水线状态变更
/ws/v1/deploy-status/:environmentId # 部署状态变更
```

**WebSocket 消息协议：**

```json
{
  "type": "log|status|deploy|heartbeat|error",
  "payload": { ... },
  "timestamp": "RFC3339",
  "seq": 12345
}
```

- 心跳：服务端每 30 秒发送 `{"type":"heartbeat"}`
- 断线检测：客户端 60 秒未收到心跳视为断开
- 断点续传：重连时发送 `lastSeq`，服务端从该序号后继续推送
- 空闲超时：10 分钟无消息服务端主动关闭
- 连接限制：单用户最多 10 个 WebSocket 连接

**错误码段分配：**

| 码段 | 模块 | 示例 |
|------|------|------|
| 40001-40099 | 认证与鉴权 | 40001 Token 过期、40002 权限不足、40003 账号已禁用 |
| 40101-40199 | 项目管理 | 40101 项目不存在、40102 项目名重复 |
| 40201-40299 | 流水线 | 40201 流水线不存在、40202 并发数超限、40203 CRD 翻译失败 |
| 40301-40399 | 环境与部署 | 40301 环境不存在、40302 Namespace 已占用 |
| 40401-40499 | Git 集成 | 40401 OAuth 授权失败、40402 Webhook 签名验证失败 |
| 40501-40599 | 变量与凭证 | 40501 变量名重复、40502 密钥解密失败 |
| 40601-40699 | 构建与产物 | 40601 镜像仓库连接失败、40602 产物不存在 |
| 50001-50099 | 系统内部错误 | 50001 数据库错误、50002 Redis 不可达、50003 K8s API 不可达 |

**统一响应格式：**

```json
{
  "code": 0,
  "message": "success",
  "data": { ... },
  "requestId": "req-xxxx"
}
```

错误响应：

```json
{
  "code": 40201,
  "message": "流水线并发数超限",
  "detail": "当前流水线已有 3 个运行中的实例，并发策略为排队等待",
  "requestId": "req-xxxx"
}
```

### Frontend Architecture

**流水线可视化编排器：@xyflow/react v12**
- 版本：@xyflow/react v12.10.1（MIT 开源）
- Stage → Group Node，Step → 子 Node
- 连线表示 Stage 执行顺序（串行/并行）
- 自定义 Node 组件渲染 Step 配置面板
- dagre 布局算法用于自动排列
- 内置缩放、拖拽、选择、小地图
- 运行时状态着色（等待灰/运行中蓝/成功绿/失败红）

**前端路由：React Router v7**

```
/login
/dashboard
/projects
/projects/:projectId
/projects/:projectId/pipelines
/projects/:projectId/pipelines/:pipelineId
/projects/:projectId/pipelines/:pipelineId/runs/:runId
/projects/:projectId/environments
/projects/:projectId/environments/:envId
/projects/:projectId/services
/projects/:projectId/variables
/admin/users
/admin/settings
/admin/integrations
```

- React Router Outlet 实现项目级布局（侧边栏 + 内容区）
- 路由守卫基于 Casbin 权限数据，无权限路由不渲染入口

**前端配置：Vite env**
- `.env.development` / `.env.production` 分环境
- `VITE_API_BASE_URL` / `VITE_WS_BASE_URL` / `VITE_APP_TITLE`
- 生产部署通过 Docker 构建参数或 `window.__ENV__` 运行时注入

### Infrastructure & Deployment

**结构化日志：Go slog（标准库）**
- JSON Handler 输出结构化日志
- `slog.LevelVar` 支持运行时动态调整日志级别（通过 admin API）
- 日志脱敏引擎作为 slog Handler 包装层
- 与 Gin 集成：自定义 Logger 中间件

**应用配置：YAML 文件 + 环境变量覆盖**
- 开发环境：`config.yaml` 文件（直观可读）
- 生产环境：环境变量优先覆盖（K8s ConfigMap/Secret 注入）
- 解析库：`gopkg.in/yaml.v3`（Go 标准 YAML 库，不引入 Viper）
- 敏感配置（密码、密钥、Token）全部通过环境变量注入，YAML 中留空

**部署方式：Helm Chart**
- `charts/zcid/` 目录结构：Chart.yaml + values.yaml + templates/
- values.yaml 参数化所有外部连接
- 支持 `helm install zcid ./charts/zcid` 一键安装
- MVP 阶段保持简单 Chart 结构

**健康检查三级端点：**

| 端点 | 用途 | 检查内容 | 超时 |
|------|------|---------|------|
| `GET /healthz` | K8s liveness probe | 进程存活 | - |
| `GET /readyz` | K8s readiness probe | DB + Redis 连接 | 3s |
| `GET /api/v1/health` | 管理界面详细报告 | DB/Redis/MinIO/K8s/Tekton/ArgoCD | 3s/项 |

状态枚举：`healthy` / `degraded`（部分不可用）/ `unhealthy`（核心不可用）

### Decision Impact Analysis

**Implementation Sequence:**
1. 应用配置管理（config.yaml + env）— 所有模块的基础
2. 结构化日志（slog）+ 日志脱敏 Handler — 开发调试基础
3. PostgreSQL 连接 + GORM + golang-migrate — 数据层基础
4. Redis 连接 + 缓存层 — 认证和策略缓存依赖
5. JWT 双 Token 认证 + Casbin RBAC — 所有 API 的前置依赖
6. 统一错误处理 + 响应格式 — API 层基础
7. REST API 路由骨架 — 业务模块开发基础
8. WebSocket 消息协议 + 连接管理 — 实时功能基础
9. 健康检查端点 — 部署验证基础
10. @xyflow/react 编排器 — 核心 UI 功能
11. Helm Chart — 部署交付

**Cross-Component Dependencies:**
- JWT 认证 ↔ Redis（Refresh Token 存储）
- Casbin RBAC ↔ PostgreSQL（策略存储）↔ Redis（Watcher 热更新）
- WebSocket 消息协议 ↔ 日志脱敏引擎（构建日志推送前脱敏）
- 错误码 ↔ 统一响应格式 ↔ 前端错误处理
- 健康检查 ↔ Redis 缓存（结果缓存 30s）
- @xyflow/react ↔ REST API（流水线配置 CRUD）↔ JSONB 存储

## Implementation Patterns & Consistency Rules

### Naming Patterns

**数据库命名：**
- 表名：snake_case 复数 — `users`、`projects`、`pipelines`、`pipeline_runs`
- 列名：snake_case — `project_id`、`created_at`、`pipeline_config`
- 外键：`{关联表单数}_id` — `project_id`、`user_id`、`pipeline_id`
- 索引：`idx_{表名}_{列名}` — `idx_pipelines_project_id`、`idx_pipeline_runs_status`
- 唯一约束：`uk_{表名}_{列名}` — `uk_users_username`、`uk_projects_name`
- GORM 模型自动映射：Go struct `PipelineRun` → 表 `pipeline_runs`（GORM 默认行为）

**API JSON 字段命名：**
- 请求/响应 Body：camelCase — `projectId`、`pipelineConfig`、`createdAt`
- Go DTO struct tag：`json:"projectId"`
- 理由：前端 JavaScript/TypeScript 惯例是 camelCase，后端 DTO 层做转换

**Go 代码命名：**
- struct/interface：PascalCase — `PipelineService`、`GitProvider`
- 公开方法：PascalCase — `CreatePipeline`、`GetByID`
- 私有方法/变量：camelCase — `buildCRD`、`pipelineConfig`
- 常量：错误码用 `ErrPipelineNotFound`，配置常量用 `DefaultPageSize`
- 包名：全小写单词 — `pipeline`、`middleware`、`response`
- 接口命名：行为动词后缀 -er 或描述性名词 — `GitProvider`、`Notifier`、`Translator`
- 错误变量：`Err` 前缀 — `ErrNotFound`、`ErrUnauthorized`

**React/TypeScript 命名：**
- 组件文件：PascalCase.tsx — `PipelineEditor.tsx`、`StepConfigPanel.tsx`
- hook 文件：useCamelCase.ts — `usePipelineStatus.ts`、`useWebSocket.ts`
- 工具文件：camelCase.ts — `formatDate.ts`、`maskSecret.ts`
- 目录名：kebab-case — `pipeline-editor/`、`log-viewer/`
- 组件命名：PascalCase — `<PipelineEditor />`
- 变量/函数：camelCase — `pipelineData`、`handleSubmit`
- 类型/接口：PascalCase + 描述性后缀 — `PipelineConfig`、`CreatePipelineRequest`、`PipelineListResponse`
- Zustand store：`use{Domain}Store` — `usePipelineStore`、`useAuthStore`

### Structure Patterns

**测试文件位置：**
- 后端：测试文件与源文件同目录 — `internal/pipeline/service_test.go`（Go 标准实践）
- 前端：测试文件与源文件同目录 — `PipelineEditor.test.tsx`（与组件共存）
- CRD 翻译测试用例：`pkg/tekton/testdata/` 目录存放 JSON 输入和期望的 YAML 输出
- 集成测试：`tests/integration/` 顶层目录（testcontainers-go 测试）

**Import 排序（Go）：**
```go
import (
    // 标准库
    "context"
    "fmt"

    // 第三方库
    "github.com/gin-gonic/gin"
    "gorm.io/gorm"

    // 项目内部
    "github.com/xjy/zcid/internal/pipeline"
    "github.com/xjy/zcid/pkg/response"
)
```

**Import 排序（TypeScript）：**
```typescript
// 第三方库
import { useQuery } from '@tanstack/react-query';
import { Button } from '@arco-design/web-react';

// 项目内部
import { usePipelineStore } from '@/stores/pipelineStore';
import { PipelineConfig } from '@/types';

// 相对路径（同模块）
import { StepNode } from './StepNode';
```

### Format Patterns

**日期时间格式：**
- 数据库：PostgreSQL `timestamptz`（UTC 存储）
- API JSON：RFC3339 — `"2026-03-01T12:00:00Z"`
- Go 内部：`time.Time`
- 前端显示：本地时区格式化（dayjs，Arco Design 内置依赖）

**分页格式：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "items": [],
    "total": 100,
    "page": 1,
    "pageSize": 20
  }
}
```
- 请求参数：`?page=1&pageSize=20`（camelCase）
- 默认值：page=1, pageSize=20
- 最大 pageSize：100

**Null 处理：**
- Go→JSON：`omitempty` tag，零值字段不输出
- 数据库 NULL：Go 中用指针类型（`*string`、`*int64`）
- 前端：`undefined` 表示未加载，`null` 表示明确为空
- API 响应中不返回 `null` 字段，直接省略（`omitempty`）

### Communication Patterns

**后端错误传播链：**

```
repo 层 → 返回 error（GORM 原始错误或自定义错误）
  ↓
service 层 → 包装为业务错误（带错误码）
  ↓
handler 层 → 调用 response.HandleError(c, err) 统一输出
  ↓
全局错误中间件 → 兜底未捕获的 panic
```

- repo 层：不翻译错误，直接返回 GORM error
- service 层：判断错误类型，翻译为业务错误码（`response.NewBizError(40201, "流水线不存在")`）
- handler 层：不做错误判断，统一调用 `response.HandleError(c, err)`
- 业务错误类型定义在 `pkg/response/errors.go`

**日志规范：**

| 级别 | 使用场景 | 示例 |
|------|---------|------|
| ERROR | 需要人工介入的异常 | 数据库连接失败、K8s API 不可达、CRD 翻译 panic |
| WARN | 可自动恢复的异常 | Webhook 签名验证失败、缓存未命中回退查库、重试中 |
| INFO | 关键业务事件 | 用户登录、流水线创建、PipelineRun 提交、部署触发 |
| DEBUG | 开发调试信息 | 请求参数、SQL 语句、CRD 翻译中间结果 |

- 日志必须包含 `requestId`（从 Gin context 获取）
- 日志中禁止打印：密码、Token、加密密钥、凭证明文
- 结构化字段规范：`slog.String("projectId", id)`, `slog.Int("statusCode", code)`

**前端状态管理分层：**

| 层 | 工具 | 职责 | 示例 |
|-----|------|------|------|
| 服务端数据 | TanStack Query | API 数据获取、缓存、乐观更新 | 流水线列表、项目详情 |
| 客户端 UI 状态 | Zustand | 纯前端状态 | 侧边栏展开/折叠、编辑器模式切换 |
| 实时数据 | WebSocket Manager | 服务器推送数据 | 构建日志、运行状态 |
| URL 状态 | React Router | 路由参数 | 当前项目 ID、当前流水线 ID |
| 表单状态 | 组件内 state | 表单输入 | Step 配置表单 |

规则：不在 Zustand 中缓存 API 数据（TanStack Query 的职责），不用 TanStack Query 管理 UI 状态（Zustand 的职责）。

### Process Patterns

**前端 Loading 状态：**
- TanStack Query 自带 `isLoading` / `isFetching` / `isError` 状态，不手动管理
- 全局 Loading：顶部细线进度条（路由切换时）
- 局部 Loading：Arco Design `<Spin />` 组件包裹内容区
- 骨架屏（Skeleton）：首次加载列表页时使用
- 按钮 Loading：提交操作时按钮进入 loading 状态，防止重复点击

**前端错误处理：**
- API 错误：TanStack Query `onError` 回调 → Arco Design `Message.error()` 全局提示
- 401 错误：Axios 拦截器统一处理 → 跳转登录页
- 网络错误：TanStack Query 自动重试 3 次，失败后展示重试按钮
- WebSocket 断开：连接状态指示器（绿/黄/红），自动重连，日志区展示"重连中..."提示
- React Error Boundary：包裹路由级组件，展示友好错误页面

**幂等键设计模式：**
- Webhook 去重：`{event_type}:{repo}:{commit_sha}:{timestamp_minute}` → Redis SETNX，TTL 5 分钟
- PipelineRun 提交：`{pipeline_id}:{trigger_type}:{trigger_id}` → 提交前检查是否已存在
- ArgoCD Application：按 `{project_id}:{environment_id}:{service_id}` 唯一标识，创建前查询是否存在
- 通知发送：`{event_type}:{pipeline_run_id}:{status}` → Redis SETNX 防止重复通知

### Enforcement Guidelines

**All AI Agents MUST:**
- 遵循上述命名约定，不自行发明新的命名风格
- 错误处理遵循 repo→service→handler 传播链，不在 handler 层做业务逻辑判断
- 日志使用 slog 结构化输出，包含 requestId，不使用 fmt.Println 或 log.Println
- 前端数据获取使用 TanStack Query，不手动管理 loading/error 状态
- JSON 字段使用 camelCase，数据库字段使用 snake_case
- 新模块遵循 internal/{module}/ 目录结构（handler.go/service.go/repo.go/model.go/dto.go）
- 测试文件与源文件同目录，命名 `*_test.go`（Go）/ `*.test.tsx`（React）

## Project Structure & Boundaries

### Requirements to Structure Mapping

| FR 能力领域 | 后端模块 | 前端页面/组件 | 公共层依赖 |
|------------|---------|-------------|-----------|
| 用户与权限管理 (FR1-5) | `internal/auth/` | `pages/login/`、`pages/admin/users/` | `pkg/middleware/auth.go` |
| 项目与资源管理 (FR6-10) | `internal/project/`、`internal/environment/`、`internal/service/` | `pages/projects/` | `pkg/middleware/project_scope.go` |
| 变量与凭证管理 (FR11-14) | `internal/variable/` | `pages/projects/[id]/variables/` | `pkg/crypto/`、`pkg/masking/` |
| Git 仓库集成 (FR15-19) | `internal/git/` | `pages/admin/integrations/` | — |
| 流水线编排与执行 (FR20-31) | `internal/pipeline/` | `pages/projects/[id]/pipelines/`、`components/pipeline-editor/` | `pkg/tekton/` |
| 构建与产物管理 (FR32-35) | `internal/pipeline/` | 复用流水线运行详情页 | `pkg/tekton/` |
| 实时日志与状态监控 (FR36-43) | `internal/pipeline/` | `pages/projects/[id]/pipelines/[pid]/runs/[rid]/` | `pkg/ws/`、`pkg/k8s/` |
| 部署与环境管理 (FR44-49) | `internal/environment/` | `pages/projects/[id]/environments/[eid]/` | `pkg/argocd/` |
| 通知与审计 (FR50-52) | `internal/notification/`、`internal/audit/` | `pages/admin/settings/` | `pkg/middleware/audit.go` |
| 全局概览与引导 (FR53-56) | `internal/dashboard/` | `pages/dashboard/` | — |
| 平台运维 (FR57-62) | `internal/admin/` | `pages/admin/settings/` | 全部 `pkg/` |

### Complete Project Directory Structure

**后端：**

```
zcid/
├── cmd/
│   └── server/
│       └── main.go                  # 入口：配置加载→DB/Redis/MinIO→K8s Client→路由注册→启动
│
├── internal/                        # 业务模块（按能力领域组织）
│   ├── auth/                        # FR1-5: 用户与权限管理
│   │   ├── handler.go               #   POST /auth/login, /auth/refresh, /auth/logout
│   │   ├── service.go               #   JWT 签发/验证、用户 CRUD、密码 bcrypt
│   │   ├── repo.go                  #   users 表操作
│   │   ├── model.go                 #   User struct
│   │   └── dto.go                   #   LoginRequest, TokenResponse
│   │
│   ├── project/                     # FR6-10: 项目与资源管理
│   │   ├── handler.go               #   /projects CRUD, /projects/:id/members
│   │   ├── service.go               #   项目创建/删除、成员管理、隔离校验
│   │   ├── repo.go                  #   projects, project_members 表操作
│   │   ├── model.go                 #   Project, ProjectMember struct
│   │   └── dto.go
│   │
│   ├── environment/                 # FR7, FR44-49: 环境与部署管理
│   │   ├── handler.go               #   /projects/:id/environments CRUD, 部署操作
│   │   ├── service.go               #   环境 CRUD、ArgoCD Application 创建/同步/回滚
│   │   ├── repo.go                  #   environments, deployments 表操作
│   │   ├── model.go                 #   Environment, Deployment struct
│   │   └── dto.go
│   │
│   ├── service/                     # FR8: 服务管理
│   │   ├── handler.go               #   /projects/:id/services CRUD
│   │   ├── service.go
│   │   ├── repo.go
│   │   ├── model.go                 #   Service struct
│   │   └── dto.go
│   │
│   ├── pipeline/                    # FR20-43: 流水线（核心模块）
│   │   ├── handler.go               #   /projects/:id/pipelines CRUD, /runs 触发/取消
│   │   ├── service.go               #   流水线 CRUD、触发逻辑、并发控制策略
│   │   ├── repo.go                  #   pipelines, pipeline_runs 表操作
│   │   ├── model.go                 #   Pipeline, PipelineRun struct
│   │   ├── dto.go
│   │   ├── executor.go              #   运行编排：变量合并→CRD翻译→K8s提交
│   │   ├── watcher.go               #   PipelineRun 状态 Watch（K8s Informer）
│   │   ├── logger.go                #   Pod 日志流管理（client-go GetLogs+Follow）
│   │   └── webhook.go               #   Webhook 接收/验证/匹配/触发
│   │
│   ├── variable/                    # FR11-14: 变量与凭证管理
│   │   ├── handler.go               #   /projects/:id/variables CRUD, 全局变量
│   │   ├── service.go               #   四级变量合并、密钥加解密
│   │   ├── repo.go                  #   variables 表操作
│   │   ├── model.go                 #   Variable struct
│   │   └── dto.go
│   │
│   ├── git/                         # FR15-16: Git 仓库集成
│   │   ├── handler.go               #   /admin/integrations/git, 仓库/分支查询
│   │   ├── service.go               #   GitProvider 调用、OAuth 流程
│   │   ├── repo.go                  #   git_connections 表操作
│   │   ├── model.go                 #   GitConnection struct
│   │   └── dto.go
│   │
│   ├── registry/                    # FR34: 镜像仓库管理
│   │   ├── handler.go               #   /admin/integrations/registry
│   │   ├── service.go               #   RegistryProvider 调用、连接测试
│   │   ├── repo.go
│   │   ├── model.go                 #   RegistryConnection struct
│   │   └── dto.go
│   │
│   ├── notification/                # FR50-51: 通知管理
│   │   ├── handler.go
│   │   ├── service.go               #   Notifier 调用、通知规则匹配
│   │   ├── repo.go                  #   notification_rules 表操作
│   │   ├── model.go
│   │   └── dto.go
│   │
│   ├── audit/                       # FR52: 审计日志
│   │   ├── handler.go               #   审计日志查询
│   │   ├── service.go
│   │   ├── repo.go                  #   audit_logs 表操作
│   │   ├── model.go                 #   AuditLog struct
│   │   └── dto.go
│   │
│   ├── dashboard/                   # FR53-56: 全局概览
│   │   ├── handler.go               #   /dashboard 聚合查询
│   │   └── service.go               #   跨模块聚合
│   │
│   └── admin/                       # FR57-62: 平台运维
│       ├── handler.go               #   /admin/settings, /health
│       ├── service.go               #   系统设置、健康检查、版本兼容检测
│       └── dto.go
│
├── pkg/                             # 公共基础层
│   ├── tekton/
│   │   ├── client.go                #   Tekton clientset 初始化
│   │   ├── translator.go            #   CRD 翻译引擎
│   │   ├── translator_test.go
│   │   └── testdata/                #   翻译测试用例
│   │       ├── go-microservice.json
│   │       ├── go-microservice.expected.yaml
│   │       ├── java-maven.json
│   │       ├── java-maven.expected.yaml
│   │       ├── frontend-node.json
│   │       ├── frontend-node.expected.yaml
│   │       └── generic-docker.json
│   ├── argocd/
│   │   ├── client.go                #   gRPC 连接初始化
│   │   └── application.go           #   Application CRUD/Sync/Rollback
│   ├── k8s/
│   │   ├── client.go                #   K8s clientset 初始化
│   │   ├── logs.go                  #   Pod 日志流
│   │   ├── secret.go                #   临时 Secret 创建/清理
│   │   └── health.go                #   健康检测
│   ├── crypto/
│   │   ├── aes.go                   #   AES-256-GCM 加解密
│   │   └── aes_test.go
│   ├── ws/
│   │   ├── manager.go               #   连接管理器（注册/注销/fan-out）
│   │   ├── hub.go                   #   频道 Hub
│   │   └── protocol.go              #   消息协议定义
│   ├── middleware/
│   │   ├── auth.go                  #   JWT 验证 + Casbin 鉴权
│   │   ├── audit.go                 #   审计日志拦截
│   │   ├── error.go                 #   全局错误处理 + panic recovery
│   │   ├── project_scope.go         #   项目隔离过滤
│   │   ├── cors.go
│   │   ├── ratelimit.go             #   敏感端点限流
│   │   ├── logger.go                #   请求日志
│   │   └── requestid.go             #   RequestID 生成
│   ├── response/
│   │   ├── response.go              #   Success/Error/Page 响应函数
│   │   ├── errors.go                #   BizError 类型 + 错误码常量
│   │   └── codes.go                 #   错误码→HTTP 状态码映射
│   ├── masking/
│   │   ├── engine.go                #   脱敏引擎
│   │   ├── handler.go               #   slog Handler 包装
│   │   └── stream.go                #   io.Reader 包装（Pod 日志流脱敏）
│   ├── config/
│   │   ├── config.go                #   Config struct 定义
│   │   └── loader.go                #   YAML 加载 + 环境变量覆盖
│   └── gitprovider/
│       ├── provider.go              #   GitProvider 接口定义
│       ├── gitlab.go
│       └── github.go
│
├── migrations/                      # 数据库迁移
│   ├── 000001_init_users.up.sql
│   ├── 000001_init_users.down.sql
│   ├── 000002_init_projects.up.sql
│   ├── 000002_init_projects.down.sql
│   ├── 000003_init_pipelines.up.sql
│   └── ...
│
├── templates/                       # 流水线模板
│   ├── go-microservice.json
│   ├── java-maven.json
│   ├── frontend-node.json
│   └── generic-docker.json
│
├── tests/integration/               # 集成测试
│   ├── pipeline_test.go
│   ├── auth_test.go
│   └── testutil/
│
├── docs/                            # swag v2 生成的 OpenAPI v3
├── charts/zcid/                     # Helm Chart
│   ├── Chart.yaml
│   ├── values.yaml
│   └── templates/
│
├── config.yaml                      # 开发环境配置
├── config.example.yaml              # 配置示例（提交到 Git）
├── docker-compose.yml
├── Dockerfile
├── Makefile
├── go.mod
├── go.sum
├── .gitignore
└── README.md
```

**前端：**

```
web/
├── src/
│   ├── App.tsx                      # 根组件：路由 + Provider
│   ├── main.tsx                     # 入口
│   ├── pages/
│   │   ├── login/
│   │   │   └── LoginPage.tsx
│   │   ├── dashboard/
│   │   │   └── DashboardPage.tsx
│   │   ├── projects/
│   │   │   ├── ProjectListPage.tsx
│   │   │   ├── ProjectLayout.tsx    # 项目级布局（侧边栏 + Outlet）
│   │   │   ├── pipelines/
│   │   │   │   ├── PipelineListPage.tsx
│   │   │   │   ├── PipelineEditPage.tsx
│   │   │   │   └── runs/
│   │   │   │       └── RunDetailPage.tsx
│   │   │   ├── environments/
│   │   │   │   ├── EnvironmentListPage.tsx
│   │   │   │   └── EnvironmentDetailPage.tsx
│   │   │   ├── services/
│   │   │   │   └── ServiceListPage.tsx
│   │   │   └── variables/
│   │   │       └── VariableListPage.tsx
│   │   └── admin/
│   │       ├── users/UserManagePage.tsx
│   │       ├── settings/SettingsPage.tsx
│   │       └── integrations/IntegrationsPage.tsx
│   ├── components/
│   │   ├── layout/
│   │   │   ├── AppLayout.tsx
│   │   │   ├── Sidebar.tsx
│   │   │   └── TopNav.tsx
│   │   ├── pipeline-editor/
│   │   │   ├── PipelineEditor.tsx
│   │   │   ├── StageNode.tsx
│   │   │   ├── StepNode.tsx
│   │   │   ├── StepConfigPanel.tsx
│   │   │   ├── TemplateSelector.tsx
│   │   │   ├── YamlEditor.tsx
│   │   │   └── ModeSwitch.tsx
│   │   ├── log-viewer/
│   │   │   ├── LogViewer.tsx
│   │   │   ├── LogLine.tsx
│   │   │   └── ConnectionStatus.tsx
│   │   └── common/
│   │       ├── ErrorBoundary.tsx
│   │       ├── PageSkeleton.tsx
│   │       └── PermissionGate.tsx
│   ├── hooks/
│   │   ├── useAuth.ts
│   │   ├── usePipelineStatus.ts
│   │   ├── useLogStream.ts
│   │   └── useDeployStatus.ts
│   ├── services/generated/          # @hey-api/openapi-ts 自动生成
│   ├── stores/
│   │   ├── authStore.ts
│   │   ├── uiStore.ts
│   │   └── editorStore.ts
│   ├── lib/ws/
│   │   ├── WebSocketManager.ts
│   │   ├── channels.ts
│   │   └── types.ts
│   ├── theme/tokens.ts
│   ├── utils/
│   │   ├── format.ts
│   │   └── permission.ts
│   └── types/index.ts
├── public/favicon.ico
├── .env.development
├── .env.production
├── .env.example
├── index.html
├── vite.config.ts
├── vitest.config.ts
├── tsconfig.json
├── openapi-ts.config.ts
├── package.json
└── .gitignore
```

### Architectural Boundaries

**层间依赖规则：**

```
前端 ←HTTP/WS→ Gin Router → Middleware → Handler → Service → Repo → PostgreSQL
                                                      ↓
                                                   pkg/ 层
                                                ↙    ↓     ↘
                                         pkg/tekton  pkg/argocd  pkg/k8s
                                             ↓           ↓          ↓
                                         Tekton API   ArgoCD gRPC  K8s API
```

- Handler 层：只做 HTTP 解析和响应输出，不含业务逻辑
- Service 层：编排业务逻辑，可调用多个 repo 和 pkg
- Repo 层：只做单表 CRUD，不做跨表 JOIN
- pkg 层：无状态工具库，不依赖 internal/

**模块间依赖规则：**
- `internal/` 模块间：service 层可调用其他模块的 service（接口注入），不直接调用其他模块的 repo
- `internal/` → `pkg/`：单向依赖
- `pipeline/` 是唯一调用 `pkg/tekton/`、`pkg/k8s/`、`pkg/ws/` 的模块
- `environment/` 是唯一调用 `pkg/argocd/` 的模块

**数据边界：**
- 每个模块管理自己的数据库表
- 项目隔离：带 `project_id` 的表查询必须经 `project_scope` 中间件过滤
- JSONB 数据：只有 `pipeline/` 模块读写 `pipeline_config` 字段

### Integration Points

**内部集成：**

| 调用方 | 被调用方 | 集成方式 | 场景 |
|--------|---------|---------|------|
| pipeline.service | variable.service | 接口调用 | 四级变量合并 |
| pipeline.service | git.service | 接口调用 | 获取 Git 信息 |
| pipeline.executor | pkg/tekton | 直接调用 | CRD 翻译 + K8s 提交 |
| pipeline.watcher | pkg/ws | 直接调用 | 状态变更推送 |
| pipeline.logger | pkg/ws | 直接调用 | 日志流推送 |
| environment.service | pkg/argocd | 直接调用 | Application CRUD/Sync/Rollback |
| admin.service | pkg/k8s | 直接调用 | 健康检查 |

**外部集成：**

| 外部系统 | 集成方式 | 封装位置 | 优先级 |
|---------|---------|---------|-------|
| Tekton | Go Typed Client | `pkg/tekton/` | P0 |
| ArgoCD | gRPC API | `pkg/argocd/` | P0 |
| K8s API Server | client-go | `pkg/k8s/` | P0 |
| PostgreSQL | GORM | 各模块 `repo.go` | P0 |
| Redis | go-redis/v9 | middleware、auth | P0 |
| MinIO | minio-go/v7 | pipeline/logger.go | P0 |
| GitLab/GitHub | REST API (OAuth) | `pkg/gitprovider/` | P0 |
| Harbor | REST API | `internal/registry/` | P0 |
| Email (SMTP) | net/smtp | `internal/notification/` | P0 |

### Development Workflow

**后端 Makefile：**

```makefile
dev:         # air 热重载
build:       # go build -o bin/zcid cmd/server/main.go
test:        # go test ./...
test-int:    # go test ./tests/integration/... -tags=integration
lint:        # golangci-lint run
swag:        # swag init -v2 --parseDependency -g cmd/server/main.go -o docs
migrate-up:  # migrate -path migrations -database $DB_URL up
migrate-new: # migrate create -ext sql -dir migrations -seq $(name)
```

**前端 package.json scripts：**

```json
{
  "dev": "vite",
  "build": "tsc && vite build",
  "test": "vitest",
  "lint": "eslint src/",
  "codegen": "openapi-ts"
}
```

**开发流程：**
1. `docker-compose up -d` → 启动 PostgreSQL/Redis/MinIO
2. `make migrate-up` → 数据库迁移
3. `make dev` → 后端热重载
4. `cd web && npm run dev` → 前端 Vite dev server
5. API 变更：`make swag` → `cd web && npm run codegen` → 前端获取新类型

## Architecture Validation Results

### Coherence Validation ✅

**Decision Compatibility：**
所有技术选型互相兼容，已验证 12 组关键组合（Go+Gin+GORM、swag v2+@hey-api/openapi-ts、React 19+@xyflow/react v12、Casbin+GORM adapter 等），无冲突。

**Pattern Consistency：**
命名约定（DB snake_case → API camelCase → Go PascalCase → React PascalCase）各层清晰无重叠。错误传播链（repo→service→handler）与三层代码组织一致。响应格式和 WebSocket 协议统一。

**Structure Alignment：**
每个 FR 能力领域有明确的 internal/ 模块对应。pkg/ 层按技术关注点划分。前端页面路由与 REST API 路由结构对齐。模块间单向依赖规则清晰。

### Requirements Coverage Validation ✅

**Functional Requirements：62/62 FR 全部有架构支撑，覆盖率 100%。**

| FR 范围 | 架构支撑 |
|---------|---------|
| FR1-5 用户与权限 | `internal/auth/` + JWT 双 Token + Casbin RBAC |
| FR6-10 项目与资源 | `internal/project/` + `environment/` + `service/` + project_scope |
| FR11-14 变量与凭证 | `internal/variable/` + `pkg/crypto/` + `pkg/masking/` |
| FR15-19 Git 集成 | `internal/git/` + `pkg/gitprovider/` + `pipeline/webhook.go` |
| FR20-31 流水线 | `internal/pipeline/` + `pkg/tekton/translator` + @xyflow/react |
| FR32-35 构建与产物 | `pipeline/executor.go` + `internal/registry/` |
| FR36-43 日志与监控 | `pipeline/logger.go` + `pipeline/watcher.go` + `pkg/ws/` |
| FR44-49 部署 | `internal/environment/` + `pkg/argocd/` |
| FR50-52 通知与审计 | `internal/notification/` + `internal/audit/` + audit 中间件 |
| FR53-56 全局概览 | `internal/dashboard/` |
| FR57-62 平台运维 | `internal/admin/` + 健康检查三级端点 |

**Non-Functional Requirements：28/28 NFR 全部有架构支撑，覆盖率 100%。**

| NFR 范围 | 架构支撑 |
|---------|---------|
| NFR1-6 性能 | GORM 连接池 + Redis 缓存 + WebSocket fan-out + 独立列索引 |
| NFR7-13 安全 | AES-256-GCM + JWT 双 Token + Casbin RBAC + 脱敏引擎 + 限流 |
| NFR14-18 可靠性 | 健康检查 + 错误传播链 + 后台任务重试 |
| NFR19-21 可扩展性 | 分页格式 + TTL 清理 + WebSocket 连接管理 |
| NFR22-25 集成 | 4 个接口抽象 + Tekton v1 兼容检测 + 统一响应格式 |
| NFR26-28 可观测性 | audit 中间件 + 健康检查 + slog 结构化日志 |

### Implementation Readiness Validation ✅

- ✅ 所有关键技术选型附带版本号并经网络验证
- ✅ 18 个核心架构决策记录，每项含理由
- ✅ 15 个实现模式规范，每项含示例
- ✅ 后端 12 个 internal 模块 + 9 个 pkg 包完整定义
- ✅ 前端完整目录树含所有页面和组件
- ✅ 开发工具链命令完整

### Gap Analysis Results

**Critical Gaps：无。**

**Important Gaps（实现阶段补充）：**
1. 数据库 Schema 详细设计 — 建表 SQL 在 Epic 实现阶段通过 golang-migrate 逐步完成
2. JSONB 内部 JSON schema — Stage/Step 结构在流水线 Epic 实现时设计
3. Tekton CRD 翻译映射规则 — 具体字段映射在实现时设计

## UX-Driven Frontend Architecture Supplement

_本章节基于 UX 设计规范（ux-design-specification.md）补充前端架构决策，与上文 Frontend Architecture 章节互补。_

### Three-Layer Component Architecture

```
┌─────────────────────────────────────────────┐
│  Domain Layer（领域组件）                      │
│  PipelineEditor, LogViewer, DiagnosisPanel,  │
│  EnvironmentHealth, GlobalSearch             │
├─────────────────────────────────────────────┤
│  Extension Layer（扩展组件）                   │
│  StatusBadge, MiniStatusBar, DynamicForm,    │
│  StageNode, StepNode, StagePreview,          │
│  ConfirmDialog, TimeAgo, ProjectSelector     │
├─────────────────────────────────────────────┤
│  Base Layer（基础层 — Arco Design）            │
│  Button, Input, Select, Table, Modal,        │
│  Message, Notification, Spin, Skeleton       │
└─────────────────────────────────────────────┘
```

**依赖规则：**
- Domain 可依赖 Extension 和 Base，不可跨 Domain 组件直接依赖
- Extension 只依赖 Base，不依赖 Domain
- Base 层不做任何修改，通过 Theme Token 定制外观

**目录映射：**
```
web/src/components/
├── common/          # Extension Layer（StatusBadge, ConfirmDialog, TimeAgo...）
├── pipeline-editor/ # Domain Layer（PipelineEditor + 子组件）
├── log-viewer/      # Domain Layer（LogViewer + 子组件）
├── layout/          # Extension Layer（AppLayout, Sidebar, TopNav）
└── forms/           # Extension Layer（DynamicForm, @rjsf 集成）
```

### PipelineRenderer Shared Base

从 PipelineEditor 抽取 `PipelineRenderer` 作为纯渲染基座，解决 StagePreview 复用问题：

```
PipelineRenderer（共享基座，只读渲染）
├── PipelineEditor extends PipelineRenderer（+ ReactFlow 交互：拖拽/连线/选择）
└── StagePreview uses PipelineRenderer（mini 模式，无交互开销）
```

**关键设计：**
- `PipelineRenderer`：接收 `stages: StageConfig[]`，输出 ReactFlow 节点/边，dagre 自动布局
- `PipelineEditor`：继承 Renderer + 添加 `onNodesChange`/`onEdgesChange`/`onConnect` 交互
- `StagePreview`：调用 Renderer 的 `mini` 模式，`fitView` + `zoomOnScroll=false` + `panOnDrag=false`
- 运行时状态着色通过 `statusMap` 注入，Renderer 本身无状态

### LogViewer Architecture

```
┌─────────────────────────────────────────┐
│ LogViewer                               │
│  ├── xterm.js Terminal（渲染层）          │
│  ├── useLogStream Hook（数据层）          │
│  │   ├── WebSocket 实时流               │
│  │   └── MinIO 归档分页（历史日志）       │
│  ├── ConnectionStatus（连接状态指示器）    │
│  └── LogToolbar（搜索/下载/全屏）         │
└─────────────────────────────────────────┘
```

**性能约束：**
- xterm.js scrollback buffer: 50,000 行（超出自动丢弃最早行）
- 10k 行渲染 < 1s，100k 行渲染 < 3s
- MinIO 归档日志按 1MB 分片分页加载，避免一次性拉取完整日志
- `useLogStream` 内部维护 `lastSeq`，断线重连时从断点续传

**懒加载：**
- xterm.js + xterm-addon-fit + xterm-addon-search 通过 `React.lazy` + `Suspense` 按需加载
- 仅在用户进入 RunDetailPage 时加载

### DynamicForm Architecture

```
@rjsf/core（JSON Schema → Form 渲染引擎）
    ↓
@zcid/rjsf-arco-theme（独立 npm 包，Arco Design 适配层）
    ↓
DynamicForm 组件（业务封装：Step 配置表单、变量表单）
```

**设计决策：**
- `@zcid/rjsf-arco-theme` 作为独立包发布（monorepo 内 `packages/rjsf-arco-theme/`）
- 适配 Arco 的 Input/Select/Switch/InputNumber/DatePicker 等表单控件
- Step 配置的 JSON Schema 存储在流水线模板中，DynamicForm 根据 schema 动态渲染
- 自定义 widget 注册机制：`customWidgets: { 'code-editor': MonacoWidget, 'secret-input': MaskedInput }`

### Monaco Editor Lazy Loading

```typescript
// web/src/components/pipeline-editor/YamlEditor.tsx
const MonacoEditor = React.lazy(() =>
  import('@monaco-editor/react').then(mod => ({ default: mod.default }))
);

// Preload 策略：用户 hover Pipeline 编辑按钮时触发
const preloadMonaco = () => import('@monaco-editor/react');
```

**加载策略：**
- 默认不加载 Monaco（~2MB gzipped），仅在用户切换到 YAML 模式时 lazy load
- Hover 预加载：`onMouseEnter` 在 ModeSwitch 按钮上触发 `preloadMonaco()`
- Fallback：加载期间显示 Arco `<Spin />` + "加载编辑器..."

### STATUS_MAP Global State Dictionary

```typescript
// web/src/constants/statusMap.ts
export const STATUS_MAP = {
  success:   { color: '#00B42A', bg: '#E8FFEA', icon: 'IconCheckCircle',  label: '成功' },
  running:   { color: '#1677FF', bg: '#E8F3FF', icon: 'IconLoading',      label: '运行中' },
  failed:    { color: '#F53F3F', bg: '#FFECE8', icon: 'IconCloseCircle',  label: '失败' },
  warning:   { color: '#FF7D00', bg: '#FFF7E8', icon: 'IconExclamation',  label: '警告' },
  pending:   { color: '#86909C', bg: '#F2F3F5', icon: 'IconClockCircle',  label: '等待中' },
  cancelled: { color: '#86909C', bg: '#F2F3F5', icon: 'IconMinusCircle',  label: '已取消' },
  timeout:   { color: '#F53F3F', bg: '#FFECE8', icon: 'IconClockCircle',  label: '超时' },
} as const satisfies Record<string, StatusStyle>;
```

**使用规则：**
- 所有状态→颜色→图标映射从 `STATUS_MAP` 读取，禁止组件内硬编码
- `StatusBadge`、`MiniStatusBar`、`StageNode` 运行时着色均引用此字典
- 新增状态只需在此处添加一行，全局生效

### Hooks Data Decoupling Layer

```typescript
// 三个核心实时数据 Hook，封装 WebSocket + TanStack Query 集成

// usePipeline — 流水线配置 CRUD + 运行列表
usePipeline(projectId, pipelineId?) → {
  pipelines,          // TanStack Query: GET /pipelines
  pipeline,           // TanStack Query: GET /pipelines/:pid
  runs,               // TanStack Query: GET /pipelines/:pid/runs
  realtimeStatus,     // WebSocket: /ws/v1/pipeline-status/:projectId
  createPipeline,     // mutation
  updatePipeline,     // mutation
  triggerRun,         // mutation
}

// useLogStream — 构建日志实时流
useLogStream(runId, stepName) → {
  terminal,           // xterm.js Terminal 实例引用
  connectionState,    // 'connecting' | 'connected' | 'disconnected' | 'archived'
  switchToArchive,    // 切换到 MinIO 归档日志
  searchInLogs,       // xterm-addon-search
}

// useDeployStatus — 部署状态实时监控
useDeployStatus(environmentId) → {
  deployments,        // TanStack Query: GET /environments/:eid
  realtimeStatus,     // WebSocket: /ws/v1/deploy-status/:environmentId
  syncApp,            // mutation: ArgoCD sync
  rollbackApp,        // mutation: ArgoCD rollback
}
```

**规则：**
- 页面组件只通过 Hooks 获取数据，不直接调用 API 或 WebSocket
- Hook 内部处理 WebSocket 消息 → TanStack Query cache invalidation（实时数据触发列表刷新）
- Hook 返回的 mutation 自带 optimistic update 和 error rollback

### Design Token Dual-Track Consumption

```
Arco Design Token（编译时）          Custom Token（运行时）
        │                                    │
   Less 变量                           CSS Variables
   ConfigProvider                      :root { --zcid-* }
        │                                    │
   Arco 组件内部样式                    自定义组件样式
   (Button, Input, Table...)           (StatusBadge, MiniStatusBar...)
```

**实现：**
- Arco 组件：通过 `ConfigProvider` + Less 变量覆盖（`theme/tokens.ts` 已定义）
- 自定义组件：通过 CSS Variables（`--zcid-status-success`、`--zcid-spacing-page` 等）
- 两套 Token 在 `theme/tokens.ts` 统一管理，确保色值一致
- 暗色模式预留：CSS Variables 支持运行时切换，Arco 通过 `body[arco-theme="dark"]`

### Responsive Breakpoint Implementation

```css
/* web/src/styles/breakpoints.css */
:root {
  --bp-mobile: 768px;
  --bp-tablet: 1024px;
  --bp-desktop: 1280px;
  --bp-wide: 1440px;
}
```

**自适应规则：**
- `< 768px`：不主动适配（管理平台非移动端场景）
- `768px - 1024px`：隐藏侧边栏，顶部汉堡菜单
- `1024px - 1280px`：侧边栏折叠为 icon-only 模式（64px）
- `1280px+`：侧边栏完整展开（240px）
- `1440px+`：内容区最大宽度 1200px 居中

**Sidebar 自动折叠：**
```typescript
// web/src/stores/uiStore.ts 扩展
interface UIStore {
  sidebarCollapsed: boolean;
  setSidebarCollapsed: (v: boolean) => void;
  // 监听 window resize，< 1280px 自动折叠
}
```

### Accessibility Testing Pipeline

**CI 集成（阻断级）：**
- `eslint-plugin-jsx-a11y`：ESLint 规则，PR 级阻断
- `axe-core`：vitest 集成，关键页面组件测试中调用 `axe(container)` 断言零 violations
- Lighthouse CI：`lhci autorun`，a11y score < 90 阻断合并

**开发时辅助：**
- Storybook `@storybook/addon-a11y`：组件开发时实时 a11y 检查（P1 阶段引入 Storybook 后）
- 手动检查清单：键盘导航（Tab/Enter/Escape）、屏幕阅读器（VoiceOver）、色盲模拟

**关键 a11y 实现：**
- 所有交互元素 `tabIndex` + `aria-label`
- PipelineEditor 节点：`role="treeitem"` + `aria-expanded`
- LogViewer：`role="log"` + `aria-live="polite"`
- StatusBadge：`aria-label` 包含状态文本（不仅依赖颜色）
- 焦点管理：Modal 打开时 trap focus，关闭时恢复焦点

### Updated Frontend Directory Structure

基于三层组件架构，补充目录：

```
web/src/
├── components/
│   ├── common/                    # Extension Layer
│   │   ├── StatusBadge.tsx
│   │   ├── MiniStatusBar.tsx
│   │   ├── ConfirmDialog.tsx
│   │   ├── TimeAgo.tsx
│   │   ├── ProjectSelector.tsx
│   │   ├── ErrorBoundary.tsx
│   │   ├── PageSkeleton.tsx
│   │   └── PermissionGate.tsx
│   ├── pipeline-editor/           # Domain Layer
│   │   ├── PipelineRenderer.tsx   # 共享渲染基座（新增）
│   │   ├── PipelineEditor.tsx     # 交互编辑器（extends Renderer）
│   │   ├── StageNode.tsx          # 三模式：edit/runtime/mini
│   │   ├── StepNode.tsx
│   │   ├── StepConfigPanel.tsx
│   │   ├── StagePreview.tsx       # mini 预览（uses Renderer）
│   │   ├── TemplateSelector.tsx
│   │   ├── YamlEditor.tsx         # Monaco lazy load
│   │   └── ModeSwitch.tsx
│   ├── log-viewer/                # Domain Layer
│   │   ├── LogViewer.tsx          # xterm.js 容器
│   │   ├── LogToolbar.tsx         # 搜索/下载/全屏
│   │   └── ConnectionStatus.tsx
│   ├── forms/                     # Extension Layer（新增）
│   │   └── DynamicForm.tsx        # @rjsf/core 封装
│   └── layout/
│       ├── AppLayout.tsx
│       ├── Sidebar.tsx
│       └── TopNav.tsx
├── constants/
│   └── statusMap.ts               # STATUS_MAP 全局字典（新增）
├── styles/
│   └── breakpoints.css            # 响应式断点 CSS Variables（新增）
├── packages/                      # monorepo 内部包（新增，项目根目录）
│   └── rjsf-arco-theme/          # @zcid/rjsf-arco-theme
│       ├── src/
│       ├── package.json
│       └── tsconfig.json
```

### UX Architecture Enforcement Rules

**All AI Agents MUST additionally:**
- 新组件按三层架构归类，Domain 组件不可跨域依赖
- 所有状态颜色/图标从 `STATUS_MAP` 读取，禁止硬编码
- 页面数据获取必须通过 Hooks 层（usePipeline/useLogStream/useDeployStatus），不直接调用 API
- 重型依赖（Monaco、xterm.js）必须 `React.lazy` 懒加载
- 自定义组件样式使用 CSS Variables（`--zcid-*`），不使用 Arco Less 变量
- 交互元素必须包含 `aria-label`，状态信息不可仅依赖颜色传达

4. ArgoCD Application 模板 — Application spec 在部署 Epic 实现时设计

**Nice-to-Have：**
- E2E 测试方案（P1 阶段）
- CI 流水线配置（项目初始化时）
- 监控指标体系（P1 阶段）

### Architecture Readiness Assessment

**Overall Status: READY FOR IMPLEMENTATION**

**Confidence Level: HIGH**

**Key Strengths:**
1. 技术选型一致性高 — Go/K8s 全生态统一，前端现代化栈
2. 模块边界清晰 — 单向依赖，职责分离
3. 核心技术难点有专项设计 — CRD 翻译可独立测试，WebSocket fan-out，三级错误分类
4. 横切关注点全面 — 10 个关注点均有方案
5. 实现模式详细 — AI Agent 实现无歧义

### Implementation Handoff

**AI Agent Guidelines：**
- 遵循本文档所有架构决策
- 使用实现模式规范保持代码一致性
- 尊重项目结构和模块边界
- 所有架构问题参考本文档

**First Implementation Priority：**
1. 项目初始化（go mod init + npm create vite）
2. docker-compose.yml + config.yaml
3. pkg/config/ 配置加载
4. pkg/middleware/ 中间件骨架
5. pkg/response/ 统一响应格式
6. internal/auth/ 认证模块（第一个业务模块）

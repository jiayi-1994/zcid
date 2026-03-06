---
stepsCompleted: [1, 2, 3, 4]
inputDocuments: []
session_topic: 'zcid 平台技术选型 — 基于 Tekton+ArgoCD 的类 Zadig CI/CD 平台'
session_goals: '核心技术选型决策，保障流水线引擎、镜像仓库对接、前后端交互等关键能力落地'
selected_approach: 'ai-recommended'
techniques_used: ['Morphological Analysis', 'First Principles Thinking', 'Constraint Mapping']
ideas_generated: 42
context_file: ''
ui_preference: 'Apple 风格，蓝白色调'
session_active: false
workflow_completed: true
---

# zcid 平台技术选型 — 头脑风暴会话成果

**Facilitator:** xjy
**Date:** 2026-03-01
**总决策点：** 42 个

---

## Session Overview

**Topic:** zcid — 基于 Tekton + ArgoCD 的类 Zadig CI/CD 平台技术选型
**Goals:** 核心技术选型决策，保障关键能力落地；UI 偏好 Apple 风格蓝白色调

### 核心约束

- CI 引擎：Tekton（已确定）
- CD 引擎：ArgoCD（已确定）
- 镜像仓库：Harbor（P0）+ 阿里云 ACR（P1）
- 最高优先级：流水线（Pipeline）能力
- 参考平台：Zadig
- UI 风格：Apple 风格，蓝白主色调

---

## Technique Selection

**Approach:** AI-Recommended Techniques
**Analysis Context:** 技术架构决策场景，需结构化拆解 + 深度分析 + 风险识别

- **形态学分析（Morphological Analysis）：** 系统列出每个技术维度的候选方案，组合最优解
- **第一性原理思维（First Principles Thinking）：** 回归本质需求，驱动核心选型决策
- **约束映射（Constraint Mapping）：** 识别真实约束，避免选型返工

**AI Rationale:** 发散→收敛→排雷，确保选型既全面又务实

---

## Technique Execution Results

### 技术 1：形态学分析（Morphological Analysis）

系统拆解 zcid 平台的每个技术维度，逐一确定候选方案。

**产出决策：**

- **#1 后端语言 — Go：** 云原生标配，Tekton/ArgoCD/K8s 全生态一致，client-go 原生支持
- **#2 后端架构 — 单体服务：** Tekton 负责执行，ArgoCD 负责部署，zcid 只是编排管理层，单体足够
- **#3 Web 框架 — Gin：** Zadig 验证过的路线，中间件生态齐全，上手成本最低
- **#4 前端 — React + TypeScript + Arco Design：** 设计语言现代简洁，蓝白主题定制成本低
- **#5 中间件栈 — PostgreSQL + Redis + MinIO：** 三个组件覆盖结构化存储+缓存队列+对象存储，比 Zadig 少两个组件
- **#6 API 层 — RESTful + WebSocket：** REST 做 CRUD，WebSocket 做实时日志推送和状态变更通知
- **#7 对象存储 — MinIO：** 构建产物、日志归档、Helm Chart、配置文件，S3 兼容协议
- **#8 认证权限 — JWT + Casbin RBAC：** 无状态认证 + 细粒度权限控制，后续加 OAuth
- **#9 K8s 交互 — Tekton Go Typed Client + ArgoCD gRPC API：** 根据两个引擎架构特点分别选最合适的交互方式
- **#10 日志 — Pod 日志直读 + MinIO 归档：** 实时通过 client-go GetLogs+Follow，历史存 MinIO，不引入 ELK

### 技术 2：第一性原理思维（First Principles Thinking）

对关键决策追问「为什么」，验证决策是否站得住脚。

**产出决策：**

- **#11 平台本质定位：** zcid = 翻译层 + 状态看板，不是执行引擎。把用户意图翻译成 K8s 资源，监听状态，展示结果。越薄越好
- **#12 混合流水线模型：** 简化 Stage→Step 模型 + 模板系统（常见场景一键创建）+ YAML 逃生舱（高级用户直编 Tekton YAML）。三层用户（新手/中级/高级）一个入口三种体验
- **#13 触发机制四级优先：** 手动+Webhook(P0)，复用 Tekton Triggers EventListener；Cron(P1)；镜像变更(P2)
- **#14 三层资源模型：** Project（权限边界）→ Environment（K8s Namespace 隔离）→ Service（实际工作负载）。ArgoCD Application 与 Environment+Service 一一映射
- **#15 镜像仓库抽象层：** RegistryProvider 接口（ListRepositories/ListTags/CheckHealth），Harbor 和 ACR 各一个 adapter。zcid 不碰镜像推拉，只管配置和查询
- **#15a 优先级调整：** Harbor 优先（私有化部署，开发调试方便），ACR 其次
- **#16 Kaniko 默认构建引擎：** 无特权容器运行，产出标准 OCI 镜像，Docker/containerd/CRI-O 通吃
- **#17 BuildKit 多架构扩展：** P1 引入，专门处理 amd64+arm64 双架构构建，与 Kaniko 并存不替换

### 技术 3：约束映射（Constraint Mapping）

系统梳理硬约束、软约束、假约束，提前排雷。

**硬约束：**

- **#18 Tekton CRD 版本：** 必须基于 v1 API，启动时检测 Tekton Pipeline v0.44+
- **#19 ArgoCD 集群注册：** 封装到 zcid 集群管理功能，一步到位
- **#20 K8s RBAC 最小权限：** 精确定义 ClusterRole，只开放必要的 resources+verbs
- **#21 凭证加密：** AES-256-GCM 加密存 PostgreSQL，运行时通过 K8s Secret 传递，日志不打印凭证
- **#22 部署形态：** 先单集群 in-cluster config，ClusterManager 接口预留多集群扩展

**软约束：**

- **#23 流水线并发控制：** maxConcurrentRuns 全局+项目级，提交 PipelineRun 前检查当前运行数
- **#24 WebSocket 连接管理：** fan-out 模式，同一 Pod 日志流只 Watch 一次，多连接共享
- **#25 JSONB 查询策略：** 高频查询字段保持独立列+索引，JSONB 只做存储不做复杂查询

**假约束（已排除）：**

- **#26 "单体扛不住"：** 管理平台 QPS 不高，Go+Gin 单体轻松应对
- **#27 "需要 ELK"：** Pod 日志+MinIO 归档覆盖 CI/CD 构建日志场景
- **#28 "必须用 Helm"：** 支持 Helm/Kustomize/YAML 三种部署方式，不强绑

### 补充探索：遗漏排查

系统扫描后补充的关键决策点。

**集成与对接：**

- **#29a Git 集成：** GitProvider 接口，GitLab+GitHub 各一个 adapter，OAuth 授权关联代码仓库
- **#30a 四级变量体系：** 全局→项目→流水线→运行时，下级覆盖上级。密钥变量 AES-256-GCM 加密，运行时临时 K8s Secret 注入，结束后自动清理
- **#31a 通知系统：** Notifier 接口，邮件+Webhook(P0)，飞书/钉钉/企微(P1)

**前端补充：**

- **#32a 状态管理：** Zustand（UI 状态）+ React Query/TanStack Query（服务端数据缓存与获取）
- **#41 国际化：** react-i18next 结构预留，第一版只做中文
- **#42 Dashboard 概览页：** Apple 风格卡片布局，构建趋势图、部署频率热力图，差异化设计

**工程基础：**

- **#33a 审计日志：** Gin 中间件拦截写操作，记录用户+时间+操作+目标+结果到 PostgreSQL 审计表
- **#34a 环境配置管理：** 复用 ArgoCD 的 Helm values overlay / Kustomize overlay
- **#35 代码结构：** 按业务模块组织（pipeline/project/environment/service/registry/auth），公共基础层 common/
- **#36 ORM：** GORM，上手快功能全
- **#37 数据库 Migration：** golang-migrate，SQL 文件驱动，显式 up/down，启动时自动执行
- **#38 API 文档：** swaggo/swag 自动生成 Swagger，前端 openapi-typescript-codegen 自动生成请求代码
- **#39 统一错误处理：** 分段错误码（400xx 认证/401xx 项目/402xx 流水线）+ 统一 JSON 响应格式
- **#40 回滚机制：** 复用 ArgoCD Rollback API，前端部署历史加「回滚到此版本」按钮

---

## 最终技术全栈总览

```
┌─────────────────┬─────────────────────────────────────┐
│ 维度             │ 选型                                 │
├─────────────────┼─────────────────────────────────────┤
│ 后端语言         │ Go                                   │
│ Web 框架         │ Gin                                  │
│ 后端架构         │ 单体 + 按业务模块组织                   │
│ ORM             │ GORM                                 │
│ DB Migration    │ golang-migrate                       │
│ API 文档         │ swaggo/swag (Swagger/OpenAPI)        │
│ 前端框架         │ React + TypeScript                   │
│ UI 组件库        │ Arco Design                          │
│ 状态管理         │ Zustand + TanStack Query             │
│ 国际化           │ react-i18next（预留，先中文）           │
│ UI 风格          │ Apple 风格，蓝白色调                   │
│ API 通信         │ RESTful + WebSocket                  │
│ 数据库           │ PostgreSQL（JSONB）                   │
│ 缓存/消息队列    │ Redis（Streams）                      │
│ 对象存储         │ MinIO                                │
│ 认证             │ JWT                                  │
│ 权限             │ Casbin RBAC（后续 OAuth）              │
│ Tekton 交互      │ 官方 Go Typed Client                  │
│ ArgoCD 交互      │ 官方 gRPC API Client                  │
│ CI 引擎          │ Tekton                               │
│ CD 引擎          │ ArgoCD                               │
│ 构建引擎         │ Kaniko(P0) + BuildKit 多架构(P1)      │
│ 镜像仓库         │ Harbor(P0) + 阿里云 ACR(P1)           │
│ Git 集成         │ GitLab + GitHub，GitProvider 接口     │
│ 流水线模型       │ 混合：简化模型 + 模板 + YAML 逃生舱     │
│ 资源模型         │ Project → Environment → Service      │
│ 触发机制         │ 手动+Webhook(P0), Cron(P1)           │
│ 变量管理         │ 四级变量 + AES-256-GCM 密钥加密        │
│ 日志             │ Pod 日志直读(实时) + MinIO 归档(历史)  │
│ 通知             │ 邮件+Webhook(P0), 飞书/钉钉(P1)      │
│ 审计             │ Gin 中间件 + PostgreSQL 审计表         │
│ 环境配置         │ ArgoCD overlay (Helm/Kustomize/YAML) │
│ 部署形态         │ 单集群(P0) → 多集群(P1)              │
│ 错误处理         │ 分段错误码 + 统一 JSON 响应             │
└─────────────────┴─────────────────────────────────────┘
```

---

## 优先级路线图

### P0 — 核心链路（MVP）

1. 用户认证（JWT）+ 权限管理（Casbin RBAC）
2. 项目 / 环境 / 服务基础 CRUD
3. 流水线可视化编排 + 模板系统
4. Tekton PipelineRun 翻译与执行
5. 实时构建日志（WebSocket + Pod 日志直读）
6. Kaniko 镜像构建 + Harbor 推送
7. ArgoCD 部署 + 环境状态展示 + 回滚
8. GitLab / GitHub Webhook 触发
9. 变量与密钥管理（四级体系）
10. 邮件 + Webhook 通知
11. 审计日志
12. Dashboard 概览页

### P1 — 完善

- 阿里云 ACR 镜像仓库支持
- BuildKit amd64+arm64 多架构构建
- Cron 定时触发
- 飞书 / 钉钉 / 企业微信通知
- 多集群管理
- OAuth 集成（PaaS 平台嵌入）
- 多语言支持（中英文）

### P2 — 增强

- 镜像变更触发 CD
- Dashboard 高级数据可视化（趋势图、热力图）
- 流水线模板市场
- 全局搜索

---

## 关键架构决策摘要

### 接口抽象设计（扩展性保障）

| 接口 | 职责 | 已实现 | 待扩展 |
|------|------|--------|--------|
| `RegistryProvider` | 镜像仓库操作 | Harbor | ACR, Docker Hub, TCR |
| `GitProvider` | 代码仓库操作 | GitLab, GitHub | Gitea, Gitee |
| `Notifier` | 通知发送 | Email, Webhook | 飞书, 钉钉, 企微 |
| `ClusterManager` | K8s 集群连接 | 单集群 in-cluster | 多集群 kubeconfig |

### 安全设计要点

- 凭证存储：AES-256-GCM 加密 → PostgreSQL
- 凭证运行时：临时 K8s Secret → Tekton Pod → 结束后清理
- 密钥变量：前端不可回显，不允许被普通变量覆盖降级
- K8s RBAC：最小权限 ClusterRole
- 日志脱敏：凭证不打印到日志

### 「不做」清单

- 不自建任务调度引擎（Tekton 负责）
- 不自建部署编排（ArgoCD 负责）
- 不引入 ELK/Loki（Pod 日志 + MinIO 够用）
- 不引入独立消息队列（Redis Streams 够用）
- 不拆微服务（管理平台单体够用）
- 不强绑 Helm（支持多部署方式）

---

## Session Summary

**创意技术：** 形态学分析 → 第一性原理思维 → 约束映射
**总决策点：** 42 个核心技术选型决策
**主题覆盖：** 9 大主题（架构定位、后端、前端、流水线引擎、资源模型、集成对接、数据通信、安全权限、运维可观测）
**约束识别：** 5 个硬约束 + 3 个软约束 + 3 个假约束已排除

**关键洞察：**
- zcid 的本质是「翻译层 + 状态看板」，保持薄是最大的架构优势
- 极简中间件栈（PostgreSQL + Redis + MinIO）比 Zadig 少两个组件，运维复杂度直接砍半
- 接口抽象（RegistryProvider / GitProvider / Notifier / ClusterManager）是扩展性的关键保障
- 混合流水线模型让三层用户（新手/中级/高级）都能高效使用

**下一步：** 基于本文档进入产品简报或 PRD 创建阶段，将技术选型转化为具体的产品需求和实施计划。

**开发阶段 Skill 提取建议：** 每完成一个开发阶段，使用 `/oh-my-claudecode:learner` 提取该阶段的技术决策、踩坑记录、架构模式、API 设计规范为 skill，供二期开发复用。

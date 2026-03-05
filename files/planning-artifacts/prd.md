---
stepsCompleted: [step-01-init, step-02-discovery, step-02b-vision, step-02c-executive-summary, step-03-success, step-04-journeys, step-05-domain, step-06-innovation-skipped, step-07-project-type, step-08-scoping, step-09-functional, step-10-nonfunctional, step-11-polish]
inputDocuments: ['files/planning-artifacts/product-brief-zcid-2026-03-01.md', 'files/brainstorming/brainstorming-session-2026-03-01-001.md']
workflowType: 'prd'
briefCount: 1
brainstormingCount: 1
researchCount: 0
projectDocsCount: 0
classification:
  projectType: 'B2B 平台工具（可私有部署）'
  domain: 'Cloud-Native DevOps Platform'
  complexity: '中高（技术集成驱动）'
  projectContext: 'greenfield'
date: 2026-03-01
author: xjy
---

# Product Requirements Document - zcid

**Author:** xjy
**Date:** 2026-03-01

## Executive Summary

zcid 是一个面向开发者和 DevOps 工程师的云原生 CI/CD 平台，基于 Tekton（CI 引擎）和 ArgoCD（CD 引擎）构建。平台的核心定位是「翻译层 + 状态看板」——将用户的可视化操作意图翻译为 Kubernetes CRD 资源，监听执行状态，展示实时结果，而不自建任何任务调度或部署编排引擎。

平台解决的核心问题：Tekton 和 ArgoCD 作为云原生 CI/CD 标杆工具能力强大，但原生界面完全面向 YAML 和命令行，对绝大多数开发者使用门槛过高，尤其不符合国内用户的操作习惯。现有替代方案要么架构老旧（Jenkins），要么社区萎缩（Zadig），要么锁定云厂商（阿里云效），市场上缺少「云原生 + 体验好 + 可私有部署 + 可商业化」四合一的解决方案。

目标用户覆盖三层：后端开发者（自助完成构建部署，不写 YAML）、DevOps 工程师（高效配置流水线，赋能开发团队）、技术管理者（全局状态可视化，支撑决策）。平台支持两条核心构建链路：容器化微服务（Kaniko→Harbor→ArgoCD 部署）和传统应用（编译打包→MinIO 对象存储）。

交付模式为双轨制：私有云独立部署和 PaaS 平台模块化嵌入，覆盖自用和商业化两个场景。

### What Makes This Special

1. **「小白都好用」的产品哲学：** 混合流水线模型让三层用户（新手选模板 / 中级可视化编排 / 高级 YAML 直编）在同一入口获得各自最优体验，差异化时刻是用户第一次不写任何 YAML 完成从代码到部署的完整闭环
2. **最薄翻译层架构：** 不重新造轮子，Tekton 负责执行，ArgoCD 负责部署，zcid 只做意图翻译和状态展示，架构越薄维护成本越低、与 K8s 生态兼容性越强
3. **极简中间件栈：** PostgreSQL + Redis + MinIO 三组件覆盖全部基础能力，比同类产品少两个中间件，运维复杂度直接砍半
4. **市场窗口明确：** K8s 成为事实标准 + Jenkins 老化 + Zadig 社区下降 + 云厂商锁定，四重因素打开了云原生 CI/CD 平台的市场空白

## Project Classification

- **项目类型：** B2B 平台工具（可私有部署）
- **领域：** Cloud-Native DevOps Platform
- **复杂度：** 中高 — 技术集成驱动（Tekton CRD 翻译、ArgoCD gRPC 集成、WebSocket 实时日志、K8s API 深度交互），非合规驱动
- **项目上下文：** Greenfield（全新产品）
- **技术栈：** Go + Gin（后端）、React + TypeScript + Arco Design（前端）、PostgreSQL + Redis + MinIO（中间件）
- **UI 风格：** Apple 风格，蓝白色调

## Success Criteria

### User Success

**开发者（小李）成功标准：**
- 从首次登录到成功运行第一条模板流水线（构建成功+产物生成）< 30 分钟（不含管理员前置配置时间）
- 能从模板自助创建流水线并成功运行，不需要 DevOps 介入（MVP 阶段定性验证：3-5 名非 DevOps 开发者独立完成；Growth 阶段切换为定量指标 > 70%）
- 查看构建日志、触发构建、确认部署状态均可在 3 次点击内完成
- 构建失败时能在日志中快速定位错误原因，无需切换到其他系统
- **Aha Moment：** 第一次全程不写 YAML 完成从代码提交到 K8s 部署的闭环

**DevOps 工程师（老张）成功标准：**
- 通过可视化界面 + 模板系统配置流水线效率不低于直接编写 Tekton YAML
- 开发者自助完成日常 CI/CD 操作，DevOps 只处理需要运维介入的事项
- 一个界面管理项目、环境、服务、流水线全流程，不在多个系统间切换
- **Aha Moment：** 开发者第一次自己完成了流水线配置和部署

**技术管理者（王总）成功标准：**
- 项目状态一览页面能一眼看到最近构建和部署状态
- 不用问人就知道各环境的健康状况

### Business Success

**3 个月目标（MVP）：**
- P0 核心链路完成，内部团队可用
- 容器化链路（Kaniko→Harbor→ArgoCD）和传统构建链路（JAR→MinIO）两条链路跑通
- 流水线执行成功率 > 95%

**6 个月目标（完善）：**
- P1 功能就绪（ACR、多架构构建、Cron 触发）
- 具备对外展示和试用的产品完整度
- 建立性能基线：基于真实运行数据建立构建时间、部署时间基线，为后续优化提供依据

**12 个月目标（商业化）：**
- PaaS 嵌入能力就绪（OAuth 集成）
- 可作为模块进入商业化流程

### Technical Success

- **CRD 翻译正确率：** 用户可视化编排翻译为 Tekton PipelineRun 后能正确执行（不因翻译错误而失败）> 99%
- **平台稳定性：** 月可用率 > 99.5%（排除 Tekton/ArgoCD 自身故障）
- **部署响应时间：** 从 zcid 触发到 ArgoCD Application 开始同步 < 30 秒（不含 ArgoCD 同步完成时间）
- **日志推送延迟：** WebSocket 推送延迟 < 2 秒，首次连接到日志流出现 < 5 秒
- **构建隔离性：** 多个流水线并发运行时，任何一条失败不影响其他流水线
- **并发构建：** 单项目多个微服务能同时触发构建，互不干扰
- **构建产物覆盖：** MVP 至少覆盖镜像（OCI）+ JAR 两种产物类型
- **凭证安全：** AES-256-GCM 加密存储，运行时临时 Secret 注入，结束自动清理，日志不打印凭证（纳入自动化测试持续回归）

### Measurable Outcomes

**MVP 验证指标（上线即验证）：**

| 指标 | 衡量方式 | 目标 |
|------|---------|------|
| 核心链路成功率 | 流水线执行成功率 | > 95% |
| CRD 翻译正确率 | 翻译后 PipelineRun 正确执行率 | > 99% |
| 上手时间 | 首次登录→模板流水线成功运行 | < 30 分钟 |

**运营指标（上线后持续跟踪）：**

| 指标 | 衡量方式 | 目标 |
|------|---------|------|
| 用户自助率 | 开发者自建流水线成功比例 | > 70% |
| 部署响应 | zcid 触发到 ArgoCD 同步 | < 30 秒 |
| 平台可用性 | 月可用率 | > 99.5% |
| 日志推送 | WebSocket 延迟 | < 2 秒 |
| 首次日志连接 | Watch 建立到日志出现 | < 5 秒 |

## Product Scope

### MVP - Minimum Viable Product

**Wave 1（基础设施层）：** JWT 认证 + RBAC 权限、项目/环境/服务 CRUD、变量与密钥管理、Git 仓库集成

**Wave 2（核心链路）：** 流水线可视化编排 + 模板系统、Tekton PipelineRun 翻译与执行、实时构建日志（WebSocket）、Kaniko 镜像构建 + Harbor 推送、传统应用构建（JAR→MinIO）

**Wave 3（闭环体验）：** ArgoCD 部署 + 状态展示 + 回滚、Webhook 触发、通知（Webhook）、审计日志（简化版）、简化版 Dashboard

### Growth Features (Post-MVP)

详见「Project Scoping & Phased Development > Post-MVP Features」章节。核心方向：镜像仓库扩展（ACR）、多架构构建、定时触发、IM 通知集成、多集群管理、性能基线建立。

### Vision (Future)

详见「Project Scoping & Phased Development > Post-MVP Features > Phase 3」章节。核心方向：OAuth/PaaS 嵌入、高级数据可视化、模板市场、全局搜索、多语言支持。

## User Journeys

### Journey 1: 小李的第一次自助部署（开发者 - 成功路径）

小李是一个 25 岁的 Go 后端开发者，入职三年，对 K8s 有基本了解但从未写过 Tekton YAML。之前每次改完代码都要找 DevOps 老张帮忙部署，老张经常很忙，一等就是半天，小李觉得自己被流程拖慢了。

**开场：** 周一早上，老张在团队群里发了一条消息：「zcid 上线了，大家注册账号自己试试，以后构建部署不用找我了。」小李半信半疑地点开链接，用公司邮箱登录。

**首次引导：** 登录后，页面顶部出现一个欢迎横幅：「欢迎来到 zcid，你已被加入 mall-platform 项目，点此查看你的流水线。」小李点击后直接进入项目页面，看到自己负责的 user-service 已经有了一条配好的流水线。

**探索：** 小李好奇地点进「创建流水线」，发现有模板列表——「Go 微服务构建」正好是他需要的。选择模板后，界面上只需要填几个参数：Git 仓库地址（下拉选择已关联的仓库）、分支（默认 main）、Dockerfile 路径、镜像名称。小李花了 3 分钟填完。

**高潮：** 小李点击「运行」，页面跳转到流水线执行视图。他看到流水线的每个 Step 像进度条一样依次亮起——代码拉取 ✅、编译构建（正在运行...）。他点开构建 Step，实时日志开始滚动，跟他在终端看 `go build` 的输出一模一样。3 分钟后，所有 Step 全部变绿，镜像成功推送到 Harbor。紧接着 ArgoCD 开始同步，部署状态从「Syncing」变成「Healthy」。

**结局：** 整个过程 8 分钟，小李全程没写一行 YAML，没问任何人。他在群里回复老张：「这个好用，以后不用烦你了。」这是他的 Aha Moment。

**权限边界：** 小李作为普通成员，只能看到自己被分配的项目，只能在 dev 环境触发部署。staging/prod 环境的部署需要项目管理员或更高权限。

**揭示的能力需求：** 首次登录引导、模板系统、参数化配置、实时日志 WebSocket、流水线状态可视化、Git 仓库下拉选择、一键运行、RBAC 权限隔离

---

### Journey 2: 小李遇到构建失败和部署失败（开发者 - 错误恢复路径）

**场景 A — 构建失败：**

周三下午，小李提交了一段代码，Webhook 自动触发了流水线。他收到一封邮件通知：「流水线执行失败」。

小李打开 zcid，在流水线列表中看到最新一次运行标红。点进去，Step 视图清晰地显示「编译构建」这个 Step 失败了，其他 Step 灰色（未执行）。他点开失败的 Step，日志输出最后几行用红色高亮显示了错误：`undefined: config.NewDBConnection`——他忘了提交一个依赖文件。

小李回到 IDE 补上遗漏的文件，push 到 Git。Webhook 再次自动触发流水线。这次所有 Step 顺利通过。

**场景 B — 部署失败：**

周五，小李的构建成功了，镜像推送到了 Harbor，但 ArgoCD 部署阶段出了问题。zcid 显示部署状态为「Degraded」，同步详情显示：「ImagePullBackOff — 镜像拉取失败」。原来是 staging 环境的 Harbor 凭证过期了。

小李没有权限修改凭证，但他在 zcid 上看到了清晰的错误信息，截图发给老张。老张更新了凭证后，小李在 zcid 上点击「重新同步」，ArgoCD 重新拉取镜像，部署成功。如果问题更严重，小李还可以查看部署历史，看到上一次成功的版本。

**揭示的能力需求：** 失败状态醒目标识、错误日志高亮、Step 级别状态展示、Webhook 自动触发、邮件通知、ArgoCD 同步状态展示、部署错误详情、重新同步、部署历史

---

### Journey 3: 老张搭建项目 CI/CD 全流程（DevOps - 配置管理路径）

老张是团队的 DevOps 负责人，30 岁，5 年经验。公司新启动一个微服务项目，有 3 个后端服务（Go）、1 个传统 Java 服务和 1 个前端应用。老张需要为整个项目搭建 CI/CD 流程。

**开场：** 老张登录 zcid，创建新项目「mall-platform」，设置三个环境：dev（开发）、staging（预发）、prod（生产），分别映射到不同的 K8s Namespace。然后在项目中添加 5 个服务：user-service、order-service、payment-service（Go 微服务）、legacy-erp-sync（Java 传统服务）、mall-web（前端）。

**容器化链路配置：** 老张为 3 个 Go 后端服务创建流水线，选择「Go 微服务构建」模板，修改少量参数（不同的 Git 仓库和镜像名称）。前端服务选择「前端 Node 构建」模板。这 4 条流水线走容器化链路：代码拉取 → 构建 → Kaniko 打镜像 → 推送 Harbor → ArgoCD 部署。

**传统构建链路配置：** legacy-erp-sync 是个传统 Java 项目，不走容器化。老张选择「Java Maven 构建」模板，流水线配置为：代码拉取 → Maven 编译打包 → JAR 上传到 MinIO 对象存储。没有镜像构建和 ArgoCD 部署环节，产出的 JAR 包由运维从 MinIO 下载后手动部署到传统虚拟机。

**变量与权限配置：** 老张在项目级别配置了公共变量（Harbor 地址、镜像前缀、MinIO 地址），在流水线级别配置了各服务的特有变量。密钥变量（Harbor 密码、MinIO Access Key）加密存储，界面上显示为 `***`。然后将团队成员添加到项目中，3 个后端开发者设为普通成员（只能操作 dev 环境），技术 Leader 设为项目管理员（可操作所有环境）。

**高潮：** 5 条流水线全部配置完成。当天下午，4 个开发者各自提交代码，4 条流水线同时触发、并行运行、互不干扰。3 个 Go 服务成功构建镜像并部署到 dev 环境，Java 服务成功打包 JAR 并上传到 MinIO。老张在 zcid 上看到所有服务状态一目了然。

**结局：** 以前搭建同样的流程需要至少一整天。现在 2 小时搞定，覆盖了容器化和传统两种构建链路，开发者能自助操作。

**揭示的能力需求：** 项目/环境/服务 CRUD、多环境管理（Namespace 映射）、流水线模板复用、容器化和传统两种构建链路、MinIO 产物上传、项目级/流水线级变量管理、密钥加密存储、多服务并行构建、成员管理与 RBAC 权限分配

---

### Journey 4: 王总的周一状态检查（技术管理者 - 全局监控路径）

王总是技术团队 CTO，35 岁。每周一早上他需要了解各项目的 CI/CD 健康状况，以前要分别问几个 DevOps，现在他打开 zcid。

**开场：** 王总登录 zcid，Dashboard 首页显示他有权限的所有项目卡片。每个项目卡片上标注了最近构建状态（成功/失败次数）和环境健康状态（绿色/黄色/红色）。

**快速扫描：** mall-platform 项目显示：dev 环境 5/5 服务健康，staging 环境 4/5（payment-service 部署失败，标红）。王总点进 staging 环境，看到 payment-service 最近一次部署失败，原因是健康检查超时。

**采取行动：** 王总查看部署历史，看到上一个版本是健康的。作为项目管理员，他直接点击「回滚到此版本」，ArgoCD 执行回滚，payment-service 恢复到上一个正常版本，状态变绿。然后他在群里通知老张和相关开发者排查新版本的问题。

**结局：** 王总用 5 分钟完成了状态检查和紧急回滚，以前需要 30 分钟沟通才能获取的信息和需要老张操作的回滚，现在自己就能完成。

**揭示的能力需求：** Dashboard 项目卡片、环境健康状态汇总、服务级别状态展示、部署历史、一键回滚、跨项目全局视图、项目管理员权限操作

---

### Journey 5: GitLab Webhook 自动触发（系统集成 - 自动化路径）

这不是人类用户的旅程，而是系统交互链路。当开发者在 GitLab 上 push 代码或创建 Merge Request 时，GitLab 通过 Webhook 通知 zcid。

**触发：** 小李在 GitLab 上将 feature 分支 push 到 remote。GitLab 根据配置的 Webhook URL 向 zcid 发送 Push Event 请求，payload 包含仓库地址、分支名、提交信息、提交者。

**处理：** zcid 接收到 Webhook 请求后：①验证 Webhook Secret 签名（失败返回 401 并记录审计日志）②幂等性检查——根据 delivery ID 去重，防止 GitLab 网络重试导致重复触发 ③匹配事件类型和仓库到对应的流水线 ④检查触发条件（分支匹配、事件类型匹配）⑤如果匹配，检查并发限制——如果同服务已有构建在运行，根据策略处理（排队等待 / 取消旧构建 / 拒绝并通知）⑥将用户的流水线配置翻译为 Tekton PipelineRun CRD 并提交到 K8s ⑦在运行时参数中注入本次触发的 Git 信息（commit SHA、分支、提交者）

**执行与反馈：** Tekton 拉起 Pod 执行构建。zcid 通过 Watch 机制监听 PipelineRun 状态变更，实时更新流水线执行状态。构建完成后，根据流水线配置的通知规则发送邮件/Webhook 通知给相关人员。

**揭示的能力需求：** Webhook 接收与签名验证、幂等性去重、事件-流水线匹配规则、触发条件配置、并发控制策略（排队/取消旧/拒绝）、Git 信息注入、PipelineRun 翻译与提交、状态监听与通知、审计日志

---

### Journey Requirements Summary

| 旅程 | 核心能力 | Wave |
|------|---------|------|
| 小李首次自助部署 | 首次引导、模板系统、参数化配置、实时日志、状态可视化、Git 仓库集成、RBAC | W1+W2 |
| 小李构建/部署失败恢复 | 错误高亮、Step 级状态、ArgoCD 状态展示、重新同步、部署历史、通知 | W2+W3 |
| 老张搭建项目 CI/CD | 项目/环境/服务管理、模板复用、两种构建链路、变量管理、密钥加密、并行构建、成员管理 | W1+W2 |
| 王总周一状态检查 | Dashboard、环境健康状态、服务级状态、部署历史、一键回滚 | W3 |
| GitLab Webhook 集成 | Webhook 验证、幂等去重、事件匹配、并发控制、CRD 翻译、状态监听 | W2+W3 |

**覆盖验证：**
- ✅ 主要用户成功路径（Journey 1）
- ✅ 主要用户错误恢复路径 — 构建失败 + 部署失败（Journey 2）
- ✅ 运维/配置管理路径 — 容器化 + 传统两种链路（Journey 3）
- ✅ 管理层监控 + 操作路径（Journey 4）
- ✅ API/系统集成路径 — 含幂等性和并发控制（Journey 5）
- ✅ 权限边界在旅程中体现（Journey 1/3）

*注：Wave 0（技术预研）在 Wave 1 之前执行，不直接产出用户功能，详见 Project Scoping & Phased Development 章节。*

## Domain-Specific Requirements

### Kubernetes 生态兼容性约束

- **Tekton CRD 版本依赖：** 必须基于 Tekton Pipeline v1 API（v0.44+），启动时通过 Discovery API 检测 `tekton.dev/v1` 是否注册
- **ArgoCD API 兼容：** 跟随 ArgoCD 主版本升级节奏，通过 Version endpoint 检测版本，gRPC API 调用需处理版本差异
- **K8s RBAC 最小权限：** zcid 的 ServiceAccount 只开放必要的 resources+verbs，精确定义 ClusterRole
- **CRD 生命周期管理：** PipelineRun/TaskRun 需设置 TTL 或定期清理策略，避免 etcd 数据膨胀（MVP 必做项）。zcid 安装文档需包含推荐的 Tekton 集群配置（feature-flags ConfigMap 等）
- **ArgoCD Application 单入口管理：** zcid 作为 ArgoCD Application 的唯一管理入口，需检测外部修改并告警，防止多入口导致状态不一致

### 安全与凭证管理

- **凭证存储：** AES-256-GCM 加密存储到 PostgreSQL，密钥管理需独立于应用配置
- **凭证运行时：** 临时 K8s Secret 注入 Tekton Pod，PipelineRun 结束后自动清理
- **日志脱敏规则：** 所有标记为 Secret 类型的变量值，在任何日志输出（构建日志、审计日志、应用日志）中替换为 `***`。覆盖范围包括凭证、Token、数据库连接串、用户自定义密钥变量
- **Webhook Secret：** GitLab/GitHub Webhook 签名验证，防止未授权触发
- **多集群 kubeconfig 安全：** 后续多集群场景下，kubeconfig 的存储安全等同于凭证管理，需 AES-256 加密（P1 预留约束）

### 实时通信约束

- **WebSocket 连接管理：** 连接超时 + 心跳检测 + 最大连接数限制，防止连接泄漏
- **前端重连与断点续传：** WebSocket 断开后自动重连，后端支持 `sinceTime` 参数从指定时间点续传日志
- **日志持久化保障：** 构建完成后日志必须归档到 MinIO，不能只依赖 Pod 日志（Pod 回收或节点异常会导致日志丢失）

### 技术集成风险

| 风险 | 影响 | 缓解措施 |
|------|------|---------|
| Tekton/ArgoCD 版本升级破坏兼容 | CRD 翻译失败 | 启动时版本检测 + 集成测试覆盖 |
| K8s API Server 不可达 | 全平台不可用 | 健康检查 + 优雅降级（展示缓存状态） |
| 大量并发 PipelineRun 压垮集群 | 集群资源耗尽 | 全局+项目级并发限制 |
| WebSocket 连接泄漏 | 服务端资源耗尽 | 连接超时 + 心跳检测 + 最大连接数限制 |
| etcd 存储 CRD 膨胀 | K8s 集群性能下降 | PipelineRun TTL + 定期归档清理（MVP 必做） |
| Harbor 镜像仓库不可达 | Kaniko 推送失败，流水线卡住 | 推送重试策略（3 次） + 超时处理 + 失败通知 |
| Tekton Pod 日志丢失 | 历史构建日志不可查 | 构建完成后归档 MinIO，实时+归档双通道 |
| ArgoCD Application 被外部修改 | 状态不一致 | 单入口管理 + 外部修改检测告警 |

## B2B 平台工具专项需求

### Project-Type Overview

zcid 作为可私有部署的 B2B 平台工具，不采用传统 SaaS 多租户架构，而是以 Project 为隔离边界的单实例部署模型。每个客户（私有云或 PaaS 嵌入）独立部署一套完整实例，项目间通过 RBAC 权限控制实现数据和操作隔离。

### 租户/隔离模型（Tenant Model）

- **部署隔离：** 每个客户独立部署，不共享数据库和中间件
- **项目隔离：** 单实例内通过 Project 实现逻辑隔离，不同项目的流水线、环境、服务、变量互不可见
- **环境隔离：** 每个 Environment 映射到独立的 K8s Namespace，资源物理隔离
- **不做的事：** 不实现跨实例的数据共享；不实现租户级别的资源配额计量（MVP 不做，PaaS 嵌入阶段 P2 需评估）

### RBAC 权限矩阵

| 操作 | 管理员 | 项目管理员 | 普通成员 |
|------|--------|-----------|---------|
| 系统设置（集群/仓库/全局变量） | ✅ | ❌ | ❌ |
| 用户管理 | ✅ | ❌ | ❌ |
| 创建/删除项目 | ✅ | ❌ | ❌ |
| 项目设置（成员/环境/通知） | ✅ | ✅ | ❌ |
| 创建/编辑流水线 | ✅ | ✅ | ❌ |
| 运行流水线（dev 环境） | ✅ | ✅ | ✅ |
| 运行流水线（staging/prod） | ✅ | ✅ | ❌ |
| 查看环境状态（所有环境） | ✅ | ✅ | ✅（只读） |
| 查看构建日志 | ✅ | ✅ | ✅ |
| 部署回滚 | ✅ | ✅ | ❌ |
| 变量管理（普通变量） | ✅ | ✅ | 只读 |
| 变量管理（密钥变量） | ✅ | ✅ | 不可见 |
| 查看审计日志 | ✅ | ✅（项目内） | ❌ |

**权限实现：** Casbin RBAC，策略存储在 PostgreSQL（Casbin PostgreSQL Adapter），支持 Watcher 机制实现策略热更新（权限变更无需重启服务）。JWT Token 中携带角色信息，Gin 中间件统一拦截鉴权。

### 集成清单（Integration List）

| 集成目标 | 接口抽象 | P0 | P1 | 交互方式 |
|---------|---------|-----|-----|---------|
| K8s API Server（单集群） | ClusterManager | ✅ | — | client-go in-cluster config |
| K8s API Server（多集群） | ClusterManager | — | ✅ | kubeconfig |
| Tekton | 直接交互 | ✅ | — | Go Typed Client |
| ArgoCD | 直接交互 | ✅ | — | gRPC API Client |
| GitLab | GitProvider | ✅ | — | REST API + OAuth |
| GitHub | GitProvider | ✅ | — | REST API + OAuth |
| Harbor | RegistryProvider | ✅ | — | REST API |
| 阿里云 ACR | RegistryProvider | — | ✅ | REST API |
| 邮件（SMTP） | Notifier | — | ✅ | SMTP |
| Webhook 通知 | Notifier | ✅ | — | HTTP POST |
| 飞书 | Notifier | — | ✅ | REST API |
| 钉钉 | Notifier | — | ✅ | REST API |
| 企业微信 | Notifier | — | ✅ | REST API |

**集成设计原则：** 所有外部系统通过接口抽象（GitProvider、RegistryProvider、Notifier、ClusterManager）接入，新增适配器不影响核心逻辑。

### 合规与安全要求

zcid 不涉及行业监管合规（非金融/医疗/政府），安全要求集中在：

- **凭证安全：** AES-256-GCM 加密存储，运行时临时 Secret，日志脱敏（详见 Domain-Specific Requirements）
- **操作审计：** 所有写操作记录到审计表（谁+时间+操作+目标+结果）
- **权限控制：** 三级 RBAC，密钥变量对普通成员不可见
- **K8s 最小权限：** zcid ServiceAccount 精确定义 ClusterRole
- **后续 PaaS 嵌入：** OAuth 2.0 集成能力预留（P1），满足嵌入平台的统一认证要求

## Project Scoping & Phased Development

### MVP Strategy & Philosophy

**MVP 类型：** Problem-Solving MVP — 验证「小白不写 YAML 也能完成云原生 CI/CD」的核心假设

**MVP 验证目标：**
- 内部团队（3-5 名开发者 + 1 名 DevOps）能完整使用 zcid 完成日常 CI/CD 工作流
- 两条构建链路（容器化 + 传统 JAR）均可跑通
- 开发者不需要 DevOps 协助即可自助完成流水线创建和运行

**MVP 设计原则：**
- 宁可功能少但体验完整，不要功能多但每个都半成品
- 核心链路（流水线→构建→部署）的体验做到极致，辅助功能做简化版
- 所有接口抽象（GitProvider、RegistryProvider、Notifier、ClusterManager）在 MVP 中就建立，但只实现一个 adapter

**资源需求评估：**

| 方案 | 配置 | 适用场景 |
|------|------|---------|
| 标准团队 | 1 后端（Go）+ 1 前端（React）+ 1 兼职 DevOps | 团队开发，按 Wave 并行推进 |
| Solo Developer | 1 人全栈 | 个人开发，按 Wave 严格串行，后端先行，前端用 Arco Design 标准组件快速搭建 |

**核心风险人员：** 后端开发者需熟悉 Tekton Go Client 和 ArgoCD gRPC API，这是最大的技术瓶颈

### 开发阶段规划

**Wave 0 — 技术预研（正式开发前）：**
- Tekton Go Client 快速原型：创建 PipelineRun，Watch 状态变更
- ArgoCD gRPC API spike：连接认证，创建 Application，获取状态（优先级最高）
- WebSocket 日志流 PoC：client-go GetLogs + Follow → WebSocket 转发
- 三个 PoC 跑通确认技术路线无 blocker 后，进入正式开发

**Wave 1 — 基础设施层：**
- JWT 认证 + Casbin RBAC
- 项目/环境/服务 CRUD
- 变量与密钥管理
- Git 仓库集成（GitLab + GitHub）
- **验收检查点：** 认证+资源管理稳固，能创建项目和管理成员

**Wave 2 — 核心链路（CI）：**
- 流水线可视化编排 + 模板系统（至少 4 个模板）
- Tekton PipelineRun 翻译与执行
- 实时构建日志（WebSocket）
- Kaniko 镜像构建 + Harbor 推送
- 传统构建 JAR→MinIO
- **验收检查点：** CI 链路可靠，两条构建链路均跑通

**Wave 3 — 闭环体验（CD + 运营）：**
- ArgoCD 部署 + 状态展示 + 回滚
- Webhook 触发
- Webhook 通知（HTTP POST，可接飞书/钉钉自定义机器人）
- 审计日志（简化版，Gin 中间件记录，不含查询界面）
- 简化版 Dashboard（项目列表+最近构建状态）
- PipelineRun TTL 清理（固定策略）
- 首次登录引导（欢迎横幅+项目快捷入口）
- **验收检查点：** CI→CD 完整闭环，内部团队可日常使用

### MVP Feature Set

**Must-Have（缺少则产品不成立）：**

| Wave | 功能 | 理由 |
|------|------|------|
| W1 | JWT 认证 + RBAC | 无认证无法使用 |
| W1 | 项目/环境/服务 CRUD | 资源模型是一切的基础 |
| W1 | 变量与密钥管理 | 流水线运行的必要依赖 |
| W1 | Git 仓库集成 | 代码来源是 CI 起点 |
| W2 | 流水线编排 + 模板 | 核心差异化体验 |
| W2 | Tekton PipelineRun 翻译 | 核心技术能力 |
| W2 | 实时构建日志 | 用户体验的关键反馈通道 |
| W2 | Kaniko 构建 + Harbor 推送 | 容器化链路核心 |
| W2 | 传统构建 JAR→MinIO | 第二条链路验证 |
| W3 | ArgoCD 部署 + 状态 + 回滚 | CD 闭环 + 安全回退能力 |
| W3 | Webhook 触发 | 自动化是 CI/CD 基本期望 |

**Nice-to-Have（MVP 做简化版）：**

| 功能 | MVP 处理方式 |
|------|-------------|
| 通知 | Webhook 通知（HTTP POST），邮件通知放 P1 |
| 审计日志 | Gin 中间件记录，不提供查询界面 |
| Dashboard | 项目列表+最近构建状态，不含趋势图 |
| 首次登录引导 | 欢迎横幅+项目快捷入口 |
| PipelineRun 清理 | 固定 TTL 策略，不支持自定义 |

### Post-MVP Features

**Phase 2（Growth — 6 个月目标）：**
- 阿里云 ACR 镜像仓库
- BuildKit amd64+arm64 多架构构建
- Cron 定时触发、ArgoCD 自动同步策略
- 飞书/钉钉/企业微信通知 + 邮件通知
- WebSocket 并发多 Pod 日志 Watch
- 审计日志查询界面
- Dashboard 完整版
- 性能基线建立
- AI 辅助构建诊断：构建/部署失败时，系统调用 AI 模型分析错误日志，自动生成「问题原因分析 + 修复建议」，帮助开发者快速定位和解决问题

**Phase 3（Expansion — 12 个月目标）：**
- 多集群管理
- OAuth 2.0 集成（PaaS 嵌入）
- 流水线模板市场、全局搜索
- 镜像变更触发 CD
- Dashboard 高级可视化
- 多语言支持（中英文）
- 资源配额计量（PaaS 场景评估）

### Risk Mitigation Strategy

**技术风险：**

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| Tekton CRD 翻译层实现难度超预期 | 高 | Wave 0 预研 + MVP 限制 5 种 Step 类型 + 实现前专项设计评审 |
| ArgoCD gRPC API 学习曲线 | 高 | Wave 0 优先做 spike + 先实现最小功能集 |
| WebSocket 实时日志稳定性 | 中 | Wave 0 PoC 验证 + MVP 简化为单连接单 Pod Watch |
| K8s 生态版本兼容问题 | 中 | 启动时版本检测 + 明确最低版本支持 |

**市场风险：**

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| 内部团队不愿迁移 | 中 | 先在一个项目试点 + 保留 YAML 逃生舱 |

**资源风险：**

| 风险 | 等级 | 缓解措施 |
|------|------|---------|
| 团队规模不足 | 高 | Solo Developer 方案兜底 + Wave 严格串行 |
| 后端不熟悉 Tekton/ArgoCD SDK | 中 | Wave 0 技术预研消除不确定性 |

**最小可行范围（资源极度紧张的 fallback）：**
- W1 + W2 + ArgoCD 基础能力（创建 Application + 查看状态）
- 不做回滚、不做通知、不做审计、不做 Dashboard
- 仍保持 CI→CD 的完整故事线，验证核心假设

## Functional Requirements

### 用户与权限管理

- FR1: 用户可以通过账号密码登录平台并获取认证凭证
- FR2: 管理员可以创建、编辑、禁用用户账号
- FR3: 管理员可以为用户分配系统级角色（管理员 / 项目管理员 / 普通成员）
- FR4: 系统根据用户角色和项目归属，控制其对资源和操作的访问权限
- FR5: 密钥类型变量对普通成员完全不可见

### 项目与资源管理

- FR6: 管理员可以创建和删除项目
- FR7: 项目管理员可以在项目内创建、编辑、删除环境，并将环境映射到 K8s Namespace
- FR8: 项目管理员可以在项目内创建、编辑、删除服务
- FR9: 项目管理员可以将用户添加到项目并分配项目内角色
- FR10: 不同项目之间的流水线、环境、服务、变量互不可见

### 变量与凭证管理

- FR11: 项目管理员可以在全局、项目、流水线三个层级创建和管理变量，下级覆盖上级
- FR12: 项目管理员可以创建密钥类型变量，密钥变量加密存储且界面不可回显
- FR13: 流水线运行时，系统自动将密钥变量以临时方式注入执行环境，运行结束后自动清理
- FR14: 系统在所有日志输出中自动脱敏密钥变量值

### Git 仓库集成

- FR15: 管理员可以配置 GitLab 和 GitHub 仓库连接（OAuth 授权）
- FR16: 用户在创建流水线时可以从已关联的仓库中选择代码仓库和分支
- FR17: 系统可以接收 GitLab/GitHub 的 Webhook 推送事件并验证签名
- FR18: 系统根据 Webhook 事件的仓库、分支、事件类型自动匹配并触发对应的流水线
- FR19: 系统对 Webhook 事件进行幂等性去重，防止重复触发

### 流水线编排与执行

- FR20: 项目管理员可以通过可视化界面编排流水线（Stage→Step 模型）
- FR21: 用户可以从预置模板（至少 4 种：Go 微服务、Java Maven、前端 Node、通用 Docker）一键创建流水线，只需填写少量参数
- FR22: 高级用户可以切换到 YAML 模式直接编辑流水线配置
- FR23: 系统将用户的流水线配置翻译为 Tekton PipelineRun CRD 并提交执行
- FR24: 用户可以手动触发流水线运行
- FR25: 用户可以为流水线配置 Webhook 自动触发规则（分支匹配、事件类型）
- FR26: 项目管理员可以为流水线配置并发控制策略（排队等待 / 取消旧构建 / 拒绝并通知）
- FR27: 多个流水线可以并行运行，互不干扰
- FR28: 流水线运行时自动注入本次触发的 Git 信息（commit SHA、分支、提交者）
- FR29: 项目管理员可以复制已有流水线配置来创建新流水线
- FR30: 用户手动触发流水线时可以临时指定或覆盖运行时参数
- FR31: 用户可以取消正在运行的流水线

### 构建与产物管理

- FR32: 系统支持容器化构建链路：代码拉取 → 编译构建 → 镜像构建 → 镜像推送到仓库
- FR33: 系统支持传统构建链路：代码拉取 → 编译打包 → 产物上传到对象存储
- FR34: 管理员可以配置镜像仓库连接（Harbor）
- FR35: 用户可以查看构建产物信息（镜像地址+Tag / 产物存储路径）

### 实时日志与状态监控

- FR36: 用户可以实时查看流水线每个 Step 的执行状态（等待/运行中/成功/失败）
- FR37: 用户可以实时查看正在运行的构建步骤的日志输出
- FR38: 构建失败时，系统醒目标识失败的 Step 并高亮显示错误日志
- FR39: 日志连接断开后可自动重连并从断点续传
- FR40: 构建完成后，系统将日志归档以供历史查看（不依赖临时 Pod）
- FR41: 用户可以查看历史构建运行的日志
- FR42: 系统持续监听已提交的流水线运行状态变更，并实时同步到平台
- FR43: 用户可以查看流水线的运行历史列表，包括每次运行的状态、触发方式、触发人、时间

### 部署与环境管理

- FR44: 系统可以将构建产物（容器镜像）部署到指定 K8s 环境
- FR45: 用户可以查看每个环境中各服务的部署状态（健康/同步中/异常等）
- FR46: 用户可以查看部署的同步详情和错误信息
- FR47: 项目管理员可以手动触发重新同步
- FR48: 项目管理员可以查看部署历史并回滚到指定版本
- FR49: 普通成员只能在 dev 环境触发部署，staging/prod 需要更高权限

### 通知与审计

- FR50: 项目管理员可以为流水线配置通知规则（构建成功/失败、部署完成）
- FR51: 系统通过 Webhook（HTTP POST）发送通知
- FR52: 系统记录所有写操作的审计日志（操作人、时间、操作类型、目标、结果）

### 全局概览与引导

- FR53: 用户登录后可以看到自己有权限的所有项目及其最近构建和环境状态概览
- FR54: 技术管理者可以跨项目查看所有环境的健康状态汇总
- FR55: 首次登录的用户可以看到自己所属的项目和快捷操作入口
- FR56: 用户可以在流水线列表、运行历史、服务列表等主要列表中进行筛选和搜索

### 平台运维

- FR57: 管理员可以配置系统级设置（K8s 集群连接、全局变量、镜像仓库、通知渠道）
- FR58: 管理员可以查看已配置的外部集成（Git 仓库、镜像仓库、通知渠道）的连接状态
- FR59: 系统启动时自动检测 Tekton 和 ArgoCD 的版本兼容性
- FR60: 系统可以检测关键依赖服务（K8s API Server、Tekton、ArgoCD）的健康状态，并在不可用时展示降级提示
- FR61: 系统自动清理过期的 PipelineRun CRD 资源（固定 TTL 策略）
- FR62: 系统检测 ArgoCD Application 被外部修改时，在相关环境状态页面展示告警信息

## Non-Functional Requirements

NFR 定义 FR 所列能力的质量标准。例如 FR37（实时日志查看）的质量由 NFR4（推送延迟 < 2 秒）约束，FR12（密钥加密存储）的标准由 NFR7（AES-256-GCM）定义。

### 性能

- **NFR1: API 响应时间** — 常规 CRUD API 请求响应时间 < 500ms（P95），列表查询含分页 < 1s（P95）
- **NFR2: CRD 翻译与提交延迟** — 从用户触发运行到 Tekton PipelineRun CRD 成功提交到 K8s < 5 秒
- **NFR3: 部署触发延迟** — 从 zcid 触发到 ArgoCD Application 开始同步 < 30 秒
- **NFR4: WebSocket 日志推送延迟** — 实时日志推送延迟 < 2 秒，首次连接到日志流出现 < 5 秒
- **NFR5: 页面加载** — 首屏加载时间 < 3 秒（生产环境），页面内导航切换 < 1 秒
- **NFR6: 并发支持** — 单实例支持至少 50 个并发用户操作，至少 20 条流水线并发运行（受集群资源限制）

### 安全

- **NFR7: 凭证加密** — 所有密钥变量使用 AES-256-GCM 加密存储，加密密钥独立于应用配置管理
- **NFR8: 传输加密** — 所有 API 通信支持 HTTPS/TLS，WebSocket 支持 WSS
- **NFR9: 认证安全** — JWT Token 设置合理过期时间，支持 Token 刷新机制，密码存储使用 bcrypt 或同等强度哈希
- **NFR10: 日志脱敏** — 所有标记为密钥类型的变量值在任何日志输出中替换为 `***`，覆盖构建日志、审计日志、应用日志，纳入自动化测试持续回归
- **NFR11: 运行时凭证隔离** — 临时 K8s Secret 注入 Tekton Pod，PipelineRun 结束后 30 秒内自动清理
- **NFR12: Webhook 签名验证** — 所有入站 Webhook 请求验证签名，验证失败返回 401 并记录审计日志
- **NFR13: K8s 最小权限** — zcid ServiceAccount 的 ClusterRole 只包含必要的 resources + verbs，不使用通配符

### 可靠性

- **NFR14: 平台可用性** — 月可用率 > 99.5%（排除 Tekton/ArgoCD 自身故障）
- **NFR15: 构建隔离** — 任何一条流水线的失败不影响其他正在运行的流水线
- **NFR16: 数据持久性** — 构建日志在 Pod 回收后仍可通过归档查看，日志归档成功率 > 99%
- **NFR17: 优雅降级** — K8s API Server 短暂不可达时，平台展示缓存状态和降级提示，不产生不可恢复错误
- **NFR18: 外部服务容错** — Harbor 镜像推送失败时自动重试（最多 3 次），超时后标记失败并通知

### 可扩展性

- **NFR19: 数据增长** — 数据库设计支持单实例存储 100+ 项目、1000+ 流水线配置、10 万+ 运行记录，无显著性能退化
- **NFR20: CRD 清理** — PipelineRun/TaskRun 按 TTL 策略自动清理，防止 etcd 数据膨胀
- **NFR21: WebSocket 连接** — 单实例支持至少 200 个并发 WebSocket 连接，超过限制时拒绝新连接并提示

### 集成

- **NFR22: 接口抽象** — 所有外部系统通过接口抽象接入（GitProvider、RegistryProvider、Notifier、ClusterManager），新增适配器不影响核心逻辑
- **NFR23: Tekton 版本兼容** — 支持 Tekton Pipeline v1 API（v0.44+），启动时自动检测版本兼容性
- **NFR24: ArgoCD 版本兼容** — 跟随 ArgoCD 主版本升级节奏，gRPC API 调用处理版本差异
- **NFR25: API 规范** — RESTful API 遵循统一响应格式，错误码段按业务模块分配，Swagger/OpenAPI 文档自动生成

### 可观测性

- **NFR26: 审计追踪** — 所有写操作记录审计日志，包含操作人、时间、操作类型、目标、结果，审计日志保留至少 90 天
- **NFR27: 健康检查** — 提供健康检查端点，覆盖数据库连接、Redis 连接、K8s API 可达性
- **NFR28: 应用日志** — 平台自身运行日志结构化输出（JSON 格式），支持日志级别动态调整

## Implementation Constraints

- **API 设计：** RESTful + Swagger 自动生成文档（swaggo/swag），前端代码生成工具实现时评估最优选（openapi-typescript-codegen / @hey-api/openapi-ts / orval）
- **错误码段分配（预留完整规划，一旦发布即为 API 合同）：**
  - 400xx：认证与权限
  - 401xx：项目管理
  - 402xx：流水线
  - 403xx：环境与服务
  - 404xx：构建与镜像
  - 405xx：Git 集成
  - 406xx：通知
  - 500xx：系统内部错误
- **数据库设计：** 高频查询字段保持独立列+索引，JSONB 只做存储不做复杂查询；流水线配置等 JSONB 数据加入 `schemaVersion` 字段，方便后续数据迁移
- **数据库迁移：** golang-migrate，SQL 文件驱动，显式 up/down，服务启动时自动执行

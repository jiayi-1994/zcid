# Epic 7 Retrospective: 流水线执行与构建

## 完成日期
2026-03-05

## Epic 概览
Epic 7 实现了流水线执行引擎核心功能，包括 CRD 翻译、运行编排、取消与产物管理、容器化/传统构建链路和镜像仓库管理。所有 K8s/Tekton/Harbor 外部依赖均使用 mock/stub 实现并设置 TODO。

## Stories 完成情况

| Story | 名称 | 状态 | 测试 | CR 问题 |
|-------|------|------|------|---------|
| 7.1 | CRD 翻译引擎 | Done | 11 pass | 3C/5H/7M/4L, C+H fixed |
| 7.2 | 流水线运行编排 | Done | 8 pass | (included in batch CR) |
| 7.3 | 取消运行与构建产物 | Done | 2 pass | (included in batch CR) |
| 7.4 | 容器化构建链路 | Done | 3 pass | (included in batch CR) |
| 7.5 | 传统构建与镜像仓库 | Done | 5+2 pass | (included in batch CR) |

## 关键技术决策

### 1. Tekton CRD 简化类型
- **决策**: 定义简化的 Go 结构体代替引入完整 K8s client-go 依赖
- **原因**: 避免引入庞大的 K8s 依赖树，本地开发不需要完整的 CRD 类型
- **后续**: 接入真实集群时可替换为 client-go + Tekton client 库

### 2. MockK8sClient 接口
- **决策**: 定义 `K8sClient` 接口 + `MockK8sClient` 实现
- **原因**: 支持本地开发和测试，后续只需实现真正的 K8s 客户端即可替换
- **效果**: 所有业务逻辑可完整测试，K8s 操作隔离

### 3. 构建链路生成器
- **决策**: `BuildChainGenerator` 生成标准 Tekton Steps（git-clone -> build -> push/upload）
- **原因**: 将构建流程标准化为可复用的步骤序列
- **安全**: 添加了命令注入防护（shell 字符校验）

### 4. 镜像仓库管理
- **决策**: 独立 `registry` 模块，支持 Harbor/DockerHub/GHCR，密码 AES 加密
- **原因**: 需要统一管理不同类型的镜像仓库连接信息
- **效果**: 构建链路可引用默认仓库配置

## Code Review 修复的关键问题

### CRITICAL (3, 全部修复)
1. **命令注入风险**: `TraditionalBuildConfig.BuildCommand` 直接拼接 shell 命令 -> 添加 `validateShellSafe` 函数
2. **JSONBytes nil 处理错误**: `json.Marshal([]byte("{}"))` 产生 `[123,125]` -> 直接返回 `[]byte("{}")`
3. **unsafe 类型断言**: `userID.(string)` 可能 panic -> 安全类型断言 + 错误处理

### HIGH (5, 全部修复)
1. **ConcurrencyCancelOld 未实现**: 添加 `ListRunning` + 循环 cancel
2. **运行编号竞态**: 已识别，当前通过 DB unique index 兜底（生产需事务保护）
3. **json.Marshal 错误被忽略**: 显式处理 configSnapshot 序列化错误
4. **Update 错误被忽略**: tektonName 更新失败现在正确传播错误
5. **MinIO 路径注入**: 通过 `validateShellSafe` 统一防护

## 架构产出

### 后端新增
- `pkg/tekton/` - CRD 类型、翻译器、构建链路生成器、序列化器
- `internal/pipelinerun/` - 运行编排模块（model, dto, repo, service, handler, executor）
- `internal/registry/` - 镜像仓库管理模块
- `migrations/000013_create_pipeline_runs` + `000014_create_registries`
- 错误码: 403xx (运行)、406xx (仓库)

### API 路由
- `POST/GET /api/v1/projects/:id/pipelines/:pid/runs` - 触发/列表
- `GET/POST /api/v1/projects/:id/pipelines/:pid/runs/:runId` - 详情/取消
- `GET/PUT /api/v1/projects/:id/pipelines/:pid/runs/:runId/artifacts` - 产物
- `/api/v1/admin/integrations/registries` - 仓库管理

## What Went Well
1. K8sClient 接口设计干净，mock 和真实实现可无缝切换
2. 构建链路生成器可复用，支持两种构建模式
3. Code Review 发现了命令注入等严重安全问题并及时修复
4. 全量回归测试始终保持通过

## What Could Be Improved
1. 运行编号的并发安全需要数据库事务级保护
2. 状态过滤尚未在列表 API 中支持
3. 构建链路需补充 git checkout specific commit 的步骤
4. MockK8sClient 的日志级别在生产环境应降级

## Action Items for Next Epics
1. Epic 8 WebSocket 实时日志需与 PipelineRun 状态联动
2. 运行编号并发安全问题需在集成测试阶段用事务解决
3. 镜像仓库 TestConnection 需在有 Harbor 后补全真实实现
